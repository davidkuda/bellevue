package main

import (
	"net/http"

	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static/dist/"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	// standard := alice.New(logRequest, commonHeaders, app.identify)
	standard := alice.New(commonHeaders, app.authenticate)
	usersOnly := alice.New(app.requireAuthentication)
	adminsOnly := alice.New(app.requireAuthentication, app.requireAdmin)

	mux.HandleFunc("GET /{$}", app.getHome)

	// activities: all require authentication:
	mux.Handle("GET /activities", usersOnly.ThenFunc(app.getActivities))
	mux.Handle("GET /activities/new", usersOnly.ThenFunc(app.getActivitiesNew))
	mux.Handle("POST /activities", usersOnly.ThenFunc(app.bellevueActivityPost))
	mux.Handle("GET /activities/{id}/edit", usersOnly.ThenFunc(app.getActivitiesIDEdit))
	mux.Handle("DELETE /activities/{id}", usersOnly.ThenFunc(app.bellevueActivityDelete))

	mux.Handle("PATCH /invoices/{id}", usersOnly.ThenFunc(app.patchInvoicesIDState))

	mux.HandleFunc("GET /login", app.getLogin)
	mux.HandleFunc("GET /login/email", app.getLoginEmail)
	mux.HandleFunc("GET /signup", app.getLoginSignup)
	mux.HandleFunc("POST /signup", app.postSignup)
	mux.HandleFunc("POST /login", app.postLogin)
	mux.HandleFunc("GET /login/dwbn", app.oidcLogin)
	mux.HandleFunc("GET /login/dwbn/callback", app.oidcCallbackHandler)

	// protected:
	mux.Handle("GET /logout", usersOnly.ThenFunc(app.getLogout))

	mux.Handle("GET /settings", adminsOnly.ThenFunc(app.getSettings))
	mux.Handle("GET /settings/products", adminsOnly.ThenFunc(app.getSettingsProducts))

	return standard.Then(mux)
}
