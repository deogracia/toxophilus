package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// respondWithRedirect gère la réponse après une création ou modification réussie.
// Si la requête provient d'HTMX, on utilise le header HX-Redirect pour rediriger.
// Sinon, on renvoie les données sous forme de JSON pour les clients d'API standard.
func respondWithRedirect(c *gin.Context, redirectURL string, data interface{}, status int) {
	if c.GetHeader("HX-Request") == "true" {
		c.Header("HX-Redirect", redirectURL)
		c.Status(status)
		return
	}
	c.JSON(status, data)
}

// respondWithDelete gère la réponse après une suppression (soft ou hard) réussie.
// Si la requête provient d'HTMX, on renvoie une chaîne vide pour faire disparaître l'élément du DOM.
// Sinon, on renvoie un message JSON standard de confirmation de suppression.
func respondWithDelete(c *gin.Context, message string) {
	if c.GetHeader("HX-Request") == "true" {
		c.String(http.StatusOK, "")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": message})
}

// RespondWithError uniformise les réponses d'erreur en JSON
func RespondWithError(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{"status": "error", "message": message})
}

// respondWithReactivate gère la réponse après une réactivation d'un élément archivé.
// Si la requête provient d'HTMX, on renvoie un statut 200 vide (HTMX va faire disparaître la ligne).
// Sinon, on renvoie un message JSON standard.
func respondWithReactivate(c *gin.Context, message string) {
	if c.GetHeader("HX-Request") == "true" {
		c.Status(http.StatusOK)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": message})
}
