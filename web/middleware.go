package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/justinas/nosurf"
)

func (st *sT) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		st.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method,
			r.URL.RequestURI())

		next.ServeHTTP(w, r)

	})
}

func (st *sT) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//only runs if there is a panic
		defer func() {

			if err := recover(); err != nil {
				w.Header().Set("Connection", "Close")
				st.serverError(w, fmt.Errorf("%s", err))

			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (st *sT) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !st.isAuthenticated(r) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		w.Header().Add("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
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

func (st *sT) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		exists := st.sessionManager.Exists(r.Context(), authenticatedUserID)
		if !exists {
			next.ServeHTTP(w, r)
			return
		}
		usr, err := st.getUserR(st.sessionManager.GetInt(r.Context(),
			authenticatedUserID))
		if errors.Is(err, errNoRecord) || !usr.Active {
			st.sessionManager.Remove(r.Context(), authenticatedUserID)
			next.ServeHTTP(w, r)
			return
		} else if err != nil {
			st.serverError(w, err)
			return
		}
		ctx := context.WithValue(r.Context(), contextKeyIsAuthenticated, true)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type plainHandler func(w http.ResponseWriter, r *http.Request)

func (st *sT) dynamicRoutes(next plainHandler) http.Handler {
	return st.recoverPanic(st.logRequest(st.authenticate(http.HandlerFunc(next))))
}

func (st *sT) dynamicAuthRoute(next plainHandler) http.Handler {
	return st.authenticate(st.requireAuthentication(http.HandlerFunc(next)))
}
