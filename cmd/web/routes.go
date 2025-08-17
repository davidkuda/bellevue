package main

import (
	"net/http"

	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	// standard := alice.New(logRequest, commonHeaders)
	standard := alice.New(commonHeaders, app.identify)
	usersOnly := alice.New(app.requireAuthentication)
	// adminsOnly := alice.New(app.requireAuthentication, app.requireAdmin)

	mux.HandleFunc("GET /", app.home)

	// Bellevue Activities (all protected):
	mux.Handle("GET /admin/new-bellevue-activity", usersOnly.ThenFunc(app.adminNewBellevueActivity))
	mux.Handle("GET /bellevue-activities", usersOnly.ThenFunc(app.bellevueActivities))
	mux.Handle("POST /bellevue-activities", usersOnly.ThenFunc(app.bellevueActivityPost))
	mux.Handle("PUT /bellevue-activities/{id}", usersOnly.ThenFunc(app.bellevueActivityPut))
	mux.Handle("DELETE /bellevue-activities/{id}", usersOnly.ThenFunc(app.bellevueActivityDelete))
	// HTMX partials
	mux.Handle("GET /bellevue-activities/{id}/edit", usersOnly.ThenFunc(app.bellevueActivityIDEdit))

	// admin:
	mux.HandleFunc("GET /admin/login", app.adminLogin)
	mux.HandleFunc("POST /admin/login", app.adminLoginPost)
	// protected:
	mux.Handle("GET /admin", usersOnly.ThenFunc(app.admin))
	mux.Handle("GET /admin/logout", usersOnly.ThenFunc(app.adminLogoutPost))

	return standard.Then(mux)
}
