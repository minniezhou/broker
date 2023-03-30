package main

import (
	"encoding/json"
	"fmt"
	"io"
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
	fmt.Printf("got /hello request\n")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	io.WriteString(w, "Hello, HTTP!\n")
}

type jsonResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (c *Config) broker(w http.ResponseWriter, r *http.Request) {
	response := jsonResponse{
		Error:   false,
		Message: "Hit the broken service",
	}
	output, _ := json.MarshalIndent(response, "", "  ")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write(output)
}
