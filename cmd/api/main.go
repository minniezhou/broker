package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Config struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

const queneName = "broker"

func main() {
	// connect to rabbit mq
	conn, err := connectToRabbit()
	if err != nil {
		log.Panic("failed to connect to rabbit mq")
	}
	defer conn.Close()
	ch, err := declareChannel(conn)
	if err != nil {
		log.Panic("failed to declare channel")
	}
	c := Config{conn: conn, ch: ch}
	h := c.Newhandler()
	log.Println("server started at port 8080...")
	err = http.ListenAndServe(":8080", h.router)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server %s \n", err)
		os.Exit(1)
	}
}

func connectToRabbit() (*amqp.Connection, error) {
	count := 1
	backoff := time.Second
	log.Println("Connecting to Rabbit...")
	for {
		conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
		if err != nil {
			count++
			backoff = time.Duration(count*count) * time.Second
			log.Println("Rabit is not ready yet, backing off...")
			time.Sleep(backoff)
		} else {
			return conn, nil
		}

		if count > 10 {
			return nil, err
		}
	}
}

func declareChannel(conn *amqp.Connection) (*amqp.Channel, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	_, err = ch.QueueDeclare(
		queneName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return nil, err
	}
	return ch, nil
}
