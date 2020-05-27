package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/justinas/nosurf"
	"github.com/saied74/toychat/pkg/broker"
)

type plainHandler func(w http.ResponseWriter, r *http.Request)

func (app *App) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method,
			r.URL.RequestURI())
		start := time.Now()
		next.ServeHTTP(w, r)
		end := time.Now()
		app.infoLog.Printf("Time difference %v", end.Sub(start))
	})
}

func (app *App) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//only runs if there is a panic
		defer func() {

			if err := recover(); err != nil {
				w.Header().Set("Connection", "Close")
				app.serverError(w, fmt.Errorf("%s", err))

			}
		}()
		next.ServeHTTP(w, r)
	})
}

// func (app *App) requireAuthentication(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		if !app.isAuthenticated(r) {
// 			http.Redirect(w, r, app.td.Login, http.StatusSeeOther)
// 			return
// 		}
// 		w.Header().Add("Cache-Control", "no-store")
// 		next.ServeHTTP(w, r)
// 	})
// }

func (app *App) requireAuthentication(next plainHandler) plainHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isAuth(r) {
			http.Redirect(w, r, app.td.Login, http.StatusSeeOther)
			return
		}
		w.Header().Add("Cache-Control", "no-store")
		next(w, r)
	}
}

func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
	})
	return csrfHandler
}

func (app *App) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		exists := app.sessionManager.Exists(r.Context(), authenticatedUserID)
		if !exists {
			next.ServeHTTP(w, r)
			return
		}
		path := strings.Split(r.URL.Path, "/")
		if len(path) < 3 {
			app.errorLog.Printf("bad path %s, short string", r.URL.Path)
		}
		usr, err := broker.GetUserR(app.table, app.sessionManager.GetInt(r.Context(),
			authenticatedUserID))
		if errors.Is(err, broker.ErrNoRecord) || !usr.Active {
			app.sessionManager.Remove(r.Context(), authenticatedUserID)
			next.ServeHTTP(w, r)
			return
		} else if err != nil {
			app.serverError(w, err)
			return
		}
		switch path[1] {
		case super:
			if usr.Role != "superadmin" {
				app.sessionManager.Remove(r.Context(), authenticatedUserID)
			}
		case admin:
			if usr.Role != admin {
				app.sessionManager.Remove(r.Context(), authenticatedUserID)
			}
		case agent:
			if usr.Role != agent {
				app.sessionManager.Remove(r.Context(), authenticatedUserID)
			}
		}
		ctx := context.WithValue(r.Context(), contextKeyIsAuthenticated, true)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *App) dynamicRoutes(next http.Handler) http.Handler {
	return noSurf(app.sessionManager.LoadAndSave(app.recoverPanic(app.logRequest(app.authenticate(next)))))
}

// func (app *App) dynamicAuthRoute(next plainHandler) http.Handler {
// 	return app.authenticate(app.requireAuthentication(http.HandlerFunc(next)))
// }
