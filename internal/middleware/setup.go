package middleware

import (
	"net/http"
	"strings"

	"github.com/deogracia/toxophilus/database"
	"github.com/deogracia/toxophilus/models"
	"github.com/gin-gonic/gin"
)

// Variable globale en mémoire vive pour optimiser les performances
var setupDoneMemory = false

// EnsureSetup force la création du premier admin si la table users est vide
func EnsureSetup() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Laissez passer les requêtes vers le setup lui-même
		if strings.HasPrefix(c.Request.URL.Path, "/setup") {
			c.Next()
			return
		}

		// 2. Si on sait déjà que c'est fait (mémoire), on passe (très rapide)
		if setupDoneMemory {
			c.Next()
			return
		}

		// 3. Sinon, on vérifie dans la base s'il y a au moins 1 utilisateur
		var count int64
		database.DB.Model(&models.User{}).Count(&count)

		if count == 0 {
			// Redirection vers le formulaire
			c.Redirect(http.StatusTemporaryRedirect, "/setup")
			c.Abort()
			return
		}

		// L'admin existe, on verrouille en mémoire pour les prochaines requêtes
		setupDoneMemory = true
		c.Next()
	}
}
