package database

import (
	"github.com/deogracia/toxophilus/models"
	"gorm.io/gorm"
)

// GormContractRepository implémente models.ContractRepository avec GORM.
type GormContractRepository struct {
	db *gorm.DB
}

// NewGormContractRepository crée une nouvelle instance de GormContractRepository.
func NewGormContractRepository(db *gorm.DB) models.ContractRepository {
	return &GormContractRepository{db: db}
}

func (r *GormContractRepository) GetAll() ([]models.Contract, error) {
	var contracts []models.Contract
	// On charge systématiquement les relations Member, Riser et Limb
	err := r.db.Preload("Member").Preload("Riser").Preload("Limb").Find(&contracts).Error
	return contracts, err
}

func (r *GormContractRepository) GetByID(id uint) (*models.Contract, error) {
	var contract models.Contract
	err := r.db.Preload("Member").Preload("Riser").Preload("Limb").First(&contract, id).Error
	if err != nil {
		return nil, err
	}
	return &contract, nil
}

func (r *GormContractRepository) Create(contract *models.Contract) error {
	// GORM va tenter de créer également les entités associées si on n'utilise pas Omit.
	// Nous omettons "Riser" et "Limb" car ils existent déjà dans le catalogue, nous voulons simplement lier leurs clés primaires.
	return r.db.Omit("Riser", "Limb").Create(contract).Error
}

func (r *GormContractRepository) Update(contract *models.Contract) error {
	return r.db.Omit("Riser", "Limb").Save(contract).Error
}

func (r *GormContractRepository) Delete(contract *models.Contract) error {
	return r.db.Delete(contract).Error
}
