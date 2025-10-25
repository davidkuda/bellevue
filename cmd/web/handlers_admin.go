package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

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

	w.Header().Set("HX-Redirect", "/admin")
	w.WriteHeader(http.StatusOK)

}

// GET /logout
func (app *application) getLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "id",
		Value:    "",
		Domain:   app.CookieDomain,
		Expires:  time.Now(),
		Secure:   true,
		HttpOnly: true,
		Path:     "/",
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (app *application) isAuthenticated(r *http.Request) bool {
	err := app.validateJWTCookie(r)
	if err == nil {
		return true
	} else {
		return false
	}
}

func (app *application) extractUserFromJWTCookie(r *http.Request) (string, error) {
	token, err := r.Cookie("id")
	if err != nil {
		return "", fmt.Errorf("couldn't find cookie: %v", err)
	}

	claims, err := jwt.HMACCheck([]byte(token.Value), []byte(app.JWT.Secret))
	if err != nil {
		return "", fmt.Errorf("detected invalid signature in jwtCookie: %v", err)
	}

	return claims.Subject, nil
}

func (app *application) validateJWTCookie(r *http.Request) error {
	token, err := r.Cookie("id")
	if err != nil {
		return fmt.Errorf("couldn't find cookie: %v", err)
	}

	claims, err := jwt.HMACCheck([]byte(token.Value), []byte(app.JWT.Secret))
	if err != nil {
		return fmt.Errorf("detected invalid signature in jwtCookie: %v", err)
	}

	if !claims.Valid(time.Now()) {
		return fmt.Errorf("token no longer valid")
	}

	if claims.Issuer != app.JWT.Issuer {
		return fmt.Errorf("token has invalid issuer: %v", err)
	}

	if !claims.AcceptAudience(app.JWT.Audience) {
		return fmt.Errorf("token is not in accepted audience: %v", err)
	}

	return nil
}
