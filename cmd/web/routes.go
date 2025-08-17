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

	mux.HandleFunc("GET /", app.getHome)

	// activities: all require authentication:
	mux.Handle("GET /activities", usersOnly.ThenFunc(app.getActivities))
	mux.Handle("GET /activities/new", usersOnly.ThenFunc(app.getActivitiesNew))
	mux.Handle("POST /activities", usersOnly.ThenFunc(app.bellevueActivityPost))
	mux.Handle("GET /activities/{id}/edit", usersOnly.ThenFunc(app.getActivitiesIDEdit))
	mux.Handle("PUT /activities/{id}", usersOnly.ThenFunc(app.putActivitiesID))
	mux.Handle("DELETE /activities/{id}", usersOnly.ThenFunc(app.bellevueActivityDelete))

	// admin:
	mux.HandleFunc("GET /login", app.getLogin)
	mux.HandleFunc("POST /login", app.postLogin)
	// protected:
	mux.Handle("GET /logout", usersOnly.ThenFunc(app.getLogout))

	return standard.Then(mux)
}
