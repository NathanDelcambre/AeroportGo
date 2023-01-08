package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

func main() {
	// Ouvre un fichier en mode écriture
	file, err := os.Create("lines.txt")
	if err != nil {
		// Gestion de l'erreur si l'ouverture du fichier échoue
		fmt.Println(err)
		return
	}
	defer file.Close()

	// Création du layout de date et d'heure
	layout := "2006-01-02:15-04-05"

	// Initialisation de la date et de l'heure de départ
	t := time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC)

	// Boucle qui itère tant que l'heure est avant 23h59m59s
	for t.Hour() < 23 {
		// Création du tableau des types de capteurs
		types := [3]string{"Heat", "Pressure", "Wind"}

		// Boucle sur les types de capteurs
		for _, typeCapteur := range types {
			var value int
			// Génération de la valeur aléatoire en fonction du type de capteur
			switch typeCapteur {
			case "Heat":
				value = rand.Intn(11) + 18
			case "Pressure":
				value = rand.Intn(101) + 1000
			case "Wind":
				value = rand.Intn(41) + 20
			}

			// Création de la ligne à écrire
			line := fmt.Sprintf("airport:NTE:%s:%s:1 %d", t.Format(layout), typeCapteur, value)

			// Écriture de la ligne dans le fichier
			_, err := file.WriteString(line + "\n")
			if err != nil {
				// Gestion de l'erreur si l'écriture échoue
				fmt.Println(err)
				return
			}
		}

		// Incrémentation de la date et de l'heure de 10 secondes
		t = t.Add(time.Second * 10)
	}
}