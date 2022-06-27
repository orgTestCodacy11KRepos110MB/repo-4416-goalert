package app

import (
	"context"
	"net/http"

	"github.com/target/goalert/switchover"
	"github.com/target/goalert/util/log"
)

func (app *App) pauseHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		err := app.requestLock.RLock(ctx)
		if err != nil {
			log.Log(ctx, err)
			return
		}
		defer app.requestLock.RUnlock()
		next.ServeHTTP(w, req)
	})
}

// LogBackgroundContext returns a context.Background with the application logger configured.
func (app *App) LogBackgroundContext() context.Context { return app.cfg.Logger.BackgroundContext() }

func (app *App) Pause(ctx context.Context) error {
	ctx = log.WithLogger(ctx, app.cfg.Logger)

	err := app.mgr.Pause(ctx)
	if err != nil {
		return err
	}
	app.db.SetMaxIdleConns(0)
	return nil
}

func (app *App) Resume() {
	app.db.SetMaxIdleConns(app.cfg.DBMaxIdle)
	app.mgr.Resume(app.LogBackgroundContext())
}

func (app *App) _pause(ctx context.Context) error {
	app.events.Stop()

	cfg := switchover.ConfigFromContext(ctx)
	if cfg.NoPauseAPI {
		return nil
	}
	err := app.requestLock.Lock(ctx)
	if err != nil {
		app.events.Start()
		return err
	}
	return nil
}

func (app *App) _resume(ctx context.Context) error {
	app.events.Start()
	app.requestLock.Unlock()
	return nil
}
