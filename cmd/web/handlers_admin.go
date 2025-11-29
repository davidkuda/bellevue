package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/davidkuda/bellevue/internal/models"
	"github.com/pascaldekloe/jwt"
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
		firstName:    r.PostForm.Get("first-name"),
		lastName:    r.PostForm.Get("last-name"),
		email:    r.PostForm.Get("email"),
		password: r.PostForm.Get("password"),
	}

	// TODO: use captcha
	// TODO: validate email
	// TODO: validate password

	user := models.User{
		FirstName: form.firstName,
		LastName: form.lastName,
		Email: form.email,
	}

	err = app.models.Users.InsertPassword(user, form.password)
	if err != nil {
		app.serverError(w, r, fmt.Errorf("failed to insert user=%+v to db: %s", user, err))
	}

	// TODO: send JWT cookie, or render login-form
	// TODO: do we need user activation emails?

	http.Redirect(w, r, "/login", http.StatusSeeOther)
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

	var claims jwt.Claims
	claims.Subject = form.email
	claims.Issued = jwt.NewNumericTime(time.Now())
	claims.NotBefore = jwt.NewNumericTime(time.Now())
	claims.Expires = jwt.NewNumericTime(time.Now().Add(24 * time.Hour))
	claims.Issuer = app.JWT.Issuer
	claims.Audiences = []string{app.JWT.Audience}

	jwtBytes, err := claims.HMACSign(jwt.HS256, []byte(app.JWT.Secret))
	if err != nil {
		log.Printf("error signing jwt: %v\n", err)
		w.Write([]byte("Login failed, incorrect credentials. Please try again."))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "id",
		Value:    string(jwtBytes),
		Domain:   app.CookieDomain,
		Expires:  time.Now().Add(10 * 24 * time.Hour),
		Secure:   true,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("HX-Redirect", "/activities")
	w.WriteHeader(http.StatusOK)

}

// GET /logout
func (app *application) getLogout(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
