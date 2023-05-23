package main

import (
	"broker/api/logging"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Handler struct {
	router *chi.Mux
}

const (
	Authorization string = "authentication"
	Logging       string = "logging"
	Send          string = "send"
)

const (
	AUTHENTICATION_SERVICE = "localhost"
	LOGGING_SERVICE        = "localhost"
	MAIL_SERVICE           = "localhost"
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
	r.Post("/grpclog", c.handleLoggingViaGRPC)
	r.Post("/handleviaqueue", c.handleEvent)
	return &Handler{
		router: r,
	}
}

func (c *Config) getHello(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	body := "Hello World!"
	err := c.ch.PublishWithContext(ctx,
		"",        // exchange
		queneName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	if err != nil {
		fmt.Println(err.Error())
		response := jsonResponse{
			Error:   true,
			Message: "Send to Queue Error!!",
		}
		c.writeJSON(w, http.StatusBadRequest, response)
		return
	}

	fmt.Printf(" [x] Sent %s\n", body)
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
	auth_url := getEnv("AUTHENTICATION_SERVICE", AUTHENTICATION_SERVICE)
	resp, err := http.Post("http://"+auth_url+":85/auth", "application/json", responseBody)
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
	log_url := getEnv("LOGGING_SERVICE", LOGGING_SERVICE)
	resp, err := http.Post("http://"+log_url+":4321/log", "application/json", responseBody)
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
	mail_url := getEnv("MAIL_SERVICE", MAIL_SERVICE)
	resp, err := http.Post("http://"+mail_url+":54321/send", "application/json", responseBody)
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

func (c *Config) handleEvent(w http.ResponseWriter, r *http.Request) {
	var request requestType
	c.readJSON(w, r, &request)
	postBody, _ := json.Marshal(request)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := c.ch.PublishWithContext(ctx,
		"",        // exchange
		queneName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(postBody),
		})
	if err != nil {
		fmt.Println(err.Error())
		response := jsonResponse{
			Error:   true,
			Message: "Send to Queue Error!!",
		}
		c.writeJSON(w, http.StatusBadRequest, response)
		return
	}
	fmt.Println("sent to queue")
	response := jsonResponse{
		Error:   false,
		Message: "Request Sent to Queue!!",
	}
	c.writeJSON(w, http.StatusAccepted, response)
}

func (c *Config) handleLoggingViaGRPC(w http.ResponseWriter, r *http.Request) {
	var request requestType
	c.readJSON(w, r, &request)
	if request.Action != Logging {
		c.ErrorJSON(w, errors.New("Error Action Type. Logging Action is needed"), http.StatusAccepted)
		return
	}
	log_url := getEnv("LOGGING_SERVICE", LOGGING_SERVICE)
	conn, err := grpc.Dial(log_url+":43210", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
		fmt.Println("connect to grcp log failed")
	}
	defer conn.Close()
	fmt.Println("New client")
	client := logging.NewLogClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := client.LogViaGRPC(ctx, &logging.LogRequest{Name: request.Log.Name, Data: request.Log.Message})

	if err != nil {
		c.ErrorJSON(w, errors.New("Log failed via GRPC"), http.StatusAccepted)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Logged via GRPC"
	payload.Data = resp.Message

	c.writeJSON(w, http.StatusAccepted, payload)

}
