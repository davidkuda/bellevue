package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
)

// GET /settings
func (app *application) getSettings(w http.ResponseWriter, r *http.Request) {
	t := app.newTemplateData(r)

	app.render(w, r, http.StatusOK, "settings.tmpl.html", &t)
}

// GET /settings/products
func (app *application) getSettingsProducts(w http.ResponseWriter, r *http.Request) {

	funcs := template.FuncMap{
		"formatDate":          formatDate,
		"formatDateFormInput": formatDateFormInput,
		"fmtDateNiceRead":     formatDateNiceRead,
		"fmtCHF":              formatCurrency,
	}


	t, err := template.New("base").Funcs(funcs).ParseGlob("./ui/html/pages/*.html")
	if err != nil {
		fmt.Errorf("Error parsing template files: %s", err.Error())
		return
	}

	page := "settings.products.tmpl.html"

	buf := bytes.Buffer{}

	data := app.newTemplateData(r)
	err = t.ExecuteTemplate(&buf, "main", data)
	if err != nil {
		errMsg := fmt.Errorf("error executing template %s: %s", page, err.Error())
		app.serverError(w, r, errMsg)
		return
	}

	w.WriteHeader(200)

	buf.WriteTo(w)
}
