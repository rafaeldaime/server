package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"code.google.com/p/goauth2/oauth"
	"github.com/dchest/uniuri" // give us random URIs
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

const authPrefix = "Basic "

var config = &oauth.Config{
	ClientId:     "560046220732101",
	ClientSecret: "18d10b619523227e65ecf5b38fc18f90",
	RedirectURL:  "http://localhost:8000/facecallback",
	Scope:        "basic_info email",
	AuthURL:      "https://www.facebook.com/dialog/oauth",
	TokenURL:     "https://graph.facebook.com/oauth/access_token",
}

var transport = &oauth.Transport{
	Config:    config,
	Transport: http.DefaultTransport,
}

func init() {
	if martini.Env == "production" {
		config.RedirectURL = "http://ira.do/facecallback"
	}
}

// It's  returns an User instance when required
type Auth interface {
	Logged() bool
	GetUser() (*User, error)
}

type UserAuth struct {
	user *User
	err  error
}

func (u *UserAuth) Logged() bool {
	if u.err != nil || u.user == nil {
		return false
	} else {
		return true
	}
}

func (u *UserAuth) GetUser() (*User, error) {
	if u.err != nil || u.user == nil {
		return nil, u.err
	} else {
		return u.user, nil
	}
}

func MeHandler(auth Auth, r render.Render) {
	user, err := auth.GetUser()
	if err != nil {
		r.JSON(http.StatusUnauthorized, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Voce nao esta logado! %s.", err)))
	} else {
		r.JSON(http.StatusOK, user)
	}
}

func AuthMiddleware(c martini.Context, db DB, r *http.Request) {
	header := r.Header.Get("Authorization")

	if header == "" {
		err := errors.New("Voce nao apresentou credenciais de autorizacao")
		userAuth := &UserAuth{nil, err}
		c.MapTo(userAuth, (*Auth)(nil))
		return
	}

	// Len of our 41 (20:20) char encoded in base64 is 56
	// (C++) long base64EncodedSize = 4 * (int)Math.Ceiling(originalSizeInBytes / 3.0);
	if (len(header) < len(authPrefix)+56) || !strings.HasPrefix(header, authPrefix) {
		err := errors.New("Credenciais de authorizacao apresentadas sao invalidas")
		userAuth := &UserAuth{nil, err}
		c.MapTo(userAuth, (*Auth)(nil))
		return
	}

	header = header[len(authPrefix):]
	auth, err := decodeAuth(header)
	if err != nil {
		userAuth := &UserAuth{nil, err}
		c.MapTo(userAuth, (*Auth)(nil))
		return
	}

	i := strings.Index(auth, ":")
	if i == -1 {
		err = errors.New("Credenciais de authorizacao estao corrompidas")
		userAuth := &UserAuth{nil, err}
		c.MapTo(userAuth, (*Auth)(nil))
		return
	}

	tokenid := auth[:i]
	userid := auth[i+1:]

	obj, err := db.Get(Token{}, tokenid)
	if err != nil {
		userAuth := &UserAuth{nil, err}
		c.MapTo(userAuth, (*Auth)(nil))
		return
	}

	token := obj.(*Token)
	if token.UserId != userid {
		err = errors.New("Usuario nao identificado pelas credenciais fornecidas")
		userAuth := &UserAuth{nil, err}
		c.MapTo(userAuth, (*Auth)(nil))
		return
	}

	obj, err = db.Get(User{}, userid)
	if err != nil {
		userAuth := &UserAuth{nil, err}
		c.MapTo(userAuth, (*Auth)(nil))
		return
	}

	user := obj.(*User)
	userAuth := &UserAuth{user: user, err: nil}
	c.MapTo(userAuth, (*Auth)(nil))
}

// Here is our BasicAuth
func encodeAuth(token *Token) string {
	auth := base64.StdEncoding.EncodeToString([]byte(token.TokenId + ":" + token.UserId))
	return authPrefix + auth
}
func decodeAuth(auth string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(auth)
	return string(decoded), err
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	next := extractPath(r.URL.Query().Get("next"))
	http.Redirect(w, r, transport.Config.AuthCodeURL(next), 302)
	return
}

func FaceCallbackHandler(db DB, r render.Render, req *http.Request) {
	//next := extractPath(r.URL.Query().Get("state"))
	code := req.URL.Query().Get("code")
	// Token = oauth.Token, err = oauth.OAuthError
	// oauthToken, err := transport.Exchange(code)
	_, err := transport.Exchange(code) //
	if err != nil {
		r.HTML(http.StatusUnauthorized, "error", fmt.Sprintf(
			"Desculpe, mas nao conseguimos autentica-lo. %s", err))
		return
	}

	// Here we have our http client configured to make calls
	client := transport.Client()

	res, err := client.Get("https://graph.facebook.com/me")
	if err != nil {
		r.HTML(http.StatusUnauthorized, "error", fmt.Sprintf(
			"Desculpe, mas nao foi possivel pegar perfil do facebook. %s", err))
		return
	}

	defer res.Body.Close()

	data := new(interface{}) // Its a pointer to an interface
	// The Decoder can handle a stream, unlike Unmarshal
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(data) // Passing a pointer
	if err != nil {
		r.HTML(http.StatusUnauthorized, "error", fmt.Sprintf(
			"Desculpe, mas a resposta do facebook nao foi um json valido. %s", err))
		return
	}

	profile, err := extractProfile(transport.Token, data)
	if err != nil {
		r.HTML(http.StatusUnauthorized, "error", fmt.Sprintf(
			"Desculpe, mas nao conseguimos acessar seu perfil. %s", err))
		return
	}

	user, err := getOrCreateUser(db, profile)
	if err != nil {
		r.HTML(http.StatusUnauthorized, "error", fmt.Sprintf(
			"Desculpe, mas nao conseguimos processar seu perfil. %s", err))
		return
	}

	// After answare the request, check if this user has an pic
	defer checkOrGetPic(db, client, user)

	token, err := newToken(db, user)

	// !!! Here we have to return JSON if not called from our iframe
	r.HTML(200, "facebookauthcallback", encodeAuth(token))
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

	err = db.Insert(token)
	if err != nil {
		return nil, err
	}

	return token, nil
}
