package config

import (
	"path/filepath"
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
	viper.SetDefault("app.data_dir", "data")
	viper.SetDefault("database.driver", "sqlite")
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.file", "toxophilus.log")
	viper.SetDefault("log.format", "texte")

	viper.SetEnvPrefix("TOXO")                             // On définit un prefixe dédié
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // On remplace les point par des underscore dans les noms des variables d'environements (lié à la config TOML)
	viper.AutomaticEnv()

	err := viper.ReadInConfig()

	// Résolution dynamique des chemins dépendants de app.data_dir s'ils ne sont pas explicitement configurés
	dataDir := viper.GetString("app.data_dir")
	if dataDir == "" {
		dataDir = "data"
		viper.Set("app.data_dir", dataDir)
	}

	if !viper.IsSet("app.upload_dir") {
		viper.Set("app.upload_dir", filepath.Join(dataDir, "uploads"))
	}
	if !viper.IsSet("app.pdf_dir") {
		viper.Set("app.pdf_dir", filepath.Join(dataDir, "pdf"))
	}
	if !viper.IsSet("database.dsn") {
		driver := viper.GetString("database.driver")
		if driver == "sqlite" {
			viper.Set("database.dsn", filepath.Join(dataDir, "toxophilus.db"))
		} else {
			viper.Set("database.dsn", "toxophilus.db")
		}
	}

	return err
}
