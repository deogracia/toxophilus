package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/deogracia/toxophilus/database"
	"github.com/deogracia/toxophilus/models"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestEnsureSetup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Redirection si 0 utilisateur", func(t *testing.T) {
		// Setup DB en mémoire
		db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		database.DB = db
		database.DB.AutoMigrate(&models.User{})
		setupDoneMemory = false // Reset de la variable globale pour le test

		r := gin.New()
		r.Use(EnsureSetup())
		r.GET("/test", func(c *gin.Context) { c.Status(200) })

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusTemporaryRedirect {
			t.Errorf("Attendu redirection (307), obtenu %d", w.Code)
		}
	})

	t.Run("Passage si utilisateur existe", func(t *testing.T) {
		setupDoneMemory = false
		database.DB.Create(&models.User{Email: "admin@test.com"})

		r := gin.New()
		r.Use(EnsureSetup())
		r.GET("/test", func(c *gin.Context) { c.Status(200) })

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Attendu OK (200), obtenu %d", w.Code)
		}
	})
}
