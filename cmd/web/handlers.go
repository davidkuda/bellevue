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

	// GetActivitiesByInvoice
	// invoice: id
	// activities: []activity
	// render table by invoice_id => templates struct
	// render past invoices underneath
	// implement an invoices page
	// template activities.tmpl.html if .UninvoicedActivities or .Invoices
	t.ViewModels.UninvoicedActivities, err = app.viewmodels.Activities.GetUninvoicedActivitiesForUser(t.User.ID)
	if err != nil {
		app.serverError(w, r, fmt.Errorf("could not get uninvoiced activities: %v", err))
		return
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

// HTMX: GET /activities/{ID}/edit
func (app *application) getActivitiesIDEdit(w http.ResponseWriter, r *http.Request) {
	activityID := r.PathValue("id")
	fmt.Println(activityID)

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
