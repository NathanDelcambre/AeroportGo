package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"strconv"
	"time"

	"archi.org/aeroportGo/internal/connection/mosquitto"
	"archi.org/aeroportGo/internal/sensors"
)

type Configuration struct {
	// Words []string `json:"words"`
	IATA       []string `json:"iata"`
	SensorID   []int    `json:"sensorId"`
	SensorType []string `json:"sensorType"`
}

func contains(val interface{}, arr interface{}) bool {
	// Vérifier si la variable est présente dans le tableau
	arrValue := reflect.ValueOf(arr)
	for i := 0; i < arrValue.Len(); i++ {
		if arrValue.Index(i).Interface() == val {
			return true
		}
	}
	return false
}

/**
	.\publisher.exe [ipAndPort] [sensorId] [iata] [sensorType] [QoS]

	@ipAndPort: localhost:1883
	@sensorId: 1, 2 or 3
	@iata: cf wikipedia
	@sensorType: Wind, Heat, Pressure
	@QoS: 2

	topic => airports/iata/sensorType/
	données => sensorId:iata:sensorType:valeur:YYYY-MM-DD-mm-ss
**/

func main() {
	// test read config files
	// data, err := ioutil.ReadFile("../../internal/conf/conf.json")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// var config Configuration
	// err = json.Unmarshal(data, &config)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// for _, word := range config.IATA {
	// 	fmt.Println(word)
	// }

	file, err := os.Open("../../internal/conf/conf.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	data := Configuration{}

	err = json.NewDecoder(file).Decode(&data)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(contains("CDG", data.IATA))
	// for i := 0; i < len(data.IATA); i++ {
		
	// 	fmt.Println(i, data.IATA[i])
	// }

	var start time.Time
	var duration float64
	var timeToWait time.Duration
	var value float64
	var date string
	var sensor *sensors.Sensor

	args := os.Args
	ipAndPort := args[1]
	sensorId, errSensor := strconv.Atoi(args[2])
	iata := args[3]
	sensorType := args[4]
	qualityOfService, errQoS := strconv.Atoi(args[5])

	fmt.Println(qualityOfService)

	client := mosquitto.Connect("tcp://" + ipAndPort, "123")

	if errSensor != nil {
		fmt.Println(errSensor)
		fmt.Println(" => le type de l'ID n'est pas un int")
		os.Exit((1))
	}

	if errQoS != nil {
		fmt.Println(errQoS)
		fmt.Println("le type du QoS n'est pas bon, utilisez un entier")
		os.Exit(1)
	}

	if qualityOfService != 2 {
		fmt.Println("Le QoS n'est pas à 2, nous suggérons de relancer l'application en utilisant 2 pour un meilleur QoS")
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
			byte(qualityOfService),
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
