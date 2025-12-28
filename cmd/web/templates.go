package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/davidkuda/bellevue/internal/models"
)

type templateData struct {
	LoggedIn          bool
	User              models.User
	IsAdmin           bool
	Title             string
	Path              string
	RootPath          string
	HTML              template.HTML
	Today             time.Time
	ActivityDay       *models.ActivityDay
	ActivityMonths    []models.ActivityMonth
	BellevueInvoices  []models.Invoice
	ProductFormConfig models.ProductFormConfig
	Form              any
	Edit              bool // used in form templates to show render a different form
	ShowUpdatedAt     bool
	HideNav           bool
	Sidebars          bool
	HighlightJS       bool
	Error             Error

	// Feature Flags
	RenderTotalsTable bool
}

type Error struct {
	HTTPStatusCode int
	HTTPStatusText string
	Method         string
	Path           string
}

func newError(r *http.Request, errorCode int) Error {
	return Error{
		HTTPStatusCode: errorCode,
		HTTPStatusText: http.StatusText(errorCode),
		Method:         r.Method,
		Path:           r.URL.Path,
	}
}

func (app *application) newTemplateData(r *http.Request) templateData {

	// TODO: if user != nil { isAuthenticated is true }
	isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
	if !ok {
		isAuthenticated = false
	}

	var user models.User
	userPointer := app.contextGetUser(r)
	if userPointer != nil {
		user = *userPointer
	} else {
		user = models.User{}
	}

	var isAdmin bool
	if isAuthenticated {
		isAdmin, ok = r.Context().Value("isAdmin").(bool)
		if !ok {
			isAdmin = false
		}
	}

	var rootPath, title string
	i := 1
	for i < len(r.URL.Path) && r.URL.Path[i] != '/' {
		i++
	}
	rootPath = r.URL.Path[0:i]
	title = strings.ToTitle(r.URL.Path[1:i])

	var renderTotalsTable bool
	renderTotalsTableEnv := os.Getenv("FEATURE_FLAG__RENDER_TOTALS_TABLE")
	if renderTotalsTableEnv == "true" {
		renderTotalsTable = true
	}

	// TODO: using empty structs with a pointer seems so wrong here. How to fix it?
	// problem is that the templates will error on render.
	return templateData{
		LoggedIn:          isAuthenticated,
		User:              user,
		IsAdmin:           isAdmin,
		Title:             title,
		RootPath:          rootPath,
		Path:              r.URL.Path,
		Sidebars:          true,
		RenderTotalsTable: renderTotalsTable,
		Today:             time.Now(),
		ProductFormConfig: app.productFormConfig,
	}
}

func (app *application) render(w http.ResponseWriter, r *http.Request, status int, page string, data *templateData) {
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("couldn't find template \"%s\" in app.templateCache", page)
		app.serverError(w, r, err)
		return
	}

	buf := bytes.Buffer{}

	var err error
	isHTMX := r.Header.Get("HX-Request") == "true"
	if isHTMX {
		err = ts.ExecuteTemplate(&buf, "main", data)
	} else {
		err = ts.ExecuteTemplate(&buf, "base", data)
	}
	if err != nil {
		errMsg := fmt.Errorf("error executing template %s: %s", page, err.Error())
		app.serverError(w, r, errMsg)
		return
	}

	w.WriteHeader(status)

	buf.WriteTo(w)
}

// named templates => [ base -> main ]
func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	funcs := template.FuncMap{
		"formatDate":          formatDate,
		"formatDateFormInput": formatDateFormInput,
		"fmtDateNiceRead":     formatDateNiceRead,
		"fmtCHF":              formatCurrency,
		"fmtMonth":            formatMonth,
	}

	pages, err := filepath.Glob("./ui/html/*.tmpl.html")
	if err != nil {
		return nil, fmt.Errorf("failed filepath.Glob for pages: %v", err)
	}

	partials, err := filepath.Glob("./ui/html/partials/*.tmpl.html")
	if err != nil {
		return nil, fmt.Errorf("failed filepath.Glob for partials: %v", err)
	}

	for _, page := range pages {
		name := filepath.Base(page)

		N := 1 + len(partials) + 1
		files := make([]string, N)
		files[0] = "./ui/html/base.tmpl.html"
		for i, partial := range partials {
			files[i+1] = partial
		}
		files[N-1] = page

		tmpl := template.New("base").Funcs(funcs)
		t, err := tmpl.ParseFiles(files...)
		if err != nil {
			return nil, fmt.Errorf("Error parsing template files: %s", err.Error())
		}

		cache[name] = t
	}

	return cache, nil
}

func formatDate(t time.Time) string {
	return t.Format("January 2, 2006")
}

func formatDateFormInput(t time.Time) string {
	return t.Format("2006-01-02")
}

func formatDateNiceRead(t time.Time) string {
	return t.Format("Mon 2.01.2006")
}

func formatMonth(t time.Time) string {
	return t.Format("January 2006")
}

// formatCurrency converts an integer (in Rappen) to a currency string like "22.50 CHF".
func formatCurrency(value int) string {
	return fmt.Sprintf("%.2f", float64(value)/100)
}
