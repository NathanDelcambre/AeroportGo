package main

import (
	"strconv"
	"encoding/json"
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

func getDatasBetweenTimeValues(w http.ResponseWriter, r *http.Request){

		query := r.URL.Query()

		// Récupération des paramètres de la requête
		airport, present := query["airport"] 
		if !present || len(airport) == 0 {
			http.Error(w, "Invalid airport IATA code, please provide valid parameter [airport - INVALID ]", http.StatusBadRequest)
			return
		}
		idQueryParam, present := query["id"]
		if !present || len(idQueryParam) == 0 {
			http.Error(w, "Invalid idCapteur, please provide valid parameter [idCapteur - INVALID ]", http.StatusBadRequest)
			return
		}
		typeQueryParam, present := query["type"]
		if !present || len(typeQueryParam) == 0 {
			http.Error(w, "Invalid typeCapteur, please provide valid parameter [typeCapteur - INVALID ]", http.StatusBadRequest)
			return
		}
		dateQueryParam1, present := query["date1"]
		if !present || len(dateQueryParam1) == 0 {
			http.Error(w, "Invalid date or time, please provide valid parameter [date1 - INVALID ]", http.StatusBadRequest)
			return
		}
		dateQueryParam2, present := query["date2"]
		if !present || len(dateQueryParam2) == 0 {
			http.Error(w, "Invalid date or time, please provide valid parameter [date2 - INVALID ]", http.StatusBadRequest)
			return
		}

		// Récuperation des paramètres de la requête
		airportParam := airport[0]
		idParam := idQueryParam[0]
		typeParam := typeQueryParam[0]
		dateParam1 := dateQueryParam1[0]
		dateParam2 := dateQueryParam2[0]


		// Conversion des dates en objet time
		layout := "02-01-2006:15-04-05"
		t1, err := time.Parse(layout, dateParam1)
		if err != nil {
			// Gestion de l'erreur si la conversion en objet time échoue
			http.Error(w, "Invalid date or time format, please provide valid parameter [dateParam1 - INVALID ]", http.StatusBadRequest)
			return
		}
		t2, err:= time.Parse(layout, dateParam2)
		if err != nil {
			// Gestion de l'erreur si la conversion en objet time échoue
			http.Error(w, "Invalid date or time format, please provide valid parameter [dateParam2 - INVALID ]", http.StatusBadRequest)
			return
		}

		// Récupération des données en fonction des deux objets time et de l'aeroport
		data := getData_between_twoTimeValues(t1, t2, airportParam, typeParam, idParam, w)
		// Envoi de la donnée au client
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(data))
}

func getAverageBetweenTwoTimeValues(w http.ResponseWriter, r *http.Request){

	query := r.URL.Query()

	// Récupération des paramètres de la requête
	airport, present := query["airport"] 
	if !present || len(airport) == 0 {
		http.Error(w, "Invalid airport IATA code, please provide a code [airport - INVALID ]", http.StatusBadRequest)
		return
	}
	dateQueryParam1, present := query["date1"]
	if !present || len(dateQueryParam1) == 0 {
		http.Error(w, "Invalid date or time, please provide a date [date1 - INVALID ]", http.StatusBadRequest)
		return
	}

	// Récuperation des paramètres de la requête
	airportParam := airport[0]
	dateParam1 := dateQueryParam1[0]

	// Conversion des dates en objet time
	layout := "02-01-2006:15-04-05"
	t1, err := time.Parse(layout, dateParam1)
	if err != nil {
		// Gestion de l'erreur si la conversion en objet time échoue
		http.Error(w, "Invalid date or time, please provide valid parameter [dateParam1 - INVALID ]", http.StatusBadRequest)
		return
	}
	
	// Récupération des données en fonction des deux objets time et de l'aeroport
	data := getAverage_between_twoTimeValues(t1, airportParam, w)
	// Envoi de la donnée au client en tant que json
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(data))
}

func getData_between_twoTimeValues(t1 time.Time, t2 time.Time, airportParam string, typeCapteur string, idCapteur string, w http.ResponseWriter) string {
	// Initialisation d'un objet json
	result := "{["
	for (t1.Before(t2)) {
		// Affichage de la date en cours de traitement
		fmt.Println("Date en cours de traitement : "+t1.Format("2006-01-02:15-04-05"))
		// Creation de la clé de recherche
		key := "airport:"+airportParam + ":" + t1.Format("2006-01-02:15-04-05")+":"+typeCapteur+":"+idCapteur
		// Récupération des données
		data := get(connexionName,key)
		// si la donnée n'existe pas on vérifie toutes les secondes un maximum de 3600 fois (1h) si la donnée existe, sinon on retourne une erreur
		if data == "" {
			for i := 0; i < 3600; i++ {
				t1 = t1.Add(time.Second * 1)
				// Creation de la clé de recherche
				key := "airport:"+airportParam + ":" + t1.Format("2006-01-02:15-04-05")+":"+typeCapteur+":"+idCapteur
				// Récupération des données
				data = get(connexionName,key)
				if data != "" {
					break
				}
			}
		}
		// Si la donnée n'existe toujours pas on retourne une erreur
		if data == "" {
			w.WriteHeader(http.StatusNotFound)
			return "404 - NOT FOUND"
		}

		// Ajout de la donnée au json en les regroupant par type de capteur avec leur clé étant la date
		result += "\""+t1.Format("2006-01-02:15-04-05")+"\" : "+data+", \n"
		// Incrémentation de la date
		t1 = t1.Add(time.Second * 10)
		
	}
	w.WriteHeader(http.StatusOK)
	return result+"]}"
}


func getAverage_between_twoTimeValues(t1 time.Time, airportParam string, w http.ResponseWriter) string {

	
	// Creation et initilisation du tableau des valeurs vide
	values := make([]int, 0)
	// Initialisation de la map des moyennes par type de capteur
	averageMap := make(map[string]string)
	// Initilisation d'une deuxieme date 24h apres t1
	t2 := t1.Add(time.Hour * 24)

	// Boucle sur les dates
	for (t1.Before(t2)) {

		// Creation du tableau des types de capteurs
		types := [3]string{"Heat", "Pressure", "Wind"}
		// Boucle sur les types de capteurs
		for _, typeCapteur := range types {

			// Creation du tableau des id des capteurs (pour l'instant 1 capteur mais on pourrait avoir une fonction qui renvoie une liste d'id en fonction du type et de l'aeroport)
			idCapteurs := [1]string{"1"}
			// Reinitialisation du tableau des valeurs
			values = values[:0]
			// Boucle sur les id des capteurs afin d'obtenir les mesures d'un capteur
			for _, idCapteur := range idCapteurs {

				// Affichage de la date en cours de traitement
				fmt.Println("Date en cours de traitement : "+t1.Format("2006-01-02:15-04-05"))
				// Creation de la clé de recherche
				key := "airport:"+airportParam + ":" + t1.Format("2006-01-02:15-04-05")+":"+typeCapteur+":"+idCapteur
				// Récupération des données
				data := get(connexionName,key)
				// Conversion de la donnée en int
				dataInt, err := strconv.Atoi(data)
				// si la conversion renvoie une erreur, on augmente de la date d'une seconde un maximum de 3600 fois jusqu'a atteindre une valeur valide, sinon on renvoie une erreur 400
				if err != nil {
					for i := 0; i < 3600; i++ {
						t1 = t1.Add(time.Second * 1)
						key := "airport:"+airportParam + ":" + t1.Format("2006-01-02:15-04-05")+":"+typeCapteur+":"+idCapteur
						data := get(connexionName,key)
						dataInt, err := strconv.Atoi(data)
						if err == nil {
							values = append(values, dataInt)
							break
						}
					}
				} else {
					// Ajout de la valeur au tableau des valeurs
					values = append(values, dataInt)
				}
				t1 = t1.Add(time.Second * 10)
			}
			if len(values) == 0 {
				w.WriteHeader(http.StatusNotFound)
				http.Error(w, "404 - NOT FOUND", http.StatusNotFound)
				return "No values or not all values found for this date, please provide valid parameter [dateParam1 - INVALID ]"
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
	w.WriteHeader(http.StatusOK)
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
	if count != 0 {
		// Calcul de la moyenne
	result = result / count
	} 
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


