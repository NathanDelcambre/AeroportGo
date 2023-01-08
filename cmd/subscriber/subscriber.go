package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
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
func dataToLog(values []string) {
	iata := values[1]
	sensorType := values[2]
	value := values[3]
	timestampNotCut := values[4]
	timestamp := strings.Join(strings.Split(timestampNotCut, "-")[:3], "-")
	path := filepath.Join("..", "..","internal", "logs", iata + "-" + timestamp + "-" + sensorType + ".csv")

	// Créer un fichier ce le fichier de log du jour du capteur n'existe pas
	if _, err := os.Stat(path); os.IsNotExist(err) {
		csvFile, err := os.Create(path)
		if err != nil {
			fmt.Printf("Erreur lors de la création du fichier .csv : %s\n", err)
			return
		}
		defer csvFile.Close()
	}

	// Ouvre le fichier en mode écriture
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("Erreur lors de l'ouverture du fichier %s : %s\n", path, err)
		return
	}
	defer file.Close()

	// Écrit dans le fichier
	_, err = file.WriteString(iata + "," + sensorType + "," + value + "," + timestampNotCut + "," + "\n")
	if err != nil {
		fmt.Printf("Erreur lors de l'écriture dans le fichier %s : %s\n", path, err)
		return
	}
	file.Sync()
}

func main() {
	var theArray [5]string
	theArray[0] = ""  // Assign a value to the first element
	theArray[1] = "CDG" // Assign a value to the second element
	theArray[2] = "Heat"  // Assign a value to the third element
	theArray[3] = "bonjour c'est la valeur"  // Assign a value to the third element
	theArray[4] = "2023-01-08-17-52-51"  // Assign a value to the third element

	dataToLog(theArray[:])
	os.Exit(1)
	///////////////////////////////
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
