package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// GET /
func (app *application) getHome(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/activities", http.StatusSeeOther)
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
	var userID int
	userID = app.contextGetUser(r).ID
	today := time.Now()
	data, _ := app.models.Activities.GetActivityDayForUser(today, userID)
	if len(data.Items) > 0 {
		endpoint := "/activities/edit?date=" + formatDateFormInput(today)
		http.Redirect(w, r, endpoint, http.StatusSeeOther)
		return
	}

	t := app.newTemplateData(r)
	t.Title = "New Bellevue Activity"
	t.Form = productForm{}
	app.render(w, r, http.StatusOK, "activities.new.tmpl.html", &t)
}

// GET /activities/edit?date=2025-11-26
func (app *application) getActivitiesEdit(w http.ResponseWriter, r *http.Request) {
	var err error

	var tm time.Time
	dateStr := r.URL.Query().Get("date")
	if dateStr != "" {
		tm, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			app.renderClientError(w, r, http.StatusBadRequest)
			return
		}
	} else {
		tm = time.Now()
	}

	var userID int
	userID = app.contextGetUser(r).ID

	activityDay, err := app.models.Activities.GetActivityDayForUser(tm, userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	t := app.newTemplateData(r)
	t.Title = "Edit Bellevue Activity"
	t.ActivityDay = activityDay
	t.ProductFormConfig = app.productFormConfig.WithValues(activityDay)
	t.Form = productForm{}
	app.render(w, r, http.StatusOK, "activities.new.tmpl.html", &t)
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

	t := app.newTemplateData(r)
	t.Edit = true
	t.Title = "New Bellevue Activity"
	t.Form = productForm{}

	app.render(w, r, http.StatusOK, "activities.new.tmpl.html", &t)
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
