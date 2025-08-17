package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/davidkuda/bellevue/internal/envcfg"
	"github.com/davidkuda/bellevue/internal/models"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type application struct {
	models models.Models

	templateCache     map[string]*template.Template
	templateCacheHTMX map[string]*template.Template

	JWT struct {
		Secret       []byte
		CookieDomain string
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	addr := flag.String("addr", ":8875", "HTTP network address")

	cookieDomain := flag.String("cookie-domain", os.Getenv("COOKIE_DOMAIN"), "localhost or kuda.ai")
	if *cookieDomain == "" {
		log.Fatal("fail startup: make sure to either pass -cookie-domain [localhost|kuda.ai] or define env var COOKIE_DOMAIN")
	}

	app := &application{}

	c := envcfg.Get()

	app.JWT.Secret = c.JWT.Secret
	app.JWT.CookieDomain = *cookieDomain

	db, err := envcfg.DB()
	if err != nil {
		log.Fatalf("could not open DB: %v\n", err)
	}
	defer db.Close()

	app.models = models.New(db)

	app.templateCache, err = newTemplateCache()
	if err != nil {
		log.Fatalf("could not initialise templateCache: %v\n", err)
	}

	app.templateCacheHTMX, err = newTemplateCacheForHTMXPartials()
	if err != nil {
		log.Fatalf("could not initialise templateCache: %v\n", err)
	}

	log.Print(fmt.Sprintf("Starting web server, listening on %s", *addr))
	err = http.ListenAndServe(*addr, app.routes())
	log.Fatal(err)
}
