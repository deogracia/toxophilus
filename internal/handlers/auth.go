package handlers

import (
	"net/http"

	"github.com/deogracia/toxophilus/database"
	"github.com/deogracia/toxophilus/internal/auth"
	"github.com/deogracia/toxophilus/models"
	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginHandler gère l'authentification et la création du cookie de session
func LoginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format d'identifiants invalide"})
		return
	}

	var user models.User
	// 1. On cherche l'utilisateur par son email
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		// Par sécurité, on ne dit pas si c'est l'email ou le mot de passe qui est faux
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Identifiants incorrects"})
		return
	}

	// 2. On vérifie le mot de passe hashé
	if !auth.CheckPasswordHash(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Identifiants incorrects"})
		return
	}

	// 3. Génération du token JWT
	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur interne lors de la création de la session"})
		return
	}

	// 4. On place le token dans un cookie sécurisé
	// Signature : SetCookie(name, value, maxAge, path, domain, secure, httpOnly)
	// maxAge: 86400 secondes = 24 heures
	// httpOnly (dernier paramètre à true) est VITAL : il empêche le JavaScript côté client de lire le token (protection XSS)
	c.SetCookie("toxo_session", token, 86400, "/", "", false, true)

	// Note: 'secure' est à false pour fonctionner sur localhost en HTTP.
	// À passer à true en production si tu utilises HTTPS.

	c.JSON(http.StatusOK, gin.H{"message": "Connexion réussie"})
}

// LogoutHandler permet de détruire la session
func LogoutHandler(c *gin.Context) {
	// On écrase le cookie avec une durée de vie négative pour forcer le navigateur à le supprimer
	c.SetCookie("toxo_session", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Déconnexion réussie"})
}
