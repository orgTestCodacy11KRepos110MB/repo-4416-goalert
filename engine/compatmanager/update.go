package compatmanager

import (
	"context"
	"fmt"
	"strings"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/user/contactmethod"
	"github.com/target/goalert/util/log"
)

// UpdateAll will process compatibility entries for the cycle.
func (db *DB) UpdateAll(ctx context.Context) error {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return err
	}
	log.Debugf(ctx, "Running compat operations.")

	tx, err := db.lock.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	type sub struct {
		ID         int
		UserID     string
		SubjectID  string
		ProviderID string
	}

	var subs []sub
	rows, err := tx.StmtContext(ctx, db.slackSubMissingCM).QueryContext(ctx)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	for rows.Next() {
		var s sub
		err = rows.Scan(&s.ID, &s.UserID, &s.SubjectID, &s.ProviderID)
		if err != nil {
			return fmt.Errorf("scan: %w", err)
		}
		subs = append(subs, s)
	}

	for _, s := range subs {
		// provider id contains the team id in the format "slack:team_id"
		// but we need to store the contact method id in the format "team_id:user_id"
		teamID := strings.TrimPrefix(s.ProviderID, "slack:")
		cm := &contactmethod.ContactMethod{
			// TODO: name must be unique
			Name:   "Slack",
			Type:   contactmethod.TypeSlackDM,
			Value:  fmt.Sprintf("%s:%s", teamID, s.SubjectID),
			UserID: s.UserID,
		}

		cm, err = db.cm.CreateTx(ctx, tx, cm)
		if err != nil {
			return fmt.Errorf("create cm: %w", err)
		}

		_, err = tx.StmtContext(ctx, db.updateSubCMID).ExecContext(ctx, s.ID, cm.ID)
		if err != nil {
			return fmt.Errorf("update sub cm_id: %w", err)
		}
	}

	return tx.Commit()
}
