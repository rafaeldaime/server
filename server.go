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

var renderOptions = render.Options{
	// Directory to load templates. Default is "templates"
	Directory: "templates",
	// Extensions to parse template files from. Defaults to [".tmpl"]
	Extensions: []string{".html"},
	// Delims sets the action delimiters to the specified strings in the Delims struct.
	Delims: render.Delims{Left: "%%", Right: "%%"},
	// Appends the given charset to the Content-Type header. Default is "UTF-8".
	Charset: "UTF-8",
	// Outputs human readable JSON
	IndentJSON: true,
}

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
	m.Use(render.Renderer(renderOptions))

	// Add the AuthMiddleware
	m.Use(OrmMiddleware)

	// Add the AuthMiddleware
	m.Use(AuthMiddleware)

	// Setup routes
	r := martini.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// Not rendering, couse of our angulars {{ }}
		http.ServeFile(w, r, "index.html")
	})

	// Just a ping route
	r.Get("/ping", func() string {
		return "pong!"
	})

	// Add the Auth Handlers
	r.Get("/login", LoginHandler)
	r.Get("/logincallback", LoginCallbackHandler)

	// Api Handlers
	r.Get("/api/me", meHandler)

	r.Get("/api/categories", getAllCategories)

	r.Post("/api/contents", addContent)
	r.Get("/api/contents", getContents)
	r.Put("/api/contents/:contentid", updateContent)
	r.Post("/api/contents/:contentid/image", changeContentImage)

	r.NotFound(func(r render.Render, req *http.Request) {
		// !!! Need to check if the call isn't to /api, so return default html not found
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
