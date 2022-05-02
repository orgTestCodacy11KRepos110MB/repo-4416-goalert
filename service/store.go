package service

import (
	"context"
	"database/sql"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"

	"github.com/google/uuid"
)

type Store struct {
	db *sql.DB

	findOne     *sql.Stmt
	findOneUp   *sql.Stmt
	findMany    *sql.Stmt
	findAll     *sql.Stmt
	findAllByEP *sql.Stmt
	insert      *sql.Stmt
	update      *sql.Stmt
	delete      *sql.Stmt
}

func NewStore(ctx context.Context, db *sql.DB) (*Store, error) {
	prep := &util.Prepare{DB: db, Ctx: ctx}
	p := prep.P

	s := &Store{db: db}
	s.findOne = p(`
		SELECT
			s.id,
			s.name,
			s.description,
			s.escalation_policy_id,
			e.name,
			fav	is distinct from null
		FROM
			services s
		JOIN escalation_policies e ON e.id = s.escalation_policy_id
		LEFT JOIN user_favorites fav ON s.id = fav.tgt_service_id AND fav.user_id = $2
		WHERE
			s.id = $1
	`)
	s.findOneUp = p(`
		SELECT
			s.id,
			s.name,
			s.description,
			s.escalation_policy_id
		FROM services s
		WHERE s.id = $1
		FOR UPDATE
	`)
	s.findMany = p(`
		SELECT
			s.id,
			s.name,
			s.description,
			s.escalation_policy_id,
			e.name,
			fav	is distinct from null
		FROM
			services s
		JOIN escalation_policies e ON e.id = s.escalation_policy_id
		LEFT JOIN user_favorites fav ON s.id = fav.tgt_service_id AND fav.user_id = $2
		WHERE
			s.id = any($1)
	`)

	s.findAll = p(`
		SELECT
			s.id,
			s.name,
			s.description,
			s.escalation_policy_id,
			e.name,
			false
		FROM
			services s,
			escalation_policies e
		WHERE
			e.id = s.escalation_policy_id
	`)
	s.findAllByEP = p(`
		SELECT
			s.id,
			s.name,
			s.description,
			s.escalation_policy_id,
			e.name,
			false
		FROM
			services s,
			escalation_policies e
		WHERE
			e.id = $1 AND
			e.id = s.escalation_policy_id
	`)
	s.insert = p(`INSERT INTO services (id,name,description,escalation_policy_id) VALUES ($1,$2,$3,$4)`)
	s.update = p(`UPDATE services SET name = $2, description = $3, escalation_policy_id = $4 WHERE id = $1`)
	s.delete = p(`DELETE FROM services WHERE id = any($1)`)

	return s, prep.Err
}

func (s *Store) FindOneForUpdate(ctx context.Context, tx *sql.Tx, id string) (*Service, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}
	err = validate.UUID("ServiceID", id)
	if err != nil {
		return nil, err
	}
	var svc Service
	err = tx.StmtContext(ctx, s.findOneUp).QueryRowContext(ctx, id).Scan(&svc.ID, &svc.Name, &svc.Description, &svc.EscalationPolicyID)
	if err != nil {
		return nil, err
	}
	return &svc, nil
}

// FindMany returns slice of Service objects given a slice of serviceIDs
func (s *Store) FindMany(ctx context.Context, ids []string) ([]Service, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}
	err = validate.ManyUUID("ServiceIDs", ids, 100)
	if err != nil {
		return nil, err
	}

	rows, err := s.findMany.QueryContext(ctx, sqlutil.UUIDArray(ids), permission.UserID(ctx))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAllFrom(rows)
}

func (s *Store) Insert(ctx context.Context, svc *Service) (*Service, error) {
	return s.CreateServiceTx(ctx, nil, svc)
}
func (s *Store) CreateServiceTx(ctx context.Context, tx *sql.Tx, svc *Service) (*Service, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return nil, err
	}

	n, err := svc.Normalize()
	if err != nil {
		return nil, err
	}

	n.ID = uuid.New().String()
	stmt := s.insert
	if tx != nil {
		stmt = tx.Stmt(stmt)
	}
	_, err = stmt.ExecContext(ctx, n.ID, n.Name, n.Description, n.EscalationPolicyID)
	if err != nil {
		return nil, err
	}

	return n, nil
}

// Delete implements the ServiceInterface interface.
func (s *Store) Delete(ctx context.Context, id string) error {
	return s.DeleteTx(ctx, nil, id)
}
func (s *Store) DeleteTx(ctx context.Context, tx *sql.Tx, id string) error {
	return s.DeleteManyTx(ctx, tx, []string{id})
}
func (s *Store) DeleteManyTx(ctx context.Context, tx *sql.Tx, ids []string) error {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}
	err = validate.ManyUUID("ServiceID", ids, 50)
	if err != nil {
		return err
	}
	stmt := s.delete
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}
	_, err = stmt.ExecContext(ctx, sqlutil.UUIDArray(ids))
	return err
}

func wrap(tx *sql.Tx, s *sql.Stmt) *sql.Stmt {
	if tx == nil {
		return s
	}
	return tx.Stmt(s)
}

// Update implements the ServiceStore interface.
func (s *Store) Update(ctx context.Context, svc *Service) error {
	return s.UpdateTx(ctx, nil, svc)
}
func (s *Store) UpdateTx(ctx context.Context, tx *sql.Tx, svc *Service) error {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}

	n, err := svc.Normalize()
	if err != nil {
		return err
	}

	err = validate.UUID("ServiceID", n.ID)
	if err != nil {
		return err
	}

	_, err = wrap(tx, s.update).ExecContext(ctx, n.ID, n.Name, n.Description, n.EscalationPolicyID)
	return err
}

func (s *Store) FindOneForUser(ctx context.Context, userID, serviceID string) (*Service, error) {
	err := validate.UUID("ServiceID", serviceID)
	if err != nil {
		return nil, err
	}

	var uid sql.NullString
	userCheck := permission.User

	if userID != "" {
		err := validate.UUID("UserID", userID)
		if err != nil {
			return nil, err
		}
		userCheck = permission.MatchUser(userID)
		uid.Valid = true
		uid.String = userID
	}

	err = permission.LimitCheckAny(ctx, userCheck, permission.System)
	if err != nil {
		return nil, err
	}

	row := s.findOne.QueryRowContext(ctx, serviceID, uid)
	var svc Service
	err = scanFrom(&svc, row.Scan)
	if err != nil {
		return nil, err
	}

	return &svc, nil
}

func (s *Store) FindOne(ctx context.Context, id string) (*Service, error) {
	// old method just calls new method
	return s.FindOneForUser(ctx, "", id)
}

func scanFrom(s *Service, f func(args ...interface{}) error) error {
	return f(&s.ID, &s.Name, &s.Description, &s.EscalationPolicyID, &s.epName, &s.isUserFavorite)
}

func scanAllFrom(rows *sql.Rows) (services []Service, err error) {
	var s Service
	for rows.Next() {
		err = scanFrom(&s, rows.Scan)
		if err != nil {
			return nil, err
		}
		services = append(services, s)
	}
	return services, nil
}

func (s *Store) FindAll(ctx context.Context) ([]Service, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return nil, err
	}

	rows, err := s.findAll.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAllFrom(rows)
}
func (s *Store) FindAllByEP(ctx context.Context, epID string) ([]Service, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return nil, err
	}

	rows, err := s.findAllByEP.QueryContext(ctx, epID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAllFrom(rows)
}
