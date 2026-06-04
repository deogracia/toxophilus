package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/deogracia/toxophilus/config"
	"github.com/deogracia/toxophilus/database"
	"github.com/deogracia/toxophilus/models"
	"github.com/gin-gonic/gin"
)

// --- REQUÊTES ---

type CreateRiserRequest struct {
	NumeroSerie string `json:"numero_serie" binding:"required"`
	Marque      string `json:"marque" binding:"required"`
	Modele      string `json:"modele"`
	Taille      string `json:"taille"`
	Lateralite  string `json:"lateralite"` // "Droitier" ou "Gaucher"
	Couleur     string `json:"couleur"`
	AnneeAchat  string `json:"annee_achat"`
	Prix        string `json:"prix"`
}

type CreateLimbRequest struct {
	NumeroSerie string `json:"numero_serie" binding:"required"`
	Marque      string `json:"marque" binding:"required"`
	Modele      string `json:"modele"`
	Taille      string `json:"taille"`
	Puissance   string `json:"puissance"` // ex: "24#"
	Commentaire string `json:"commentaire"`
	AnneeAchat  string `json:"annee_achat"`
	Prix        string `json:"prix"`
}

func parseEquipementFields(anneeStr string, prixStr string) (anneeAchat int, prix float64, err error) {
	// Gestion des champs AnneeAchat et prix de la requêtte -> la bdd
	// 1. Gestion de AnneeAchat, string -> int
	anneeInt := 0
	if anneeStr != "" {
		var err error
		anneeInt, err = strconv.Atoi(anneeStr)
		if err != nil {
			return 0, 0.0, errors.New("Erreur de conversion: l'année d'achat doit être un nombre entier valide.")
		}
	}

	// 2. gestion de Prix, string -> float64
	prixFloat := 0.0
	if prixStr != "" {
		var err error
		prixFloat, err = strconv.ParseFloat(prixStr, 64)
		if err != nil {
			return 0, 0.0, errors.New("Erreur de conversion: le prix doit être un montant valide (ex: 153.52).")
		}
	}

	return anneeInt, prixFloat, nil
}

// --- POIGNÉES (RISERS) ---

func CreateRiser(c *gin.Context) {
	var req CreateRiserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides", "details": err.Error()})
		return
	}
	var (
		anneeInt  int     = 0
		prixFloat float64 = 0.0
		err       error   = nil
	)

	// Gestion des champs AnneeAchat et prix de la requêtte -> la bdd

	anneeInt, prixFloat, err = parseEquipementFields(req.AnneeAchat, req.Prix)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	riser := models.Riser{
		NumeroSerie:   req.NumeroSerie,
		Marque:        req.Marque,
		Modele:        req.Modele,
		Taille:        req.Taille,
		Lateralite:    req.Lateralite,
		Couleur:       req.Couleur,
		AnneeAchat:    anneeInt,
		Prix:          prixFloat,
		Disponibilite: true, // Par défaut, un matériel neuf est disponible !
	}

	if err := database.DB.Create(&riser).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Impossible d'ajouter la poignée (Numéro de série déjà existant ?)"})
		return
	}

	// On envoie un ordre prioritaire à HTMX pour forcer le changement de page
	c.Header("HX-Redirect", "/equipement")

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

	var (
		anneeInt  int     = 0
		prixFloat float64 = 0.0
		err       error   = nil
	)

	anneeInt, prixFloat, err = parseEquipementFields(req.AnneeAchat, req.Prix)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	riser.NumeroSerie = req.NumeroSerie
	riser.Marque = req.Marque
	riser.Modele = req.Modele
	riser.Taille = req.Taille
	riser.Lateralite = req.Lateralite
	riser.Couleur = req.Couleur
	riser.AnneeAchat = anneeInt
	riser.Prix = prixFloat

	database.DB.Save(&riser)

	// On envoie un ordre prioritaire à HTMX pour forcer le changement de page
	c.Header("HX-Redirect", "/equipement")
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

	// 1. Si la requête vient de l'interface web (HTMX)
	if c.GetHeader("HX-Request") == "true" {
		// On renvoie du vide pour que la ligne du tableau s'évapore proprement
		c.String(http.StatusOK, "")
		return
	}

	// 2. Si la requête vient d'ailleurs (Postman, future application mobile, etc.)
	// On renvoie un vrai message JSON clair pour l'API
	c.JSON(http.StatusOK, gin.H{"message": "Poignée supprimée du catalogue"})
}

// GetEditRiserPage affiche le formulaire de modification d'une poignée
func GetEditRiserPage(c *gin.Context) {
	id := c.Param("id")
	var riser models.Riser
	if err := database.DB.First(&riser, id).Error; err != nil {
		// Tu peux créer un error.html plus tard, ou rediriger
		c.HTML(http.StatusNotFound, "index.html", gin.H{"error": "Poignée introuvable"})
		return
	}

	c.HTML(http.StatusOK, "equipement_edit.html", gin.H{
		"titre":   "Modifier la poignée " + riser.NumeroSerie,
		"type":    "riser",
		"item":    riser,
		"Version": config.AppVersion,
	})
}

// ReactivateRiser annule le soft-delete d'une poignée
func ReactivateRiser(c *gin.Context) {
	id := c.Param("id")
	// On remet le champ deleted_at à NULL pour restaurer l'objet
	if err := database.DB.Unscoped().Model(&models.Riser{}).Where("id = ?", id).Update("deleted_at", nil).Error; err != nil {
		c.String(http.StatusInternalServerError, "Erreur lors de la restauration")
		return
	}
	// On renvoie un statut 200 (OK) vide. HTMX va remplacer la ligne du tableau (tr) par ce vide (donc l'effacer).
	c.Status(http.StatusOK)
}

// HardDeleteRiser supprime définitivement la poignée de la base
func HardDeleteRiser(c *gin.Context) {
	id := c.Param("id")
	// Unscoped().Delete() effectue un vrai DELETE SQL
	if err := database.DB.Unscoped().Delete(&models.Riser{}, id).Error; err != nil {
		c.String(http.StatusInternalServerError, "Erreur lors de la suppression définitive")
		return
	}
	c.Status(http.StatusOK)
}

// --- BRANCHES (LIMBS) ---

func CreateLimb(c *gin.Context) {
	var req CreateLimbRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides", "details": err.Error()})
		return
	}

	var (
		anneeInt  int     = 0
		prixFloat float64 = 0.0
		err       error   = nil
	)

	anneeInt, prixFloat, err = parseEquipementFields(req.AnneeAchat, req.Prix)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	limb := models.Limb{
		NumeroSerie:   req.NumeroSerie,
		Marque:        req.Marque,
		Modele:        req.Modele,
		Taille:        req.Taille,
		Puissance:     req.Puissance,
		Commentaire:   req.Commentaire,
		AnneeAchat:    anneeInt,
		Prix:          prixFloat,
		Disponibilite: true,
	}

	if err := database.DB.Create(&limb).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Impossible d'ajouter les branches (Numéro de série déjà existant ?)"})
		return
	}

	// On envoie un ordre prioritaire à HTMX pour forcer le changement de page
	c.Header("HX-Redirect", "/equipement")

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

	var (
		anneeInt  int     = 0
		prixFloat float64 = 0.0
		err       error   = nil
	)

	anneeInt, prixFloat, err = parseEquipementFields(req.AnneeAchat, req.Prix)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	limb.NumeroSerie = req.NumeroSerie
	limb.Marque = req.Marque
	limb.Modele = req.Modele
	limb.Taille = req.Taille
	limb.Puissance = req.Puissance
	limb.Commentaire = req.Commentaire
	limb.AnneeAchat = anneeInt
	limb.Prix = prixFloat

	database.DB.Save(&limb)

	// On envoie un ordre prioritaire à HTMX pour forcer le changement de page
	c.Header("HX-Redirect", "/equipement")
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

	// 1. Si la requête vient de l'interface web (HTMX)
	if c.GetHeader("HX-Request") == "true" {
		// On renvoie du vide pour que la ligne du tableau s'évapore proprement
		c.String(http.StatusOK, "")
		return
	}

	// 2. Si la requête vient d'ailleurs (Postman, future application mobile, etc.)
	// On renvoie un vrai message JSON clair pour l'API
	c.JSON(http.StatusOK, gin.H{"message": "Branches supprimées du catalogue"})
}

// GetMaterielPage affiche la page d'inventaire avec les poignées et les branches
func GetEquipementPage(c *gin.Context) {
	var risers []models.Riser
	var limbs []models.Limb

	// Récupération de tout le matériel
	database.DB.Find(&risers)
	database.DB.Find(&limbs)

	c.HTML(http.StatusOK, "equipement.html", gin.H{
		"titre":    "Inventaire Matériel - Toxophilus",
		"active":   "equipement", // Allume l'onglet "Matériel" dans la navigation
		"poignees": risers,
		"branches": limbs,
		"Version":  config.AppVersion,
	})
}

// GetEditLimbPage affiche le formulaire de modification de branches
func GetEditLimbPage(c *gin.Context) {
	id := c.Param("id")
	var limb models.Limb
	if err := database.DB.First(&limb, id).Error; err != nil {
		c.HTML(http.StatusNotFound, "index.html", gin.H{"error": "Branches introuvables"})
		return
	}

	c.HTML(http.StatusOK, "equipement_edit.html", gin.H{
		"titre":   "Modifier les branches " + limb.NumeroSerie, // <-- Corrigé ici !
		"type":    "limb",
		"item":    limb,
		"Version": config.AppVersion,
	})
}

// ReactivateLimb annule le soft-delete de branches
func ReactivateLimb(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB.Unscoped().Model(&models.Limb{}).Where("id = ?", id).Update("deleted_at", nil).Error; err != nil {
		c.String(http.StatusInternalServerError, "Erreur lors de la restauration")
		return
	}
	c.Status(http.StatusOK)
}

// HardDeleteLimb supprime définitivement les branches de la base
func HardDeleteLimb(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB.Unscoped().Delete(&models.Limb{}, id).Error; err != nil {
		c.String(http.StatusInternalServerError, "Erreur lors de la suppression définitive")
		return
	}
	c.Status(http.StatusOK)
}

// GetEquipementArchivesPage affiche la page des archives du matériel
func GetEquipementArchivesPage(c *gin.Context) {
	var risers []models.Riser
	var limbs []models.Limb

	// Unscoped() permet de voir les lignes supprimées.
	// On filtre avec "deleted_at IS NOT NULL" pour ne prendre QUE la corbeille.
	database.DB.Unscoped().Where("deleted_at IS NOT NULL").Find(&risers)
	database.DB.Unscoped().Where("deleted_at IS NOT NULL").Find(&limbs)

	c.HTML(http.StatusOK, "equipement_archives.html", gin.H{
		"titre":   "Archives du Matériel - Club Toxophilus",
		"active":  "equipement",
		"Risers":  risers,
		"Limbs":   limbs,
		"Version": config.AppVersion,
	})
}
