package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/deogracia/toxophilus/database"
	"github.com/deogracia/toxophilus/models"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// Cherche la racine du projet de manière dynamique
func goToProjectRoot() {
	dir, _ := os.Getwd()
	for {
		// Si on trouve le fichier go.mod, c'est qu'on est à la racine !
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			os.Chdir(dir)
			return
		}
		// Sinon, on remonte d'un niveau parent
		parent := filepath.Dir(dir)
		if parent == dir {
			break // On est arrivé à la racine du système de fichiers (sécurité)
		}
		dir = parent
	}
}

// setupMockDB crée une base SQLite en mémoire pour éviter que les routes ne paniquent
func setupMockDB() {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	database.DB = db

	// on force la migration que des models nécessaires pour les tests
	database.DB.AutoMigrate(&models.User{}, &models.Setting{})
	// on créé un utilisateur
	database.DB.Create(&models.User{
		Email:    "test@example.test",
		Password: "password",
	})
}

func TestEnvironments(t *testing.T) {
	// On remplace os.Chdir("../..") par notre chercheur de dossier racine.
	goToProjectRoot()

	setupMockDB()
	gin.SetMode(gin.TestMode)

	// On va tester les deux modes d'un seul coup
	tests := []struct {
		name string
		env  string
	}{
		{"Mode Développement", "development"},
		{"Mode Production", "production"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Initialisation du routeur dans l'environnement donné
			r := setupRouter(tt.env, nil)

			// 2. Test du Favicon
			reqFavicon, _ := http.NewRequest("GET", "/favicon.ico", nil)
			wFavicon := httptest.NewRecorder()
			r.ServeHTTP(wFavicon, reqFavicon)

			if wFavicon.Code != http.StatusOK {
				t.Errorf("Attendu 200 OK pour favicon en %s, obtenu %d", tt.env, wFavicon.Code)
			}

			// 3. Test des Templates (Si /login renvoie 200, c'est que les templates ont bien été parsés)
			reqLogin, _ := http.NewRequest("GET", "/login", nil)
			wLogin := httptest.NewRecorder()
			r.ServeHTTP(wLogin, reqLogin)

			if wLogin.Code != http.StatusOK {
				t.Errorf("Attendu 200 OK pour la page login en %s, obtenu %d", tt.env, wLogin.Code)
			}
		})
	}
}
