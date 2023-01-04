package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"fmt"
	"net/http"
	"time"
	"github.com/gomodule/redigo/redis"
)

var connexionName redis.Conn

func ConnectRedis(host string) redis.Conn {
    fmt.Println("Tentative de connexion à REDIS: " + host)
    conn, err := redis.Dial("tcp", host)
    if err != nil {
        log.Fatal(err)
    }
    return conn
}

func DisconnetRedis(conn redis.Conn) {
    conn.Close()
}

func main() {
	// Connexion à la base de données REDIS
	connexionName = ConnectRedis("localhost:6379")
	// Vérification de la connexion
	ping(connexionName)
	// Définition des routes
	setRoutes()
	// Lancement du serveur
	log.Fatal(http.ListenAndServe(":8080", nil))
	// Déconnexion de la base de données REDIS
	DisconnetRedis(connexionName)
}


func setRoutes(){
	http.HandleFunc("/getDatasBetweenTimeValues", getDatasBetweenTimeValues)
}

type body_DataBetweenTwoTimeValues_Response struct {
	airport string;
	typeCapteur string;
	idCapteur string;
	date1 string;
	date2 string;
}

func getDatasBetweenTimeValues(w http.ResponseWriter, r *http.Request){

		
		// Récupération des dates en paramètre du body
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		var body map[string]interface{} 
		json.Unmarshal(b, &body)

		// Récuperation des paramètres de la requête
		airportParam := body["airport"].(string) 
		typeParam := body["typeCapteur"].(string) 
		dateParam1 := body["date1"].(string)
		dateParam2 := body["date2"].(string)

		// Conversion des dates en objet time
		layout := "02-01-2006:15-04-05"
		t1, err := time.Parse(layout, dateParam1)
		if err != nil {
			// Gestion de l'erreur si la conversion en objet time échoue
			http.Error(w, "Invalid date or time, please provide valid parameter [dateParam1 - INVALID ]", http.StatusBadRequest)
			return
		}
		t2, err:= time.Parse(layout, dateParam2)
		if err != nil {
			// Gestion de l'erreur si la conversion en objet time échoue
			http.Error(w, "Invalid date or time, please provide valid parameter [dateParam2 - INVALID ]", http.StatusBadRequest)
			return
		}

		// Récupération des données en fonction des deux objets time et de l'aeroport
		data := getData_between_twoTimeValues(t1, t2, airportParam, typeParam, idParam)

		// Envoi de la donnée au client
		w.Write([]byte(data))
}

func getData_between_twoTimeValues(t1 time.Time, t2 time.Time, airportParam string, typeCapteur string, idCapteur string) string {
	// Initialisation de la variable de résultat
	result := ""
	for (t1.Before(t2)) {
		// Affichage de la date en cours de traitement
		fmt.Println(t1.Format("2006-01-02:15-04-00"))
		// Creation de la clé de recherche
		key := "airport:"+airportParam + ":" + t1.Format("2006-01-02:15-04-00")+":"+typeCapteur+":"+idCapteur
		// Récupération des données
		data := get(connexionName,key)
		// Incrémentation de la date
		t1 = t1.Add(time.Minute)
		// Ajout de la donnée au résultat
		result += data
	}
	return result
}

func get(connexion redis.Conn, _key string) string {
	key := _key
	s, err := redis.String(connexion.Do("GET", key))
	if err == redis.ErrNil {
		fmt.Printf("%s : Alert! this Key does not exist\n", key)
	} else if err != nil {
		fmt.Printf("%s : Error! %s\n", key, err)
	} else {
		fmt.Printf("%s : %s\n", key, s)
	}
	return s
}

func ping(c redis.Conn) error {
	s, err := redis.String(c.Do("PING"))
	if err != nil {
	return err
	}
	
	fmt.Printf("PING Response = %s\n", s)
	
	return nil
}
