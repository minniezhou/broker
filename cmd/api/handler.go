package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Handler struct {
	router *chi.Mux
}

func (c *Config) Newhandler() *Handler {
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http//*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Use(middleware.Heartbeat("/ping"))
	r.Use(middleware.Logger)
	r.Post("/", c.broker)
	r.Get("/hello", c.getHello)
	return &Handler{
		router: r,
	}
}

func (c *Config) getHello(w http.ResponseWriter, r *http.Request) {
	response := jsonResponse{
		Error:   false,
		Message: "Hello Http!!",
	}
	c.writeJSON(w, http.StatusAccepted, response)
}

func (c *Config) broker(w http.ResponseWriter, r *http.Request) {
	response := jsonResponse{
		Error:   false,
		Message: "Hit the broken service",
	}
	c.writeJSON(w, http.StatusAccepted, response)
}
