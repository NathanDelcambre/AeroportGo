package main

import (
	"fmt"
	// "token"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	//1 connection
	client := connect("tcp://localhost:1883", "toto")
	//2 subscription
	client.Subscribe("/topic1", 0, nil)
	//3 publication
	token := client.Publish("topic1", 0, false, "ceci est mon ")

	
	token.Wait()
}

func test() string {
	return "1"
}

func createClientOptions(brokerURI string, clientId string) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerURI)

	opts.SetClientID(clientId)
	return opts
}

func connect(brokerURI string, clientId string) mqtt.Client {
	fmt.Println("Tentative de connection ( " + brokerURI + ", " + clientId + ") ...")
	opts := createClientOptions(brokerURI, clientId)
	client := mqtt.NewClient(opts)
	token := client.Connect()
	for !token.WaitTimeout(3 * time.Second) {

	}
	if err := token.Error(); err != nil {
		log.Fatal(err)
	}
	return client
}
