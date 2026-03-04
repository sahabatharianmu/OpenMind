package service

import (
	"context"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/audit_log/dto"
	"github.com/sahabatharianmu/OpenMind/internal/modules/audit_log/entity"
	"github.com/sahabatharianmu/OpenMind/internal/modules/audit_log/repository"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/datatypes"
)

type AuditLogService interface {
	Log(
		ctx context.Context,
		action string,
		resourceType string,
		resourceID *uuid.UUID,
		userID uuid.UUID,
		orgID uuid.UUID,
		details map[string]interface{},
		ipAddress *string,
		userAgent *string,
	) error
	List(
		ctx context.Context,
		organizationID uuid.UUID,
		page, pageSize int,
		filters *dto.FilterOptions,
	) ([]dto.AuditLogResponse, int64, error)
	GetOrganizationID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
}

type auditLogService struct {
	repo repository.AuditLogRepository
	log  logger.Logger
}

func NewAuditLogService(repo repository.AuditLogRepository, log logger.Logger) AuditLogService {
	return &auditLogService{
		repo: repo,
		log:  log,
	}
}

func (s *auditLogService) Log(
	ctx context.Context,
	action string,
	resourceType string,
	resourceID *uuid.UUID,
	userID uuid.UUID,
	orgID uuid.UUID,
	details map[string]interface{},
	ipAddress *string,
	userAgent *string,
) error {
	var detailsJSON datatypes.JSON
	if details != nil {
		jsonData, err := sonic.Marshal(details)
		if err != nil {
			s.log.Error("Failed to marshal audit log details", zap.Error(err))
			return err
		}
		detailsJSON = jsonData
	}

	auditLog := &entity.AuditLog{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         userID,
		Action:         action,
		ResourceType:   resourceType,
		ResourceID:     resourceID,
		Details:        detailsJSON,
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
	}

	if err := s.repo.Create(auditLog); err != nil {
		s.log.Error("Failed to create audit log", zap.Error(err))
		// Don't fail the request if audit logging fails
		return nil
	}

	return nil
}

func (s *auditLogService) List(
	ctx context.Context,
	organizationID uuid.UUID,
	page, pageSize int,
	filters *dto.FilterOptions,
) ([]dto.AuditLogResponse, int64, error) {
	offset := (page - 1) * pageSize
	logs, total, err := s.repo.List(organizationID, pageSize, offset, filters)
	if err != nil {
		return nil, 0, err
	}

	var responses []dto.AuditLogResponse
	for _, log := range logs {
		responses = append(responses, *s.mapEntityToResponse(&log))
	}

	return responses, total, nil
}

func (s *auditLogService) GetOrganizationID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	return s.repo.GetOrganizationID(userID)
}

func (s *auditLogService) mapEntityToResponse(log *entity.AuditLog) *dto.AuditLogResponse {
	var details interface{}
	if log.Details != nil {
		_ = sonic.Unmarshal(log.Details, &details)
	}

	return &dto.AuditLogResponse{
		ID:             log.ID,
		OrganizationID: log.OrganizationID,
		UserID:         log.UserID,
		UserName:       "", // Will be populated by joining with users table if needed
		Action:         log.Action,
		ResourceType:   log.ResourceType,
		ResourceID:     log.ResourceID,
		Details:        details,
		IPAddress:      log.IPAddress,
		CreatedAt:      log.CreatedAt,
	}
}
