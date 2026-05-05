package main

import (
	"github.com/deogracia/toxophilus/config"
	"github.com/deogracia/toxophilus/database"
	"github.com/deogracia/toxophilus/services"
)

func main() {
	config.LoadConfig()
	database.Connect()
	services.InitDefaultSettings()

	// TODO: Initialiser le routeur Gin ici lors de la prochaine session
}
