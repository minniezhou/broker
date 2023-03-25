package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
)

type Config struct{}

func main() {
	c := Config{}
	h := c.Newhandler()
	log.Println("server started at port 8080...")
	err := http.ListenAndServe(":8080", h.router)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server %s \n", err)
		os.Exit(1)
	}
}
