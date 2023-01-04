package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"time"
	"archi.org/aeroportGo/internal/sensors"
	"archi.org/aeroportGo/internal/connection/mosquitto"

)

// TO DO rajouter QoS dans args

/**
	.\publisher.exe [ipAndPort] [sensorId] [iata] [sensorType]

	@ipAndPort: localhost:1883
	@sensorId: 1, 2 or 3
	@iata: cf wikipedia
	@sensorType: Wind, Heat, Pressure

	topic => airports/iata/sensorType/
	données => sensorId:iata:sensorType:valeur:YYYY-MM-DD-mm-ss
**/

func main() {
	var start time.Time
	var duration float64
	var timeToWait time.Duration
	var value float64
	var date string
	var sensor *sensors.Sensor

	args := os.Args
	ipAndPort := args[1]
	sensorId, err := strconv.Atoi(args[2])
	iata := args[3]
	sensorType := args[4]

	client := mosquitto.Connect("tcp://" + ipAndPort, "123")

	if err != nil {
		fmt.Println(err)
		fmt.Println(" => le type de l'ID n'est pas un int")
		os.Exit((1))
	}

	switch sensorType {
	case "Wind":
		sensor = sensors.NewSensor(sensorId, iata, &sensors.WindSensor)
	case "Heat":
		sensor = sensors.NewSensor(sensorId, iata, &sensors.HeatSensor)
		value = math.Round(sensor.GenerateNextData()*100)/100
	case "Pressure":
		sensor = sensors.NewSensor(sensorId, iata, &sensors.PressureSensor)
		value = math.Round(sensor.GenerateNextData()*100)/100
	}
	date = time.Now().Format("2006-01-02-15-04-05")

	for {
		value = math.Round(sensor.GenerateNextData()*100)/100

		start = time.Now()
		token := client.Publish(
			"airports/" + iata + "/" + sensorType,
			2,
			false,
			strconv.Itoa(sensorId) + ":" + iata + ":"  + sensorType + ":"  + strconv.FormatFloat(value, 'f', -1, 64) + ":"  + date,
		)

		token.Wait()
		token.Error()

		fmt.Println("Message envoyé : " + strconv.Itoa(sensorId) + ":" + iata + ":"  + sensorType + ":"  + strconv.FormatFloat(value, 'f', -1, 64) + ":"  + date + "\n" + 
					"Sur le topic : " + "airports/" + iata + "/" + sensorType)
		
		duration = time.Now().Sub(start).Seconds()
		timeToWait = time.Duration(10 - int(duration))
		time.Sleep(timeToWait * time.Second)
	}
}
