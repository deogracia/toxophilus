package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/deogracia/toxophilus/database"
	"github.com/deogracia/toxophilus/models"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestProcessSetup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	database.DB = db
	database.DB.AutoMigrate(&models.User{})

	r := gin.New()
	r.POST("/setup", ProcessSetup)

	// Données de test
	body := SetupRequest{
		AdminEmail:    "new@admin.com",
		AdminPassword: "password123",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/setup", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Attendu 200, obtenu %d. Body: %s", w.Code, w.Body.String())
	}

	// Vérifier que l'utilisateur est bien en base
	var user models.User
	database.DB.First(&user)
	if user.Email != "new@admin.com" {
		t.Error("L'utilisateur n'a pas été créé en base de données")
	}
}
