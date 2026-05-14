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

func setupEquipmentTestDB() *gin.Engine {
	gin.SetMode(gin.TestMode)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	database.DB = db
	database.DB.AutoMigrate(&models.Riser{}, &models.Limb{})

	r := gin.New()
	r.POST("/risers", CreateRiser)
	r.GET("/risers", ListRisers)
	r.PUT("/risers/:id", UpdateRiser)
	r.DELETE("/risers/:id", DeleteRiser)

	r.POST("/limbs", CreateLimb)
	r.GET("/limbs", ListLimbs)
	r.PUT("/limbs/:id", UpdateLimb)
	r.DELETE("/limbs/:id", DeleteLimb)

	return r
}

func TestRiserCRUD(t *testing.T) {
	r := setupEquipmentTestDB()

	// 1. Create
	t.Run("Create Riser", func(t *testing.T) {
		body := CreateRiserRequest{
			NumeroSerie: "R-TEST-01",
			Marque:      "WNS",
			Lateralite:  "Droitier",
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/risers", bytes.NewBuffer(jsonBody))
		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Attendu 201 Created, obtenu %d", w.Code)
		}
	})

	// 2. List
	t.Run("List Risers", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/risers", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Attendu 200 OK, obtenu %d", w.Code)
		}
		if !bytes.Contains(w.Body.Bytes(), []byte("R-TEST-01")) {
			t.Error("La poignée créée n'est pas dans la liste")
		}
	})

	// 3. Update
	t.Run("Update Riser", func(t *testing.T) {
		body := CreateRiserRequest{
			NumeroSerie: "R-TEST-01-MOD",
			Marque:      "Kinetic",
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/risers/1", bytes.NewBuffer(jsonBody))
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Attendu 200 OK, obtenu %d", w.Code)
		}
	})

	// 4. Delete
	t.Run("Delete Riser", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/risers/1", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Attendu 200 OK, obtenu %d", w.Code)
		}
	})
}

func TestLimbCRUD(t *testing.T) {
	r := setupEquipmentTestDB()

	// 1. Create
	t.Run("Create Limb", func(t *testing.T) {
		body := CreateLimbRequest{
			NumeroSerie: "L-TEST-01",
			Marque:      "Sanlida",
			Puissance:   "24#",
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/limbs", bytes.NewBuffer(jsonBody))
		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Attendu 201 Created, obtenu %d", w.Code)
		}
	})

	// 2. List
	t.Run("List Limbs", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/limbs", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Attendu 200 OK, obtenu %d", w.Code)
		}
		if !bytes.Contains(w.Body.Bytes(), []byte("L-TEST-01")) {
			t.Error("Les branches créées ne sont pas dans la liste")
		}
	})

	// 3. Update
	t.Run("Update Limb", func(t *testing.T) {
		body := CreateLimbRequest{
			NumeroSerie: "L-TEST-01-MOD",
			Marque:      "Kinetic",
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/limbs/1", bytes.NewBuffer(jsonBody))
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Attendu 200 OK, obtenu %d", w.Code)
		}
	})

	// 4. Delete (pour valider le soft delete sur les branches)
	t.Run("Delete Limb", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/limbs/1", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Attendu 200 OK, obtenu %d", w.Code)
		}
	})
}

func TestGetEditRiserPage(t *testing.T) {
	r := setupEquipmentTestDB()

	// Chargement des templates pour que c.HTML fonctionne
	// (Le chemin correspond à l'arborescence depuis le dossier handlers)
	r.LoadHTMLGlob("../../templates/partials/*.html")
	r.LoadHTMLGlob("../../templates/*.html")

	// Ajout de la route à tester
	r.GET("/equipement/edit/riser/:id", GetEditRiserPage)

	// Création d'une poignée directement en base
	riser := models.Riser{
		NumeroSerie: "EDIT-RISER-001",
		Marque:      "Win&Win",
	}
	database.DB.Create(&riser)

	// --- TEST A : La poignée existe (Succès) ---
	t.Run("Succès - Poignée trouvée", func(t *testing.T) {
		// riser.ID devrait être 1 car la base est recréée vide
		req, _ := http.NewRequest("GET", "/equipement/edit/riser/1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Attendu 200 OK, obtenu %d", w.Code)
		}
	})

	// --- TEST B : La poignée n'existe pas (Erreur 404) ---
	t.Run("Erreur 404 - Poignée introuvable", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/equipement/edit/riser/9999", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Attendu 404 Not Found, obtenu %d", w.Code)
		}
	})
}

func TestGetEditLimbPage(t *testing.T) {
	r := setupEquipmentTestDB()

	r.LoadHTMLGlob("../../templates/partials/*.html")
	r.LoadHTMLGlob("../../templates/*.html")

	r.GET("/equipement/edit/limb/:id", GetEditLimbPage)

	// Création de branches directement en base
	limb := models.Limb{
		NumeroSerie: "EDIT-LIMB-001",
		Puissance:   "24 lbs",
	}
	database.DB.Create(&limb)

	// --- TEST A : Les branches existent (Succès) ---
	t.Run("Succès - Branches trouvées", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/equipement/edit/limb/1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Attendu 200 OK, obtenu %d", w.Code)
		}
	})

	// --- TEST B : Les branches n'existent pas (Erreur 404) ---
	t.Run("Erreur 404 - Branches introuvables", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/equipement/edit/limb/9999", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Attendu 404 Not Found, obtenu %d", w.Code)
		}
	})
}
