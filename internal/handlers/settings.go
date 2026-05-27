package handlers

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/deogracia/toxophilus/database"
	"github.com/deogracia/toxophilus/models"
	"github.com/gin-gonic/gin"
)

// GetSettingsPage affiche le formulaire avec les valeurs actuelles de la BDD
func GetSettingsPage(c *gin.Context) {
	var settingsList []models.Setting
	database.DB.Find(&settingsList)

	// On transforme la liste en Map pour la lire facilement dans le template
	settingsMap := make(map[string]string)
	for _, s := range settingsList {
		settingsMap[s.Cle] = s.Valeur
	}

	c.HTML(http.StatusOK, "settings.html", gin.H{
		"titre":    "Configuration du Club - Toxophilus",
		"active":   "settings",
		"Settings": settingsMap,
	})
}

// ProcessSettingsSave traite l'envoi du formulaire et l'upload du fichier
func ProcessSettingsSave(c *gin.Context) {
	// 1. Récupération des champs textes classiques
	fields := []string{
		"pdf_club_name", "pdf_club_subtitle",
		"club_address", "club_website",
		"pdf_footer_ligne1", "pdf_footer_ligne2",
		"pdf_clauses_location",
	}

	for _, field := range fields {
		valeur := c.PostForm(field)
		saveSetting(field, valeur)
		// On sauvegarde ou met à jour dans la table settings

	}

	// Gestion spécifique de la case à cocher (si vide, on stocke "false")
	showContact := c.PostForm("pdf_show_contact_footer")
	if showContact != "true" {
		showContact = "false"
	}
	saveSetting("pdf_show_contact_footer", showContact)

	// 2. Gestion de l'upload du bandeau image
	file, err := c.FormFile("header_image")
	if err == nil {
		// Un fichier a bien été soumis !
		// Étape Open Source essentielle : on crée un dossier "data/uploads" à la racine de l'exécution
		uploadDir := filepath.Join("data", "uploads")
		_ = os.MkdirAll(uploadDir, os.ModePerm)

		// On forge le nom du fichier (on garde son extension d'origine .jpg ou .png)
		ext := filepath.Ext(file.Filename)
		dst := filepath.Join(uploadDir, "bandeau_club"+ext)

		// Gin sauvegarde le fichier physiquement sur le disque dur du club
		if err := c.SaveUploadedFile(file, dst); err == nil {
			// On enregistre ce chemin dans les réglages
			var setting models.Setting
			if database.DB.Where("cle = ?", "pdf_header_image").First(&setting).Error != nil {
				database.DB.Create(&models.Setting{Cle: "pdf_header_image", Valeur: dst})
			} else {
				setting.Valeur = dst
				database.DB.Save(&setting)
			}
		}
	}

	// Redirection vers la page des paramètres avec un message de succès
	c.Redirect(http.StatusSeeOther, "/settings")
}

// Fonction utilitaire interne pour éviter la répétition du bloc GORM
func saveSetting(cle, valeur string) {
	var setting models.Setting
	result := database.DB.Where("cle = ?", cle).First(&setting)
	if result.Error != nil {
		// Nouvelle clé
		database.DB.Create(&models.Setting{Cle: cle, Valeur: valeur})
	} else {
		// Mise à jour
		setting.Valeur = valeur
		database.DB.Save(&setting)
	}
}
