package repository

import (
	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/organization/entity"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type OrganizationRepository interface {
	GetByID(id uuid.UUID) (*entity.Organization, error)
	GetByUserID(userID uuid.UUID) (*entity.Organization, error)
	GetMemberCount(orgID uuid.UUID) (int64, error)
	GetMemberRole(organizationID, userID uuid.UUID) (string, error)
	Update(org *entity.Organization) error
}

type organizationRepository struct {
	db  *gorm.DB
	log logger.Logger
}

func NewOrganizationRepository(db *gorm.DB, log logger.Logger) OrganizationRepository {
	return &organizationRepository{
		db:  db,
		log: log,
	}
}

func (r *organizationRepository) GetByID(id uuid.UUID) (*entity.Organization, error) {
	var org entity.Organization
	if err := r.db.First(&org, "id = ?", id).Error; err != nil {
		r.log.Error("Failed to get organization by ID", zap.Error(err), zap.String("id", id.String()))
		return nil, err
	}
	return &org, nil
}

func (r *organizationRepository) GetByUserID(userID uuid.UUID) (*entity.Organization, error) {
	var org entity.Organization
	err := r.db.Table("organizations").
		Joins("JOIN organization_members ON organization_members.organization_id = organizations.id").
		Where("organization_members.user_id = ?", userID).
		First(&org).Error

	if err != nil {
		r.log.Error("Failed to get organization by user ID", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, err
	}

	return &org, nil
}

func (r *organizationRepository) GetMemberCount(orgID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&entity.OrganizationMember{}).
		Where("organization_id = ?", orgID).
		Count(&count).Error

	if err != nil {
		r.log.Error("Failed to count organization members", zap.Error(err), zap.String("org_id", orgID.String()))
		return 0, err
	}

	return count, nil
}

func (r *organizationRepository) GetMemberRole(organizationID, userID uuid.UUID) (string, error) {
	var member entity.OrganizationMember
	err := r.db.Where("organization_id = ? AND user_id = ?", organizationID, userID).
		First(&member).Error

	if err != nil {
		r.log.Error("Failed to get member role", zap.Error(err), 
			zap.String("organization_id", organizationID.String()),
			zap.String("user_id", userID.String()))
		return "", err
	}

	return member.Role, nil
}

func (r *organizationRepository) Update(org *entity.Organization) error {
	if err := r.db.Save(org).Error; err != nil {
		r.log.Error("Failed to update organization", zap.Error(err), zap.String("id", org.ID.String()))
		return err
	}
	return nil
}
