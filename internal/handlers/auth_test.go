package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/deogracia/toxophilus/database"
	"github.com/deogracia/toxophilus/internal/auth"
	"github.com/deogracia/toxophilus/models"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

func TestAuthHandlers(t *testing.T) {
	// Configuration de l'environnement de test
	gin.SetMode(gin.TestMode)
	viper.Set("app.secret_key", "super_secret_test")
	defer viper.Reset()

	// Initialisation de la BDD en mémoire
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	database.DB = db
	database.DB.AutoMigrate(&models.User{})

	// Création d'un utilisateur de test avec un mot de passe haché
	hashedPassword, _ := auth.HashPassword("MotDePasse123")
	database.DB.Create(&models.User{
		Email:    "test@club.com",
		Password: hashedPassword,
	})

	// Création du routeur
	r := gin.New()
	r.POST("/login", LoginHandler)
	r.POST("/logout", LogoutHandler)

	t.Run("Connexion réussie", func(t *testing.T) {
		body := LoginRequest{Email: "test@club.com", Password: "MotDePasse123"}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Attendu 200 OK, obtenu %d", w.Code)
		}

		// Vérifier que le cookie a bien été envoyé
		cookies := w.Result().Cookies()
		if len(cookies) == 0 || cookies[0].Name != "toxo_session" {
			t.Error("Le cookie de session n'a pas été créé")
		}
	})

	t.Run("Connexion échouée (mauvais mot de passe)", func(t *testing.T) {
		body := LoginRequest{Email: "test@club.com", Password: "FauxPassword"}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Attendu 401 Unauthorized, obtenu %d", w.Code)
		}
	})

	t.Run("Déconnexion (Logout)", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/logout", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Attendu 200 OK, obtenu %d", w.Code)
		}

		cookies := w.Result().Cookies()
		if len(cookies) > 0 {
			// On vérifie que la durée de vie du cookie est bien écrasée
			if cookies[0].MaxAge > 0 {
				t.Error("Le cookie n'a pas été correctement invalidé")
			}
		}
	})
}
