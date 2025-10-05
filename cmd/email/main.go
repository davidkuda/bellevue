package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"os"
	"text/template"
	"time"

	"github.com/davidkuda/bellevue/internal/envcfg"
	"github.com/davidkuda/bellevue/internal/models"
)

type email struct {
	from    string
	to      []string
	subject string
	body    []byte
}

type TemplateData struct {
	Subject     string
	To          string
	From        string
	Name        string
	Date        string
	SenderName  string
	SenderEmail string

	Recipient     BankAccount
	Zahlungszweck string

	User       models.User
	Invoice    models.Invoice
	Activities models.BellevueActivities
}

func main() {

	cfg := loadConfigFromEnv()

	db, err := envcfg.DB()
	if err != nil {
		log.Fatalf("could not open DB: %v\n", err)
	}
	defer db.Close()

	m := models.New(db)

	funcs := template.FuncMap{
		"fmtCHF":  formatCurrency,
		"fmtDate": formatDate,
	}

	// Parse template file
	tmpl, err := template.New("email.tmpl").Funcs(funcs).ParseFiles("email.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	users, err := m.Users.GetAll()
	if err != nil {
		log.Fatalf("failed fetching users from DB: %v", err)
	}

	for _, user := range users {
		fmt.Println("sending invoice to", user.Email)

		invoice, err := m.Invoices.GetInvoiceOfLastMonth(user)
		if err != nil {
			log.Fatalf("failed getting invoices of user %d: %v\n", user.ID, err)
		}

		activities, err := m.BellevueActivities.GetActivitiesOfPreviousMonth(user.ID)
		if err != nil {
			// TODO: Don't fatal, but exit the loop, try next candidate
			log.Fatalf("failed getting activites of last month: %v", err)
		}

		data := TemplateData{
			Subject:       "Deine Rechnung für den letzten Monat im Bellevue",
			To:            user.Email,
			From:          cfg.SMTP.User,
			Date:          time.Now().Format(time.RFC1123Z),
			SenderName:    cfg.SenderName,
			SenderEmail:   cfg.SenderEmail,
			Recipient:     cfg.Recipient,
			Zahlungszweck: zahlungszweck(invoice, user),
			User:          user,
			Invoice:       invoice,
			Activities:    activities,
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			log.Fatal(err)
		}

		em := email{
			from:    cfg.SMTP.User,
			to:      []string{os.Getenv("EMAIL_TO")},
			subject: data.Subject,
			// body:    normalizeCRLF(buf.Bytes()),
			body: buf.Bytes(),
		}

		if err := sendViaImplicitTLS(cfg, em); err != nil {
			log.Fatal(err)
		}

		// fmt.Println(&em.from)
		// fmt.Println(string(em.body))

		log.Printf("Sent invoice with total sum of %v CHF to %s via implicit TLS SMTP.\n", formatCurrency(data.Invoice.TotalPrice), user.Email)
	}
}

func zahlungszweck(invoice models.Invoice, user models.User) string {
	var positions string
	if invoice.TotalEating > 0 {
		positions += fmt.Sprintf("Essen %s, ", formatCurrency(invoice.TotalEating))
	}
	if invoice.TotalLectures > 0 {
		positions += fmt.Sprintf("Vorträge %s, ", formatCurrency(invoice.TotalLectures))
	}
	if invoice.TotalCoffees > 0 {
		positions += fmt.Sprintf("Kaffee %s, ", formatCurrency(invoice.TotalCoffees))
	}
	if invoice.TotalSaunas > 0 {
		positions += fmt.Sprintf("Sauna %s, ", formatCurrency(invoice.TotalSaunas))
	}
	if invoice.TotalKiosk > 0 {
		positions += fmt.Sprintf("Kiosk %s, ", formatCurrency(invoice.TotalKiosk))
	}

	positions = positions[:len(positions)-2]

	return fmt.Sprintf("%s: %s: %s", user.FirstName, invoice.Period.Format("2006-01"), positions)
}

func sendViaImplicitTLS(cfg config, em email) error {
	tlsCfg := &tls.Config{
		ServerName: cfg.SMTP.Host,
		MinVersion: tls.VersionTLS12,
	}

	log.Println("tls.Dial...")
	conn, err := tls.Dial("tcp", net.JoinHostPort(cfg.SMTP.Host, cfg.SMTP.Port), tlsCfg)
	if err != nil {
		return fmt.Errorf("tls dial: %w", err)
	}
	defer conn.Close()

	log.Println("smtp.NewClient...")
	c, err := smtp.NewClient(conn, cfg.SMTP.Host)
	if err != nil {
		return fmt.Errorf("smtp newclient: %w", err)
	}
	defer c.Quit()

	log.Println("smtp.PlainAuth...")
	auth := smtp.PlainAuth("", cfg.SMTP.User, cfg.SMTP.Pass, cfg.SMTP.Host)
	if err := c.Auth(auth); err != nil {
		return fmt.Errorf("auth: %w", err)
	}

	if err := c.Mail(em.from); err != nil {
		return fmt.Errorf("MAIL FROM: %w", err)
	}
	for _, rcpt := range em.to {
		if err := c.Rcpt(rcpt); err != nil {
			return fmt.Errorf("RCPT TO %s: %w", rcpt, err)
		}
	}

	writer, err := c.Data()
	if err != nil {
		return fmt.Errorf("DATA: %w", err)
	}

	if _, err := writer.Write(em.body); err != nil {
		return fmt.Errorf("write body: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("close data: %w", err)
	}

	return nil
}

// emails according to RFC5322 require \r\n line endings, not just \n.
// therefore, we need to normalize the line endings, which means to
// make sure that they end in \r\n. nice little leetcodish challenge :)
//
// DEC HEX
//
//	10   A  LF => NL line feed, new line
//	13   D  CR => Carriage Return
//	32  20  space
//	92  5C  \
//
// 110  6E  n
// 114  72  r
func normalizeCRLF(in []byte) []byte {
	var newLinesCount int
	for i := range in {
		if in[i] == '\n' {
			newLinesCount++
		}
	}

	out := make([]byte, len(in)+newLinesCount, len(in)+newLinesCount)

	var k int // outPointer
	for i := range in {
		if in[i] == '\n' {
			if i == 0 || in[i-1] != '\r' {
				out[k] = '\r'
				k++
			}
		}
		out[k] = in[i]
		k++
	}
	return out
}

// formatCurrency converts an integer (in Rappen) to a currency string like "22.50 CHF".
func formatCurrency(value int) string {
	return fmt.Sprintf("%.2f", float64(value)/100)
}

func formatDate(t time.Time) string {
	return t.Format("Mon 2.01.2006")
}
