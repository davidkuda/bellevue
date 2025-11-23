package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/davidkuda/bellevue/internal/envcfg"
	"github.com/davidkuda/bellevue/internal/models"

	"github.com/coreos/go-oidc/v3/oidc"
	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/oauth2"
)

type application struct {
	models models.Models

	productFormConfig  models.ProductFormConfig
	priceCategoryIDMap models.PriceCategoryIDMap
	productIDMap       models.ProductIDMap

	templateCache map[string]*template.Template
	OIDC          openIDConnect

	CookieDomain string

	JWT struct {
		Secret   []byte
		Issuer   string
		Audience string // TODO: should this be []string?
	}
}

var (
	clientID     = os.Getenv("OIDC_CLIENT_ID")
	clientSecret = os.Getenv("OIDC_CLIENT_SECRET")
	issuer       = os.Getenv("OIDC_ISSUER")
	redirectURL  = os.Getenv("OIDC_REDIRECT_URL")
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	addr := flag.String("addr", ":8875", "HTTP network address")
	flag.Parse()

	cookieDomain := flag.String("cookie-domain", os.Getenv("COOKIE_DOMAIN"), "localhost or kuda.ai")
	if *cookieDomain == "" {
		log.Fatal("fail startup: make sure to either pass -cookie-domain [localhost|kuda.ai] or define env var COOKIE_DOMAIN")
	}

	app := &application{}

	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		log.Fatal(err)
	}
	app.OIDC.provider = provider

	oidcConfig := &oidc.Config{
		ClientID: clientID,
	}
	app.OIDC.verifier = provider.Verifier(oidcConfig)

	app.OIDC.config = oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  redirectURL,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

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

	app.priceCategoryIDMap, err = app.models.PriceCategories.GetPriceCatMap()
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

	log.Print(fmt.Sprintf("Starting web server, listening on %s", *addr))
	err = http.ListenAndServe(*addr, app.routes())
	log.Fatal(err)
}
