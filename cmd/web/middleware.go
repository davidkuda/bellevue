package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/davidkuda/bellevue/internal/models"
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

func (app *application) identify(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "no-store")

		var user models.User
		var userEmail string
		var userID int

		userEmail, err := app.extractUserFromJWTCookie(r)
		if err != nil {
			ctx := r.Context()
			ctx = context.WithValue(ctx, "isAuthenticated", false)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
			return
		}

		user, err = app.models.Users.GetUserByEmail(userEmail)
		if err != nil {
			err = fmt.Errorf("could not get user with email %s: %v\n", userEmail, err)
			app.serverError(w, r, err)
			return
		}

		userID = user.ID

		ctx := r.Context()
		ctx = context.WithValue(ctx, "isAuthenticated", true)
		ctx = context.WithValue(ctx, "userEmail", userEmail)
		ctx = context.WithValue(ctx, "userID", userID)
		ctx = context.WithValue(ctx, "user", user)

		// TODO: implement nice permission management...
		if userID == 1 {
			ctx = context.WithValue(ctx, "isAdmin", true)
		}

		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: think this over, this makes another JWT decoding
		if !app.isAuthenticated(r) {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isAdmin, ok := r.Context().Value("isAdmin").(bool)
		if !ok {
			app.renderClientError(w, r, http.StatusForbidden)
			return
		}
		if !isAdmin {
			app.renderClientError(w, r, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// see https://owasp.org/www-project-secure-headers/
// see https://developer.mozilla.org/en-US/docs/Web/HTTP/Guides/CSP
func commonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// JS: script-src:
		setTheme := "'sha256-d0p7Z2OKW9F6H7+KJP42Xcw2Tb90XTuKIILK5NffXgQ='"
		highlightJS := "'sha256-KuW8nrMYej09eTtZkBNDwTy8Yn05dABB5v2dLSEPgTY='"

		// style-src:
		htmx := "'sha256-faU7yAF8NxuMTNEwVmBz+VcYeIoBQ2EMHW3WaVxCvnk='"

		w.Header().Set(
			"Content-Security-Policy",
			"default-src 'self';"+
				"img-src 'self' images.ctfassets.net;"+
				fmt.Sprintf("script-src 'self' cdnjs.cloudflare.com cdn.jsdelivr.net %s %s;", setTheme, highlightJS)+
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
