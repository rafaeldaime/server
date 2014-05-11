package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

var Env struct {
	Url        string
	Port       int
	Production bool
}

// The only one martini instance
var m *martini.Martini

func init() {
	if martini.Env == "production" {
		log.Println("Server in production")
		Env.Port = 8000
		Env.Url = "http://tur.ma/"
		Env.Production = true
	} else {
		log.Println("Server in development")
		Env.Port = 8000
		Env.Url = fmt.Sprintf("http://localhost:%d/", Env.Port)
		Env.Production = true
	}

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

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	// Just a ping route
	r.Get("/ping", func() string {
		return "pong!"
	})

	// Add the Auth Handlers
	r.Get("/login", LoginHandler)
	r.Get("/logincallback", LoginCallbackHandler)
	r.Get("/me", MeHandler)

	r.Get("/channel", getAllChannels)

	r.Post("/content", createNewContent)

	r.NotFound(func(r render.Render, req *http.Request) {

		r.JSON(http.StatusNotFound, NewError(ErrorCodeDefault, fmt.Sprintf(
			"Desculpe, mas nao ha nada no endereço requisitado. [%s] %s", req.Method, req.RequestURI)))
	})

	// Add the routers to my martini instances
	m.Action(r.Handle)
}

func main() {
	// Starting de HTTPS server in a new goroutine

	// log.Println("Starting HTTPs server in port 8001...")
	// go func() {
	// 	if err := http.ListenAndServeTLS(":8001", "cert.pem", "key.key", m); err != nil {
	// 		log.Fatal(err)
	// 	}
	// }()

	// Starting de HTTP server
	log.Println("Starting HTTP server in " + Env.Url + " ...")
	if err := http.ListenAndServe(fmt.Sprintf(":%d", Env.Port), m); err != nil {
		log.Fatal(err)
	}

}
