package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSlogLogger(t *testing.T) {
	// 1. Configuration de Gin en mode test
	gin.SetMode(gin.TestMode)

	// 2. Création d'un routeur vierge avec notre middleware
	router := gin.New()
	router.Use(SlogLogger())

	// 3. Création de routes factices pour le test
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	router.GET("/erreur", func(c *gin.Context) {
		c.Status(http.StatusInternalServerError)
	})

	router.GET("/static/test.css", func(c *gin.Context) {
		c.String(http.StatusOK, "css content")
	})

	// ==========================================
	// TEST A : Requête classique réussie (200 OK)
	// ==========================================
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("❌ Test A: Code HTTP attendu 200, obtenu %d", w1.Code)
	}

	// ==========================================
	// TEST B : Requête en erreur (500 Internal Server Error)
	// ==========================================
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/erreur", nil)
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusInternalServerError {
		t.Errorf("❌ Test B: Code HTTP attendu 500, obtenu %d", w2.Code)
	}

	// ==========================================
	// TEST C : Fichier statique (exclusion des logs)
	// ==========================================
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodGet, "/static/test.css", nil)
	router.ServeHTTP(w3, req3)

	if w3.Code != http.StatusOK {
		t.Errorf("❌ Test C: Code HTTP attendu 200, obtenu %d", w3.Code)
	}
}
