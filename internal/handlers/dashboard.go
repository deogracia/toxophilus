package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetDashboardPage affiche la page d'accueil
func GetDashboardPage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"titre":  "Accueil - Toxophilus",
		"active": "accueil",
	})
}
