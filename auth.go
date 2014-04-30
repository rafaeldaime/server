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
	"os"
	"time"

	"code.google.com/p/goauth2/oauth"
	"github.com/dchest/uniuri" // give us random URIs
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
	// tk := unmarshallToken(s)
	// if tk != nil {
	// 	// check if the access token is expired
	// 	if tk.IsExpired() && tk.Refresh() == "" {
	// 		s.Delete("token")
	// 		tk = nil
	// 	}
	// }
	// Inject tokens.
	// ... c.MapTo(tk, (*Tokens)(nil))
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

// Handler that redirects user to the login page
// if user is not logged in.
func LoginRequiredHandler(s sessions.Session, w http.ResponseWriter, r *http.Request) {
	// Token := nil //unmarshallToken(s)
	// if Token == nil || Token.IsExpired() {
	// 	next := url.QueryEscape(r.URL.RequestURI())
	// 	http.Redirect(w, r, "/login?next="+next, 302)
	// }
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

func AuthCallbackHandler(db DB, enc Encoder, s sessions.Session, w http.ResponseWriter, r *http.Request) (int, string) {
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

	profile, err := getProfile(client)
	if err != nil {
		return http.StatusUnauthorized, Must(enc.Encode(
			NewError(ErrorCodeDefault, fmt.Sprintf(
				"Desculpe, mas nao conseguimos acessar seu perfil. %s", err))))
	}

	user, err := getOrCreateUser(db, profile)
	if err != nil {
		return http.StatusUnauthorized, Must(enc.Encode(
			NewError(ErrorCodeDefault, fmt.Sprintf(
				"Desculpe, mas nao conseguimos processar seu perfil. %s", err))))
	}

	// After answare the request, check if this user has an pic
	defer checkOrGetPic(db, client, user)

	token, err := newToken(db, user)

	// THIS TOKEN SHOULD BE SET BY JAVASCRIPT!
	// USER IS AUTHENTICATED!

	return 200, Must(enc.Encode(token))
}

// func unmarshallToken(s sessions.Session) (t *Token) {
// 	if s.Get("Token") == nil {
// 		return
// 	}
// 	data := s.Get("Token").([]byte)
// 	var tk oauth.Token
// 	json.Unmarshal(data, &tk)
// 	return &Token{tk}
// }

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
	// I can know that is facebook by my token? maybe...
	profile.Source = "facebook"
	// UserId is not inserted know, but is required
	profile.UserName = p["username"].(string)
	profile.Email = p["email"].(string)
	profile.FullName = p["name"].(string)
	profile.Gender = p["gender"].(string)
	profile.ProfileUrl = p["link"].(string)
	profile.Language = p["locale"].(string)
	profile.Verified = p["verified"].(bool)
	profile.FirstName = p["first_name"].(string)
	profile.LastName = p["last_name"].(string)

	timeLayout := "2006-01-02T15:04:05+0000"
	updateTime, err := time.Parse(timeLayout, p["updated_time"].(string))
	if err != nil {
		return nil, err
	}
	profile.SourceUpdate = updateTime

	profile.AccessToken = token.AccessToken
	profile.RefreshToken = token.RefreshToken
	profile.Scope = config.Scope // A package config variable
	profile.TokenExpiry = token.Expiry
	// Creation and LastUpdate will be auto-added by database

	return profile, nil
}

// Take a http client with a transport pre-configured with the user credentials
// and return de name of the pic saved in public/pics/
func getPic(db DB, client *http.Client, user *User) (*Pic, error) {
	// OK WE HAVE THE PROFILE, BUT WE DON'T HAVE USER'S PICTURE
	// LETS TAKE IT
	res, err := client.Get("https://graph.facebook.com/me/picture")
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	img, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	picId := uniuri.NewLen(20)
	p, err := db.Get(Pic{}, picId)
	for err == nil && p != nil {
		// Shit, we generated an existing picId!
		// Lotto, where are you?
		picId := uniuri.NewLen(20)
		p, err = db.Get(Pic{}, picId)
	}

	if err != nil {
		return nil, err
	}

	fo, err := os.Create("public/pic/" + picId + ".png")
	if err != nil {
		return nil, err
	}

	// close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	fo.Write(img)

	pic := new(Pic)
	pic.PicId = picId
	pic.UserId = user.UserId
	pic.Creation = time.Now()
	pic.Deleted = false // Lol...

	err = db.Insert(pic)
	if err != nil {
		return nil, err
	}

	return pic, nil
}

func checkOrGetPic(db DB, client *http.Client, user *User) {
	count, err := db.SelectInt("select count(*) from pic where userid=?", user.UserId)
	if err != nil || count == 0 {
		pic, err := getPic(db, client, user)
		if err != nil {
			log.Printf("Error getting users pic: %v", err)
		}
		log.Printf("New pic %s saved from user %s", pic.PicId, user.UserId)
	}
}

func getProfile(client *http.Client) (*Profile, error) {
	res, err := client.Get("https://graph.facebook.com/me")
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	data := new(interface{}) // Its a pointer to an interface
	// The Decoder can handle a stream, unlike Unmarshal
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(data) // Passing a pointer
	if err != nil {
		return nil, err
	}

	return extractProfile(transport.Token, data)
}

func newUser(db DB, profile *Profile) (*User, error) {

	userId := uniuri.NewLen(20)
	userName := profile.UserName

	var users []User
	var increment = 0

	query := "select * from user where userid=? or username=?"
	_, err := db.Select(&users, query, userId, userName)

	for err == nil && len(users) != 0 {

		log.Printf("U[%v]: %#v", len(users), users)
		//return nil, nil

		u := users[0]
		users = users[:0] // Clear my slice

		// I will play in lottery if we see it happennig
		// just one time in 10^32
		if u.UserId == userId {
			userId = uniuri.NewLen(20)
		}

		// If already exists an user with this userName
		if u.UserName == userName {
			increment += 1
			userName = fmt.Sprintf("%s-%d", profile.UserName, increment)
		}

		// Check if still existing an user like this
		_, err = db.Select(&users, query, userId, userName)
	}

	// Checks if main loop didn't finish for an error
	if err != nil {
		return nil, err
	}

	user := new(User)
	user.UserId = userId
	user.UserName = userName
	user.Creation = time.Now()
	user.LastUpdate = time.Now()

	err = db.Insert(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func getOrCreateUser(db DB, profile *Profile) (*User, error) {

	p, err := db.Get(Profile{}, profile.ProfileId, profile.Source)
	if err != nil {
		return nil, err
	}

	if p != nil {
		profileSaved := p.(*Profile)
		profile.UserId = profileSaved.UserId
		profile.LastUpdate = time.Now()

		_, err := db.Update(profile)
		if err != nil {
			return nil, err
		}

		u, err := db.Get(User{}, profile.UserId)
		if err != nil {
			return nil, err
		}

		user := u.(*User)

		return user, nil
	}

	user, err := newUser(db, profile)
	if err != nil {
		return nil, err
	}

	// Now this new profile has an user account
	profile.UserId = user.UserId
	profile.Creation = time.Now()
	profile.LastUpdate = time.Now()

	err = db.Insert(profile)
	if err != nil {
		return nil, err
	}

	// !!! REMEMBER TO GET USERS PIC!

	return user, nil
}

func newToken(db DB, user *User) (*Token, error) {
	token := new(Token)

	tokenId := uniuri.NewLen(20)
	t, err := db.Get(Token{}, tokenId)
	for err == nil && t != nil {
		// Shit, we generated an existing token...
		// just one time in 10^32
		// Let's play the lottery!
		tokenId := uniuri.NewLen(20)
		t, err = db.Get(Token{}, tokenId)
	}

	if err != nil {
		return nil, err
	}

	token.TokenId = tokenId
	token.UserId = user.UserId
	token.Creation = time.Now()
	return token, nil
}
