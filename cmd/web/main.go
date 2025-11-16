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

	productFormConfig models.ProductFormConfig
	priceCategoryMap  models.PriceCategoryMap
	productIDMap      models.ProductIDMap

	templateCache         map[string]*template.Template
	templateCacheHTMX     map[string]*template.Template
	templateCacheSettings map[string]*template.Template

	CookieDomain string

	JWT struct {
		Secret   []byte
		Issuer   string
		Audience string // TODO: should this be []string?
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	addr := flag.String("addr", ":8875", "HTTP network address")
	flag.Parse()

	cookieDomain := flag.String("cookie-domain", os.Getenv("COOKIE_DOMAIN"), "localhost or kuda.ai")
	if *cookieDomain == "" {
		log.Fatal("fail startup: make sure to either pass -cookie-domain [localhost|kuda.ai] or define env var COOKIE_DOMAIN")
	}

	app := &application{}

	c := envcfg.Get()

	app.CookieDomain = *cookieDomain
	app.JWT.Secret = c.JWT.Secret
	app.JWT.Issuer = c.JWT.Issuer
	app.JWT.Audience = c.JWT.Audience

	db, err := envcfg.DB()
	if err != nil {
		log.Fatalf("could not open DB: %v\n", err)
	}
	defer db.Close()

	app.models = models.New(db)

	app.productFormConfig, err = app.models.Products.GetProductFormConfig()
	if err != nil {
		log.Fatalf("could not load productFormConfig: %v\n", err)
	}

	app.priceCategoryMap, err = app.models.PriceCategories.GetPriceCatMap()
	if err != nil {
		log.Fatalf("could not load app.priceCategoryMap: %v\n", err)
	}

	app.productIDMap, err = app.models.Products.GetProductIDMap()
	if err != nil {
		log.Fatalf("could not load app.productIDMap: %v\n", err)
	}

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
