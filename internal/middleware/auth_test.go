package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/deogracia/toxophilus/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func TestAuthRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)
	viper.Set("app.secret_key", "super_secret_test")
	defer viper.Reset()

	// On crée un routeur avec le middleware appliqué
	r := gin.New()
	r.Use(AuthRequired())

	// Route de test qui s'exécute uniquement si le middleware laisse passer
	r.GET("/protected", func(c *gin.Context) {
		_, exists := c.Get("userID")
		if !exists {
			c.Status(http.StatusInternalServerError) // Ne devrait jamais arriver si le middleware marche
			return
		}
		// On renvoie l'ID dans le header pour vérifier qu'il a bien été transmis
		c.Header("X-User-ID", "trouvé")
		c.Status(http.StatusOK)
	})

	t.Run("Sans Cookie", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protected", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Attendu 401, obtenu %d", w.Code)
		}
	})

	t.Run("Avec un Cookie Invalide", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.AddCookie(&http.Cookie{Name: "toxo_session", Value: "ceci_nest_pas_un_token_valide"})
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Attendu 401 pour un token invalide, obtenu %d", w.Code)
		}
	})

	t.Run("Avec un Cookie Valide", func(t *testing.T) {
		// 1. On génère un vrai token valide pour l'utilisateur ID 99
		validToken, _ := auth.GenerateToken(99)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.AddCookie(&http.Cookie{Name: "toxo_session", Value: validToken})
		r.ServeHTTP(w, req)

		// 2. Le middleware doit laisser passer (200 OK)
		if w.Code != http.StatusOK {
			t.Errorf("Attendu 200 OK, obtenu %d", w.Code)
		}

		// 3. On vérifie que le middleware a bien extrait et passé l'ID au contexte
		if w.Header().Get("X-User-ID") != "trouvé" {
			t.Error("L'ID de l'utilisateur n'a pas été transmis dans le contexte Gin")
		}
	})
}
