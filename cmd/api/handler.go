package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Handler struct {
	router *chi.Mux
}

const (
	Authorization string = "authentication"
	Logging       string = "logging"
	Send          string = "send"
)

type requestType struct {
	Action string   `json:"action"`
	Auth   authType `json:"auth,omitempty"`
	Log    logType  `json:"log,omitempty"`
	Send   sendType `json:"send,omitempty"`
}

type authType struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type logType struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

type sendType struct {
	From       string   `josn:"from,omitempty"`
	FromName   string   `josn:"from_name,omitempty"`
	To         string   `josn:"to"`
	Subject    string   `josn:"subject"`
	Body       string   `json:"body"`
	Attachment []string `josn:"attachments,omitempty"`
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
	r.Post("/handle", c.handle)
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

func (c *Config) handle(w http.ResponseWriter, r *http.Request) {
	var request requestType
	c.readJSON(w, r, &request)
	switch request.Action {
	case Authorization:
		c.handleAuthorization(request.Auth, w)
	case Logging:
		c.handleLogging(request.Log, w)
	case Send:
		c.handleSendEmail(request.Send, w)
	default:
		c.ErrorJSON(w, errors.New("Unknown action type"))
	}
}

func (c *Config) handleAuthorization(request authType, w http.ResponseWriter) {
	postBody, _ := json.Marshal(request)
	responseBody := bytes.NewBuffer(postBody)
	resp, err := http.Post("http://localhost:80/auth", "application/json", responseBody)
	if err != nil {
		c.ErrorJSON(w, errors.New("Authrization error, request failed"), http.StatusAccepted)
		return
	}
	defer resp.Body.Close()

	var payloadfromService jsonResponse
	err = json.NewDecoder(resp.Body).Decode(&payloadfromService)
	if err != nil {
		c.ErrorJSON(w, errors.New("Authentication failed"), http.StatusAccepted)
		return
	}

	if payloadfromService.Error {
		c.ErrorJSON(w, errors.New("Authentication failed"), http.StatusAccepted)
		return
	}
	var payload jsonResponse
	payload.Error = false
	payload.Message = "Authenticated"
	payload.Data = payloadfromService.Data

	c.writeJSON(w, http.StatusAccepted, payload)

}

func (c *Config) handleLogging(request logType, w http.ResponseWriter) {
	postBody, _ := json.Marshal(request)
	responseBody := bytes.NewBuffer(postBody)
	resp, err := http.Post("http://localhost:4321/log", "application/json", responseBody)
	if err != nil {
		c.ErrorJSON(w, errors.New("Logging error"), http.StatusAccepted)
		return
	}
	defer resp.Body.Close()

	var payloadfromService jsonResponse
	err = json.NewDecoder(resp.Body).Decode(&payloadfromService)
	if err != nil {
		c.ErrorJSON(w, errors.New("Log failed"), http.StatusAccepted)
		return
	}

	if payloadfromService.Error {
		c.ErrorJSON(w, errors.New("Log failed"), http.StatusAccepted)
		return
	}
	var payload jsonResponse
	payload.Error = false
	payload.Message = "Logged"
	payload.Data = payloadfromService.Data

	c.writeJSON(w, http.StatusAccepted, payload)

}

func (c *Config) handleSendEmail(request sendType, w http.ResponseWriter) {
	postBody, _ := json.Marshal(request)
	responseBody := bytes.NewBuffer(postBody)
	resp, err := http.Post("http://localhost:54321/send", "application/json", responseBody)
	if err != nil {
		c.ErrorJSON(w, errors.New("Sending Email error"), http.StatusAccepted)
		return
	}
	defer resp.Body.Close()

	var payloadfromService jsonResponse
	err = json.NewDecoder(resp.Body).Decode(&payloadfromService)
	if err != nil {
		c.ErrorJSON(w, errors.New("Sending Email failed"), http.StatusAccepted)
		return
	}

	if payloadfromService.Error {
		c.ErrorJSON(w, errors.New("Sending Email failed"), http.StatusAccepted)
		return
	}
	var payload jsonResponse
	payload.Error = false
	payload.Message = "Email Sent"
	payload.Data = payloadfromService.Data

	c.writeJSON(w, http.StatusAccepted, payload)

}
