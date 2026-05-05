package config

import (
	"log"

	"github.com/spf13/viper"
)

func LoadConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("⚠️ Utilisation des variables d'environnement (config.toml introuvable): %v\n", err)
	} else {
		log.Println("✅ Fichier config.toml chargé.")
	}
}
