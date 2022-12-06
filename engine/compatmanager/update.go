package compatmanager

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/target/goalert/permission"
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
		// but we need to store the contact method id in the format "team_id:subject_id"
		teamID := strings.TrimPrefix(s.ProviderID, "slack:")
		value := fmt.Sprintf("%s:%s", teamID, s.SubjectID)

		cmID := uuid.New()
		// TODO: name must be unique
		_, err = tx.StmtContext(ctx, db.insertCM).ExecContext(ctx, cmID, "Slack", "SLACK_DM", value, s.UserID)
		if err != nil {
			return fmt.Errorf("insert cm: %w", err)
		}

		_, err = tx.StmtContext(ctx, db.updateSubCMID).ExecContext(ctx, s.ID, cmID)
		if err != nil {
			return fmt.Errorf("update sub cm_id: %w", err)
		}
	}

	return tx.Commit()
}
