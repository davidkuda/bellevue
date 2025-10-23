package main

import (
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/davidkuda/bellevue/internal/models"
)

var (
	FieldError = errors.New("FieldError")
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
	t.BellevueOfferings = t.BellevueActivity.NewBellevueOfferings()
	t.Form = bellevueActivityForm{}

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
	t.Form = bellevueActivityForm{}

	if isHTMX {
		app.renderHTMXPartial(w, r, http.StatusOK, "activities.new.tmpl.html", &t)
	} else {
		app.render(w, r, http.StatusOK, "activities.new.tmpl.html", &t)
	}
}

// PUT /activity/:id
func (app *application) putActivitiesID(w http.ResponseWriter, r *http.Request) {
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
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("failed converting idStr to id (int); idStr=%s:, %v\n", idStr, err)
		app.renderClientError(w, r, http.StatusBadRequest)
		return
	}

	err = r.ParseForm()
	if err != nil {
		log.Printf("Failed parsing form: %v", err)
		app.renderClientError(w, r, http.StatusBadRequest)
		return
	}

	form := bellevueActivityForm{}
	err = form.parseRequest(r)
	if err != nil {
		if err == FieldError {
			t := app.newTemplateDataBellevueActivity(r, form)
			app.render(w, r, http.StatusUnprocessableEntity, "activities.new.tmpl.html", &t)
			return
		} else {
			log.Println(fmt.Errorf("failed parsing form bellevue activity: %v", err))
			app.renderClientError(w, r, http.StatusUnprocessableEntity)
			return
		}
	}

	form.ID = id

	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		err = errors.New("could not get userID from request.Context")
		app.serverError(w, r, err)
		return
	}
	form.UserID = userID

	authorized, err := app.models.BellevueActivities.ActivityOwnedByUserID(form.ID, form.UserID)
	if err != nil {
		log.Printf("PUT /bellevue-activity/%d: ActivityOwnedByUserID(%d, %d) failed: %v\n", id, id, form.UserID, err)
		app.serverError(w, r, err)
		return
	}

	// TODO: I really need to setup testing with all the stuff implemented...
	if !authorized {
		log.Printf("PUT /bellevue-activity/%d: ActivityOwnedByUserID(%d, %d): unauthorized request\n", form.ID, form.ID, form.UserID)
		app.serverError(w, r, err)
		return
	}

	a := form.toModel()

	err = app.models.BellevueActivities.Update(a)
	if err != nil {
		err = fmt.Errorf("PUT /bellevue-activity/%d: failed app.models.BellevueActivites.Update: %v", id, err)
		// TODO: Now with HTMX, app.serverError does no longer give feedback to the user
		// app.clientError does not work, either. needs partials.
		app.serverError(w, r, err)
		return
	}

	t := app.newTemplateData(r)

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
	app.renderHTMXPartial(w, r, http.StatusOK, "activities.tmpl.html", &t)
}

// POST /activity
func (app *application) bellevueActivityPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Printf("Failed parsing form: %v", err)
		app.renderClientError(w, r, http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		log.Println("post /bellevue-activity: could not get userID from request.Context")
		app.serverError(w, r, err)
		return
	}

	form := bellevueActivityForm{}
	err = form.parseRequest(r)
	if err != nil {
		if err == FieldError {
			t := app.newTemplateDataBellevueActivity(r, form)
			app.render(w, r, http.StatusUnprocessableEntity, "admin.new_bellevue_activity.tmpl.html", &t)
			return
		} else {
			log.Println(fmt.Errorf("failed parsing form bellevue activity: %v", err))
			app.renderClientError(w, r, http.StatusUnprocessableEntity)
			return
		}
	}

	b := form.toModel()

	b.UserID = userID

	err = app.models.BellevueActivities.Insert(b)
	if err != nil {
		log.Printf("app.bellevueActivities.Insert(): %v\n", err)
		app.serverError(w, r, err)
		return
	}

	// TODO: send some notification (Toast) to the UI (successfully submitted)
	http.Redirect(w, r, "/bellevue-activities", http.StatusSeeOther)
	return
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

func (app *application) someHTMXPartial(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<p>I am a partial.</p>"))
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


func (app *application) newTemplateDataBellevueActivity(r *http.Request, form bellevueActivityForm) templateData {
	t := app.newTemplateData(r)
	t.BellevueActivity = form.toModel()
	t.BellevueActivity.PopulateItems()
	t.BellevueOfferings = t.BellevueActivity.NewBellevueOfferings()
	t.Form = form
	return t
}

type bellevueActivityForm struct {
	ID          int
	UserID      int
	Date        time.Time
	Breakfasts  int
	Lunches     int
	Dinners     int
	Coffees     int
	Saunas      int
	Lectures    int
	SnacksCHF   int
	Comment     string
	FieldErrors map[string]string
}

func (form *bellevueActivityForm) parseRequest(r *http.Request) error {
	f := r.PostForm

	dateStr := f.Get("bellevue-activity-date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("form.Date: stconv.Atoi: someone wants to write non-integers: value: %v, err: %v", f.Get("bellevue-activity-breakfasts"), err)
	}

	errorTemplate := "form.%v: stconv.Atoi: someone wants to write non-integers: value: %v, err: %v"

	// although the form has client side validation of integers by using
	// <input type="number">, a malicious actor could still place a POST
	// request not via the web form.
	breakfasts, err := strconv.Atoi(f.Get("bellevue-activity-breakfasts"))
	if err != nil {
		return fmt.Errorf(errorTemplate, "Breakfasts", f.Get("bellevue-activity-breakfasts"), err)
	}

	lunches, err := strconv.Atoi(f.Get("bellevue-activity-lunches"))
	if err != nil {
		return fmt.Errorf(errorTemplate, "Lunches", f.Get("bellevue-activity-lunches"), err)
	}

	dinners, err := strconv.Atoi(f.Get("bellevue-activity-dinners"))
	if err != nil {
		return fmt.Errorf(errorTemplate, "Dinners", f.Get("bellevue-activity-dinners"), err)
	}

	coffees, err := strconv.Atoi(f.Get("bellevue-activity-coffees"))
	if err != nil {
		return fmt.Errorf(errorTemplate, "Coffees", f.Get("bellevue-activity-coffees"), err)
	}

	saunas, err := strconv.Atoi(f.Get("bellevue-activity-saunas"))
	if err != nil {
		return fmt.Errorf(errorTemplate, "Saunas", f.Get("bellevue-activity-saunas"), err)
	}

	lectures, err := strconv.Atoi(f.Get("bellevue-activity-lectures"))
	if err != nil {
		return fmt.Errorf(errorTemplate, "Lectures", f.Get("bellevue-activity-lectures"), err)
	}

	var snacksCHF int
	snackCHFString := f.Get("bellevue-activity-snacks")
	if len(snackCHFString) > 0 {
		priceFloat, err := strconv.ParseFloat(snackCHFString, 64)
		if err != nil {
			return fmt.Errorf("failed parsing string \"%s\" to float:", snackCHFString)
		}
		snacksCHF = int(math.Round(priceFloat * 100))
	}

	form.Date = date
	form.Breakfasts = breakfasts
	form.Lunches = lunches
	form.Dinners = dinners
	form.Coffees = coffees
	form.Saunas = saunas
	form.Lectures = lectures
	form.SnacksCHF = snacksCHF
	form.Comment = f.Get("bellevue-activity-comment")
	form.FieldErrors = map[string]string{}

	if form.hasNegativeNumbers() {
		form.FieldErrors["negatives"] = "you may not send negative numbers."
	}

	if form.hasOnlyZeroes() {
		form.FieldErrors["zeroes"] = "you have all 0 and therefore not any activity to upload."
	}

	if len(form.FieldErrors) > 0 {
		return FieldError
	}

	return nil
}

func (b *bellevueActivityForm) hasNegativeNumbers() bool {
	if b.Breakfasts < 0 {
		return true
	}
	if b.Lunches < 0 {
		return true
	}
	if b.Dinners < 0 {
		return true
	}
	if b.Coffees < 0 {
		return true
	}
	if b.Saunas < 0 {
		return true
	}
	if b.Lectures < 0 {
		return true
	}
	if b.SnacksCHF < 0 {
		return true
	}
	return false
}

func (b *bellevueActivityForm) hasOnlyZeroes() bool {
	if b.Breakfasts > 0 {
		return false
	}
	if b.Lunches > 0 {
		return false
	}
	if b.Dinners > 0 {
		return false
	}
	if b.Coffees > 0 {
		return false
	}
	if b.Saunas > 0 {
		return false
	}
	if b.Lectures > 0 {
		return false
	}
	if b.SnacksCHF > 0 {
		return false
	}
	return true
}

func (b *bellevueActivityForm) toModel() *models.BellevueActivity {
	return &models.BellevueActivity{
		ID:         b.ID,
		UserID:     b.UserID,
		Date:       b.Date,
		Breakfasts: b.Breakfasts,
		Lunches:    b.Lunches,
		Dinners:    b.Dinners,
		Coffees:    b.Coffees,
		Saunas:     b.Saunas,
		SnacksCHF:  b.SnacksCHF,
		Lectures:   b.Lectures,
		Comment:    b.Comment,
	}
}
