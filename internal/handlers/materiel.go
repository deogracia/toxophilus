package handlers

import (
	"net/http"

	"github.com/deogracia/toxophilus/database"
	"github.com/deogracia/toxophilus/models"
	"github.com/gin-gonic/gin"
)

// GetMaterielPage affiche la page d'inventaire avec les poignées et les branches
func GetMaterielPage(c *gin.Context) {
	var poignees []models.Poignee
	var branches []models.Branche

	// Récupération de tout le matériel
	database.DB.Find(&poignees)
	database.DB.Find(&branches)

	c.HTML(http.StatusOK, "materiel.html", gin.H{
		"titre":    "Inventaire Matériel - Toxophilus",
		"active":   "materiel", // Allume l'onglet "Matériel" dans la navigation
		"poignees": poignees,
		"branches": branches,
	})
}
