package services

import (
	"errors"
	"log"

	"github.com/deogracia/toxophilus/database"
	"github.com/deogracia/toxophilus/models"
	"gorm.io/gorm"
)

// GetSetting récupère la valeur du paramètre `cle`
// S'il n'est pas défini, il retourne `defaultVal`
func GetSetting(cle string, defaultVal string) string {
	var setting models.Setting
	if err := database.DB.Where("cle = ?", cle).First(&setting).Error; err != nil {
		return defaultVal
	}
	return setting.Valeur
}

// SetSetting sauve en BDD le couple `cle`/`valeur`
func SetSetting(cle string, valeur string) error {
	var setting models.Setting
	result := database.DB.Where("cle = ?", cle).First(&setting)

	if result.Error != nil && errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return database.DB.Create(&models.Setting{Cle: cle, Valeur: valeur}).Error
	}
	setting.Valeur = valeur
	return database.DB.Save(&setting).Error
}

// InitDefaultSettings charge en base quelques valeurs par défaut
func InitDefaultSettings() {
	defaults := map[string]string{
		"montant_arc_nu":    "120.00",
		"duree_location_an": "1",
		"montant_caution":   "300.00",
	}

	for cle, val := range defaults {
		var existing models.Setting
		if database.DB.Where("cle = ?", cle).First(&existing).Error != nil {
			if err := SetSetting(cle, val); err != nil {
				log.Fatalf("❌ Impossible de mettre la valeur par défaut %s: %s. \nErreur: %v", cle, val, err)
			}
		}
	}
}
