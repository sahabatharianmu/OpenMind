package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/notification/entity"
	notificationRepo "github.com/sahabatharianmu/OpenMind/internal/modules/notification/repository"
	"github.com/sahabatharianmu/OpenMind/internal/modules/patient/dto"
	handoffEntity "github.com/sahabatharianmu/OpenMind/internal/modules/patient/entity"
	handoffRepo "github.com/sahabatharianmu/OpenMind/internal/modules/patient/repository"
	userRepo "github.com/sahabatharianmu/OpenMind/internal/modules/user/repository"
	"github.com/sahabatharianmu/OpenMind/pkg/email"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
	"go.uber.org/zap"
)

type PatientHandoffService interface {
	RequestHandoff(
		ctx context.Context,
		patientID, receivingClinicianID, requestingClinicianID, organizationID uuid.UUID,
		message, requestedRole *string,
		baseURL string,
	) (*dto.HandoffResponse, error)
	ApproveHandoff(
		ctx context.Context,
		handoffID, approvingClinicianID, organizationID uuid.UUID,
		reason *string,
		baseURL string,
	) error
	RejectHandoff(
		ctx context.Context,
		handoffID, rejectingClinicianID, organizationID uuid.UUID,
		reason *string,
		baseURL string,
	) error
	CancelHandoff(
		ctx context.Context,
		handoffID, cancellingClinicianID, organizationID uuid.UUID,
	) error
	GetHandoff(ctx context.Context, handoffID, userID, organizationID uuid.UUID) (*dto.HandoffResponse, error)
	ListHandoffs(ctx context.Context, patientID, userID, organizationID uuid.UUID) ([]dto.HandoffResponse, error)
	ListPendingHandoffs(ctx context.Context, clinicianID, organizationID uuid.UUID) ([]dto.HandoffResponse, error)
}

type patientHandoffService struct {
	handoffRepo      handoffRepo.PatientHandoffRepository
	patientRepo      handoffRepo.PatientRepository
	patientSvc       PatientService
	userRepo         userRepo.UserRepository
	emailService     *email.EmailService
	notificationRepo notificationRepo.NotificationRepository
	log              logger.Logger
}

func NewPatientHandoffService(
	handoffRepo handoffRepo.PatientHandoffRepository,
	patientRepo handoffRepo.PatientRepository,
	patientSvc PatientService,
	userRepo userRepo.UserRepository,
	emailService *email.EmailService,
	notificationRepo notificationRepo.NotificationRepository,
	log logger.Logger,
) PatientHandoffService {
	return &patientHandoffService{
		handoffRepo:      handoffRepo,
		patientRepo:      patientRepo,
		patientSvc:       patientSvc,
		userRepo:         userRepo,
		emailService:     emailService,
		notificationRepo: notificationRepo,
		log:              log,
	}
}

func (s *patientHandoffService) RequestHandoff(
	ctx context.Context,
	patientID, receivingClinicianID, requestingClinicianID, organizationID uuid.UUID,
	message, requestedRole *string,
	baseURL string,
) (*dto.HandoffResponse, error) {
	// 1. Validate patient exists and belongs to organization
	patient, err := s.patientRepo.FindByID(patientID)
	if err != nil {
		s.log.Error("Failed to find patient for handoff", zap.Error(err), zap.String("patient_id", patientID.String()))
		return nil, response.NewBadRequest("Patient not found or you do not have access to this patient")
	}
	if patient == nil {
		return nil, response.NewBadRequest("Patient not found or you do not have access to this patient")
	}
	if patient.OrganizationID != organizationID {
		return nil, response.NewBadRequest("Patient not found or you do not have access to this patient")
	}

	// 2. Validate requesting clinician is assigned to patient
	isAssigned, err := s.patientRepo.IsPatientAssignedToClinician(patientID, requestingClinicianID)
	if err != nil {
		return nil, fmt.Errorf("failed to check assignment: %w", err)
	}
	if !isAssigned {
		return nil, response.NewBadRequest("You must be assigned to this patient to request a handoff")
	}

	// 3. Validate receiving clinician exists and is in same organization
	receivingOrgID, err := s.patientRepo.GetOrganizationID(receivingClinicianID)
	if err != nil || receivingOrgID != organizationID {
		return nil, response.NewBadRequest("Receiving clinician does not belong to this organization")
	}

	// 4. Cannot request handoff to self
	if requestingClinicianID == receivingClinicianID {
		return nil, response.NewBadRequest("Cannot request handoff to yourself")
	}

	// 5. Check if there's already a pending handoff for this patient/clinician combination
	existingHandoff, err := s.handoffRepo.GetPendingByPatientAndClinician(patientID, requestingClinicianID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing handoff: %w", err)
	}
	if existingHandoff != nil {
		return nil, response.NewBadRequest("You already have a pending handoff request for this patient")
	}

	// 6. Get requesting clinician's current role
	assignments, err := s.patientRepo.GetAssignedClinicians(patientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get assignments: %w", err)
	}

	var requestingRole string
	for _, assignment := range assignments {
		if assignment.ClinicianID == requestingClinicianID {
			requestingRole = assignment.Role
			break
		}
	}

	if requestingRole == "" {
		return nil, response.NewBadRequest("Requesting clinician is not assigned to this patient")
	}

	// 7. Determine role for receiving clinician (use requested role or inherit)
	roleToAssign := requestedRole
	if roleToAssign == nil || *roleToAssign == "" {
		roleToAssign = &requestingRole
	}

	// 8. Create handoff record
	handoff := &handoffEntity.PatientHandoff{
		ID:                    uuid.New(),
		PatientID:             patientID,
		RequestingClinicianID: requestingClinicianID,
		ReceivingClinicianID:  receivingClinicianID,
		Status:                handoffEntity.StatusRequested,
		RequestedRole:         roleToAssign,
		Message:             message,
		RequestedAt:           time.Now(),
	}

	if err := s.handoffRepo.Create(handoff); err != nil {
		s.log.Error("Failed to create handoff", zap.Error(err),
			zap.String("patient_id", patientID.String()),
			zap.String("requesting_clinician_id", requestingClinicianID.String()),
			zap.String("receiving_clinician_id", receivingClinicianID.String()))
		
		// Check if it's a foreign key constraint error
		if strings.Contains(err.Error(), "foreign key constraint") {
			// Verify patient still exists
			patientCheck, errCheck := s.patientRepo.FindByID(patientID)
			if errCheck != nil || patientCheck == nil {
				return nil, response.NewBadRequest("Patient not found. Please refresh and try again.")
			}
			return nil, response.NewInternalServerError("Failed to create handoff request. Please try again.")
		}
		
		return nil, fmt.Errorf("failed to create handoff: %w", err)
	}

	// 9. Get user details for notifications
	requestingUser, err := s.userRepo.GetByID(requestingClinicianID)
	if err != nil {
		s.log.Error("Failed to get requesting user", zap.Error(err))
		return nil, fmt.Errorf("failed to get requesting user: %w", err)
	}

	receivingUser, err := s.userRepo.GetByID(receivingClinicianID)
	if err != nil {
		s.log.Error("Failed to get receiving user", zap.Error(err))
		return nil, fmt.Errorf("failed to get receiving user: %w", err)
	}

	patientName := fmt.Sprintf("%s %s", patient.FirstName, patient.LastName)
	handoffURL := fmt.Sprintf("%s/dashboard/patients/%s", baseURL, patientID.String())

	// 10. Send email notification to receiving clinician
	if err := s.emailService.SendHandoffRequestEmail(
		receivingUser.Email,
		patientName,
		requestingUser.FullName,
		handoffURL,
	); err != nil {
		s.log.Warn("Failed to send handoff request email", zap.Error(err))
		// Don't fail the request if email fails
	}

	// 11. Create in-app notification for receiving clinician
	handoffEntityType := entity.RelatedEntityTypePatientHandoff
	if err := s.notificationRepo.Create(&entity.Notification{
		UserID:            receivingClinicianID,
		Type:              entity.TypeHandoffRequest,
		Title:             fmt.Sprintf("Patient Handoff Request: %s", patientName),
		Message:           fmt.Sprintf("%s has requested to hand off patient %s to you.", requestingUser.FullName, patientName),
		RelatedEntityType: &handoffEntityType,
		RelatedEntityID:   &handoff.ID,
		IsRead:            false,
	}); err != nil {
		s.log.Warn("Failed to create in-app notification", zap.Error(err))
		// Don't fail the request if notification creation fails
	}

	// 12. Build response
	response := &dto.HandoffResponse{
		ID:                     handoff.ID,
		PatientID:              handoff.PatientID,
		PatientName:            patientName,
		RequestingClinicianID:  handoff.RequestingClinicianID,
		RequestingClinicianName: requestingUser.FullName,
		RequestingClinicianEmail: requestingUser.Email,
		ReceivingClinicianID:   handoff.ReceivingClinicianID,
		ReceivingClinicianName: receivingUser.FullName,
		ReceivingClinicianEmail: receivingUser.Email,
		Status:                 handoff.Status,
		RequestedRole:          handoff.RequestedRole,
		Message:                handoff.Message,
		RequestedAt:            handoff.RequestedAt,
		RespondedAt:            handoff.RespondedAt,
		RespondedBy:            handoff.RespondedBy,
		CreatedAt:              handoff.CreatedAt,
		UpdatedAt:              handoff.UpdatedAt,
	}

	s.log.Info("Patient handoff requested", zap.String("handoff_id", handoff.ID.String()),
		zap.String("patient_id", patientID.String()),
		zap.String("requesting_clinician_id", requestingClinicianID.String()),
		zap.String("receiving_clinician_id", receivingClinicianID.String()))

	return response, nil
}

func (s *patientHandoffService) ApproveHandoff(
	ctx context.Context,
	handoffID, approvingClinicianID, organizationID uuid.UUID,
	reason *string,
	baseURL string,
) error {
	// 1. Get handoff
	handoff, err := s.handoffRepo.GetByID(handoffID)
	if err != nil {
		return fmt.Errorf("failed to get handoff: %w", err)
	}
	if handoff == nil {
		return response.ErrNotFound
	}

	// 2. Verify handoff belongs to organization
	patient, err := s.patientRepo.FindByID(handoff.PatientID)
	if err != nil {
		return response.ErrNotFound
	}
	if patient.OrganizationID != organizationID {
		return response.ErrNotFound
	}

	// 3. Verify handoff can be approved
	if !handoff.CanBeApproved() {
		return response.NewBadRequest("Handoff cannot be approved in its current state")
	}

	// 4. Verify approving clinician is the receiving clinician
	if handoff.ReceivingClinicianID != approvingClinicianID {
		return response.ErrForbidden
	}

	// 5. Get requesting clinician's current role
	assignments, err := s.patientRepo.GetAssignedClinicians(handoff.PatientID)
	if err != nil {
		return fmt.Errorf("failed to get assignments: %w", err)
	}

	var requestingRole string
	var requestingIsPrimary bool
	for _, assignment := range assignments {
		if assignment.ClinicianID == handoff.RequestingClinicianID {
			requestingRole = assignment.Role
			requestingIsPrimary = assignment.Role == "primary"
			break
		}
	}

	if requestingRole == "" {
		return response.NewBadRequest("Requesting clinician is no longer assigned to this patient")
	}

	// 6. Determine role for receiving clinician
	roleToAssign := handoff.RequestedRole
	if roleToAssign == nil || *roleToAssign == "" {
		roleToAssign = &requestingRole
	}

	// 7. Validate: Ensure at least one primary clinician remains after handoff
	primaryCount, err := s.patientRepo.CountPrimaryClinicians(handoff.PatientID)
	if err != nil {
		return fmt.Errorf("failed to count primary clinicians: %w", err)
	}

	if requestingIsPrimary && primaryCount <= 1 && *roleToAssign != "primary" {
		return response.NewBadRequest("Cannot approve handoff: would leave patient without a primary clinician")
	}

	// 8. Update assignments: Assign receiving first, then unassign requesting
	// This order ensures we don't violate the "last primary clinician" constraint
	// First assign receiving clinician
	if err := s.patientSvc.AssignClinician(ctx, handoff.PatientID, dto.AssignClinicianRequest{
		ClinicianID: handoff.ReceivingClinicianID,
		Role:        *roleToAssign,
	}, organizationID, approvingClinicianID); err != nil {
		s.log.Error("Failed to assign receiving clinician", zap.Error(err))
		return fmt.Errorf("failed to assign receiving clinician: %w", err)
	}

	// Then unassign requesting clinician (now safe because receiving is already assigned)
	if err := s.patientSvc.UnassignClinician(ctx, handoff.PatientID, handoff.RequestingClinicianID, organizationID, approvingClinicianID); err != nil {
		s.log.Error("Failed to unassign requesting clinician", zap.Error(err))
		// Try to rollback: unassign receiving clinician
		_ = s.patientSvc.UnassignClinician(ctx, handoff.PatientID, handoff.ReceivingClinicianID, organizationID, approvingClinicianID)
		return fmt.Errorf("failed to unassign requesting clinician: %w", err)
	}

	// 9. Update handoff status
	now := time.Now()
	handoff.Status = handoffEntity.StatusApproved
	handoff.RespondedAt = &now
	handoff.RespondedBy = &approvingClinicianID

	if err := s.handoffRepo.Update(handoff); err != nil {
		s.log.Error("Failed to update handoff status", zap.Error(err))
		return fmt.Errorf("failed to update handoff: %w", err)
	}

	// 10. Get user details for notifications
	requestingUser, err := s.userRepo.GetByID(handoff.RequestingClinicianID)
	if err != nil {
		s.log.Error("Failed to get requesting user", zap.Error(err))
		// Continue even if user fetch fails
		requestingUser = nil
	}

	receivingUser, err := s.userRepo.GetByID(handoff.ReceivingClinicianID)
	if err != nil {
		s.log.Error("Failed to get receiving user", zap.Error(err))
		receivingUser = nil
	}

	patientName := fmt.Sprintf("%s %s", patient.FirstName, patient.LastName)

	// 11. Send email notifications
	if requestingUser != nil {
		if err := s.emailService.SendHandoffApprovedEmail(
			requestingUser.Email,
			patientName,
			receivingUser.FullName,
		); err != nil {
			s.log.Warn("Failed to send handoff approved email to requesting clinician", zap.Error(err))
		}
	}

	// 12. Create in-app notifications
	handoffEntityType := entity.RelatedEntityTypePatientHandoff
	if requestingUser != nil && receivingUser != nil {
		if err := s.notificationRepo.Create(&entity.Notification{
			UserID:            handoff.RequestingClinicianID,
			Type:              entity.TypeHandoffApproved,
			Title:             fmt.Sprintf("Patient Handoff Approved: %s", patientName),
			Message:           fmt.Sprintf("%s has approved your handoff request for patient %s.", receivingUser.FullName, patientName),
			RelatedEntityType: &handoffEntityType,
			RelatedEntityID:   &handoff.ID,
			IsRead:            false,
		}); err != nil {
			s.log.Warn("Failed to create in-app notification for requesting clinician", zap.Error(err))
		}
	}

	s.log.Info("Patient handoff approved", zap.String("handoff_id", handoffID.String()),
		zap.String("patient_id", handoff.PatientID.String()))

	return nil
}

func (s *patientHandoffService) RejectHandoff(
	ctx context.Context,
	handoffID, rejectingClinicianID, organizationID uuid.UUID,
	reason *string,
	baseURL string,
) error {
	// 1. Get handoff
	handoff, err := s.handoffRepo.GetByID(handoffID)
	if err != nil {
		return fmt.Errorf("failed to get handoff: %w", err)
	}
	if handoff == nil {
		return response.ErrNotFound
	}

	// 2. Verify handoff belongs to organization
	patient, err := s.patientRepo.FindByID(handoff.PatientID)
	if err != nil {
		return response.ErrNotFound
	}
	if patient.OrganizationID != organizationID {
		return response.ErrNotFound
	}

	// 3. Verify handoff can be rejected
	if !handoff.CanBeRejected() {
		return response.NewBadRequest("Handoff cannot be rejected in its current state")
	}

	// 4. Verify rejecting clinician is the receiving clinician
	if handoff.ReceivingClinicianID != rejectingClinicianID {
		return response.ErrForbidden
	}

	// 5. Update handoff status
	now := time.Now()
	handoff.Status = handoffEntity.StatusRejected
	handoff.RespondedAt = &now
	handoff.RespondedBy = &rejectingClinicianID

	if err := s.handoffRepo.Update(handoff); err != nil {
		s.log.Error("Failed to update handoff status", zap.Error(err))
		return fmt.Errorf("failed to update handoff: %w", err)
	}

	// 6. Get user details for notifications
	requestingUser, err := s.userRepo.GetByID(handoff.RequestingClinicianID)
	if err != nil {
		s.log.Error("Failed to get requesting user", zap.Error(err))
		requestingUser = nil
	}

	rejectingUser, err := s.userRepo.GetByID(handoff.ReceivingClinicianID)
	if err != nil {
		s.log.Error("Failed to get rejecting user", zap.Error(err))
		rejectingUser = nil
	}

	patientName := fmt.Sprintf("%s %s", patient.FirstName, patient.LastName)
	reasonText := ""
	if reason != nil {
		reasonText = *reason
	}

	// 7. Send email notification
	if requestingUser != nil {
		if err := s.emailService.SendHandoffRejectedEmail(
			requestingUser.Email,
			patientName,
			rejectingUser.FullName,
			reasonText,
		); err != nil {
			s.log.Warn("Failed to send handoff rejected email", zap.Error(err))
		}
	}

	// 8. Create in-app notification
	handoffEntityType := entity.RelatedEntityTypePatientHandoff
	if requestingUser != nil {
		rejectionMessage := fmt.Sprintf("%s has rejected your handoff request for patient %s.", rejectingUser.FullName, patientName)
		if reasonText != "" {
			rejectionMessage += fmt.Sprintf(" Reason: %s", reasonText)
		}

		if err := s.notificationRepo.Create(&entity.Notification{
			UserID:            handoff.RequestingClinicianID,
			Type:              entity.TypeHandoffRejected,
			Title:             fmt.Sprintf("Patient Handoff Rejected: %s", patientName),
			Message:           rejectionMessage,
			RelatedEntityType: &handoffEntityType,
			RelatedEntityID:   &handoff.ID,
			IsRead:            false,
		}); err != nil {
			s.log.Warn("Failed to create in-app notification", zap.Error(err))
		}
	}

	s.log.Info("Patient handoff rejected", zap.String("handoff_id", handoffID.String()),
		zap.String("patient_id", handoff.PatientID.String()))

	return nil
}

func (s *patientHandoffService) CancelHandoff(
	ctx context.Context,
	handoffID, cancellingClinicianID, organizationID uuid.UUID,
) error {
	// 1. Get handoff
	handoff, err := s.handoffRepo.GetByID(handoffID)
	if err != nil {
		return fmt.Errorf("failed to get handoff: %w", err)
	}
	if handoff == nil {
		return response.ErrNotFound
	}

	// 2. Verify handoff belongs to organization
	patient, err := s.patientRepo.FindByID(handoff.PatientID)
	if err != nil {
		return response.ErrNotFound
	}
	if patient.OrganizationID != organizationID {
		return response.ErrNotFound
	}

	// 3. Verify handoff can be cancelled
	if !handoff.CanBeCancelled() {
		return response.NewBadRequest("Handoff cannot be cancelled in its current state")
	}

	// 4. Verify cancelling clinician is the requesting clinician
	if handoff.RequestingClinicianID != cancellingClinicianID {
		return response.ErrForbidden
	}

	// 5. Update handoff status
	now := time.Now()
	handoff.Status = handoffEntity.StatusCancelled
	handoff.RespondedAt = &now
	handoff.RespondedBy = &cancellingClinicianID

	if err := s.handoffRepo.Update(handoff); err != nil {
		s.log.Error("Failed to update handoff status", zap.Error(err))
		return fmt.Errorf("failed to update handoff: %w", err)
	}

	// 6. Create in-app notification for receiving clinician
	patientName := fmt.Sprintf("%s %s", patient.FirstName, patient.LastName)
	handoffEntityType := entity.RelatedEntityTypePatientHandoff
	if err := s.notificationRepo.Create(&entity.Notification{
		UserID:            handoff.ReceivingClinicianID,
		Type:              entity.TypeHandoffCancelled,
		Title:             fmt.Sprintf("Patient Handoff Cancelled: %s", patientName),
		Message:           fmt.Sprintf("The handoff request for patient %s has been cancelled.", patientName),
		RelatedEntityType: &handoffEntityType,
		RelatedEntityID:   &handoff.ID,
		IsRead:            false,
	}); err != nil {
		s.log.Warn("Failed to create in-app notification for receiving clinician", zap.Error(err))
	}

	s.log.Info("Patient handoff cancelled", zap.String("handoff_id", handoffID.String()),
		zap.String("patient_id", handoff.PatientID.String()))

	return nil
}

func (s *patientHandoffService) GetHandoff(ctx context.Context, handoffID, userID, organizationID uuid.UUID) (*dto.HandoffResponse, error) {
	handoff, err := s.handoffRepo.GetByID(handoffID)
	if err != nil {
		return nil, fmt.Errorf("failed to get handoff: %w", err)
	}
	if handoff == nil {
		return nil, response.ErrNotFound
	}

	// Verify handoff belongs to organization
	patient, err := s.patientRepo.FindByID(handoff.PatientID)
	if err != nil {
		return nil, response.ErrNotFound
	}
	if patient.OrganizationID != organizationID {
		return nil, response.ErrNotFound
	}

	// Verify user has access (must be requesting or receiving clinician, or admin/owner)
	// For now, allow if user is in the organization
	userOrgID, err := s.patientRepo.GetOrganizationID(userID)
	if err != nil || userOrgID != organizationID {
		return nil, response.ErrForbidden
	}

	return s.mapHandoffToResponse(handoff, patient)
}

func (s *patientHandoffService) ListHandoffs(ctx context.Context, patientID, userID, organizationID uuid.UUID) ([]dto.HandoffResponse, error) {
	// Verify patient belongs to organization
	patient, err := s.patientRepo.FindByID(patientID)
	if err != nil {
		return nil, response.ErrNotFound
	}
	if patient.OrganizationID != organizationID {
		return nil, response.ErrNotFound
	}

	handoffs, err := s.handoffRepo.GetByPatientID(patientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get handoffs: %w", err)
	}

	responses := make([]dto.HandoffResponse, len(handoffs))
	for i, handoff := range handoffs {
		resp, err := s.mapHandoffToResponse(&handoff, patient)
		if err != nil {
			s.log.Warn("Failed to map handoff to response", zap.Error(err))
			continue
		}
		responses[i] = *resp
	}

	return responses, nil
}

func (s *patientHandoffService) ListPendingHandoffs(ctx context.Context, clinicianID, organizationID uuid.UUID) ([]dto.HandoffResponse, error) {
	handoffs, err := s.handoffRepo.GetPendingByReceivingClinician(clinicianID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending handoffs: %w", err)
	}

	var responses []dto.HandoffResponse
	for _, handoff := range handoffs {
		// Verify handoff belongs to organization
		patient, err := s.patientRepo.FindByID(handoff.PatientID)
		if err != nil {
			continue
		}
		if patient.OrganizationID != organizationID {
			continue
		}

		resp, err := s.mapHandoffToResponse(&handoff, patient)
		if err != nil {
			s.log.Warn("Failed to map handoff to response", zap.Error(err))
			continue
		}
		responses = append(responses, *resp)
	}

	return responses, nil
}

func (s *patientHandoffService) mapHandoffToResponse(handoff *handoffEntity.PatientHandoff, patient *handoffEntity.Patient) (*dto.HandoffResponse, error) {
	requestingUser, err := s.userRepo.GetByID(handoff.RequestingClinicianID)
	if err != nil {
		return nil, fmt.Errorf("failed to get requesting user: %w", err)
	}

	receivingUser, err := s.userRepo.GetByID(handoff.ReceivingClinicianID)
	if err != nil {
		return nil, fmt.Errorf("failed to get receiving user: %w", err)
	}

	patientName := fmt.Sprintf("%s %s", patient.FirstName, patient.LastName)

	return &dto.HandoffResponse{
		ID:                     handoff.ID,
		PatientID:             handoff.PatientID,
		PatientName:            patientName,
		RequestingClinicianID:  handoff.RequestingClinicianID,
		RequestingClinicianName: requestingUser.FullName,
		RequestingClinicianEmail: requestingUser.Email,
		ReceivingClinicianID:   handoff.ReceivingClinicianID,
		ReceivingClinicianName: receivingUser.FullName,
		ReceivingClinicianEmail: receivingUser.Email,
		Status:                 handoff.Status,
		RequestedRole:          handoff.RequestedRole,
		Message:                handoff.Message,
		RequestedAt:            handoff.RequestedAt,
		RespondedAt:            handoff.RespondedAt,
		RespondedBy:            handoff.RespondedBy,
		CreatedAt:              handoff.CreatedAt,
		UpdatedAt:              handoff.UpdatedAt,
	}, nil
}

