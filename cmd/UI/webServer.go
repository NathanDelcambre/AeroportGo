package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
)

type Sensor struct {
	ID                 int
	AirportID          string
	SensorType         string
	LastGeneratedValue float64
	LastGeneratedTime  string
}

func main() {
	http.HandleFunc("/airports/", func(w http.ResponseWriter, r *http.Request) {

		// Récupérez l'identifiant de l'aéroport
		iata := strings.TrimPrefix(r.URL.Path, "/airports/")
		// Ouvrez une connexion à Redis
		conn, err := redis.Dial("tcp", "localhost:6379")
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		// Initilisez la date de maintenant
		now := time.Now()
		// Définissez la date de début à 10 minutes avant maintenant
		start := now.Add(-10 * time.Minute)

		// Créer une map pour stocker les données
		var data []Sensor
		// Créez une boucle qui itère sur les 10 dernières minutes
		for t := start; t.Before(now); t = t.Add(time.Second * 10) {

			types := [3]string{"Heat", "Pressure", "Wind"}

			// Récupérez les données pour chque type de capteur
			for _, typeCapteur := range types {
				sensor, err := getSensorValue(conn, iata, t, typeCapteur, 1)
				if err != nil {
					// Affichage dans la console en erreur de l'erreur
					fmt.Println(err)
				}
				// Ajoutez les données à la map
				data = append(data, sensor)
			}

		}

		tmpl, err := template.ParseFiles("template.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := tmpl.Execute(w, reverse(data)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	http.ListenAndServe(":8080", nil)
}

func getSensorValue(conn redis.Conn, iata string, t time.Time, sensorType string, id int) (Sensor, error) {
	layout := "2006-01-02:15-04-05"
	key := fmt.Sprintf("airport:%s:%s:%s:%d", iata, t.Format(layout), sensorType, id)
	value, err := redis.Float64(conn.Do("GET", key))
	layout2 := "2006-01-02 à 15:04:05"
	return Sensor{
		ID:                 id,
		AirportID:          iata,
		SensorType:         sensorType,
		LastGeneratedValue: value,
		LastGeneratedTime:  t.Format(layout2),
	}, err
}

func reverse(numbers []Sensor) []Sensor {
	newNumbers := make([]Sensor, 0, len(numbers))
	for i := len(numbers) - 1; i >= 0; i-- {
		newNumbers = append(newNumbers, numbers[i])
	}
	return newNumbers
}
