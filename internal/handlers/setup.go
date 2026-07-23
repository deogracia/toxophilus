package handlers

import (
	"net/http"

	"github.com/deogracia/toxophilus/database"
	"github.com/deogracia/toxophilus/internal/auth"
	"github.com/deogracia/toxophilus/models"
	"github.com/gin-gonic/gin"
)

// SetupRequest définit les données minimales pour l'admin
type SetupRequest struct {
	AdminEmail    string `json:"admin_email" binding:"required,email"`
	AdminPassword string `json:"admin_password" binding:"required,min=8"`
}

// ProcessSetup traite la création du premier utilisateur
func ProcessSetup(c *gin.Context) {
	// 1. Double sécurité : vérifier qu'il n'y a vraiment aucun utilisateur
	var count int64
	database.DB.Model(&models.User{}).Count(&count)

	if count > 0 {
		RespondWithError(c, http.StatusForbidden, "L'administrateur a déjà été créé.")
		return
	}

	// 2. Valider les données reçues
	var req SetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, http.StatusBadRequest, "Données invalides : "+err.Error())
		return
	}

	// 3. Hasher le mot de passe et créer l'utilisateur
	hashedPassword, _ := auth.HashPassword(req.AdminPassword)
	adminUser := models.User{
		Email:    req.AdminEmail,
		Password: hashedPassword,
	}

	if err := database.DB.Create(&adminUser).Error; err != nil {
		RespondWithError(c, http.StatusInternalServerError, "Impossible de créer l'administrateur")
		return
	}

	if c.GetHeader("HX-Request") == "true" {
		// On force la redirection vers la page de connexion
		c.Header("HX-Redirect", "/login")
		c.Status(http.StatusNoContent)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Administrateur créé avec succès ! Vous pouvez maintenant vous connecter."})
}

// GetSetupPage affiche la page de création du premier compte
func GetSetupPage(c *gin.Context) {
	c.HTML(http.StatusOK, "setup.html", gin.H{
		"titre": "Premier lancement - Toxophilus",
	})
}
