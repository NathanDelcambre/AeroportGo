package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var wg sync.WaitGroup

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

func handler(client mqtt.Client, message mqtt.Message) {
	fmt.Println(string(message.Payload()))
}

func main() {
	client := connect("tcp://localhost:1883", "1234")
	client.Subscribe("a/b/c", 0, handler)
	wg.Add(1)
	wg.Wait()
}
