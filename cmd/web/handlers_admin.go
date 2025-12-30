package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/davidkuda/bellevue/internal/models"
)

// GET /login
func (app *application) getLogin(w http.ResponseWriter, r *http.Request) {
	if app.isAuthenticated(r) {
		http.Redirect(w, r, "/activities", http.StatusSeeOther)
		return
	}

	t := app.newTemplateData(r)
	app.render(w, r, 200, "login.tmpl.html", &t)
}

// GET /login/email: Partial: Login form that asks for email and password.
func (app *application) getLoginEmail(w http.ResponseWriter, r *http.Request) {
	if app.isAuthenticated(r) {
		http.Redirect(w, r, "/activities", http.StatusSeeOther)
		return
	}

	t := app.newTemplateData(r)
	app.render(w, r, 200, "login.email.tmpl.html", &t)
}

// GET /signup: Partial: Signup form that asks for name, email and password.
func (app *application) getLoginSignup(w http.ResponseWriter, r *http.Request) {
	if app.isAuthenticated(r) {
		http.Redirect(w, r, "/activities", http.StatusSeeOther)
		return
	}

	t := app.newTemplateData(r)
	app.render(w, r, 200, "login.signup.tmpl.html", &t)
}

// POST /signup
func (app *application) postSignup(w http.ResponseWriter, r *http.Request) {
	var err error

	type userSignupForm struct {
		firstName string
		lastName  string
		email     string
		password  string
	}
	err = r.ParseForm()
	if err != nil {
		app.serverError(w, r, fmt.Errorf("failed parsing form: %e", err))
		return
	}
	form := userSignupForm{
		firstName: r.PostForm.Get("first-name"),
		lastName:  r.PostForm.Get("last-name"),
		email:     r.PostForm.Get("email"),
		password:  r.PostForm.Get("password"),
	}

	// TODO: use captcha
	// TODO: validate email
	// TODO: validate password

	user := models.User{
		FirstName: form.firstName,
		LastName:  form.lastName,
		Email:     form.email,
	}

	userID, err := app.models.Users.InsertPassword(user, form.password)
	if err != nil {
		app.serverError(w, r, fmt.Errorf("failed to insert user=%+v to db: %s", user, err))
	}
	app.sessionManager.Put(r.Context(), "UserID", userID)
	// TODO: do we need user activation emails?

	http.Redirect(w, r, "/activities", http.StatusSeeOther)
}

// POST /login
func (app *application) postLogin(w http.ResponseWriter, r *http.Request) {
	type userLoginForm struct {
		email    string
		password string
	}
	err := r.ParseForm()
	if err != nil {
		log.Printf("Failed parsing form: %v", err)
		w.Write([]byte("Login failed, incorrect credentials. Please try again."))
		return
	}
	form := userLoginForm{
		email:    r.PostForm.Get("email"),
		password: r.PostForm.Get("password"),
	}

	err = app.models.Users.Authenticate(form.email, form.password)
	if err != nil {
		log.Printf("error authenticating user with username %s and password %s: %v\n", form.email, form.password, err)
		w.Write([]byte("Login failed, incorrect credentials. Please try again."))
		return
	}

	u, err := app.models.Users.GetUserByEmail(form.email)
	if err != nil {
		app.serverError(w, r, fmt.Errorf("could not get userID by email=%s: %s", form.email, err))
		return
	}
	userID := u.ID

	app.sessionManager.Put(r.Context(), "UserID", userID)
	w.Header().Set("HX-Redirect", "/activities")
}

// GET /logout
func (app *application) getLogout(w http.ResponseWriter, r *http.Request) {
	app.sessionManager.Destroy(r.Context())
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
