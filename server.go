package main

import (
	"log"
	"net/http"
	"github.com/go-martini/martini"
)

func main(){
	m := martini.New()

	// Setup middleware
	m.Use(martini.Recovery())
	m.Use(martini.Logger())

	r := martini.NewRouter()
	r.Get("/", func() string {
		return "Hello world"
	})

	m.Action(r.Handle)

	if err := http.ListenAndServe(":8080", m); err != nil {
		log.Fatal(err)
	}
}
