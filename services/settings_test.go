package services

import (
	"testing"

	"github.com/deogracia/toxophilus/database" // N'oublie pas d'adapter le nom de ton module
	"github.com/deogracia/toxophilus/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB initialise une base de données SQLite en mémoire vive pour les tests
func setupTestDB() {
	// ":memory:" indique à SQLite de ne rien écrire sur le disque
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Rend les tests moins bavards
	})
	if err != nil {
		panic("Échec de la connexion à la base de test : " + err.Error())
	}

	// On remplace l'instance globale par notre base de test
	database.DB = db

	// On migre uniquement la table dont on a besoin pour ce test
	database.DB.AutoMigrate(&models.Setting{})
}

func TestGetSetting(t *testing.T) {
	setupTestDB()

	// Préparation (Arrange) : on insère une fausse donnée
	database.DB.Create(&models.Setting{Cle: "test_cle", Valeur: "test_valeur"})

	// Action & Vérification 1 : la clé existe
	val := GetSetting("test_cle", "defaut")
	if val != "test_valeur" {
		t.Errorf("Attendu 'test_valeur', obtenu '%s'", val)
	}

	// Action & Vérification 2 : la clé n'existe pas (doit renvoyer la valeur par défaut)
	valDefaut := GetSetting("cle_inconnue", "valeur_secours")
	if valDefaut != "valeur_secours" {
		t.Errorf("Attendu 'valeur_secours', obtenu '%s'", valDefaut)
	}
}

func TestSetSetting(t *testing.T) {
	setupTestDB()

	// Test 1 : Création d'un nouveau paramètre
	err := SetSetting("nouvelle_cle", "nouvelle_valeur")
	if err != nil {
		t.Errorf("Erreur inattendue lors de la création : %v", err)
	}

	// Vérification en base
	var setting models.Setting
	database.DB.Where("cle = ?", "nouvelle_cle").First(&setting)
	if setting.Valeur != "nouvelle_valeur" {
		t.Errorf("La valeur n'a pas été correctement sauvegardée")
	}

	// Test 2 : Mise à jour d'un paramètre existant
	SetSetting("nouvelle_cle", "valeur_modifiee")
	database.DB.Where("cle = ?", "nouvelle_cle").First(&setting)
	if setting.Valeur != "valeur_modifiee" {
		t.Errorf("La valeur n'a pas été correctement mise à jour, obtenu '%s'", setting.Valeur)
	}
}

func TestInitDefaultSettings(t *testing.T) {
	setupTestDB()

	// On modifie manuellement une valeur qui fait partie des valeurs par défaut
	// pour s'assurer que InitDefaultSettings ne l'écrase pas.
	SetSetting("montant_caution", "500.00")

	// On lance l'initialisation
	InitDefaultSettings()

	// Vérification 1 : Une valeur par défaut a bien été créée
	arcNu := GetSetting("montant_arc_nu", "")
	if arcNu != "120.00" {
		t.Errorf("Le paramètre par défaut montant_arc_nu n'a pas été initialisé correctement")
	}

	// Vérification 2 : Notre valeur modifiée manuellement n'a pas été écrasée
	caution := GetSetting("montant_caution", "")
	if caution != "500.00" {
		t.Errorf("InitDefaultSettings a écrasé une valeur existante ! Obtenu: '%s'", caution)
	}
}
