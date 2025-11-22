package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// GET /
func (app *application) getHome(w http.ResponseWriter, r *http.Request) {
	if app.isAuthenticated(r) {
		http.Redirect(w, r, "/activities", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

// GET /activities
func (app *application) getActivities(w http.ResponseWriter, r *http.Request) {
	var err error

	t := app.newTemplateData(r)

	t.ActivityMonths, err = app.models.Activities.GetActivityMonths(t.User.ID)
	if err != nil {
		app.serverError(w, r, fmt.Errorf("could not get activity months: %e", err))
	}

	app.render(w, r, http.StatusOK, "activities.tmpl.html", &t)
}

// GET /activities/new
func (app *application) getActivitiesNew(w http.ResponseWriter, r *http.Request) {
	t := app.newTemplateData(r)
	t.Title = "New Bellevue Activity"
	t.ProductFormConfig  = app.productFormConfig
	t.Form = productForm{}

	isHTMX := r.Header.Get("HX-Request") == "true"
	if isHTMX {
		app.renderHTMXPartial(w, r, http.StatusOK, "activities.new.tmpl.html", &t)
	} else {
		app.render(w, r, http.StatusOK, "activities.new.tmpl.html", &t)
	}

}

// HTMX: GET /activities/{ID}/edit
func (app *application) getActivitiesIDEdit(w http.ResponseWriter, r *http.Request) {
	// get activity ID:
	parts := strings.Split(r.URL.Path, "/")

	// We expect: ["", "bellevue-activities", "{ID}", "edit"]
	if len(parts) != 4 {
		log.Println("failed splitting request URL")
		app.renderClientError(w, r, http.StatusBadRequest)
		return
	}


	isHTMX := r.Header.Get("HX-Request") == "true"

	t := app.newTemplateData(r)
	t.Edit = true
	t.Title = "New Bellevue Activity"
	t.Form = productForm{}

	if isHTMX {
		app.renderHTMXPartial(w, r, http.StatusOK, "activities.new.tmpl.html", &t)
	} else {
		app.render(w, r, http.StatusOK, "activities.new.tmpl.html", &t)
	}
}

// DELETE /bellevue-activity/:id
func (app *application) bellevueActivityDelete(w http.ResponseWriter, r *http.Request) {

	// get ID from URL:
	parts := strings.Split(r.URL.Path, "/")

	// We expect: ["", "bellevue-activities", "{ID}"]
	if len(parts) != 3 {
		log.Println("failed splitting request URL")
		app.renderClientError(w, r, http.StatusBadRequest)
		return
	}

	// w.Header().Add("HX-Trigger-After-Settle", `{"refresh-table": {"reason":"item-deleted"}}"`)
	w.Header().Add("HX-Trigger-After-Settle", "refresh-table")
}

// PATCH /invoices/{id}?set-state={state}
func (app *application) patchInvoicesIDState(w http.ResponseWriter, r *http.Request) {
	// get ID from request URL:
	path := strings.TrimPrefix(r.URL.Path, "/invoices/")
	id, err := strconv.Atoi(path)
	if err != nil {
		// TODO: Should I send error to app.renderClientError for logging? or log in an err block?
		log.Printf("failed converting path to id (int); path=%s:, %v\n", path, err)
		app.renderClientError(w, r, http.StatusBadRequest)
		return
	}

	// Query param: set-state
	state := r.URL.Query().Get("set-state")
	log.Println("state:", state)

	// TODO: get enum from postgres, maybe put it in a map[string]bool and check with if _, ok := map[state]; !ok {}
	if state != "unpaid" && state != "paid" {
		log.Printf("received invalid state: state=%s\n", state)
		app.renderClientError(w, r, http.StatusBadRequest)
		return
	}

	// TODO: check if user has permission to change state

	// TODO: update state in postgres

	log.Printf("id=%s, state=%s", id, state)
}

// for HTMX: GET /activities?month="2025-05"
