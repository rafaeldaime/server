package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"code.google.com/p/goauth2/oauth"
	"github.com/dchest/uniuri" // give us random URIs
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

var config = &oauth.Config{
	ClientId:     "560046220732101",
	ClientSecret: "18d10b619523227e65ecf5b38fc18f90",
	//RedirectURL:   // It is setted dynamicaly on LoginHandler based on server current url
	Scope:    "basic_info email",
	AuthURL:  "https://www.facebook.com/dialog/oauth",
	TokenURL: "https://graph.facebook.com/oauth/access_token",
}

var transport = &oauth.Transport{
	Config:    config,
	Transport: http.DefaultTransport,
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
	if len(header) < 56 {
		err := errors.New("Credenciais de authorizacao apresentadas sao invalidas.")
		userAuth := &UserAuth{nil, err}
		c.MapTo(userAuth, (*Auth)(nil))
		return
	}

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

	if obj == nil {
		// This tokenid was not found
		err = errors.New("Credenciais fornecidas nao foram encontradas")
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
	return auth
}
func decodeAuth(auth string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(auth)
	return string(decoded), err
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// The state will inform if the response will be json or html
	// Html is the default request and will do 2 things:
	// Set the cookie "credentials" and return the visitor to the main page
	state := r.URL.Query().Get("state")
	if state == "" { // The default is a direct html request
		state = "html"
	}

	// Set based on the server current url
	transport.Config.RedirectURL = Env.Url + "logincallback/"

	http.Redirect(w, r, transport.Config.AuthCodeURL(state), 302)
	return
}

func renderHtmlOrJson(r render.Render, w http.ResponseWriter, req *http.Request, state string, err error, message string) {
	if state == "json" {
		r.JSON(http.StatusUnauthorized, NewError(ErrorCodeDefault, fmt.Sprintf(message+" %s.", err)))
	} else {
		cookieError := &http.Cookie{Name: "error", Value: fmt.Sprintf(message+" %s.", err), MaxAge: 60 * 30, Path: "/"}
		http.SetCookie(w, cookieError)
		http.Redirect(w, req, "/", 302)
	}
}

func LoginCallbackHandler(db DB, r render.Render, w http.ResponseWriter, req *http.Request) {
	// The state indicates if it is an browser request, or an default json request
	state := req.URL.Query().Get("state")

	code := req.URL.Query().Get("code")

	// Token = oauth.Token, err = oauth.OAuthError
	_, err := transport.Exchange(code)
	if err != nil {
		renderHtmlOrJson(r, w, req, state, err, "Desculpe, mas nao conseguimos autentica-lo.")
		return
	}

	// Here we have our http client configured to make calls
	client := transport.Client()

	res, err := client.Get("https://graph.facebook.com/me")
	if err != nil {
		renderHtmlOrJson(r, w, req, state, err, "Desculpe, mas nao foi possivel pegar perfil do facebook.")
		return
	}

	defer res.Body.Close()

	data := new(interface{}) // Its a pointer to an interface
	// The Decoder can handle a stream, unlike Unmarshal
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(data) // Passing a pointer
	if err != nil {
		renderHtmlOrJson(r, w, req, state, err, "Desculpe, mas a resposta do facebook nao foi um json valido.")
		return
	}

	profile, err := extractProfile(transport.Token, data)
	if err != nil {
		renderHtmlOrJson(r, w, req, state, err, "Desculpe, mas nao conseguimos acessar seu perfil.")
		return
	}

	user, err := getOrCreateUser(db, profile)
	if err != nil {
		renderHtmlOrJson(r, w, req, state, err, "Desculpe, mas nao conseguimos processar seu perfil")
		return
	}

	// After answare the request, check if this user has an pic
	err = checkOrGetPic(db, client, user)
	if err != nil {
		renderHtmlOrJson(r, w, req, state, err, "Desculpe, mas ocorreu um erro ao processar sua foto do perfil.")
		return
	}

	token, err := newToken(db, user)
	credentials := encodeAuth(token)

	if state == "json" {
		r.JSON(http.StatusOK, map[string]interface{}{"credentials": credentials})
		return
	}
	// COOKIES CAN'T HAVE VALUES WITH COMMA OR SPACE !!!
	cookieCredentials := &http.Cookie{Name: "credentials", Value: credentials, MaxAge: 60 * 60 * 24 * 30 * 12, Path: "/"}
	cookieMessage := &http.Cookie{Name: "message", Value: "Logou-se!", MaxAge: 60 * 30, Path: "/"}
	http.SetCookie(w, cookieCredentials)
	http.SetCookie(w, cookieMessage)
	http.Redirect(w, req, "/", 302)
}

func extractProfile(token *oauth.Token, data *interface{}) (*Profile, error) {
	// Here we are taking the value in where this variable is pointing to,
	// and we cast it to an map[string] interface
	p := (*data).(map[string]interface{})

	profile := new(Profile)

	// My Facebook profile
	profile.ProfileId = p["id"].(string)
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
func getPic(db DB, client *http.Client) (*Pic, error) {
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

	fo, err := os.Create("public/img/pic/" + picId + ".png")
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
	pic.Creation = time.Now()
	pic.Deleted = false // Lol...

	err = db.Insert(pic)
	if err != nil {
		return nil, err
	}

	return pic, nil
}

func checkOrGetPic(db DB, client *http.Client, user *User) error {
	// Check if this user has a default pic yet
	if user.PicId == "default" {
		pic, err := getPic(db, client)
		if err != nil {
			return err
		}

		log.Printf("New pic %s saved from user %s", pic.PicId, user.UserId)

		user.PicId = pic.PicId
		count, err := db.Update(user)
		if err != nil {
			return err
		}
		if count == 0 {
			user.PicId = "default"
			return errors.New("Nenhum usuario foi atualizado com a foto salva")
		}
	}
	return nil
}

func newUser(db DB, profile *Profile) (*User, error) {

	userId := uniuri.NewLen(20)
	userName := profile.UserName

	var users []User
	var increment = 0

	query := "select * from user where userid=? or username=?"
	_, err := db.Select(&users, query, userId, userName)

	for err == nil && len(users) != 0 {

		log.Printf("Trying to create another userid U[%v]: %#v \n", len(users), users)

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
			userName = fmt.Sprintf("%s%d", profile.UserName, increment)
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
	user.PicId = "default" // Reference to default pic
	user.FullName = profile.FullName
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

	p, err := db.Get(Profile{}, profile.ProfileId)
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
