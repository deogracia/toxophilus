package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/deogracia/toxophilus/config"
	"github.com/deogracia/toxophilus/models"
	"github.com/gin-gonic/gin"
)

// MemberHandler gère les requêtes HTTP pour les membres.
// Il utilise l'injection de dépendances via l'interface MemberRepository.
type MemberHandler struct {
	repo models.MemberRepository
}

// NewMemberHandler crée une nouvelle instance de MemberHandler.
func NewMemberHandler(repo models.MemberRepository) *MemberHandler {
	return &MemberHandler{repo: repo}
}

// CreateMemberRequest définit ce qu'on attend du formulaire (Postman/Frontend)
type CreateMemberRequest struct {
	CodeAdherent  string `json:"code_adherent" binding:"required"`
	Nom           string `json:"nom" binding:"required"`
	Prenom        string `json:"prenom" binding:"required"`
	DateNaissance string `json:"date_naissance" binding:"required"` // Attendu: AAAA-MM-JJ
	Email         string `json:"email" binding:"required,email"`
	Telephone     string `json:"telephone"`
	NumeroRue     string `json:"numero_rue"`
	Rue           string `json:"rue"`
	Ville         string `json:"ville"`
	CodePostal    string `json:"code_postal"`
}

// CreateMember ajoute un nouvel adhérent dans la base
func (h *MemberHandler) CreateMember(c *gin.Context) {
	var req CreateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides", "details": err.Error()})
		return
	}

	// Conversion de la date (le format "2006-01-02" est la date de référence fixe en Go)
	dateNaissance, err := time.Parse("2006-01-02", req.DateNaissance)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format de date invalide. Utilisez AAAA-MM-JJ"})
		return
	}

	member := models.Member{
		CodeAdherent:  req.CodeAdherent,
		Nom:           req.Nom,
		Prenom:        req.Prenom,
		DateNaissance: dateNaissance,
		Email:         req.Email,
		Telephone:     req.Telephone,
		NumeroRue:     req.NumeroRue,
		Rue:           req.Rue,
		Ville:         req.Ville,
		CodePostal:    req.CodePostal,
	}

	// Insertion en base de données via le repository
	if err := h.repo.Create(&member); err != nil {
		// L'erreur typique ici serait un doublon sur le CodeAdherent (qui est en uniqueIndex)
		c.JSON(http.StatusConflict, gin.H{"error": "Impossible de créer le membre (Ce code adhérent existe peut-être déjà)"})
		return
	}

	// On retourne le membre fraîchement créé (avec son ID généré) et le code 201 (Created)
	c.JSON(http.StatusCreated, member)
}

// ListMembers renvoie tous les adhérents du club
func (h *MemberHandler) ListMembers(c *gin.Context) {
	members, err := h.repo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des membres"})
		return
	}

	c.JSON(http.StatusOK, members)
}

// UpdateMember met à jour les informations d'un adhérent existant
func (h *MemberHandler) UpdateMember(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalide"})
		return
	}

	// 1. Vérifier si le membre existe via le repository
	member, err := h.repo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Membre introuvable"})
		return
	}

	// 2. Récupérer les nouvelles données
	var req CreateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides", "details": err.Error()})
		return
	}

	dateNaissance, err := time.Parse("2006-01-02", req.DateNaissance)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format de date invalide. Utilisez AAAA-MM-JJ"})
		return
	}

	// 3. Mettre à jour les champs
	member.CodeAdherent = req.CodeAdherent
	member.Nom = req.Nom
	member.Prenom = req.Prenom
	member.DateNaissance = dateNaissance
	member.Email = req.Email
	member.Telephone = req.Telephone
	member.NumeroRue = req.NumeroRue
	member.Rue = req.Rue
	member.Ville = req.Ville
	member.CodePostal = req.CodePostal

	// 4. Sauvegarder via le repository
	if err := h.repo.Update(member); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Impossible de modifier : ce Code Adhérent est peut être déjà utilisé par un autre archer."})
		return
	}

	c.JSON(http.StatusOK, member)
}

// DeleteMember supprime un adhérent
func (h *MemberHandler) DeleteMember(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalide"})
		return
	}

	member, err := h.repo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Membre introuvable"})
		return
	}

	// Soft delete via le repository
	if err := h.repo.Delete(member); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de supprimer ce membre."})
		return
	}

	respondWithDelete(c, "Adhérent supprimé avec succès")
}

// ReactivateMember restaure un membre supprimé (Soft Delete)
func (h *MemberHandler) ReactivateMember(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalide"})
		return
	}

	if err := h.repo.Reactivate(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de réactiver ce membre."})
		return
	}

	respondWithReactivate(c, "Membre réactivé avec succès")
}

// HardDeleteMember supprime définitivement un membre de la base de données
func (h *MemberHandler) HardDeleteMember(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalide"})
		return
	}

	if err := h.repo.HardDelete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la suppression définitive"})
		return
	}

	respondWithDelete(c, "Adhérent supprimé définitivement")
}

// ExportMemberData exporte les données d'un membre en JSON
func (h *MemberHandler) ExportMemberData(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalide"})
		return
	}

	// On récupère le membre (même s'il est archivé)
	member, err := h.repo.GetByIDWithUnscoped(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Membre introuvable"})
		return
	}

	// On définit les headers pour forcer le navigateur à télécharger un fichier
	fileName := "export_" + member.Nom + "_" + member.Prenom + ".json"
	c.Header("Content-Disposition", "attachment; filename="+fileName)

	// On renvoie simplement l'objet member en JSON
	c.JSON(http.StatusOK, member)
}

// GetMembersPage affiche la liste des membres actifs
func (h *MemberHandler) GetMembersPage(c *gin.Context) {
	members, err := h.repo.GetAll()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "index.html", gin.H{"error": "Erreur lors du chargement des membres"})
		return
	}

	c.HTML(http.StatusOK, "members.html", gin.H{
		"titre":   "Gestion des Membres - Toxophilus",
		"membres": members,
		"active":  "membres",
		"Version": config.AppVersion,
	})
}

// GetMemberEditPage affiche le formulaire de modification d'un membre
func (h *MemberHandler) GetMemberEditPage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/members")
		return
	}

	member, err := h.repo.GetByID(uint(id))
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/members")
		return
	}

	c.HTML(http.StatusOK, "member_edit.html", gin.H{
		"titre":   "Modifier - Toxophilus",
		"active":  "membres",
		"member":  member,
		"Version": config.AppVersion,
	})
}

// GetMemberArchivesPage affiche la liste des membres supprimés
func (h *MemberHandler) GetMemberArchivesPage(c *gin.Context) {
	archivedMembers, err := h.repo.GetArchived()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "index.html", gin.H{"error": "Erreur lors du chargement des archives"})
		return
	}

	c.HTML(http.StatusOK, "member_archives.html", gin.H{
		"titre":   "Archives - Les membres supprimés - Toxophilus",
		"active":  "membres",
		"membres": archivedMembers,
		"Version": config.AppVersion,
	})
}
