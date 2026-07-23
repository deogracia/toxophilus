package database

import (
	"github.com/deogracia/toxophilus/models"
	"gorm.io/gorm"
)

// GormMemberRepository implémente models.MemberRepository avec GORM.
type GormMemberRepository struct {
	db *gorm.DB
}

// NewGormMemberRepository crée une nouvelle instance de GormMemberRepository.
func NewGormMemberRepository(db *gorm.DB) models.MemberRepository {
	return &GormMemberRepository{db: db}
}

func (r *GormMemberRepository) GetAll() ([]models.Member, error) {
	var members []models.Member
	err := r.db.Find(&members).Error
	return members, err
}

func (r *GormMemberRepository) GetArchived() ([]models.Member, error) {
	var archivedMembers []models.Member
	err := r.db.Unscoped().Where("deleted_at IS NOT NULL").Find(&archivedMembers).Error
	return archivedMembers, err
}

func (r *GormMemberRepository) GetByID(id uint) (*models.Member, error) {
	var member models.Member
	err := r.db.First(&member, id).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *GormMemberRepository) GetByIDWithUnscoped(id uint) (*models.Member, error) {
	var member models.Member
	err := r.db.Unscoped().First(&member, id).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *GormMemberRepository) Create(member *models.Member) error {
	return r.db.Create(member).Error
}

func (r *GormMemberRepository) Update(member *models.Member) error {
	return r.db.Save(member).Error
}

func (r *GormMemberRepository) Delete(member *models.Member) error {
	return r.db.Delete(member).Error
}

func (r *GormMemberRepository) Reactivate(id uint) error {
	return r.db.Unscoped().Model(&models.Member{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

func (r *GormMemberRepository) HardDelete(id uint) error {
	return r.db.Unscoped().Delete(&models.Member{}, id).Error
}
