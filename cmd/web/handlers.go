package main

import (
	"errors"
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
	t := app.newTemplateData(r)

	invoices, err := app.models.Invoices.GetAllInvoicesOfUser(t.User.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	t.BellevueInvoices = invoices

	BAOs, err := app.models.BellevueActivities.NewBellevueActivityOverviews(t.User.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	t.BellevueActivityOverviews = BAOs

	bas, err := app.models.BellevueActivities.GetAllByUser(t.User.ID)
	if err != nil {
		log.Println(fmt.Errorf("failed reading bellevue activities: %v", err))
	}
	t.BellevueActivityOverview.BellevueActivities = bas
	t.BellevueActivityOverview.CalculateTotalPrice()
	app.render(w, r, http.StatusOK, "activities.tmpl.html", &t)
}

// GET /activities/new
func (app *application) getActivitiesNew(w http.ResponseWriter, r *http.Request) {
	t := app.newTemplateData(r)
	t.Title = "New Bellevue Activity"
	t.ProductFormConfig  = app.productFormConfig
	t.BellevueOfferings = t.BellevueActivity.NewBellevueOfferings()
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

	idStr := parts[2]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("failed converting idStr to id (int); idStr=%s:, %v\n", idStr, err)
		app.renderClientError(w, r, http.StatusBadRequest)
		return
	}

	// TODO: compare maxID

	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		err = errors.New("could not get userID from request.Context")
		app.serverError(w, r, err)
		return
	}

	activity, err := app.models.BellevueActivities.GetByID(id)
	if err != nil {
		err = fmt.Errorf("failed fetching activity by ID; id=%d: %v", id, err)
		app.serverError(w, r, err)
		return
	}

	if activity.UserID != userID {
		app.renderClientError(w, r, http.StatusUnauthorized)
		return
	}

	isHTMX := r.Header.Get("HX-Request") == "true"

	t := app.newTemplateData(r)
	t.Edit = true
	t.BellevueActivity = activity
	t.Title = "New Bellevue Activity"
	t.BellevueOfferings = activity.NewBellevueOfferings()
	t.Form = productForm{}

	if isHTMX {
		app.renderHTMXPartial(w, r, http.StatusOK, "activities.new.tmpl.html", &t)
	} else {
		app.render(w, r, http.StatusOK, "activities.new.tmpl.html", &t)
	}
}

// DELETE /bellevue-activity/:id
func (app *application) bellevueActivityDelete(w http.ResponseWriter, r *http.Request) {
	var err error

	// get ID from URL:
	parts := strings.Split(r.URL.Path, "/")

	// We expect: ["", "bellevue-activities", "{ID}"]
	if len(parts) != 3 {
		log.Println("failed splitting request URL")
		app.renderClientError(w, r, http.StatusBadRequest)
		return
	}

	idStr := parts[2]
	activityID, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("failed converting idStr to id (int); idStr=%s:, %v\n", idStr, err)
		app.renderClientError(w, r, http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		err = errors.New("could not get userID from request.Context")
		app.serverError(w, r, err)
		return
	}

	authorized, err := app.models.BellevueActivities.ActivityOwnedByUserID(activityID, userID)
	if err != nil {
		log.Printf("DELETE /bellevue-activity/%d: ActivityOwnedByUserID(%d, %d) failed: %v\n", activityID, activityID, userID, err)
		app.serverError(w, r, err)
		return
	}

	// TODO: I really need to setup testing with all the stuff implemented...
	if !authorized {
		log.Printf("DELETE /bellevue-activity/%d: ActivityOwnedByUserID(%d, %d): unauthorized request\n", activityID, activityID, userID)
		app.renderClientError(w, r, http.StatusForbidden)
		return
	}

	err = app.models.BellevueActivities.Delete(activityID)
	if err != nil {
		app.serverError(w, r, fmt.Errorf("failed BellevueActivities.Delete(actiityID=%d): %v", activityID, err))
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

// HTMX Partial: GET /activities/overview-by-months
func (app *application) getActivitiesOverviewByMonths(w http.ResponseWriter, r *http.Request) {
	t := app.newTemplateData(r)
	BAOs, err := app.models.BellevueActivities.NewBellevueActivityOverviews(t.User.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	t.BellevueActivityOverviews = BAOs
	app.renderHTMXPartial(w, r, http.StatusOK, "htmx.partial.activities.overview-by-month.tmpl.html", &t)
}

// HTMX Partial: GET /activities/by-month
func (app *application) getActivitiesByMonths(w http.ResponseWriter, r *http.Request) {
	t := app.newTemplateData(r)
	BAOs, err := app.models.BellevueActivities.NewBellevueActivityOverviews(t.User.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	t.BellevueActivityOverviews = BAOs
	app.renderHTMXPartial(w, r, http.StatusOK, "htmx.partial.activities.by-month.tmpl.html", &t)
}
