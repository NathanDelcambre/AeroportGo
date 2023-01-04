package main

import (
	"fmt"
	// "time"
	"os"
	"strconv"
	"time"
	mosquitto "archi.org/aeroportGo/internal/connection/mosquitto"

)

// commande du publisher en cmd => .\publisher.exe [sensorId] [iata] [sensorType] 
func main() {
	args := os.Args

	ipAndPort := args[1]
	sensorId, err := strconv.Atoi(args[2])
	iata := args[3]
	sensorType := args[4]


	
	if err != nil {
		fmt.Println(err)
		fmt.Println(" => le type de l'ID n'est pas un int")
		os.Exit((1))
	}

	// fmt.Println("tout va bien\n")

	// fmt.Println("sensorid =" , sensorId )
	// fmt.Println("sensortype = " + sensorType)
	// fmt.Println("iata = " + iata)

	client := mosquitto.Connect("tcp://" + ipAndPort, "123")

	var start time.Time
	var duration float64
	var timeToWait time.Duration

	// topic aeroports/iata/sensorType/
	// données : sensorId + iata + sensorType + valeur  + YYY-MM-DD-mm-ss
	for {
		start = time.Now()
		token := client.Publish(
			// "aeropots/" + iata + "/" + sensorType,
			"a/b/c",
			0,
			false,
			strconv.Itoa(sensorId),
		)
		token.Wait()
		token.Error()
		fmt.Println("Message envoyé sur le topic : " + "aeropots/" + iata + "/" + sensorType)
		duration = time.Now().Sub(start).Seconds()
		timeToWait = time.Duration(10 - int(duration))
		time.Sleep(timeToWait * time.Second)
	}
}
