package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"time"

	"github.com/davidkuda/bellevue/internal/models"
	"github.com/davidkuda/bellevue/internal/viewmodels"
)

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

	Recipient     BankAccount
	Zahlungszweck string

	User        *models.User
	Invoice     *models.InvoiceV2
	ViewInvoice *viewmodels.Invoice
}

func newTemplateData(
	cfg config,
	user *models.User,
	invoice *models.InvoiceV2,
	viewInvoice *viewmodels.Invoice,
) *TemplateData {
	data := TemplateData{
		Subject:       "Deine Rechnung für den Februar 2026 im Bellevue",
		To:            user.Email,
		From:          cfg.SMTP.User,
		Date:          time.Now().Format(time.RFC1123Z),
		SenderName:    cfg.SenderName,
		SenderEmail:   cfg.SenderEmail,
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
