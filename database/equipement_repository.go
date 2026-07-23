package database

import (
	"github.com/deogracia/toxophilus/models"
	"gorm.io/gorm"
)

// GormRiserRepository implémente models.RiserRepository avec GORM.
type GormRiserRepository struct {
	db *gorm.DB
}

// NewGormRiserRepository crée une nouvelle instance de GormRiserRepository.
func NewGormRiserRepository(db *gorm.DB) models.RiserRepository {
	return &GormRiserRepository{db: db}
}

func (r *GormRiserRepository) GetAll() ([]models.Riser, error) {
	var risers []models.Riser
	err := r.db.Find(&risers).Error
	return risers, err
}

func (r *GormRiserRepository) GetArchived() ([]models.Riser, error) {
	var archivedRisers []models.Riser
	err := r.db.Unscoped().Where("deleted_at IS NOT NULL").Find(&archivedRisers).Error
	return archivedRisers, err
}

func (r *GormRiserRepository) GetByID(id uint) (*models.Riser, error) {
	var riser models.Riser
	err := r.db.First(&riser, id).Error
	if err != nil {
		return nil, err
	}
	return &riser, nil
}

func (r *GormRiserRepository) GetByIDWithUnscoped(id uint) (*models.Riser, error) {
	var riser models.Riser
	err := r.db.Unscoped().First(&riser, id).Error
	if err != nil {
		return nil, err
	}
	return &riser, nil
}

func (r *GormRiserRepository) Create(riser *models.Riser) error {
	return r.db.Create(riser).Error
}

func (r *GormRiserRepository) Update(riser *models.Riser) error {
	return r.db.Save(riser).Error
}

func (r *GormRiserRepository) Delete(riser *models.Riser) error {
	return r.db.Delete(riser).Error
}

func (r *GormRiserRepository) Reactivate(id uint) error {
	return r.db.Unscoped().Model(&models.Riser{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

func (r *GormRiserRepository) HardDelete(id uint) error {
	return r.db.Unscoped().Delete(&models.Riser{}, id).Error
}

// GormLimbRepository implémente models.LimbRepository avec GORM.
type GormLimbRepository struct {
	db *gorm.DB
}

// NewGormLimbRepository crée une nouvelle instance de GormLimbRepository.
func NewGormLimbRepository(db *gorm.DB) models.LimbRepository {
	return &GormLimbRepository{db: db}
}

func (r *GormLimbRepository) GetAll() ([]models.Limb, error) {
	var limbs []models.Limb
	err := r.db.Find(&limbs).Error
	return limbs, err
}

func (r *GormLimbRepository) GetArchived() ([]models.Limb, error) {
	var archivedLimbs []models.Limb
	err := r.db.Unscoped().Where("deleted_at IS NOT NULL").Find(&archivedLimbs).Error
	return archivedLimbs, err
}

func (r *GormLimbRepository) GetByID(id uint) (*models.Limb, error) {
	var limb models.Limb
	err := r.db.First(&limb, id).Error
	if err != nil {
		return nil, err
	}
	return &limb, nil
}

func (r *GormLimbRepository) GetByIDWithUnscoped(id uint) (*models.Limb, error) {
	var limb models.Limb
	err := r.db.Unscoped().First(&limb, id).Error
	if err != nil {
		return nil, err
	}
	return &limb, nil
}

func (r *GormLimbRepository) Create(limb *models.Limb) error {
	return r.db.Create(limb).Error
}

func (r *GormLimbRepository) Update(limb *models.Limb) error {
	return r.db.Save(limb).Error
}

func (r *GormLimbRepository) Delete(limb *models.Limb) error {
	return r.db.Delete(limb).Error
}

func (r *GormLimbRepository) Reactivate(id uint) error {
	return r.db.Unscoped().Model(&models.Limb{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

func (r *GormLimbRepository) HardDelete(id uint) error {
	return r.db.Unscoped().Delete(&models.Limb{}, id).Error
}
