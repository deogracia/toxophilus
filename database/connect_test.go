package database

import (
	"testing"

	"github.com/spf13/viper"
)

func TestConnect(t *testing.T) {
	// On force les paramètres de Viper spécialement pour ce test
	viper.Set("database.driver", "sqlite")
	viper.Set("database.dsn", ":memory:") // Base éphémère en RAM

	// On lance notre fonction de connexion
	Connect()

	// 1. Vérifier que la variable globale DB a bien été initialisée
	if DB == nil {
		t.Fatal("La base de données n'a pas été initialisée (DB est nil)")
	}

	// 2. Vérifier que l'AutoMigrate a bien fonctionné
	// S'il a fonctionné, la table "users" (ou "members") devrait exister en base
	if !DB.Migrator().HasTable("users") {
		t.Error("La migration a échoué : la table 'users' n'existe pas")
	}

	if !DB.Migrator().HasTable("contracts") {
		t.Error("La migration a échoué : la table 'contracts' n'existe pas")
	}
}
