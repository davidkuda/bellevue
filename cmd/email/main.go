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
)

type config struct {
	SMTP SMTPConfig
}

type SMTPConfig struct {
	Host string
	Port string
	User string
	Pass string
}

type email struct {
	from    string
	to      []string
	subject string
	body    []byte
}

type TemplateData struct {
	Subject string
	To      string
	From    string
	Name    string
}

func main() {

	cfg := loadConfigFromEnv()

	// Parse template file
	tmpl, err := template.ParseFiles("email.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	data := TemplateData{
		Subject: "Hello From Go!",
		From:    cfg.SMTP.User,
		Name:    "David",
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		log.Fatal(err)
	}

	em := email{
		from:    cfg.SMTP.User,
		to:      []string{os.Getenv("EMAIL_TO")},
		subject: "Hello From Go!",
		body:    body.Bytes(),
	}

	// DEC HEX
	//  10   A  LF => NL line feed, new line
	//  13   D  CR => Carriage Return
	//  32  20  space
	//  92  5C  \
	// 110  6E  n
	// 114  72  r
	log.Println("Unnormalized:")
	log.Println(em.body)
	r := normalizeCRLF(em.body)
	log.Println("Normalized:")
	// here, you should see a bunch of 13 10:
	log.Println([]byte(r))

	// if err := sendViaImplicitTLS(cfg, em); err != nil {
	// 	log.Fatal(err)
	// }

	log.Println("Sent OK (implicit TLS).")
}

func loadConfigFromEnv() config {
	c := config{
		SMTP: SMTPConfig{
			Host: os.Getenv("SMTP_HOST"),
			Port: os.Getenv("SMTP_PORT"),
			User: os.Getenv("SMTP_USER"),
			Pass: os.Getenv("SMTP_PASS"),
		},
	}

	var fail bool

	if c.SMTP.Host == "" {
		fail = true
		log.Print("Could not read env var SMTP_HOST")
	}

	if c.SMTP.Port == "" {
		fail = true
		log.Print("Could not read env var SMTP_PORT")
	}

	if c.SMTP.User == "" {
		fail = true
		log.Print("Could not read env var SMTP_USER")
	}

	if c.SMTP.Pass == "" {
		fail = true
		log.Print("Could not read env var SMTP_PASS")
	}

	if fail {
		os.Exit(1)
	}

	return c
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
