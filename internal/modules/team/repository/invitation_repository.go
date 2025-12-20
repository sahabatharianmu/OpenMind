package repository

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/team/entity"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TeamInvitationRepository interface {
	Create(invitation *entity.TeamInvitation) error
	GetByID(id uuid.UUID) (*entity.TeamInvitation, error)
	GetByToken(token string) (*entity.TeamInvitation, error)
	GetByEmailAndOrganization(email string, organizationID uuid.UUID) (*entity.TeamInvitation, error)
	ListByOrganization(organizationID uuid.UUID, limit, offset int) ([]entity.TeamInvitation, int64, error)
	Update(invitation *entity.TeamInvitation) error
	Delete(id uuid.UUID) error
	CancelPendingInvitations(email string, organizationID uuid.UUID) error
}

type teamInvitationRepository struct {
	db  *gorm.DB
	log logger.Logger
}

func NewTeamInvitationRepository(db *gorm.DB, log logger.Logger) TeamInvitationRepository {
	return &teamInvitationRepository{
		db:  db,
		log: log,
	}
}

func (r *teamInvitationRepository) Create(invitation *entity.TeamInvitation) error {
	if err := r.db.Create(invitation).Error; err != nil {
		r.log.Error("Failed to create team invitation", zap.Error(err),
			zap.String("email", invitation.Email),
			zap.String("organization_id", invitation.OrganizationID.String()))
		return err
	}
	return nil
}

func (r *teamInvitationRepository) GetByID(id uuid.UUID) (*entity.TeamInvitation, error) {
	var invitation entity.TeamInvitation
	if err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&invitation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("invitation not found: %s", id.String())
		}
		r.log.Error("Failed to get invitation by ID", zap.Error(err), zap.String("id", id.String()))
		return nil, err
	}
	return &invitation, nil
}

func (r *teamInvitationRepository) GetByToken(token string) (*entity.TeamInvitation, error) {
	var invitation entity.TeamInvitation
	if err := r.db.Where("token = ? AND deleted_at IS NULL", token).First(&invitation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("invitation not found for token")
		}
		r.log.Error("Failed to get invitation by token", zap.Error(err))
		return nil, err
	}
	return &invitation, nil
}

func (r *teamInvitationRepository) GetByEmailAndOrganization(email string, organizationID uuid.UUID) (*entity.TeamInvitation, error) {
	var invitation entity.TeamInvitation
	if err := r.db.Where("email = ? AND organization_id = ? AND deleted_at IS NULL", email, organizationID).
		Order("created_at DESC").First(&invitation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("invitation not found for email %s in organization %s", email, organizationID.String())
		}
		r.log.Error("Failed to get invitation by email and organization", zap.Error(err),
			zap.String("email", email),
			zap.String("organization_id", organizationID.String()))
		return nil, err
	}
	return &invitation, nil
}

func (r *teamInvitationRepository) ListByOrganization(organizationID uuid.UUID, limit, offset int) ([]entity.TeamInvitation, int64, error) {
	var invitations []entity.TeamInvitation
	var total int64

	if err := r.db.Model(&entity.TeamInvitation{}).
		Where("organization_id = ? AND deleted_at IS NULL", organizationID).
		Count(&total).Error; err != nil {
		r.log.Error("Failed to count invitations", zap.Error(err))
		return nil, 0, err
	}

	if err := r.db.Where("organization_id = ? AND deleted_at IS NULL", organizationID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&invitations).Error; err != nil {
		r.log.Error("Failed to list invitations", zap.Error(err))
		return nil, 0, err
	}

	return invitations, total, nil
}

func (r *teamInvitationRepository) Update(invitation *entity.TeamInvitation) error {
	if err := r.db.Save(invitation).Error; err != nil {
		r.log.Error("Failed to update invitation", zap.Error(err), zap.String("id", invitation.ID.String()))
		return err
	}
	return nil
}

func (r *teamInvitationRepository) Delete(id uuid.UUID) error {
	if err := r.db.Delete(&entity.TeamInvitation{}, "id = ?", id).Error; err != nil {
		r.log.Error("Failed to delete invitation", zap.Error(err), zap.String("id", id.String()))
		return err
	}
	return nil
}

func (r *teamInvitationRepository) CancelPendingInvitations(email string, organizationID uuid.UUID) error {
	if err := r.db.Model(&entity.TeamInvitation{}).
		Where("email = ? AND organization_id = ? AND status = 'pending' AND deleted_at IS NULL", email, organizationID).
		Update("status", "cancelled").Error; err != nil {
		r.log.Error("Failed to cancel pending invitations", zap.Error(err),
			zap.String("email", email),
			zap.String("organization_id", organizationID.String()))
		return err
	}
	return nil
}

