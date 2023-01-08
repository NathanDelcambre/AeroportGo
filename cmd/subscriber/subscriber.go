package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	redisConn "archi.org/aeroportGo/internal/connection/redis"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gomodule/redigo/redis"
)

var wg sync.WaitGroup
var conn redis.Conn

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
	fmt.Println("Connexion réussi !")
	return client
}

func messageHandler(client mqtt.Client, message mqtt.Message) {
	if len(strings.Split(message.Topic(), "/")[1]) > 3 {
		fmt.Println("Code IATA invalide")
		return
	}
	values := strings.Split(string(message.Payload()), ":")
	if values[1] != strings.Split(message.Topic(), "/")[1] {
		fmt.Println("Erreur entre le IATA du topic et le IATA de la donnée")
		return
	}
	log.Println("Ecriture dans la bdd")
	conn.Do("SET", "airport:"+values[1]+":"+values[4]+":"+values[2]+":value", values[3]+":"+values[0])
	// id du capteur : iata : sensor type : value : timestamp


}

// To do
// Envoie les data dans un fichier de log csv
func dataToLog(values string) {
	
	// filePath := filepath.Join("..", "..","internal", "logs", values + ".csv")
	// "airport:"+values[1]+";"+values[4]+";"+values[2]+";value", values[3]+";"+values[0]
}

func main() {
	args := os.Args
	if len(args) != 3 {
		log.Fatalln("Veuillez saisir tous les arguments")
	}
	ipAndPortMosquitto := args[1]
	ipAndPortRedis := args[2]
	conn = redisConn.ConnectRedis(ipAndPortRedis)
	defer redisConn.DisconnetRedis(conn)
	client := connect("tcp://"+ipAndPortMosquitto, "1")
	client.Subscribe("airports/#", 2, messageHandler)
	wg.Add(1)
	wg.Wait()
}
