package compatmanager

import (
	"context"
	"database/sql"

	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/util"
)

// DB handles keeping compatibility-related data in sync.
type DB struct {
	db   *sql.DB
	lock *processinglock.Lock

	slackSubMissingCM *sql.Stmt
	updateSubCMID     *sql.Stmt
	insertCM          *sql.Stmt
}

// Name returns the name of the module.
func (db *DB) Name() string { return "Engine.CompatManager" }

// NewDB creates a new DB.
func NewDB(ctx context.Context, db *sql.DB) (*DB, error) {
	lock, err := processinglock.NewLock(ctx, db, processinglock.Config{
		Version: 1,
		Type:    processinglock.TypeCompat,
	})
	if err != nil {
		return nil, err
	}

	p := &util.Prepare{Ctx: ctx, DB: db}

	return &DB{
		db:   db,
		lock: lock,

		// get all entries missing cm_id where provider_id starts with "slack:"
		slackSubMissingCM: p.P(`
			select id, user_id, subject_id, provider_id from auth_subjects where
				provider_id like 'slack:%' and cm_id is null
			for update skip locked
			limit 25
		`),

		// update cm_id for a given user_id and subject_id
		updateSubCMID: p.P(`update auth_subjects set cm_id = $2 where id = $1`),

		insertCM: p.P(`
			insert into user_contact_methods (id, name, type, value, user_id)
			values ($1, $2, $3, $4, $5)
		`),
	}, p.Err
}
