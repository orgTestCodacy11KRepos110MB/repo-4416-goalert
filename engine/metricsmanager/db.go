package metricsmanager

import (
	"context"
	"database/sql"

	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/util"
)

const engineVersion = 2

// DB handles updating metrics
type DB struct {
	db   *sql.DB
	lock *processinglock.Lock

	boundNow *sql.Stmt

	scanLogs      *sql.Stmt
	insertMetrics *sql.Stmt

	nextDailyMetricsDate *sql.Stmt
	insertDailyMetrics   *sql.Stmt
}

// Name returns the name of the module.
func (db *DB) Name() string { return "Engine.MetricsManager" }

// NewDB creates a new DB.
func NewDB(ctx context.Context, db *sql.DB) (*DB, error) {
	lock, err := processinglock.NewLock(ctx, db, processinglock.Config{
		Version: engineVersion,
		Type:    processinglock.TypeMetrics,
	})
	if err != nil {
		return nil, err
	}

	p := &util.Prepare{Ctx: ctx, DB: db}

	return &DB{
		db:   db,
		lock: lock,

		// NOTE: this buffer provides time for in-flight requests to settle
		boundNow: p.P(`select now() - '2 minutes'::interval`),

		scanLogs: p.P(`
			select alert_id, timestamp, id 
			from alert_logs 
			where event='closed' and timestamp < $3 and (timestamp > $1 or (timestamp = $1 and id > $2)) 
			order by timestamp, id 
			limit 500`),

		insertMetrics: p.P(`
			insert into alert_metrics (alert_id, service_id, time_to_ack, time_to_close, escalated, closed_at)
			select
				a.id,
				a.service_id,
				(select timestamp - a.created_at from alert_logs where alert_id = a.id and event = 'acknowledged' order by timestamp limit 1),
				(select timestamp - a.created_at from alert_logs where alert_id = a.id and event = 'closed'       order by timestamp limit 1),
				(select count(*) > 1             from alert_logs where alert_id = a.id and event = 'escalated'),
				(select timestamp                from alert_logs where alert_id = a.id and event = 'closed'       order by timestamp limit 1)
			from alerts a
			where a.id = any($1) and a.service_id is not null
			on conflict do nothing
		`),

		nextDailyMetricsDate: p.P(`
			select (date(timezone('UTC'::text, closed_at))) from alert_metrics 
			where  (date(timezone('UTC'::text, closed_at))) > $1::date 
			and    (date(timezone('UTC'::text, closed_at))) < $2::date
			order by (date(timezone('UTC'::text, closed_at)))
			limit 1;
		`),

		insertDailyMetrics: p.P(`
			insert into daily_alert_metrics (date, service_id, alert_count, avg_time_to_ack, avg_time_to_close, escalated_count)
			select 
				$1::date, 
				service_id, 
				count(*), 
				avg(time_to_ack), 
				avg(time_to_close), 
				count(*) filter (where escalated=true)
			from alert_metrics
			where (date(timezone('UTC'::text, closed_at))) = $1
			group by service_id;
		`),
	}, p.Err
}
