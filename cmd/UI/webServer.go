package main

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/gomodule/redigo/redis"
)

func main() {
	http.HandleFunc("/airports/", func(w http.ResponseWriter, r *http.Request) {
		iata := strings.TrimPrefix(r.URL.Path, "/airports/")
		conn, err := redis.Dial("tcp", "localhost:6379")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		values, err := redis.Values(conn.Do("HGETALL", "airport:"+iata))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var data []Sensor
		for i := 0; i < len(values); i += 2 {
			key := values[i].(string)
			val := values[i+1].(string)
			parts := strings.Split(key, ":")
			value, err := strconv.ParseFloat(parts[3], 64)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			id, err := strconv.Atoi(parts[0])
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			data = append(data, Sensor{ID: id, AirportID: parts[1], SensorType: parts[2], LastGeneratedValue: value})
		}

		tmpl, err := template.ParseFiles("template.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	http.ListenAndServe(":8080", nil)
}
