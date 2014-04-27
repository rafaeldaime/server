package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"code.google.com/p/goauth2/oauth"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/sessions"
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
			s.Delete("token")
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

type Token struct {
	oauth.Token
}

func (t *Token) ExtraData() map[string]string {
	return t.Extra
}

// Returns the access Token.
func (t *Token) Access() string {
	return t.AccessToken
}

// Returns the refresh Token.
func (t *Token) Refresh() string {
	return t.RefreshToken
}

// Returns whether the access Token is
// expired or not.
func (t *Token) IsExpired() bool {
	if t == nil {
		return true
	}
	return t.Expired()
}

// Returns the expiry time of the user's
// access Token.
func (t *Token) ExpiryTime() time.Time {
	return t.Expiry
}

// Formats Tokens into string.
func (t *Token) String() string {
	return fmt.Sprintf("Tokens: %v", t)
}

// Handler that redirects user to the login page
// if user is not logged in.
func LoginRequiredHandler(s sessions.Session, w http.ResponseWriter, r *http.Request) {
	Token := unmarshallToken(s)
	if Token == nil || Token.IsExpired() {
		next := url.QueryEscape(r.URL.RequestURI())
		http.Redirect(w, r, "/login?next="+next, 302)
	}
}

func LoginHandler(s sessions.Session, w http.ResponseWriter, r *http.Request) {
	next := extractPath(r.URL.Query().Get("next"))
	if s.Get("Token") == nil {
		// User is not logged in.
		http.Redirect(w, r, transport.Config.AuthCodeURL(next), 302)
		return
	}
	// No need to login, redirect to the next page.
	http.Redirect(w, r, next, 302)
}

func LogoutHandler(s sessions.Session, w http.ResponseWriter, r *http.Request) {
	next := extractPath(r.URL.Query().Get("next"))
	s.Delete("Token")
	http.Redirect(w, r, next, 302)
}

func AuthCallbackHandler(enc Encoder, s sessions.Session, w http.ResponseWriter, r *http.Request) (int, string) {
	//next := extractPath(r.URL.Query().Get("state"))
	code := r.URL.Query().Get("code")
	// Token = oauth.Token, err = oauth.OAuthError
	// oauthToken, err := transport.Exchange(code)
	_, err := transport.Exchange(code) //
	if err != nil {
		return http.StatusUnauthorized, Must(enc.Encode(
			NewError(ErrorCodeDefault, fmt.Sprintf(
				"Desculpe, mas nao conseguimos autentica-lo. %s", err))))
	}

	// Here we have our http client configured to make calls
	client := transport.Client()
	res, err := client.Get("https://graph.facebook.com/me")
	if err != nil {
		return http.StatusBadRequest, Must(enc.Encode(
			NewError(ErrorCodeDefault, fmt.Sprintf(
				"Desculpe, mas nao conseguimos acessar seus dados. %s", err))))
	}

	defer res.Body.Close()

	data := new(interface{}) // Its a pointer to an interface
	// The Decoder can handle a stream, unlike Unmarshal
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(data) // Passing a pointer
	if err != nil {
		return http.StatusBadRequest, Must(enc.Encode(
			NewError(ErrorCodeDefault, fmt.Sprintf(
				"Desculpe, a mensagem do outro servidor nao esta no formato json. %s", err))))
	}

	profile, err := extractProfile(transport.Token, data)
	if err != nil {
		return http.StatusBadRequest, Must(enc.Encode(
			NewError(ErrorCodeDefault, fmt.Sprintf(
				"Desculpe, mas nao conseguimos extrair o seu perfil pelos dados fornecidos. %s", err))))
	}

	// OK WE HAVE THE PROFILE, BUT WE DON'T HAVE USER'S PICTURE
	// LETS TAKE IT

	//http.Redirect(w, r, next, 302)
	return 200, Must(enc.Encode(profile))
}

func unmarshallToken(s sessions.Session) (t *Token) {
	if s.Get("Token") == nil {
		return
	}
	data := s.Get("Token").([]byte)
	var tk oauth.Token
	json.Unmarshal(data, &tk)
	return &Token{tk}
}

func extractPath(next string) string {
	n, err := url.Parse(next)
	if err != nil {
		return "/"
	}
	return n.Path
}

func extractProfile(token *oauth.Token, data *interface{}) (*Profile, error) {
	// Here we are taking the value in where this variable is pointing to,
	// and we cast it to an map[string] interface
	p := (*data).(map[string]interface{})

	profile := new(Profile)

	// Just Facebook for now
	profile.ProfileId = p["id"].(string)
	profile.Source = "facebook"
	profile.UserId = "" // We don't know yet
	profile.UserName = p["username"].(string)
	profile.Email = p["email"].(string)
	profile.FullName = p["name"].(string)
	profile.Gender = p["gender"].(string)
	profile.ProfileUrl = p["link"].(string)
	profile.ImageUrl = "" // I don't know how to take it yet
	profile.Language = p["locale"].(string)
	profile.Verified = p["verified"].(bool)
	profile.FirstName = p["first_name"].(string)
	profile.LastName = p["last_name"].(string)

	timeLayout := "2006-01-02T15:04:05+0000"
	updateTime, err := time.Parse(timeLayout, p["updated_time"].(string))
	if err != nil {
		return profile, err
	}
	profile.LastUpdateTime = updateTime

	profile.AccessToken = token.AccessToken
	profile.RefreshToken = token.RefreshToken
	profile.Expiry = token.Expiry
	profile.Scope = config.Scope // A package config variable

	return profile, nil
}
