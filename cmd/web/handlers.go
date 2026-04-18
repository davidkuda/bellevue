package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/davidkuda/bellevue/internal/email"
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

	t.ViewModels.UninvoicedActivities, err = app.viewmodels.Activities.GetUninvoicedActivitiesForUser(t.User.ID)
	if err != nil {
		app.serverError(w, r, fmt.Errorf("could not get uninvoiced activities: %v", err))
		return
	}

	t.ViewModels.SentInvoices, err = app.viewmodels.Activities.GetAllInvoicesForUser(t.User.ID)
	if err != nil {
		app.serverError(w, r, fmt.Errorf("could not get sent invoices: %v", err))
		return
	}

	app.render(w, r, http.StatusOK, "activities.tmpl.html", &t)
}

// GET /activities/new
func (app *application) getActivitiesNew(w http.ResponseWriter, r *http.Request) {
	t := app.newTemplateData(r)
	t.Title = "New Bellevue Activity"
	t.Form = productForm{}
	app.render(w, r, http.StatusOK, "activities.new.tmpl.html", &t)
}

// HTMX: GET /activities/{ID}/edit
func (app *application) getActivitiesIDEdit(w http.ResponseWriter, r *http.Request) {
	activityIDString := r.PathValue("id")
	activityID, err := strconv.Atoi(activityIDString)
	if err != nil {
		app.serverError(w, r, fmt.Errorf("invalid activityID in path, could not parse: %v", err))
		return
	}

	t := app.newTemplateData(r)

	viewActivity, err := app.viewmodels.Activities.GetActivityByIDForUser(activityID, t.User.ID)
	if err != nil {
		app.serverError(w, r, fmt.Errorf("could not get uninvoiced activities: %v", err))
		return
	}

	t.ViewModels.Activity = viewActivity
	t.Edit = true
	t.Title = "Edit Bellevue Activity"
	t.ProductFormConfig = app.productFormConfig.WithValues(viewActivity)
	t.Form = productForm{}

	app.render(w, r, http.StatusOK, "activities.new.tmpl.html", &t)
}

// DELETE /activities/{id}
func (app *application) bellevueActivityDelete(w http.ResponseWriter, r *http.Request) {
	activityIDString := r.PathValue("id")
	activityID, err := strconv.Atoi(activityIDString)
	if err != nil {
		app.serverError(w, r, fmt.Errorf("invalid activityID in path, could not parse: %v", err))
		return
	}

	ctx := context.TODO()
	tx, err := app.db.BeginTx(ctx, nil)
	if err != nil {
		app.serverError(w, r, fmt.Errorf("failed starting transaction: %e", err))
		return
	}
	defer tx.Rollback()

	// NOTE: If there was a cascade delete, I wouldn't need a transaction and two funcs.
	// however, I don't want ease at deleting consumptions.
	app.models.Consumptions.DeleteByActivityID(activityID, tx)
	app.models.Activities.Delete(activityID, tx)

	if err := tx.Commit(); err != nil {
		app.serverError(w, r, fmt.Errorf("failed committing transaction: %s", err))
		return
	}
}

// POST /invoices
func (app *application) invoicePost(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
	userID := user.ID

	ctx := context.TODO()
	tx, err := app.db.BeginTx(ctx, nil)
	if err != nil {
		app.serverError(w, r, fmt.Errorf("failed starting transaction: %e", err))
		return
	}
	defer tx.Rollback()

	invoice, err := app.models.InvoicesV2.NewInvoiceTx(userID, tx)
	if err != nil {
		err = fmt.Errorf("could not create new invoice: %v", err)
		app.serverError(w, r, err)
		return
	}

	app.models.InvoicesV2.AssignAllOpenActivitiesToInvoiceTx(userID, invoice.ID, tx)

	tx.Commit()

	viewInvoice, err := app.viewmodels.Activities.GetInvoiceForUser(invoice.ID, user.ID)
	if err != nil {
		err = fmt.Errorf("could not get invoice invoiceID=%v userID=%v: %v", invoice.ID, user.ID, err)
		app.serverError(w, r, err)
		return
	}

	email.Send(app.EmailConfig, user, &invoice, viewInvoice)
	w.Header().Set("HX-Redirect", "/activities")
}
