// inspired by: https://github.com/coreos/go-oidc/blob/v3/example/userinfo/app.go
package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/davidkuda/bellevue/internal/models"
	"golang.org/x/oauth2"
)

type openIDConnect struct {
	provider    *oidc.Provider
	verifier    *oidc.IDTokenVerifier
	config      oauth2.Config
	accessToken string
}

type claims struct {
	Email      string `json:"email"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	SUB        string `json:"sub"`
}

func (app *application) oidcLogin(w http.ResponseWriter, r *http.Request) {
	state, err := randString(16)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	nonce, err := randString(16)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	setCallbackCookie(w, r, "state", state)
	setCallbackCookie(w, r, "nonce", nonce)

	http.Redirect(w, r, app.OIDC.config.AuthCodeURL(state, oidc.Nonce(nonce)), http.StatusFound)
}

func (app *application) oidcCallbackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()
	state, err := r.Cookie("state")
	if err != nil {
		http.Error(w, "state not found", http.StatusBadRequest)
		return
	}

	if r.URL.Query().Get("state") != state.Value {
		http.Error(w, "state did not match", http.StatusBadRequest)
		return
	}

	oauth2Token, err := app.OIDC.config.Exchange(ctx, r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	userInfo, err := app.OIDC.provider.UserInfo(ctx, oauth2.StaticTokenSource(oauth2Token))
	if err != nil {
		http.Error(w, "Failed to get userinfo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	c := claims{}
	err = userInfo.Claims(&c)
	if err != nil {
		app.serverError(w, r, fmt.Errorf("could not unmarshal claims: %s", err))
		return
	}

	userID, err := app.models.Users.GetUserIDBySUB(c.SUB)
	if err != nil {
		if err == sql.ErrNoRows {
			u := models.User{
				Email:     c.Email,
				FirstName: c.GivenName,
				LastName:  c.FamilyName,
				SUB:       c.SUB,
			}
			err = app.models.Users.InsertOIDC(u)
			if err != nil {
				app.serverError(w, r, fmt.Errorf("failed inserting new user: %s", err))
				return
			}
		} else {
			app.serverError(w, r, fmt.Errorf("failed getting id with sub=%s:", c.SUB))
			return
		}
	}

	app.sessionManager.Put(r.Context(), "UserID", userID)

	http.Redirect(w, r, "/activities", http.StatusSeeOther)
}

func randString(nByte int) (string, error) {
	b := make([]byte, nByte)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func setCallbackCookie(w http.ResponseWriter, r *http.Request, name, value string) {
	c := &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   int(time.Hour.Seconds()),
		Secure:   r.TLS != nil,
		HttpOnly: true,
	}
	http.SetCookie(w, c)
}
