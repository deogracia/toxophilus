package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

func LoadConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	// On cherche dans le dossier courant (utile pour l'appli) et le parent (utile pour les tests)
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")

	// --- Valeurs par défaut ---
	viper.SetDefault("app.port", "8080")
	viper.SetDefault("database.driver", "sqlite")
	viper.SetDefault("database.dsn", "toxophilus.db") // Force la création d'un fichier réel
	viper.SetDefault("app.upload_dir", "data/uploads")

	viper.SetEnvPrefix("TOXO")                             // On définit un prefixe dédié
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // On remplace les point par des underscore dans les noms des variables d'environements (lié à la config TOML)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("ℹ️ Utilisation des variables d'environnement et/ou des valeurs par défaut (config.toml introuvable): %v\n", err)
	} else {
		log.Println("✅ Fichier config.toml chargé.")
	}
}
