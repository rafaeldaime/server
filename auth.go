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

//
// Return true if user is logged, or false
//
func (u *UserAuth) Logged() bool {
	if u.err != nil {
		return false
	}
	return true
}

//
// Return the user in this context or an anonymous user and an error
//
func (u *UserAuth) GetUser() (*User, error) {
	return u.user, u.err
}

//
// Response returned to AuthMiddleware if the handler aborted, cause it needs an user authenticated
// It returns 419 if users is sending an invalid Authorization header
// Or returns 403 saying that this handler requires an Authorization
//
func AccessDeniedHandler(r render.Render, req *http.Request, err error) {
	if req.Header.Get("Authorization") != "" {
		// Client thiks he is logged, force him to logout
		r.JSON(419, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Suas credenciais de autenticacao nao sao validas! %s.", err)))
		return
	}
	// Client didn't send credentials, say him it's required
	r.JSON(403, NewError(ErrorCodeDefault, fmt.Sprintf(
		"Voce precisa estar autenticado pelo sistema! %s.", err)))
}

//
// Middleware called every call that requires user Authentication
// It injects the Auth object in handlers that require it to work
// Return the user in this context or an anonymous user and an error
// Should be used .Logged to identify if it's an anonymous user
// and use .GetUser to get this context user or the anonymous user the error object
//
func AuthMiddleware(c martini.Context, db DB, r render.Render, req *http.Request) {
	anonymous := &User{
		UserId:     "anonymous",
		UserName:   "Anonimo",
		PicId:      "default",
		FullName:   "Usuario Anonimo",
		LikeCount:  0,
		Creation:   time.Now(),
		LastUpdate: time.Now(),
		Deleted:    false,
		Admin:      false,
	}

	// Try to get the credentials from Authorization header or credentials cookie
	credentials := req.Header.Get("Authorization")
	if credentials == "" {
		cookie, err := req.Cookie("credentials")
		if err == nil {
			credentials = cookie.Value
		}
	}

	if credentials == "" {
		err := errors.New("Voce nao apresentou credenciais de autorizacao")
		userAuth := &UserAuth{anonymous, err}
		c.MapTo(userAuth, (*Auth)(nil))
		c.Next()          // Call the next handler
		if !c.Written() { // Handler aborted
			// If the nex handler aborted, by an empty return
			// So he can't go on without user, let's alert user
			AccessDeniedHandler(r, req, err)
		}
		return
	}

	// Len of our 41 (20:20) char encoded in base64 is 56
	// (C++) long base64EncodedSize = 4 * (int)Math.Ceiling(originalSizeInBytes / 3.0);
	if len(credentials) != 56 {
		err := errors.New("Credenciais de authorizacao apresentadas sao invalidas.")
		userAuth := &UserAuth{anonymous, err}
		c.MapTo(userAuth, (*Auth)(nil))
		c.Next()          // Call the next handler
		if !c.Written() { // Handler aborted
			AccessDeniedHandler(r, req, err)
		}
		return
	}

	token, err := decodeAuth(credentials)
	if err != nil {
		userAuth := &UserAuth{anonymous, err}
		c.MapTo(userAuth, (*Auth)(nil))
		c.Next()          // Call the next handler
		if !c.Written() { // Handler aborted
			AccessDeniedHandler(r, req, err)
		}
		return
	}

	obj, err := db.Get(Token{}, token.TokenId, token.UserId)
	if err != nil {
		userAuth := &UserAuth{anonymous, err}
		c.MapTo(userAuth, (*Auth)(nil))
		c.Next()          // Call the next handler
		if !c.Written() { // Handler aborted
			AccessDeniedHandler(r, req, err)
		}
		return
	}

	if obj == nil {
		// This tokenid was not found
		err = errors.New("Credenciais fornecidas nao foram encontradas")
		userAuth := &UserAuth{anonymous, err}
		c.MapTo(userAuth, (*Auth)(nil))
		c.Next()          // Call the next handler
		if !c.Written() { // Handler aborted
			AccessDeniedHandler(r, req, err)
		}
		return
	}

	token = obj.(*Token)

	obj, err = db.Get(User{}, token.UserId)
	if err != nil {
		userAuth := &UserAuth{anonymous, err}
		c.MapTo(userAuth, (*Auth)(nil))
		c.Next()          // Call the next handler
		if !c.Written() { // Handler aborted
			AccessDeniedHandler(r, req, err)
		}
		return
	}

	user := obj.(*User)
	userAuth := &UserAuth{user: user, err: nil}
	// Here the next handler has got the user
	c.MapTo(userAuth, (*Auth)(nil))
}

//
// Receives a Token object and returns an encoded Authorization header
//
func encodeAuth(token *Token) string {
	authorization := base64.StdEncoding.EncodeToString([]byte(token.TokenId + ":" + token.UserId))
	return authorization
}

//
// Receives an encothed Authorization header and returns a Token object
//
func decodeAuth(authorization string) (*Token, error) {
	bytesDecoded, err := base64.StdEncoding.DecodeString(authorization)
	if err != nil {
		err = errors.New("Credenciais de authorizacao estao corrompidas")
		return nil, err
	}

	authorization = string(bytesDecoded)

	i := strings.Index(authorization, ":")
	if i == -1 {
		err = errors.New("Credenciais de authorizacao foram modificadas")
		return nil, err
	}

	token := &Token{
		TokenId: authorization[:i],
		UserId:  authorization[i+1:],
	}

	return token, err
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// The state will inform if the response will be json or html
	// Html is the default request and will do 2 things:
	// Set the cookie "credentials" and return the visitor to the main page
	state := r.URL.Query().Get("method")
	if state == "" { // The default is Json method, but it could be html too
		state = "json"
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
	// log.Printf("user: %#v \n", user)
	// log.Printf("Token: %#v \n", token)
	if err != nil {
		renderHtmlOrJson(r, w, req, state, err, "Desculpe, mas ocorreu um erro ao criar seu token de acesso.")
		return
	}
	credentials := encodeAuth(token)

	if state == "json" {
		r.JSON(http.StatusOK, map[string]interface{}{"credentials": credentials})
		return
	}
	// COOKIES CAN'T HAVE VALUES WITH COMMA OR SPACE !!!
	cookieCredentials := &http.Cookie{Name: "credentials", Value: credentials, MaxAge: 60 * 60 * 24 * 30 * 12, Path: "/"}
	cookieMessage := &http.Cookie{Name: "message", Value: "Bem-Vindo!", MaxAge: 60 * 30, Path: "/"}
	http.SetCookie(w, cookieCredentials)
	http.SetCookie(w, cookieMessage)
	http.Redirect(w, req, "/", 302)
}

func extractProfile(token *oauth.Token, data *interface{}) (*Profile, error) {
	// Here we are taking the value in where this variable is pointing to,
	// and we cast it to an map[string] interface
	p := (*data).(map[string]interface{})

	// Extracting updated_time from Facebook return
	timeLayout := "2006-01-02T15:04:05+0000"
	updateTime, err := time.Parse(timeLayout, p["updated_time"].(string))
	if err != nil {
		return nil, err
	}
	// The Facebook profile
	profile := &Profile{
		ProfileId:    p["id"].(string),
		UserId:       "", // UserId is not inserted know, but is required
		UserName:     p["username"].(string),
		Email:        p["email"].(string),
		FullName:     p["name"].(string),
		Gender:       p["gender"].(string),
		ProfileUrl:   p["link"].(string),
		Language:     p["locale"].(string),
		Verified:     p["verified"].(bool),
		FirstName:    p["first_name"].(string),
		LastName:     p["last_name"].(string),
		SourceUpdate: updateTime,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Scope:        config.Scope, // A package config variable
		TokenExpiry:  token.Expiry,
		Creation:     time.Now(),
		LastUpdate:   time.Now(),
	}

	return profile, nil
}

// Take a http client with a transport pre-configured with the user credentials
// and return the Pic saved in public/img/pic/ and stored in DB
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

	pic := &Pic{
		PicId:    picId,
		Creation: time.Now(),
		Deleted:  false,
	}

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

	user := &User{
		UserId:     userId,
		UserName:   userName,
		PicId:      "default", // Reference to default pic
		FullName:   profile.FullName,
		LikeCount:  0,
		Creation:   time.Now(),
		LastUpdate: time.Now(),
		Deleted:    false,
		Admin:      false,
	}

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

	return user, nil
}

func newToken(db DB, user *User) (*Token, error) {

	tokenId := uniuri.NewLen(20)
	t, err := db.Get(Token{}, tokenId, user.UserId)
	for err == nil && t != nil {
		// Shit, we generated an existing token...
		// just one time in 10^32
		// Let's play the lottery!
		tokenId := uniuri.NewLen(20)
		t, err = db.Get(Token{}, tokenId, user.UserId)
	}

	if err != nil {
		return nil, err
	}

	token := &Token{
		TokenId:  tokenId,
		UserId:   user.UserId,
		Creation: time.Now(),
	}

	err = db.Insert(token)
	if err != nil {
		return nil, err
	}

	return token, nil
}
