package main

import (
	"strconv"
	"encoding/json"
	"io/ioutil"
	"log"
	"fmt"
	"net/http"
	"time"
	"github.com/gomodule/redigo/redis"
)

var connexionName redis.Conn
var averageMap map[string]string

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

func setRoutes(){
	http.HandleFunc("/getDatasBetweenTimeValues", getDatasBetweenTimeValues)
	http.HandleFunc("/getAverageBetweenTimeValues", getAverageBetweenTwoTimeValues)
}

func ping(c redis.Conn) error {
	s, err := redis.String(c.Do("PING"))
	if err != nil {
	return err
	}
	
	fmt.Printf("PING Response = %s\n", s)
	
	return nil
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
		idParam := body["idCapteur"].(string)
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

func getAverageBetweenTwoTimeValues(w http.ResponseWriter, r *http.Request){
	// Récupération des dates en paramètre du body
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	var body map[string]interface{}
	json.Unmarshal(b, &body)

	// Récuperation des paramètres de la requête
	airportParam := body["airport"].(string)
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
	data := getAverage_between_twoTimeValues(t1, t2, airportParam)
	// Envoi de la donnée au client
	w.Write([]byte(data))
}

func getData_between_twoTimeValues(t1 time.Time, t2 time.Time, airportParam string, typeCapteur string, idCapteur string) string {
	// Initialisation de la variable de résultat
	result := ""
	for (t1.Before(t2)) {
		// Affichage de la date en cours de traitement
		fmt.Println("Date en cours de traitement : "+t1.Format("2006-01-02:15-04-05"))
		// Creation de la clé de recherche
		key := "airport:"+airportParam + ":" + t1.Format("2006-01-02:15-04-00")+":"+typeCapteur+":"+idCapteur
		// Récupération des données
		data := get(connexionName,key)
		// Incrémentation de la date
		t1 = t1.Add(time.Second * 10)
		// Ajout de la donnée au résultat
		result += data
	}
	return result
}


func getAverage_between_twoTimeValues(t1 time.Time, t2 time.Time, airportParam string) string {

	
	// Creation et initilisation du tableau des valeurs vide
	values := make([]int, 0)
	// Initialisation de la map des moyennes par type de capteur
	averageMap := make(map[string]string)

	// Boucle sur les dates
	for (t1.Before(t2)) {

		// Creation du tableau des types de capteurs
		types := [3]string{"Heat", "Pressure", "Wind"}
		// Boucle sur les types de capteurs
		for _, typeCapteur := range types {

			// Creation du tableau des id des capteurs
			idCapteurs := [3]string{"1", "2", "3"}
			// Reinitialisation du tableau des valeurs
			values = values[:0]
			// Boucle sur les id des capteurs afin d'obtenir les mesures d'un capteur
			for _, idCapteur := range idCapteurs {

				// Affichage de la date en cours de traitement
				fmt.Println("Date en cours de traitement : "+t1.Format("2006-01-02:15-04-05"))
				// Creation de la clé de recherche
				key := "airport:"+airportParam + ":" + t1.Format("2006-01-02:15-04-00")+":"+typeCapteur+":"+idCapteur
				// Récupération des données
				data := get(connexionName,key)
				// Conversion de la donnée en int
				dataInt, err := strconv.Atoi(data)
				for  {	
					// Récupération des nouvelles données
					data = get(connexionName,key)
					// Conversion de la donnée en int
					dataInt, err = strconv.Atoi(data)
					if ((err != nil)) {
						// Gestion de l'erreur si la conversion en int échoue
						break
					}
					// ajout d'une seconde en cas de clé inexistante pour maintenir la cohérence des données
					t1 = t1.Add(time.Second)
				}
				// Ajout de la donnée au résultat
				values = append(values, dataInt)
				// Incrémentation de la date
				t1 = t1.Add(time.Second * 10)

			}

			// Calcul de la moyenne des valeurs
			average := getAverageAsString(values)
			// Ajout de la moyenne au résultat
			averageMap[typeCapteur] = average

		}
	}
	// Conversion de la map en json
	averageJson, errJson := json.Marshal(averageMap)
	// Gestion de l'erreur si la conversion en json échoue
	if errJson != nil {
		fmt.Println(errJson)
	}
	// Retour du résultat
	return string(averageJson)

}



func getAverageAsString(values []int) string {
	// Initialisation de la variable de résultat
	var result int
	// Initialisation de la variable de comptage
	var count int
	// Boucle sur les valeurs
	for _, value := range values {
		// Ajout de la valeur au résultat
		result += value
		// Incrémentation du compteur
		count++
	}
	// Calcul de la moyenne
	result = result / count
	// Conversion du résultat en string
	resultString := strconv.Itoa(result)
	// Retour du résultat
	return resultString
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


