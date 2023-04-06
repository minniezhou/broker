package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type jsonResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (c *Config) readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1024 * 1024
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(data)
	if err != nil {
		return err
	}

	err = decoder.Decode(new(struct{}))
	if err != io.EOF {
		return errors.New("The input should just only have one json block")
	}
	return nil
}

func (c *Config) writeJSON(w http.ResponseWriter, statusCode int, data any, headers ...http.Header) error {
	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(statusCode)
	_, err = w.Write(output)
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) ErrorJSON(w http.ResponseWriter, err error, statusCode ...int) {
	payLoad := jsonResponse{
		Error:   true,
		Message: err.Error(),
	}
	status := http.StatusBadRequest
	if len(statusCode) > 0 {
		status = statusCode[0]
	}
	c.writeJSON(w, status, payLoad)
}
