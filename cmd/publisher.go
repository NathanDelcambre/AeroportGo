package main

import (
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func createClientOptions(brokerURI string, clientId string) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerURI)

	opts.SetClientID(clientId)
	return opts
}

func connect(brokerURI string, clientId string) mqtt.Client {
	fmt.Println("Tentative de connection (" + brokerURI + ", " + clientId + ")...")
	opts := createClientOptions(brokerURI, clientId)
	client := mqtt.NewClient(opts)
	token := client.Connect()
	for !token.WaitTimeout(3 * time.Second) {
		if err := token.Error(); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println("test")
	return client
}

func test() bool {
	return false
}

func main() {
	client := connect("tcp://localhost:1883", "123")
	token := client.Publish("a/b/c", 0, false, "Test golang")
	token.Wait()
}
