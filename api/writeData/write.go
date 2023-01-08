package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/gomodule/redigo/redis"
)

func main() {
	// Ouvre le fichier en mode lecture
	file, err := os.Open("lines.txt")
	if err != nil {
		// Gestion de l'erreur si l'ouverture du fichier échoue
		fmt.Println(err)
		return
	}
	defer file.Close()

	// Création d'un scanneur pour lire le fichier ligne par ligne
	scanner := bufio.NewScanner(file)

	// Connexion à Redis
	conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		// Gestion de l'erreur si la connexion échoue
		fmt.Println(err)
		return
	}
	defer conn.Close()

	// Boucle qui itère tant qu'il y a une ligne à lire
	for scanner.Scan() {
		// Récupération de la ligne
		line := scanner.Text()
		// Envoi de la commande SET à Redis
		_, err := conn.Do("SET", line)
		if err != nil {
			// Gestion de l'erreur si l'envoi de la commande échoue
			fmt.Println(err)
			return
		}
	}

	// Vérification des erreurs après la fin de la boucle
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
		return
	}
}
