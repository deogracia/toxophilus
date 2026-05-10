package handlers

import (
	"net/http"
	"time"

	"github.com/deogracia/toxophilus/database"
	"github.com/deogracia/toxophilus/models"
	"github.com/gin-gonic/gin"
)

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
func CreateMember(c *gin.Context) {
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

	// Insertion en base de données
	if err := database.DB.Create(&member).Error; err != nil {
		// L'erreur typique ici serait un doublon sur le CodeAdherent (qui est en uniqueIndex)
		c.JSON(http.StatusConflict, gin.H{"error": "Impossible de créer le membre (Ce code adhérent existe peut-être déjà)"})
		return
	}

	// On retourne le membre fraîchement créé (avec son ID généré) et le code 201 (Created)
	c.JSON(http.StatusCreated, member)
}

// ListMembers renvoie tous les adhérents du club
func ListMembers(c *gin.Context) {
	var members []models.Member

	// Find récupère tout par défaut
	if err := database.DB.Find(&members).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des membres"})
		return
	}

	c.JSON(http.StatusOK, members)
}

// UpdateMember met à jour les informations d'un adhérent existant
func UpdateMember(c *gin.Context) {
	// On récupère l'ID passé dans l'URL (ex: /api/members/1)
	id := c.Param("id")
	var member models.Member

	// 1. Vérifier si le membre existe
	if err := database.DB.First(&member, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Membre introuvable"})
		return
	}

	// 2. Récupérer les nouvelles données
	// On réutilise intelligemment la même structure que pour la création
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

	// 4. Sauvegarder en base
	if err := database.DB.Save(&member).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Impossible de modifier : ce Code Adhérent est peut être déjà utilisé par un autre archer."})
		return
	}

	c.JSON(http.StatusOK, member)
}

// DeleteMember supprime un adhérent
func DeleteMember(c *gin.Context) {
	id := c.Param("id")
	var member models.Member

	if err := database.DB.First(&member, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Membre introuvable"})
		return
	}

	// Magie de GORM : Puisque notre modèle inclut gorm.Model,
	// cela ne supprime pas vraiment la ligne, mais remplit le champ "deleted_at".
	// C'est ce qu'on appelle un "Soft Delete". Très sécurisant pour ne pas casser l'historique !
	database.DB.Delete(&member)

	c.JSON(http.StatusOK, gin.H{"message": "Adhérent supprimé avec succès"})
}

// ReactivateMember restaure un membre supprimé (Soft Delete)
func ReactivateMember(c *gin.Context) {
	id := c.Param("id")

	// Unscoped() permet d'ignorer le filtre de suppression de GORM.
	// On remet le champ deleted_at à NULL.
	if err := database.DB.Unscoped().Model(&models.Member{}).Where("id = ?", id).Update("deleted_at", nil).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de réactiver ce membre."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Membre réactivé avec succès"})
}

// HardDeleteMember supprime définitivement un membre de la base de données
func HardDeleteMember(c *gin.Context) {
	id := c.Param("id")

	// L'utilisation de Unscoped().Delete() supprime la ligne physiquement
	if err := database.DB.Unscoped().Delete(&models.Member{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la suppression définitive"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Adhérent supprimé définitivement"})
}

func ExportMemberData(c *gin.Context) {
	id := c.Param("id")
	var member models.Member

	// On récupère le membre (même s'il est en archives)
	if err := database.DB.Unscoped().First(&member, id).Error; err != nil {
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
func GetMembersPage(c *gin.Context) {
	var members []models.Member
	database.DB.Find(&members)

	c.HTML(http.StatusOK, "members.html", gin.H{
		"titre":   "Gestion des Membres - Toxophilus",
		"membres": members,
		"active":  "membres",
	})
}

// GetMemberEditPage affiche le formulaire de modification d'un membre
func GetMemberEditPage(c *gin.Context) {
	id := c.Param("id")
	var member models.Member

	if err := database.DB.First(&member, id).Error; err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/members")
		return
	}

	c.HTML(http.StatusOK, "member_edit.html", gin.H{
		"titre":  "Modifier - Toxophilus",
		"active": "membres",
		"member": member,
	})
}

// GetMemberArchivesPage affiche la liste des membres supprimés
func GetMemberArchivesPage(c *gin.Context) {
	var archivedMembers []models.Member
	database.DB.Unscoped().Where("deleted_at IS NOT NULL").Find(&archivedMembers)

	c.HTML(http.StatusOK, "member_archives.html", gin.H{
		"titre":   "Archives - Les membres supprimés - Toxophilus",
		"active":  "membres",
		"membres": archivedMembers,
	})
}
