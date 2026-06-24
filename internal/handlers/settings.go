package handlers

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/deogracia/toxophilus/config"
	"github.com/deogracia/toxophilus/models"
	"github.com/gin-gonic/gin"
)

// SettingHandler gère les requêtes HTTP pour les réglages métiers.
type SettingHandler struct {
	repo models.SettingRepository
}

// NewSettingHandler crée une nouvelle instance de SettingHandler.
func NewSettingHandler(repo models.SettingRepository) *SettingHandler {
	return &SettingHandler{repo: repo}
}

// GetSettingsPage affiche le formulaire avec les valeurs actuelles de la BDD
func (h *SettingHandler) GetSettingsPage(c *gin.Context) {
	settingsList, err := h.repo.GetAll()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "index.html", gin.H{"error": "Erreur lors du chargement des réglages"})
		return
	}

	// On transforme la liste en Map pour la lire facilement dans le template
	settingsMap := make(map[string]string)
	for _, s := range settingsList {
		settingsMap[s.Cle] = s.Valeur
	}

	c.HTML(http.StatusOK, "settings.html", gin.H{
		"titre":    "Configuration du Club - Toxophilus",
		"active":   "settings",
		"Settings": settingsMap,
		"Version":  config.AppVersion,
	})
}

// ProcessSettingsSave traite l'envoi du formulaire et l'upload du fichier
func (h *SettingHandler) ProcessSettingsSave(c *gin.Context) {
	// 1. Récupération des champs textes classiques
	fields := []string{
		"pdf_club_name", "pdf_club_subtitle",
		"club_address", "club_website",
		"pdf_footer_ligne1", "pdf_footer_ligne2",
		"pdf_clause_mise_disposition", "pdf_clause_conditions_utilisation",
		"pdf_clause_entretien_reparations", "pdf_clause_depot_garantie",
		"pdf_clause_duree_restitution",
		"montant_caution", "montant_loyer",
	}

	settingsToSave := make(map[string]string)
	for _, field := range fields {
		settingsToSave[field] = c.PostForm(field)
	}

	// Gestion spécifique de la case à cocher (si vide, on stocke "false")
	showContact := c.PostForm("pdf_show_contact_footer")
	if showContact != "true" {
		showContact = "false"
	}
	settingsToSave["pdf_show_contact_footer"] = showContact

	// Enregistrement par lot via le repository (transactionnel)
	if err := h.repo.SaveAll(settingsToSave); err != nil {
		c.HTML(http.StatusInternalServerError, "settings.html", gin.H{"error": "Impossible d'enregistrer les réglages."})
		return
	}

	// 2. Gestion de l'upload du bandeau image
	file, err := c.FormFile("header_image")
	if err == nil {
		// Un fichier a bien été soumis !
		// Étape Open Source essentielle : on crée un dossier "data/uploads" à la racine de l'exécution
		uploadDir := filepath.Join("data", "uploads")
		_ = os.MkdirAll(uploadDir, 0750)

		// On forge le nom du fichier (on garde son extension d'origine .jpg ou .png)
		ext := filepath.Ext(file.Filename)
		dst := filepath.Join(uploadDir, "bandeau_club"+ext)

		// Gin sauvegarde le fichier physiquement sur le disque dur du club
		if err := c.SaveUploadedFile(file, dst); err == nil {
			// On enregistre ce chemin dans les réglages via le repository
			setting, err := h.repo.GetByKey("pdf_header_image")
			if err != nil {
				// Non trouvé, on le crée
				_ = h.repo.Save(&models.Setting{Cle: "pdf_header_image", Valeur: dst})
			} else {
				setting.Valeur = dst
				_ = h.repo.Save(setting)
			}
		}
	}

	// Redirection vers la page des paramètres avec un message de succès
	c.Redirect(http.StatusSeeOther, "/settings")
}
