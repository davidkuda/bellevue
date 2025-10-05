package main

import (
	"log"
	"os"
)

type config struct {
	SMTP SMTPConfig

	SenderName  string
	SenderEmail string

	Recipient BankAccount
}

type SMTPConfig struct {
	Host string
	Port string
	User string
	Pass string
}

type BankAccount struct {
	IBAN   string
	Name   string
	Street string
	PLZOrt string
}

func loadConfigFromEnv() config {
	c := config{
		SenderName:  os.Getenv("SENDER_NAME"),
		SenderEmail: os.Getenv("SENDER_EMAIL_ADDRESS"),
		SMTP: SMTPConfig{
			Host: os.Getenv("SMTP_HOST"),
			Port: os.Getenv("SMTP_PORT"),
			User: os.Getenv("SMTP_USER"),
			Pass: os.Getenv("SMTP_PASS"),
		},
		Recipient: BankAccount{
			IBAN: os.Getenv("RECIPIENT_IBAN"),
			Name: os.Getenv("RECIPIENT_NAME"),
			Street: os.Getenv("RECIPIENT_STREET"),
			PLZOrt: os.Getenv("RECIPIENT_PLZ_ORT"),
		},
	}

	var fail bool

	if c.SenderName == "" {
		fail = true
		log.Print("Could not read env var SENDER_NAME")
	}

	if c.SenderEmail == "" {
		fail = true
		log.Print("Could not read env var SENDER_EMAIL_ADDRESS")
	}

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

	if c.Recipient.IBAN == "" {
		fail = true
		log.Print("Could not read env var RECIPIENT_IBAN")
	}

	if fail {
		os.Exit(1)
	}

	return c
}
