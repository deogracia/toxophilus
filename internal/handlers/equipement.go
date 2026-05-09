package handlers

import (
	"net/http"

	"github.com/deogracia/toxophilus/database"
	"github.com/deogracia/toxophilus/models"
	"github.com/gin-gonic/gin"
)

// --- REQUÊTES ---

type CreateRiserRequest struct {
	NumeroSerie string  `json:"numero_serie" binding:"required"`
	Marque      string  `json:"marque" binding:"required"`
	Modele      string  `json:"modele"`
	Taille      string  `json:"taille"`
	Lateralite  string  `json:"lateralite"` // "Droitier" ou "Gaucher"
	Couleur     string  `json:"couleur"`
	AnneeAchat  int     `json:"annee_achat"`
	Prix        float64 `json:"prix"`
}

type CreateLimbRequest struct {
	NumeroSerie string  `json:"numero_serie" binding:"required"`
	Marque      string  `json:"marque" binding:"required"`
	Modele      string  `json:"modele"`
	Taille      string  `json:"taille"`
	Puissance   string  `json:"puissance"` // ex: "24#"
	Commentaire string  `json:"commentaire"`
	AnneeAchat  int     `json:"annee_achat"`
	Prix        float64 `json:"prix"`
}

// --- POIGNÉES (RISERS) ---

func CreateRiser(c *gin.Context) {
	var req CreateRiserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides", "details": err.Error()})
		return
	}

	riser := models.Riser{
		NumeroSerie:   req.NumeroSerie,
		Marque:        req.Marque,
		Modele:        req.Modele,
		Taille:        req.Taille,
		Lateralite:    req.Lateralite,
		Couleur:       req.Couleur,
		AnneeAchat:    req.AnneeAchat,
		Prix:          req.Prix,
		Disponibilite: true, // Par défaut, un matériel neuf est disponible !
	}

	if err := database.DB.Create(&riser).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Impossible d'ajouter la poignée (Numéro de série déjà existant ?)"})
		return
	}

	c.JSON(http.StatusCreated, riser)
}

func ListRisers(c *gin.Context) {
	var risers []models.Riser
	database.DB.Find(&risers)
	c.JSON(http.StatusOK, risers)
}

func UpdateRiser(c *gin.Context) {
	id := c.Param("id")
	var riser models.Riser

	if err := database.DB.First(&riser, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Poignée introuvable"})
		return
	}

	var req CreateRiserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides", "details": err.Error()})
		return
	}

	riser.NumeroSerie = req.NumeroSerie
	riser.Marque = req.Marque
	riser.Modele = req.Modele
	riser.Taille = req.Taille
	riser.Lateralite = req.Lateralite
	riser.Couleur = req.Couleur
	riser.AnneeAchat = req.AnneeAchat
	riser.Prix = req.Prix

	database.DB.Save(&riser)
	c.JSON(http.StatusOK, riser)
}

func DeleteRiser(c *gin.Context) {
	id := c.Param("id")
	var riser models.Riser

	if err := database.DB.First(&riser, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Poignée introuvable"})
		return
	}

	database.DB.Delete(&riser)
	c.JSON(http.StatusOK, gin.H{"message": "Poignée supprimée du catalogue"})
}

// --- BRANCHES (LIMBS) ---

func CreateLimb(c *gin.Context) {
	var req CreateLimbRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides", "details": err.Error()})
		return
	}

	limb := models.Limb{
		NumeroSerie:   req.NumeroSerie,
		Marque:        req.Marque,
		Modele:        req.Modele,
		Taille:        req.Taille,
		Puissance:     req.Puissance,
		Commentaire:   req.Commentaire,
		AnneeAchat:    req.AnneeAchat,
		Prix:          req.Prix,
		Disponibilite: true,
	}

	if err := database.DB.Create(&limb).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Impossible d'ajouter les branches (Numéro de série déjà existant ?)"})
		return
	}

	c.JSON(http.StatusCreated, limb)
}

func ListLimbs(c *gin.Context) {
	var limbs []models.Limb
	database.DB.Find(&limbs)
	c.JSON(http.StatusOK, limbs)
}

func UpdateLimb(c *gin.Context) {
	id := c.Param("id")
	var limb models.Limb

	if err := database.DB.First(&limb, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Branches introuvables"})
		return
	}

	var req CreateLimbRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides", "details": err.Error()})
		return
	}

	limb.NumeroSerie = req.NumeroSerie
	limb.Marque = req.Marque
	limb.Modele = req.Modele
	limb.Taille = req.Taille
	limb.Puissance = req.Puissance
	limb.Commentaire = req.Commentaire
	limb.AnneeAchat = req.AnneeAchat
	limb.Prix = req.Prix

	database.DB.Save(&limb)
	c.JSON(http.StatusOK, limb)
}

func DeleteLimb(c *gin.Context) {
	id := c.Param("id")
	var limb models.Limb

	if err := database.DB.First(&limb, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Branches introuvables"})
		return
	}

	database.DB.Delete(&limb)
	c.JSON(http.StatusOK, gin.H{"message": "Branches supprimées du catalogue"})
}
