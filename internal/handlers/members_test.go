package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/deogracia/toxophilus/database"
	"github.com/deogracia/toxophilus/models"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupMemberTestDB initialise une base SQLite en mémoire vierge pour chaque test
func setupMemberTestDB() *gin.Engine {
	gin.SetMode(gin.TestMode)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	database.DB = db
	database.DB.AutoMigrate(&models.Member{})

	r := gin.New()
	r.POST("/members", CreateMember)
	r.GET("/members", ListMembers)
	r.PUT("/members/:id", UpdateMember)
	r.DELETE("/members/:id", DeleteMember)

	return r
}

func TestCreateMember(t *testing.T) {
	r := setupMemberTestDB()

	t.Run("Succès", func(t *testing.T) {
		body := CreateMemberRequest{
			CodeAdherent:  "TEST-001",
			Nom:           "Stark",
			Prenom:        "Arya",
			DateNaissance: "1995-04-15",
			Email:         "arya@winterfell.com",
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/members", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Attendu 201 Created, obtenu %d", w.Code)
		}
	})

	t.Run("Erreur de format de date", func(t *testing.T) {
		body := CreateMemberRequest{
			CodeAdherent:  "TEST-002",
			Nom:           "Snow",
			Prenom:        "Jon",
			DateNaissance: "15/04/1995", // MAUVAIS FORMAT
			Email:         "jon@wall.com",
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/members", bytes.NewBuffer(jsonBody))
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Attendu 400 Bad Request, obtenu %d", w.Code)
		}
	})
}

func TestListMembers(t *testing.T) {
	r := setupMemberTestDB()

	// Insertion manuelle d'un membre pour le test
	database.DB.Create(&models.Member{Nom: "Lannister", Prenom: "Tyrion"})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/members", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Attendu 200 OK, obtenu %d", w.Code)
	}

	// On vérifie que le JSON retourné contient bien notre membre
	if !bytes.Contains(w.Body.Bytes(), []byte("Tyrion")) {
		t.Error("La liste ne contient pas le membre attendu")
	}
}

func TestUpdateMember(t *testing.T) {
	r := setupMemberTestDB()

	// Création d'un membre initial
	date, _ := time.Parse("2006-01-02", "1990-01-01")
	member := models.Member{Nom: "Targaryen", Prenom: "Daenerys", DateNaissance: date}
	database.DB.Create(&member)

	// Payload de mise à jour
	body := CreateMemberRequest{
		CodeAdherent:  "DRAG-01",
		Nom:           "Targaryen",
		Prenom:        "Dany", // Modification ici
		DateNaissance: "1990-01-01",
		Email:         "dany@dragon.com",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	// Attention, on met bien l'ID "1" dans l'URL
	req, _ := http.NewRequest("PUT", "/members/1", bytes.NewBuffer(jsonBody))
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Attendu 200 OK, obtenu %d. Body: %s", w.Code, w.Body.String())
	}

	var updatedMember models.Member
	database.DB.First(&updatedMember, 1)
	if updatedMember.Prenom != "Dany" {
		t.Errorf("La mise à jour a échoué en base. Attendu 'Dany', obtenu '%s'", updatedMember.Prenom)
	}
}

func TestDeleteMember(t *testing.T) {
	r := setupMemberTestDB()

	database.DB.Create(&models.Member{Nom: "Baratheon", Prenom: "Robert"})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/members/1", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Attendu 200 OK, obtenu %d", w.Code)
	}

	// Vérification du "Soft Delete"
	var count int64
	// On demande explicitement à GORM de compter même les supprimés (Unscoped)
	database.DB.Unscoped().Model(&models.Member{}).Where("id = ?", 1).Count(&count)
	if count == 0 {
		t.Error("Le membre a été hard-deleted de la base au lieu d'un soft-delete !")
	}

	// Mais si on fait une requête normale, il ne doit plus apparaître
	database.DB.Model(&models.Member{}).Where("id = ?", 1).Count(&count)
	if count > 0 {
		t.Error("Le membre n'a pas été soft-deleted")
	}
}
