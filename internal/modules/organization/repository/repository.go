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
	IsUserMemberOfAnyOrganization(userID uuid.UUID) (bool, error)
	AddMember(member *entity.OrganizationMember) error
	Update(org *entity.Organization) error
	ListMembers(orgID uuid.UUID) ([]entity.OrganizationMember, error)
	UpdateMemberRole(organizationID, userID uuid.UUID, role string) error
	RemoveMember(organizationID, userID uuid.UUID) error
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

func (r *organizationRepository) IsUserMemberOfAnyOrganization(userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&entity.OrganizationMember{}).
		Where("user_id = ?", userID).
		Count(&count).Error

	if err != nil {
		r.log.Error("Failed to check if user is member of any organization", zap.Error(err),
			zap.String("user_id", userID.String()))
		return false, err
	}

	return count > 0, nil
}

func (r *organizationRepository) AddMember(member *entity.OrganizationMember) error {
	if err := r.db.Create(member).Error; err != nil {
		r.log.Error("Failed to add organization member", zap.Error(err),
			zap.String("organization_id", member.OrganizationID.String()),
			zap.String("user_id", member.UserID.String()))
		return err
	}
	return nil
}

func (r *organizationRepository) Update(org *entity.Organization) error {
	if err := r.db.Save(org).Error; err != nil {
		r.log.Error("Failed to update organization", zap.Error(err), zap.String("id", org.ID.String()))
		return err
	}
	return nil
}

func (r *organizationRepository) ListMembers(orgID uuid.UUID) ([]entity.OrganizationMember, error) {
	var members []entity.OrganizationMember
	err := r.db.Where("organization_id = ?", orgID).Find(&members).Error
	if err != nil {
		r.log.Error("Failed to list organization members", zap.Error(err), zap.String("org_id", orgID.String()))
		return nil, err
	}
	return members, nil
}

func (r *organizationRepository) UpdateMemberRole(organizationID, userID uuid.UUID, role string) error {
	err := r.db.Model(&entity.OrganizationMember{}).
		Where("organization_id = ? AND user_id = ?", organizationID, userID).
		Update("role", role).Error
	if err != nil {
		r.log.Error("Failed to update member role", zap.Error(err),
			zap.String("organization_id", organizationID.String()),
			zap.String("user_id", userID.String()),
			zap.String("role", role))
		return err
	}
	return nil
}

func (r *organizationRepository) RemoveMember(organizationID, userID uuid.UUID) error {
	err := r.db.Where("organization_id = ? AND user_id = ?", organizationID, userID).
		Delete(&entity.OrganizationMember{}).Error
	if err != nil {
		r.log.Error("Failed to remove organization member", zap.Error(err),
			zap.String("organization_id", organizationID.String()),
			zap.String("user_id", userID.String()))
		return err
	}
	return nil
}
