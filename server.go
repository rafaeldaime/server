package main

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/sessions"
	"log"
	"net/http"
)

// The only one acess token just for testing https
const AuthToken = "token"
const AuthPass = ""

var FacebookOAuth = Facebook(&Options{
	ClientId:     "560046220732101",
	ClientSecret: "18d10b619523227e65ecf5b38fc18f90",
	RedirectURL:  "http://localhost:8000/oauth2callback",
	Scopes:       []string{"basic_info", "email"},
})

// The only one martini instance
var m *martini.Martini

func init() {
	m = martini.New()

	// Setup middleware
	m.Use(martini.Recovery())
	m.Use(martini.Logger())
	m.Use(sessions.Sessions("session", sessions.NewCookieStore([]byte("CookieIrado"))))
	m.Use(FacebookOAuth)

	// Setup routes
	r := martini.NewRouter()
	r.Get("/", func() string {
		return "Hello world"
	})

	r.Get("/test", func() string {
		return ` 
		<p>Home</p>
		 
		<ul>
		<li><a href="/login?next=/admin">admin</a></li>
		<li><a href="/logout">logout</a></li>
		</ul>
		`
	})

	r.Get("/admin", LoginRequired, func() string {
		return `
		<p>Admin</p>
		 
		<ul>
		<li><a href="/logout">logout</a></li>
		</ul>
		`
	})

	// tokens are injected to the handlers
	r.Get("/token", func(tokens Tokens) (int, string) {
		if tokens != nil {
			return 200, tokens.Access()
		}
		return 403, "not authenticated"
	})

	// testing https secure
	r.Get("/secure", BasicAuth(AuthToken, AuthPass), func() string {
		return "You are authenticated by Authorization Header!"
	})

	// Just a ping route
	r.Get("/ping", func() string {
		return "pong!"
	})

	// Add the routesr action
	m.Action(r.Handle)
}

func main() {
	log.Println("Starting server...")
	// Starting de HTTPS server in a new goroutine
	go func() {
		if err := http.ListenAndServeTLS(":8001", "cert.pem", "key.pem", m); err != nil {
			log.Fatal(err)
		}
	}()

	// Starting de HTTPS server
	if err := http.ListenAndServe(":8000", m); err != nil {
		log.Fatal(err)
	}
}
