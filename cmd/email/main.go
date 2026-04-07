package main

import (
	"bytes"
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

		// Comment in/out one of the next 3 blocks:

		// Either: Assign all uninvoiced activities:
		// N, err := app.models.InvoicesV2.AssignAllOpenActivitiesToInvoiceTx(user.ID, invoice.ID, tx)

		// Or: Assign activities by month:
		// MONTH := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
		// N, err := app.models.InvoicesV2.AssignOpenActivitiesByMonthToInvoiceForUserTx(
		// 	MONTH, user.ID, invoice.ID, tx,
		// )

		// Or: Assign activities by range (Q1 2026):
		start := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
		N, err := app.models.InvoicesV2.AssignOpenActivitiesByRangeToInvoiceForUserTx(
			start, end, user.ID, invoice.ID, tx,
		)

		app.config.EmailSubject = "Deine Rechnung für das Q1 2026 im Bellevue"

		if err != nil {
			log.Fatalf("could not assign activities to invoice userID=%d invoiceID=%d: %s\n", user.ID, invoice.ID, err)
		}

		log.Printf("number of invoiced activities for %s: %d\n", user.Email, N)
		if N == 0 {
			log.Fatalln("no activities, skipping")
			// log.Println("no activities, skipping")
			// continue
		}

		tx.Commit()

		viewInvoice, err := app.viewmodels.Activities.GetInvoiceForUser(invoice.ID, user.ID)
		if viewInvoice == nil {
			log.Fatal("for this to work, you need an invoice...")
		}

		data := newTemplateData(
			app.config, &user, &invoice, viewInvoice,
		)
		var buf bytes.Buffer
		if err := app.templates.ExecuteTemplate(&buf, "email", data); err != nil {
			log.Fatal(err)
		}

		em := email{
			from:    app.config.SMTP.User,
			to:      []string{user.Email},
			subject: data.Subject,
			body:    buf.Bytes(),
			// body:    normalizeCRLF(buf.Bytes()),
		}

		// fmt.Println(buf.String())

		if err := sendViaImplicitTLS(app.config, em); err != nil {
			log.Fatal(err)
		}
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
	files := []string{
		"email.tmpl",
		"email.txt.tmpl",
		"email.html.tmpl",
	}
	tmpl := template.New("email").Funcs(funcs)
	t, err := tmpl.ParseFiles(files...)
	if err != nil {
		log.Fatal(err)
	}
	app.templates = t

	return app
}

// formatCurrency converts an integer (in Rappen) to a currency string like "22.50 CHF".
func formatCurrency(value int) string {
	return fmt.Sprintf("%.2f", float64(value)/100)
}

func formatDate(t time.Time) string {
	return t.Format("2.01.2006")
}
