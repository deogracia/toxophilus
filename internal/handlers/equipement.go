package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/deogracia/toxophilus/config"
	"github.com/deogracia/toxophilus/models"
	"github.com/gin-gonic/gin"
)

// EquipementHandler gère les requêtes HTTP pour l'équipement (Risers & Limbs).
type EquipementHandler struct {
	riserRepo models.RiserRepository
	limbRepo  models.LimbRepository
}

// NewEquipementHandler crée une nouvelle instance de EquipementHandler.
func NewEquipementHandler(riserRepo models.RiserRepository, limbRepo models.LimbRepository) *EquipementHandler {
	return &EquipementHandler{
		riserRepo: riserRepo,
		limbRepo:  limbRepo,
	}
}

// --- REQUÊTES ---

type CreateRiserRequest struct {
	NumeroSerie    string `json:"numero_serie" binding:"required"`
	Marque         string `json:"marque" binding:"required"`
	Modele         string `json:"modele"`
	Taille         string `json:"taille"`
	Lateralite     string `json:"lateralite"` // "Droitier" ou "Gaucher"
	Couleur        string `json:"couleur"`
	AnneeAchat     string `json:"annee_achat"`
	DateInventaire string `json:"date_inventaire"`
	Prix           string `json:"prix"`
}

type CreateLimbRequest struct {
	NumeroSerie    string `json:"numero_serie" binding:"required"`
	Marque         string `json:"marque" binding:"required"`
	Modele         string `json:"modele"`
	Taille         string `json:"taille"`
	Puissance      string `json:"puissance"` // ex: "24#"
	Commentaire    string `json:"commentaire"`
	AnneeAchat     string `json:"annee_achat"`
	DateInventaire string `json:"date_inventaire"`
	Prix           string `json:"prix"`
}

func parseEquipementFields(anneeStr string, dateInventaireStr string, prixStr string) (anneeAchat int, DateInventaire int, prix float64, err error) {
	// Gestion des champs AnneeAchat, DateInventaire et prix de la requêtte -> la bdd
	// 1. Gestion de AnneeAchat, string -> int
	anneeInt := 0
	if anneeStr != "" {
		var err error
		anneeInt, err = strconv.Atoi(anneeStr)
		if err != nil {
			return 0, 0, 0.0, errors.New("Erreur de conversion: AnneeAchat - l'année d'achat doit être un nombre entier valide.")
		}
	}

	// 2. Gestion de AnneeAchat, string -> int
	dateInventaireInt := 0
	if dateInventaireStr != "" {
		var err error
		dateInventaireInt, err = strconv.Atoi(dateInventaireStr)
		if err != nil {
			return 0, 0, 0.0, errors.New("Erreur de conversion: DateInventaire - l'année d'achat doit être un nombre entier valide.")
		}
	}

	// 3. gestion de Prix, string -> float64
	prixFloat := 0.0
	if prixStr != "" {
		var err error
		prixFloat, err = strconv.ParseFloat(prixStr, 64)
		if err != nil {
			return 0, 0, 0.0, errors.New("Erreur de conversion: Prix - le prix doit être un montant valide (ex: 153.52).")
		}
	}

	return anneeInt, dateInventaireInt, prixFloat, nil
}

// --- POIGNÉES (RISERS) ---

func (h *EquipementHandler) CreateRiser(c *gin.Context) {
	var req CreateRiserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides", "details": err.Error()})
		return
	}
	var (
		anneeInt          int     = 0
		dateInventaireInt int     = 0
		prixFloat         float64 = 0.0
		err               error   = nil
	)

	anneeInt, dateInventaireInt, prixFloat, err = parseEquipementFields(req.AnneeAchat, req.DateInventaire, req.Prix)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	riser := models.Riser{
		NumeroSerie:    req.NumeroSerie,
		Marque:         req.Marque,
		Modele:         req.Modele,
		Taille:         req.Taille,
		Lateralite:     req.Lateralite,
		Couleur:        req.Couleur,
		AnneeAchat:     anneeInt,
		DateInventaire: dateInventaireInt,
		Prix:           prixFloat,
		Disponibilite:  true, // Par défaut, un matériel neuf est disponible !
	}

	if err := h.riserRepo.Create(&riser); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Impossible d'ajouter la poignée (Numéro de série déjà existant ?)"})
		return
	}

	respondWithRedirect(c, "/equipement", riser, http.StatusCreated)
}

func (h *EquipementHandler) ListRisers(c *gin.Context) {
	risers, err := h.riserRepo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des poignées"})
		return
	}
	c.JSON(http.StatusOK, risers)
}

func (h *EquipementHandler) UpdateRiser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalide"})
		return
	}

	riser, err := h.riserRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Poignée introuvable"})
		return
	}

	var req CreateRiserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides", "details": err.Error()})
		return
	}

	var (
		anneeInt          int     = 0
		dateInventaireInt         = 0
		prixFloat         float64 = 0.0
		errConv           error   = nil
	)

	anneeInt, dateInventaireInt, prixFloat, errConv = parseEquipementFields(req.AnneeAchat, req.DateInventaire, req.Prix)
	if errConv != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errConv.Error()})
		return
	}

	riser.NumeroSerie = req.NumeroSerie
	riser.Marque = req.Marque
	riser.Modele = req.Modele
	riser.Taille = req.Taille
	riser.Lateralite = req.Lateralite
	riser.Couleur = req.Couleur
	riser.AnneeAchat = anneeInt
	riser.DateInventaire = dateInventaireInt
	riser.Prix = prixFloat

	if err := h.riserRepo.Update(riser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la mise à jour"})
		return
	}

	respondWithRedirect(c, "/equipement", riser, http.StatusOK)
}

func (h *EquipementHandler) DeleteRiser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalide"})
		return
	}

	riser, err := h.riserRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Poignée introuvable"})
		return
	}

	if err := h.riserRepo.Delete(riser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la suppression"})
		return
	}

	respondWithDelete(c, "Poignée supprimée du catalogue")
}

// GetEditRiserPage affiche le formulaire de modification d'une poignée
func (h *EquipementHandler) GetEditRiserPage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.HTML(http.StatusNotFound, "index.html", gin.H{"error": "ID invalide"})
		return
	}

	riser, err := h.riserRepo.GetByID(uint(id))
	if err != nil {
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
func (h *EquipementHandler) ReactivateRiser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "ID invalide")
		return
	}

	if err := h.riserRepo.Reactivate(uint(id)); err != nil {
		c.String(http.StatusInternalServerError, "Erreur lors de la restauration")
		return
	}
	respondWithReactivate(c, "Poignée restaurée avec succès")
}

// HardDeleteRiser supprime définitivement la poignée de la base
func (h *EquipementHandler) HardDeleteRiser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "ID invalide")
		return
	}

	if err := h.riserRepo.HardDelete(uint(id)); err != nil {
		c.String(http.StatusInternalServerError, "Erreur lors de la suppression définitive")
		return
	}
	respondWithDelete(c, "Poignée supprimée définitivement")
}

// --- BRANCHES (LIMBS) ---

func (h *EquipementHandler) CreateLimb(c *gin.Context) {
	var req CreateLimbRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides", "details": err.Error()})
		return
	}

	var (
		anneeInt          int     = 0
		dateInventaireInt int     = 0
		prixFloat         float64 = 0.0
		err               error   = nil
	)

	anneeInt, dateInventaireInt, prixFloat, err = parseEquipementFields(req.AnneeAchat, req.DateInventaire, req.Prix)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	limb := models.Limb{
		NumeroSerie:    req.NumeroSerie,
		Marque:         req.Marque,
		Modele:         req.Modele,
		Taille:         req.Taille,
		Puissance:      req.Puissance,
		Commentaire:    req.Commentaire,
		AnneeAchat:     anneeInt,
		DateInventaire: dateInventaireInt,
		Prix:           prixFloat,
		Disponibilite:  true,
	}

	if err := h.limbRepo.Create(&limb); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Impossible d'ajouter les branches (Numéro de série déjà existant ?)"})
		return
	}

	respondWithRedirect(c, "/equipement", limb, http.StatusCreated)
}

func (h *EquipementHandler) ListLimbs(c *gin.Context) {
	limbs, err := h.limbRepo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des branches"})
		return
	}
	c.JSON(http.StatusOK, limbs)
}

func (h *EquipementHandler) UpdateLimb(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalide"})
		return
	}

	limb, err := h.limbRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Branches introuvables"})
		return
	}

	var req CreateLimbRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides", "details": err.Error()})
		return
	}

	var (
		anneeInt          int     = 0
		dateInventaireInt int     = 0
		prixFloat         float64 = 0.0
		errConv           error   = nil
	)

	anneeInt, dateInventaireInt, prixFloat, errConv = parseEquipementFields(req.AnneeAchat, req.DateInventaire, req.Prix)
	if errConv != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errConv.Error()})
		return
	}

	limb.NumeroSerie = req.NumeroSerie
	limb.Marque = req.Marque
	limb.Modele = req.Modele
	limb.Taille = req.Taille
	limb.Puissance = req.Puissance
	limb.Commentaire = req.Commentaire
	limb.AnneeAchat = anneeInt
	limb.DateInventaire = dateInventaireInt
	limb.Prix = prixFloat

	if err := h.limbRepo.Update(limb); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la mise à jour"})
		return
	}

	respondWithRedirect(c, "/equipement", limb, http.StatusOK)
}

func (h *EquipementHandler) DeleteLimb(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalide"})
		return
	}

	limb, err := h.limbRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Branches introuvables"})
		return
	}

	if err := h.limbRepo.Delete(limb); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la suppression"})
		return
	}

	respondWithDelete(c, "Branches supprimées du catalogue")
}

// GetEquipementPage affiche la page d'inventaire avec les poignées et les branches
func (h *EquipementHandler) GetEquipementPage(c *gin.Context) {
	risers, errR := h.riserRepo.GetAll()
	limbs, errL := h.limbRepo.GetAll()

	if errR != nil || errL != nil {
		c.HTML(http.StatusInternalServerError, "index.html", gin.H{"error": "Erreur lors de la récupération du matériel"})
		return
	}

	c.HTML(http.StatusOK, "equipement.html", gin.H{
		"titre":    "Inventaire Matériel - Toxophilus",
		"active":   "equipement", // Allume l'onglet "Matériel" dans la navigation
		"poignees": risers,
		"branches": limbs,
		"Version":  config.AppVersion,
	})
}

// GetEditLimbPage affiche le formulaire de modification de branches
func (h *EquipementHandler) GetEditLimbPage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.HTML(http.StatusNotFound, "index.html", gin.H{"error": "ID invalide"})
		return
	}

	limb, err := h.limbRepo.GetByID(uint(id))
	if err != nil {
		c.HTML(http.StatusNotFound, "index.html", gin.H{"error": "Branches introuvables"})
		return
	}

	// Nettoyage de la puissance pour le formulaire HTML d'édition (qui est de type "number")
	// On ne garde que les chiffres pour éviter que le navigateur n'affiche une zone de saisie vide !
	puissanceChiffres := ""
	for _, char := range limb.Puissance {
		if char >= '0' && char <= '9' {
			puissanceChiffres += string(char)
		}
	}
	if puissanceChiffres != "" {
		limb.Puissance = puissanceChiffres
	}

	c.HTML(http.StatusOK, "equipement_edit.html", gin.H{
		"titre":   "Modifier les branches " + limb.NumeroSerie,
		"type":    "limb",
		"item":    limb,
		"Version": config.AppVersion,
	})
}

// ReactivateLimb annule le soft-delete de branches
func (h *EquipementHandler) ReactivateLimb(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "ID invalide")
		return
	}

	if err := h.limbRepo.Reactivate(uint(id)); err != nil {
		c.String(http.StatusInternalServerError, "Erreur lors de la restauration")
		return
	}
	respondWithReactivate(c, "Branches restaurées avec succès")
}

// HardDeleteLimb supprime définitivement les branches de la base
func (h *EquipementHandler) HardDeleteLimb(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "ID invalide")
		return
	}

	if err := h.limbRepo.HardDelete(uint(id)); err != nil {
		c.String(http.StatusInternalServerError, "Erreur lors de la suppression définitive")
		return
	}
	respondWithDelete(c, "Branches supprimées définitivement")
}

// GetEquipementArchivesPage affiche la page des archives du matériel
func (h *EquipementHandler) GetEquipementArchivesPage(c *gin.Context) {
	risers, errR := h.riserRepo.GetArchived()
	limbs, errL := h.limbRepo.GetArchived()

	if errR != nil || errL != nil {
		c.HTML(http.StatusInternalServerError, "index.html", gin.H{"error": "Erreur lors du chargement des archives"})
		return
	}

	c.HTML(http.StatusOK, "equipement_archives.html", gin.H{
		"titre":   "Archives du Matériel - Club Toxophilus",
		"active":  "equipement",
		"Risers":  risers,
		"Limbs":   limbs,
		"Version": config.AppVersion,
	})
}
