package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"code.google.com/p/goauth2/oauth"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/sessions"
)

// Here start our simplest OAuth2

const (
	codeRedirect = 302
	keyToken     = "oauth2_token"
	keyNextPage  = "next"
)

var config = &oauth.Config{
	ClientId:     "560046220732101",
	ClientSecret: "18d10b619523227e65ecf5b38fc18f90",
	RedirectURL:  "http://localhost:8000/authcallback",
	Scope:        "basic_info email",
	AuthURL:      "https://www.facebook.com/dialog/oauth",
	TokenURL:     "https://graph.facebook.com/oauth/access_token",
}

var transport = &oauth.Transport{
	Config:    config,
	Transport: http.DefaultTransport,
}

func AuthMiddleware(s sessions.Session, c martini.Context, w http.ResponseWriter, r *http.Request) {
	tk := unmarshallToken(s)
	if tk != nil {
		// check if the access token is expired
		if tk.IsExpired() && tk.Refresh() == "" {
			s.Delete(keyToken)
			tk = nil
		}
	}
	// Inject tokens.
	c.MapTo(tk, (*Tokens)(nil))
}

// Here is our BasicAuth
func BasicAuth(username string, password string) http.HandlerFunc {
	var siteAuth = base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	return func(res http.ResponseWriter, req *http.Request) {
		auth := req.Header.Get("Authorization")
		if !SecureCompare(auth, "Basic "+siteAuth) {
			res.Header().Set("WWW-Authenticate", "Basic realm=\"Autorizacao requereida\"")
			http.Error(res, "Not Authorized", http.StatusUnauthorized)
		}
	}
}

// SecureCompare performs a constant time compare of two strings to limit timing attacks.
func SecureCompare(given string, actual string) bool {
	givenSha := sha256.Sum256([]byte(given))
	actualSha := sha256.Sum256([]byte(actual))

	return subtle.ConstantTimeCompare(givenSha[:], actualSha[:]) == 1
}

// Represents a container that contains
// user's OAuth 2.0 access and refresh tokens.
type Tokens interface {
	Access() string
	Refresh() string
	IsExpired() bool
	ExpiryTime() time.Time
	ExtraData() map[string]string
}

type token struct {
	oauth.Token
}

func (t *token) ExtraData() map[string]string {
	return t.Extra
}

// Returns the access token.
func (t *token) Access() string {
	return t.AccessToken
}

// Returns the refresh token.
func (t *token) Refresh() string {
	return t.RefreshToken
}

// Returns whether the access token is
// expired or not.
func (t *token) IsExpired() bool {
	if t == nil {
		return true
	}
	return t.Expired()
}

// Returns the expiry time of the user's
// access token.
func (t *token) ExpiryTime() time.Time {
	return t.Expiry
}

// Formats tokens into string.
func (t *token) String() string {
	return fmt.Sprintf("tokens: %v", t)
}

// Handler that redirects user to the login page
// if user is not logged in.
func LoginRequiredHandler(s sessions.Session, w http.ResponseWriter, r *http.Request) {
	token := unmarshallToken(s)
	if token == nil || token.IsExpired() {
		next := url.QueryEscape(r.URL.RequestURI())
		http.Redirect(w, r, "/login?next="+next, codeRedirect)
	}
}

func LoginHandler(s sessions.Session, w http.ResponseWriter, r *http.Request) {
	next := extractPath(r.URL.Query().Get(keyNextPage))
	if s.Get(keyToken) == nil {
		// User is not logged in.
		http.Redirect(w, r, transport.Config.AuthCodeURL(next), codeRedirect)
		return
	}
	// No need to login, redirect to the next page.
	http.Redirect(w, r, next, codeRedirect)
}

func LogoutHandler(s sessions.Session, w http.ResponseWriter, r *http.Request) {
	next := extractPath(r.URL.Query().Get(keyNextPage))
	s.Delete(keyToken)
	http.Redirect(w, r, next, codeRedirect)
}

func AuthCallbackHandler(enc Encoder, s sessions.Session, w http.ResponseWriter, r *http.Request) (int, string) {
	//next := extractPath(r.URL.Query().Get("state"))
	code := r.URL.Query().Get("code")
	// token = oauth.Token, err = oauth.OAuthError
	oauthToken, err := transport.Exchange(code)
	if err != nil {
		return http.StatusUnauthorized, string(Must(enc.Encode(
			NewError(ErrorCodeDefault, fmt.Sprintf("Desculpe, mas nao conseguimos autentica-lo. %s", err)))))
	}

	// Here we have our http client configured to make calls
	client := transport.Client()
	res, err := client.Get("https://graph.facebook.com/me")
	if err != nil {
		return http.StatusBadRequest, string(Must(enc.Encode(
			NewError(ErrorCodeDefault, fmt.Sprintf("Desculpe, mas nao conseguimos acessar seus dados. %s", err)))))
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return http.StatusBadRequest, string(Must(enc.Encode(
			NewError(ErrorCodeDefault, fmt.Sprintf("Desculpe, ocorreu um erro ao ler a resposta do outro servidor. %s", err)))))
	}

	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return http.StatusBadRequest, string(Must(enc.Encode(
			NewError(ErrorCodeDefault, fmt.Sprintf("Desculpe, a mensagem do outro servidor nao esta no formato json. %s", err)))))
	}

	// Store the credentials in the session.
	value, _ := json.Marshal(oauthToken)
	s.Set(keyToken, value)
	//http.Redirect(w, r, next, codeRedirect)
	log.Println(fmt.Sprintf("transport.Token: %#v", transport.Token))
	return 200, Must(enc.Encode(data))
}

func unmarshallToken(s sessions.Session) (t *token) {
	if s.Get(keyToken) == nil {
		return
	}
	data := s.Get(keyToken).([]byte)
	var tk oauth.Token
	json.Unmarshal(data, &tk)
	return &token{tk}
}

func extractPath(next string) string {
	n, err := url.Parse(next)
	if err != nil {
		return "/"
	}
	return n.Path
}
