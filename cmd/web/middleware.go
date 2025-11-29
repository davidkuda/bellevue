package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/davidkuda/bellevue/internal/models"
)

type contextKey string

const (
	userContextKey            contextKey = "user"
	isAuthenticatedContextKey contextKey = "isAuthenticated"
)

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			ip     = r.RemoteAddr
			proto  = r.Proto
			method = r.Method
			uri    = r.URL.RequestURI()
			platf  = r.Header.Get("Sec-Ch-Ua-Platform")
			// verbose Header:
			// uagent = r.Header.Get("User-Agent")
		)

		// caddy will set X-Forwarded-For with original src IP when reverse proxying.
		// r.RemoteAddr will be localhost, in that case.
		xff := r.Header.Get("X-Forwarded-For")
		if xff != "" {
			ip = xff
		}

		log.Printf("msg=ReceivedRequest ip=%v proto=%v method=%v uri=%v platf=%v", ip, proto, method, uri, platf)

		// verbose Header:
		// log.Printf("msg=ReceivedRequest ip=%v proto=%v method=%v uri=%v platf=%v user-agent=%v", ip, proto, method, uri, platf, uagent)

		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := app.sessionManager.GetInt(r.Context(), "UserID")
		if userID == 0 {
			ctx := r.Context()
			ctx = context.WithValue(ctx, "isAuthenticated", false)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
			return
		}

		user, err := app.models.Users.GetUserByID(userID)
		if err != nil {
			// if user vanished, nuke the session and continue unauthenticated
			app.sessionManager.Remove(r.Context(), "userID")
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, isAuthenticatedContextKey, true)
		ctx = context.WithValue(ctx, userContextKey, &user)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := app.isAuthenticated(r)
		if !auth {
			app.renderClientError(w, r, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement permissions management
		user := app.contextGetUser(r)
		if user == nil {
			app.renderClientError(w, r, http.StatusUnauthorized)
			return
		}
		// TODO: right now, user 1 is the admin x)
		if user.ID != 1 {
			app.renderClientError(w, r, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) contextGetUser(r *http.Request) *models.User {
	user, ok := r.Context().Value(userContextKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}

func (app *application) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
	if !ok {
		return false
	}
	return isAuthenticated
}

// see https://owasp.org/www-project-secure-headers/
// see https://developer.mozilla.org/en-US/docs/Web/HTTP/Guides/CSP
func commonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// JS: script-src:
		setTheme := "'sha256-iPh555NGYFuqXa3x4Etpt6REdQ/TiOrBh3UPr3/vH5s='"

		// style-src:
		htmx := "'sha256-faU7yAF8NxuMTNEwVmBz+VcYeIoBQ2EMHW3WaVxCvnk='"

		w.Header().Set(
			"Content-Security-Policy",
			"default-src 'self';"+
				"img-src 'self' images.ctfassets.net;"+
				fmt.Sprintf("script-src 'self' cdnjs.cloudflare.com cdn.jsdelivr.net %s;", setTheme)+
				fmt.Sprintf("style-src 'self' cdnjs.cloudflare.com fonts.googleapis.com %s;", htmx)+
				"font-src fonts.gstatic.com",
		)
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")

		w.Header().Set("Server", "Go")

		next.ServeHTTP(w, r)
	})
}
