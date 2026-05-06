package config

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestLoadConfig_OnlyEnv(t *testing.T) {
	viper.Reset()
	// On s'assure qu'aucun fichier n'est lu dans ce répertoire de test
	// (Viper ne trouvera pas de "./config/config.toml" car il n'existe pas encore ici)

	viper.SetEnvPrefix("TOXO")                             // On définit un prefixe dédié
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // On remplace les point par des underscore dans les noms des variables d'environements (lié à la config TOML)

	os.Setenv("TOXO_APP_PORT", "9999")
	defer os.Unsetenv("TOXO_APP_PORT")

	LoadConfig()

	if viper.GetString("app.port") != "9999" {
		t.Errorf("SCÉNARIO ENV SEUL : Attendu port 9999, obtenu '%s'", viper.GetString("app.port"))
	}
}

func TestLoadConfig_WithFile(t *testing.T) {
	viper.Reset()

	viper.SetEnvPrefix("TOXO")                             // On définit un prefixe dédié
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // On remplace les point par des underscore dans les noms des variables d'environements (lié à la config TOML)

	// Création du fichier temporaire
	fauxConfig := "[app]\nport = 7777"
	os.WriteFile("config.toml", []byte(fauxConfig), 0644)
	defer os.Remove("config.toml")

	LoadConfig()

	if viper.GetInt("app.port") != 7777 {
		t.Errorf("SCÉNARIO FICHIER : Attendu port 7777, obtenu %d", viper.GetInt("app.port"))
	}
}

func TestLoadConfig_Override(t *testing.T) {
	viper.Reset()

	viper.SetEnvPrefix("TOXO")                             // On définit un prefixe dédié
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // On remplace les point par des underscore dans les noms des variables d'environements (lié à la config TOML)

	// 1. Fichier présent
	fauxConfig := "[app]\nport = 7777"
	os.WriteFile("config.toml", []byte(fauxConfig), 0644)
	defer os.Remove("config.toml")

	// 2. Env Var présente (doit être prioritaire)
	os.Setenv("TOXO_APP_PORT", "8888")
	defer os.Unsetenv("TOXO_APP_PORT")

	LoadConfig()

	if viper.GetString("app.port") != "8888" {
		t.Errorf("SCÉNARIO OVERRIDE : L'env var 8888 aurait dû écraser le fichier 7777. Obtenu : '%s'", viper.GetString("app.port"))
	}
}
