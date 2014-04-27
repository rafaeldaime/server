package main

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/sessions"
	"log"
	"net/http"
)

// The only one acess token just for testing https
const AuthToken = "token2"
const AuthPass = ""

// The only one martini instance
var m *martini.Martini

func init() {
	m = martini.New()

	// Setup middleware
	m.Use(martini.Recovery())
	m.Use(martini.Logger())
	m.Use(sessions.Sessions("session", sessions.NewCookieStore([]byte("CookieIrado"))))

	// Add de EncoderMiddleware for Json encode
	m.Use(MapEncoder)

	// Add the AuthMiddleware
	m.Use(AuthMiddleware)

	// Setup routes
	r := martini.NewRouter()

	// Add the Auth Handlers
	r.Get("/login", LoginHandler)
	r.Get("/logout", LogoutHandler)
	r.Get("/authcallback", AuthCallbackHandler)

	r.Get("/", func() string {
		return ` 
		<p>Home</p>
		<ul>
		<li><a href="/admin.html">admin</a></li>
		<li><a href="/login">login</a></li>
		<li><a href="/token">token</a></li>
		<li><a href="/logout">logout</a></li>
		</ul>`
	})

	r.Get("/admin.html", LoginRequiredHandler, func() string {
		return `
		<p>Admin</p>
		<ul>
		<li><a href="/">Home</a></li>
		<li><a href="/logout">logout</a></li>
		</ul>
		`
	})

	// tokens are injected to the handlers
	r.Get("/token", func(enc Encoder, tokens Tokens) (int, string) {
		if tokens != nil {
			return http.StatusOK, Must(enc.Encode(tokens)) //.Access()
		}
		return 403, Must(enc.Encode("Nao autenticado"))
	})

	// testing https secure
	r.Get("/secure", BasicAuth(AuthToken, AuthPass), func() string {
		return "Voce foi autenticado pelo Authorization Header!"
	})

	// Just a ping route
	r.Get("/ping", func() string {
		return "pong!"
	})

	r.NotFound(func(r *http.Request) (int, string) {
		return http.StatusNotFound, "Pagina nao encontrada " + r.URL.Path
	})

	// Add the routesr action
	m.Action(r.Handle)
}

func main() {
	log.Println("Starting server...")
	// Starting de HTTPS server in a new goroutine
	// go func() {
	// 	if err := http.ListenAndServeTLS(":8001", "cert.pem", "key.pem", m); err != nil {
	// 		log.Fatal(err)
	// 	}
	// }()

	// Starting de HTTPS server
	if err := http.ListenAndServe(":8000", m); err != nil {
		log.Fatal(err)
	}
}
