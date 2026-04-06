package main

import (
	"database/sql"
	"fmt"
	"log"
	"text/template"
	"time"

	"github.com/davidkuda/bellevue/internal/envcfg"
	"github.com/davidkuda/bellevue/internal/models"
)

type application struct {
	db        *sql.DB
	models    models.Models
	templates *template.Template
	config    config
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

	return

	for _, user := range users {

		if user.Email != app.config.TestEmail {
			continue
		}

		log.Printf("starting invoicing flow for user id=%d email=%s\n", user.ID, user.Email)

		numOfConsumptions, err := app.models.Consumptions.CountOpenConsumptionsForUser(user.ID)
		if err != nil {
			log.Fatalf("could not count consumptions: %s\n", err)
		}
		if numOfConsumptions == 0 {
			// TODO: maybe I can send a reminder here to advertise for this app?
			log.Println("no consumptions, skipping")
			continue
		}

		log.Printf("sending invoice with %d consumptions to %s (userID=%d)\n", numOfConsumptions, user.Email, user.ID)

		// TODO: the next few calls should probably all be in a transaction...
		invoice, err := app.models.InvoicesV2.NewInvoice(user.ID)
		if err != nil {
			log.Fatalf("could not create a new invoice user.ID=%d: %s\n", user.ID, err)
		}

		// TODO: assign open consumptions by date
		numOfAffectedConsumptions, err := app.models.InvoicesV2.AssignAllOpenConsumptionsToInvoice(user.ID, invoice.ID)
		if err != nil {
			log.Fatalf("could not assign consumptions to invoice user.ID=%d invoice.ID=%d: %s\n", user.ID, invoice.ID, err)
		}
		log.Printf("number of invoiced consumptions for %s: %d\n", user.Email, numOfConsumptions)
		if numOfAffectedConsumptions == 0 {
			log.Println("no consumptions, skipping")
			continue
		}

		// TODO: reimplement, also the email template...
		// priceCats, err := app.models.InvoicesV2.CalculatePriceCategoriesByInvoiceID(invoice.ID)
		// if err != nil {
		// 	log.Fatalf("failed retrieving price cat sums: %s\n", err)
		// }

		// send email
		// email must have price cat sums and all consumptions

		// set status to sent

		fmt.Println(invoice)
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
