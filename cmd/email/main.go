package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"text/template"
	"time"

	"github.com/davidkuda/bellevue/internal/envcfg"
	"github.com/davidkuda/bellevue/internal/models"
	"github.com/davidkuda/bellevue/internal/viewmodels"
)

type application struct {
	db         *sql.DB
	models     models.Models
	viewmodels viewmodels.Models
	templates  *template.Template
	config     config
}

func main() {
	log.Println("starting invoice and email flow")

	app := newApplication()
	defer app.db.Close()

	users, err := app.models.Users.GetAll()
	if err != nil {
		log.Fatalf("failed fetching users from DB: %v", err)
	}

	for _, user := range users {
		fmt.Println(user.Email)
	}

	for _, user := range users {

		if app.config.TestEmail != "" {
			if user.Email != app.config.TestEmail {
				continue
			}
		}

		log.Printf("starting invoicing flow for user id=%d email=%s\n", user.ID, user.Email)

		numUninvoicedActivities, err := app.models.Activities.CountUninvoicedActivitiesForUser(user.ID)
		if err != nil {
			log.Fatalf("could not count activities: %s\n", err)
		}
		if numUninvoicedActivities == 0 {
			// TODO: maybe I can send a reminder here to advertise for this app?
			log.Println("no consumptions, skipping")
			continue
		}

		log.Printf("sending invoice with %d activities to %s (userID=%d)\n", numUninvoicedActivities, user.Email, user.ID)

		ctx := context.TODO()
		tx, err := app.db.BeginTx(ctx, nil)
		if err != nil {
			log.Fatalf("failed starting transaction: %v\n", err)
			return
		}
		defer tx.Rollback()

		invoice, err := app.models.InvoicesV2.NewInvoiceTx(user.ID, tx)
		if err != nil {
			log.Fatalf("could not create a new invoice user.ID=%d: %s\n", user.ID, err)
		}

		MONTH := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)
		N, err := app.models.InvoicesV2.AssignOpenActivitiesByMonthToInvoiceForUserTx(
			MONTH, user.ID, invoice.ID, tx,
		)
		if err != nil {
			log.Fatalf("could not assign activities to invoice user.ID=%d invoice.ID=%d: %s\n", user.ID, invoice.ID, err)
		}

		log.Printf("number of invoiced activities for %s: %d\n", user.Email, N)
		if N == 0 {
			log.Fatalln("no activities, skipping")
			// log.Println("no activities, skipping")
			// continue
		}

		// TODO: uncomment once ready
		// tx.Commit()
		tx.Rollback()

		// TODO: rename function to GetInvoiceForUser, drop word Sent
		viewInvoice, err := app.viewmodels.Activities.GetSentInvoiceForUser(invoice.ID, user.ID)
		fmt.Println(viewInvoice)

		// send email
		// email must have price cat sums and all consumptions

		// set status to sent

	}
}

func newApplication() application {
	app := application{}

	cfg := loadConfigFromEnv()
	app.config = cfg

	db, err := envcfg.DB()
	if err != nil {
		log.Fatalf("could not open DB: %v\n", err)
	}
	app.db = db

	m := models.New(db)
	app.models = m

	vm := viewmodels.New(db)
	app.viewmodels = vm

	funcs := template.FuncMap{
		"fmtCHF":  formatCurrency,
		"fmtDate": formatDate,
	}

	// Parse template file
	tmpl, err := template.New("email.tmpl").Funcs(funcs).ParseFiles("email.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	app.templates = tmpl

	return app
}

// formatCurrency converts an integer (in Rappen) to a currency string like "22.50 CHF".
func formatCurrency(value int) string {
	return fmt.Sprintf("%.2f", float64(value)/100)
}

func formatDate(t time.Time) string {
	return t.Format("Mon 2.01.2006")
}
