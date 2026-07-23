package database

import (
	"github.com/deogracia/toxophilus/models"
	"gorm.io/gorm"
)

// GormSettingRepository implémente models.SettingRepository avec GORM.
type GormSettingRepository struct {
	db *gorm.DB
}

// NewGormSettingRepository crée une nouvelle instance de GormSettingRepository.
func NewGormSettingRepository(db *gorm.DB) models.SettingRepository {
	return &GormSettingRepository{db: db}
}

func (r *GormSettingRepository) GetAll() ([]models.Setting, error) {
	var settings []models.Setting
	err := r.db.Find(&settings).Error
	return settings, err
}

func (r *GormSettingRepository) GetByKey(key string) (*models.Setting, error) {
	var setting models.Setting
	err := r.db.Where("cle = ?", key).First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *GormSettingRepository) Save(setting *models.Setting) error {
	return r.db.Save(setting).Error
}

func (r *GormSettingRepository) SaveAll(settings map[string]string) error {
	// Nous utilisons une transaction pour garantir que tous les réglages soient sauvegardés de façon atomique
	return r.db.Transaction(func(tx *gorm.DB) error {
		for key, value := range settings {
			var setting models.Setting
			// On cherche si la clé existe déjà, sinon on l'instancie
			err := tx.Where("cle = ?", key).First(&setting).Error
			if err != nil {
				// Si non trouvée, on crée une nouvelle instance
				setting = models.Setting{
					Cle:    key,
					Valeur: value,
				}
			} else {
				// Si trouvée, on met à jour la valeur
				setting.Valeur = value
			}

			if err := tx.Save(&setting).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
