package email

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"text/template"
	"time"

	"github.com/davidkuda/bellevue/internal/models"
	"github.com/davidkuda/bellevue/internal/viewmodels"
)

func Send(
	cfg EmailConfig,
	user *models.User,
	invoice *models.InvoiceV2,
	viewInvoice *viewmodels.Invoice,
) error {
	files := []string{
		"internal/email/email.tmpl",
		"internal/email/email.txt.tmpl",
		"internal/email/email.html.tmpl",
	}

	funcs := template.FuncMap{
		"fmtCHF":  formatCurrency,
		"fmtDate": formatDate,
	}

	tmpl := template.New("email").Funcs(funcs)
	t, err := tmpl.ParseFiles(files...)
	if err != nil {
		log.Fatal(err)
	}

	data := newTemplateData(
		cfg, user, invoice, viewInvoice,
	)

	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "email", data); err != nil {
		err = fmt.Errorf("could not execute template: %v", err)
		return err
	}

	em := email{
		from:    cfg.SMTP.User,
		to:      []string{user.Email},
		subject: data.Subject,
		body:    buf.Bytes(),
	}

	if err := sendViaImplicitTLS(cfg, em); err != nil {
		err = fmt.Errorf("could not send email: %v", err)
		return err
	}

	return nil
}

type email struct {
	from    string
	to      []string
	subject string
	body    []byte
}

func newEmail() {}

type TemplateData struct {
	Subject     string
	To          string
	From        string
	Name        string
	Date        string
	SenderName  string
	SenderEmail string
	Domain      string

	Recipient     BankAccount
	Zahlungszweck string

	User        *models.User
	Invoice     *models.InvoiceV2
	ViewInvoice *viewmodels.Invoice
}

func newTemplateData(
	cfg EmailConfig,
	user *models.User,
	invoice *models.InvoiceV2,
	viewInvoice *viewmodels.Invoice,
) *TemplateData {
	data := TemplateData{
		Subject:       cfg.EmailSubject,
		To:            user.Email,
		From:          cfg.SMTP.User,
		Date:          time.Now().Format(time.RFC1123Z),
		SenderName:    cfg.SenderName,
		SenderEmail:   cfg.SenderEmail,
		Domain:        cfg.Domain,
		Recipient:     cfg.Recipient,
		Zahlungszweck: zahlungszweck(viewInvoice, user),
		User:          user,
		Invoice:       invoice,
		ViewInvoice:   viewInvoice,
	}
	return &data
}

func zahlungszweck(invoice *viewmodels.Invoice, user *models.User) string {
	var positions string

	for _, cat := range invoice.Categories {
		positions += fmt.Sprintf("%s %s, ", cat.Name, formatCurrency(cat.TotalPrice))
	}

	// remove trailing comma and space char
	positions = positions[:len(positions)-2]

	positions = fmt.Sprintf("%s: %s", user.FirstName, positions)

	// TODO: it would be nice to add something like 2026-03 or 2026-Q1 to the string

	return positions
}

func sendViaImplicitTLS(cfg EmailConfig, em email) error {
	tlsCfg := &tls.Config{
		ServerName: cfg.SMTP.Host,
		MinVersion: tls.VersionTLS12,
	}

	conn, err := tls.Dial("tcp", net.JoinHostPort(cfg.SMTP.Host, cfg.SMTP.Port), tlsCfg)
	if err != nil {
		return fmt.Errorf("tls dial: %w", err)
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, cfg.SMTP.Host)
	if err != nil {
		return fmt.Errorf("smtp newclient: %w", err)
	}
	defer c.Quit()

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

// formatCurrency converts an integer (in Rappen) to a currency string like "22.50 CHF".
func formatCurrency(value int) string {
	return fmt.Sprintf("%.2f", float64(value)/100)
}

func formatDate(t time.Time) string {
	return t.Format("2.01.2006")
}
