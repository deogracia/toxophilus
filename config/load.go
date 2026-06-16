package config

import (
	"strings"

	"github.com/spf13/viper"
)

// LoadConfig charge la configuration et met en place les valeurs par défaut
//
// Il s'appuie sur viper pour charger le nécessaire et faire la fusion des configurations.
func LoadConfig() error {
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
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.file", "toxophilus.log")
	viper.SetDefault("log.format", "texte")

	viper.SetEnvPrefix("TOXO")                             // On définit un prefixe dédié
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // On remplace les point par des underscore dans les noms des variables d'environements (lié à la config TOML)
	viper.AutomaticEnv()

	return viper.ReadInConfig()
}
