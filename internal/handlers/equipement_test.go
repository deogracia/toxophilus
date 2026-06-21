package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/deogracia/toxophilus/database"
	"github.com/deogracia/toxophilus/models"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupEquipmentTestDB() (*gin.Engine, *EquipementHandler) {
	gin.SetMode(gin.TestMode)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	database.DB = db
	database.DB.AutoMigrate(&models.Riser{}, &models.Limb{})

	riserRepo := database.NewGormRiserRepository(db)
	limbRepo := database.NewGormLimbRepository(db)
	h := NewEquipementHandler(riserRepo, limbRepo)

	r := gin.New()
	r.POST("/risers", h.CreateRiser)
	r.GET("/risers", h.ListRisers)
	r.PUT("/risers/:id", h.UpdateRiser)
	r.DELETE("/risers/:id", h.DeleteRiser)

	r.POST("/limbs", h.CreateLimb)
	r.GET("/limbs", h.ListLimbs)
	r.PUT("/limbs/:id", h.UpdateLimb)
	r.DELETE("/limbs/:id", h.DeleteLimb)

	return r, h
}

func TestRiserCRUD(t *testing.T) {
	r, _ := setupEquipmentTestDB()

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
	r, _ := setupEquipmentTestDB()

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
	r, h := setupEquipmentTestDB()

	// Chargement des templates pour que c.HTML fonctionne
	// (Le chemin correspond à l'arborescence depuis le dossier handlers)
	r.LoadHTMLGlob("../../templates/partials/*.html")
	r.LoadHTMLGlob("../../templates/*.html")

	// Ajout de la route à tester
	r.GET("/equipement/edit/riser/:id", h.GetEditRiserPage)

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
	r, h := setupEquipmentTestDB()

	r.LoadHTMLGlob("../../templates/partials/*.html")
	r.LoadHTMLGlob("../../templates/*.html")

	r.GET("/equipement/edit/limb/:id", h.GetEditLimbPage)

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

func TestEquipementArchives(t *testing.T) {
	r, h := setupEquipmentTestDB()

	// 1. Chargement des templates pour la vue HTML
	templ := template.Must(template.ParseGlob("../../templates/*.html"))
	template.Must(templ.ParseGlob("../../templates/partials/*.html"))
	r.SetHTMLTemplate(templ)

	// 2. Déclaration des routes (Poignées ET Branches)
	r.GET("/equipement/archives", h.GetEquipementArchivesPage)
	r.PUT("/api/risers/:id/reactivate", h.ReactivateRiser)
	r.DELETE("/api/risers/:id/hard", h.HardDeleteRiser)
	r.PUT("/api/limbs/:id/reactivate", h.ReactivateLimb)
	r.DELETE("/api/limbs/:id/hard", h.HardDeleteLimb)

	// ==========================================
	// PRÉPARATION DES DONNÉES
	// ==========================================

	// A. Création d'une poignée
	riser := models.Riser{NumeroSerie: "ARCHIVE-RISER", Marque: "Hoyt"}
	if err := database.DB.Create(&riser).Error; err != nil {
		t.Fatalf("❌ Échec de la création de la poignée : %v", err)
	}
	if err := database.DB.Delete(&riser).Error; err != nil {
		t.Fatalf("❌ Échec du soft-delete de la poignée : %v", err)
	}

	// B. Création de branches
	limb := models.Limb{NumeroSerie: "ARCHIVE-LIMB", Marque: "Uukha"}
	if err := database.DB.Create(&limb).Error; err != nil {
		t.Fatalf("❌ Échec de la création des branches : %v", err)
	}
	if err := database.DB.Delete(&limb).Error; err != nil {
		t.Fatalf("❌ Échec du soft-delete des branches : %v", err)
	}

	// 4. VERIFICATION DE SÉCURITÉ : On s'assure que la base contient bien l'archive
	var checkCount int64
	database.DB.Unscoped().Model(&models.Riser{}).Where("deleted_at IS NOT NULL").Count(&checkCount)
	if checkCount == 0 {
		t.Fatalf("❌ La poignée a été supprimée, mais n'apparaît pas comme archivée dans la base de données !")
	}
	database.DB.Unscoped().Model(&models.Limb{}).Where("deleted_at IS NOT NULL").Count(&checkCount)
	if checkCount == 0 {
		t.Fatalf("❌ Les branches ont été supprimées, mais n'apparaîssent pas comme archivées dans la base de données !")
	}
	// ==========================================
	// EXÉCUTION DES TESTS
	// ==========================================

	// --- TEST 1 : Affichage de la page d'archives ---
	t.Run("Affichage Archives", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/equipement/archives", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Attendu 200 OK, obtenu %d", w.Code)
		}

		if !bytes.Contains(w.Body.Bytes(), []byte("ARCHIVE-RISER")) {
			t.Errorf("La poignée archivée n'apparaît pas sur la page")
		}
		if !bytes.Contains(w.Body.Bytes(), []byte("ARCHIVE-LIMB")) {
			t.Errorf("Les branches archivées n'apparaissent pas sur la page")
		}
	})

	// --- TEST 2 : Restauration ---
	t.Run("Restauration Matériel", func(t *testing.T) {
		// Restauration Poignée
		reqRiser, _ := http.NewRequest("PUT", fmt.Sprintf("/api/risers/%d/reactivate", riser.ID), nil)
		wRiser := httptest.NewRecorder()
		r.ServeHTTP(wRiser, reqRiser)

		if wRiser.Code != http.StatusOK {
			t.Errorf("Attendu 200 OK pour la poignée, obtenu %d", wRiser.Code)
		}

		// Restauration Branches
		reqLimb, _ := http.NewRequest("PUT", fmt.Sprintf("/api/limbs/%d/reactivate", limb.ID), nil)
		wLimb := httptest.NewRecorder()
		r.ServeHTTP(wLimb, reqLimb)

		if wLimb.Code != http.StatusOK {
			t.Errorf("Attendu 200 OK pour les branches, obtenu %d", wLimb.Code)
		}

		// Re-suppression pour le test de hard delete
		database.DB.Delete(&riser)
		database.DB.Delete(&limb)
	})

	// --- TEST 3 : Suppression définitive (Hard Delete) ---
	t.Run("Hard Delete Matériel", func(t *testing.T) {
		// Hard Delete Poignée
		reqRiser, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/risers/%d/hard", riser.ID), nil)
		wRiser := httptest.NewRecorder()
		r.ServeHTTP(wRiser, reqRiser)

		var countRiser int64
		database.DB.Unscoped().Model(&models.Riser{}).Where("id = ?", riser.ID).Count(&countRiser)
		if countRiser > 0 {
			t.Errorf("La poignée existe toujours en base après un Hard Delete")
		}

		// Hard Delete Branches
		reqLimb, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/limbs/%d/hard", limb.ID), nil)
		wLimb := httptest.NewRecorder()
		r.ServeHTTP(wLimb, reqLimb)

		var countLimb int64
		database.DB.Unscoped().Model(&models.Limb{}).Where("id = ?", limb.ID).Count(&countLimb)
		if countLimb > 0 {
			t.Errorf("Les branches existent toujours en base après un Hard Delete")
		}
	})
}
