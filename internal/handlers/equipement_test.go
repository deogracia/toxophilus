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
		body := CreateRiserRequest{
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
