package main

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
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
	// Serving public directory
	m.Use(martini.Static("public"))

	// Render html templates from /templates directory
	m.Use(render.Renderer(render.Options{
		Extensions: []string{".html"},
	}))

	// Add the AuthMiddleware
	m.Use(OrmMiddleware)

	// Add the AuthMiddleware
	m.Use(AuthMiddleware)

	// Setup routes
	r := martini.NewRouter()

	// Add the Auth Handlers
	r.Get("/login", LoginHandler)
	r.Get("/facecallback", FaceCallbackHandler)
	r.Get("/me", MeHandler)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	// Just a ping route
	r.Get("/ping", func() string {
		return "pong!"
	})

	// Just a ping route
	r.Get("/api", func() string {
		return "its my api!"
	})

	// r.NotFound(func(r *http.Request) (int, string) {
	// 	return http.StatusNotFound, "Pagina nao encontrada " + r.URL.Path
	// })

	// Add the routesr action
	m.Action(r.Handle)
}

func main() {
	if martini.Env == "production" {
		log.Println("Server in production")
	} else {
		log.Println("Server in development")
	}
	// Starting de HTTPS server in a new goroutine

	// log.Println("Starting HTTPs server in port 8001...")
	// go func() {
	// 	if err := http.ListenAndServeTLS(":8001", "cert.pem", "key.key", m); err != nil {
	// 		log.Fatal(err)
	// 	}
	// }()

	// Starting de HTTP server
	log.Println("Starting HTTP server in port 8000...")
	if err := http.ListenAndServe(":8000", m); err != nil {
		log.Fatal(err)
	}

}
