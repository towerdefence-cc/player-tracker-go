package servers

import (
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"net/url"
	"os"
	"player-tracker-go/rabbitmq/model"
	"player-tracker-go/service"
	"strings"
)

const (
	rabbitMqUriFormat = "amqp://%v:%v@%s:5672//"
)

func InitRabbitMq() {
	uri := createUri()
	conn, err := amqp.Dial(uri)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	log.Print("Connected to RabbitMQ. Starting consumer...")

	msgs, err := ch.Consume(
		"player-tracker-connections", // queue
		"",                           // consumer
		true,                         // auto-ack
		false,                        // exclusive
		false,                        // no-local
		false,                        // no-wait
		nil,                          // args
	)
	if err != nil {
		panic(err)
	}

	var forever chan struct{}

	go func() {
		for d := range msgs {
			processMessage(d)
		}
	}()

	<-forever
}

func processMessage(d amqp.Delivery) {
	if d.Type == "connect" {
		serverId := d.AppId

		var event model.ConnectEventDataPackage
		err := json.Unmarshal(d.Body, &event)
		if err != nil {
			log.Printf("Failed to unmarshal connect event (%s): %s", d.Body, err)
			return
		}

		if strings.HasPrefix(serverId, "velocity-") {
			err = service.UpdatePlayerProxy(context.Background(), event.PlayerId, event.Username, serverId)
		} else {
			err = service.UpdatePlayerServer(context.Background(), event.PlayerId, event.Username, serverId)
		}
		if err != nil {
			log.Printf("Failed to update player server (pid: %s, sid: %s): %s", event.PlayerId, serverId, err)
		}
	} else if d.Type == "disconnect" {
		serverId := d.AppId

		var event model.DisconnectEventDataPackage
		err := json.Unmarshal(d.Body, &event)
		if err != nil {
			log.Printf("Failed to unmarshal disconnect event (%s): %s", d.Body, err)
			return
		}

		if strings.HasPrefix(serverId, "velocity-") {
			err = service.ProxyPlayerDisconnect(context.Background(), event.PlayerId)
			if err != nil {
				log.Printf("Failed to update player proxy (pid: %s, sid: %s): %s", event.PlayerId, serverId, err)
			}
		} else {
			log.Printf("Disconnect event received for non-velocity server (%s)", serverId)
		}
	} else {
		log.Printf("Unknown message type: %s", d.Type)
		log.Printf("Message: %s", d.Body)
	}
}

func createUri() string {
	username, ok := os.LookupEnv("RABBITMQ_USERNAME")
	if ok {
		username = url.QueryEscape(username)
	} else {
		username = "guest"
	}

	password, ok := os.LookupEnv("RABBITMQ_PASSWORD")
	if ok {
		password = url.QueryEscape(password)
	} else {
		password = "password"
	}

	address, ok := os.LookupEnv("RABBITMQ_ADDRESS") // In prod rabbitmq-default.towerdefence.svc
	if !ok {
		address = "localhost"
	}
	return fmt.Sprintf(rabbitMqUriFormat, username, password, address)
}
