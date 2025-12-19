package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/clinical_note/dto"
	"github.com/sahabatharianmu/OpenMind/internal/modules/clinical_note/entity"
	"github.com/sahabatharianmu/OpenMind/internal/modules/clinical_note/repository"
	"github.com/sahabatharianmu/OpenMind/pkg/crypto"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
	"go.uber.org/zap"
)

type ClinicalNoteService interface {
	Create(
		ctx context.Context,
		req dto.CreateClinicalNoteRequest,
		organizationID uuid.UUID,
	) (*dto.ClinicalNoteResponse, error)
	Update(
		ctx context.Context,
		id uuid.UUID,
		organizationID uuid.UUID,
		req dto.UpdateClinicalNoteRequest,
	) (*dto.ClinicalNoteResponse, error)
	Delete(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error
	Get(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*dto.ClinicalNoteResponse, error)
	List(ctx context.Context, organizationID uuid.UUID, page, pageSize int) ([]dto.ClinicalNoteResponse, int64, error)
	AddAddendum(
		ctx context.Context,
		noteID uuid.UUID,
		organizationID uuid.UUID,
		req dto.AddAddendumRequest,
	) (*dto.AddendumResponse, error)
	UploadAttachment(
		ctx context.Context,
		noteID uuid.UUID,
		organizationID uuid.UUID,
		fileName string,
		contentType string,
		data []byte,
	) (*dto.AttachmentResponse, error)
	DownloadAttachment(
		ctx context.Context,
		attachmentID uuid.UUID,
		organizationID uuid.UUID,
	) (string, []byte, string, error)
	GetOrganizationID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
}

type clinicalNoteService struct {
	repo       repository.ClinicalNoteRepository
	encryptSvc *crypto.EncryptionService
	log        logger.Logger
}

func NewClinicalNoteService(
	repo repository.ClinicalNoteRepository,
	encryptSvc *crypto.EncryptionService,
	log logger.Logger,
) ClinicalNoteService {
	return &clinicalNoteService{
		repo:       repo,
		encryptSvc: encryptSvc,
		log:        log,
	}
}

type clinicalNoteContent struct {
	Subjective *string `json:"subjective,omitempty"`
	Objective  *string `json:"objective,omitempty"`
	Assessment *string `json:"assessment,omitempty"`
	Plan       *string `json:"plan,omitempty"`
}

func (s *clinicalNoteService) Create(
	ctx context.Context,
	req dto.CreateClinicalNoteRequest,
	organizationID uuid.UUID,
) (*dto.ClinicalNoteResponse, error) {
	var signedAt *time.Time
	if req.IsSigned {
		now := time.Now()
		signedAt = &now
	}

	note := &entity.ClinicalNote{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		PatientID:      req.PatientID,
		ClinicianID:    req.ClinicianID,
		AppointmentID:  req.AppointmentID,
		NoteType:       req.NoteType,
		Subjective:     req.Subjective,
		Objective:      req.Objective,
		Assessment:     req.Assessment,
		Plan:           req.Plan,
		IsSigned:       req.IsSigned,
		SignedAt:       signedAt,
	}

	if err := s.encryptNote(note, organizationID); err != nil {
		return nil, fmt.Errorf("failed to encrypt note: %w", err)
	}

	if err := s.repo.Create(note); err != nil {
		return nil, err
	}

	return s.mapEntityToResponse(note), nil
}

func (s *clinicalNoteService) Update(
	ctx context.Context,
	id uuid.UUID,
	organizationID uuid.UUID,
	req dto.UpdateClinicalNoteRequest,
) (*dto.ClinicalNoteResponse, error) {
	note, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if note.OrganizationID != organizationID {
		return nil, response.ErrNotFound
	}

	// Note Locking: Once "Signed", a note becomes immutable.
	if note.IsSigned {
		return nil, response.NewForbidden("Cannot update a signed clinical note")
	}

	if req.NoteType != "" {
		note.NoteType = req.NoteType
	}
	if req.Subjective != nil {
		note.Subjective = req.Subjective
	}
	if req.Objective != nil {
		note.Objective = req.Objective
	}
	if req.Assessment != nil {
		note.Assessment = req.Assessment
	}
	if req.Plan != nil {
		note.Plan = req.Plan
	}
	if req.IsSigned != nil {
		note.IsSigned = *req.IsSigned
		if *req.IsSigned && note.SignedAt == nil {
			now := time.Now()
			note.SignedAt = &now
		} else if !*req.IsSigned {
			note.SignedAt = nil
		}
	}

	if err := s.encryptNote(note, organizationID); err != nil {
		return nil, fmt.Errorf("failed to encrypt note: %w", err)
	}

	if err := s.repo.Update(note); err != nil {
		return nil, err
	}

	return s.mapEntityToResponse(note), nil
}

func (s *clinicalNoteService) Delete(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	note, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if note.OrganizationID != organizationID {
		return response.ErrNotFound
	}

	if note.IsSigned {
		return response.NewForbidden("Cannot delete a signed clinical note")
	}

	return s.repo.Delete(id)
}

func (s *clinicalNoteService) Get(
	ctx context.Context,
	id uuid.UUID,
	organizationID uuid.UUID,
) (*dto.ClinicalNoteResponse, error) {
	note, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if note.OrganizationID != organizationID {
		return nil, response.ErrNotFound
	}

	if err := s.decryptNote(note, organizationID); err != nil {
		return nil, fmt.Errorf("failed to decrypt note: %w", err)
	}

	return s.mapEntityToResponse(note), nil
}

func (s *clinicalNoteService) List(
	ctx context.Context,
	organizationID uuid.UUID,
	page, pageSize int,
) ([]dto.ClinicalNoteResponse, int64, error) {
	offset := (page - 1) * pageSize
	notes, total, err := s.repo.List(organizationID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	var responses []dto.ClinicalNoteResponse
	for i := range notes {
		if err := s.decryptNote(&notes[i], organizationID); err != nil {
			s.log.Error("Failed to decrypt note", zap.String("note_id", notes[i].ID.String()))
		}
		responses = append(responses, *s.mapEntityToResponse(&notes[i]))
	}

	return responses, total, nil
}

func (s *clinicalNoteService) AddAddendum(
	ctx context.Context,
	noteID uuid.UUID,
	organizationID uuid.UUID,
	req dto.AddAddendumRequest,
) (*dto.AddendumResponse, error) {
	note, err := s.repo.FindByID(noteID)
	if err != nil {
		return nil, err
	}

	if note.OrganizationID != organizationID {
		return nil, response.ErrNotFound
	}

	if !note.IsSigned {
		return nil, response.NewBadRequest("Cannot add addendum to an unsigned note")
	}

	addendum := &entity.Addendum{
		ID:          uuid.New(),
		NoteID:      noteID,
		ClinicianID: req.ClinicianID,
		Content:     req.Content,
	}

	if err := s.encryptAddendum(addendum, organizationID); err != nil {
		return nil, fmt.Errorf("failed to encrypt addendum: %w", err)
	}

	if err := s.repo.AddAddendum(addendum); err != nil {
		return nil, err
	}

	return s.mapAddendumEntityToResponse(addendum), nil
}

func (s *clinicalNoteService) UploadAttachment(
	ctx context.Context,
	noteID uuid.UUID,
	organizationID uuid.UUID,
	fileName string,
	contentType string,
	data []byte,
) (*dto.AttachmentResponse, error) {
	note, err := s.repo.FindByID(noteID)
	if err != nil {
		return nil, err
	}

	if note.OrganizationID != organizationID {
		return nil, response.ErrNotFound
	}

	if note.IsSigned {
		return nil, response.NewForbidden("Cannot add attachment to a signed note")
	}

	// Encrypt the file content with tenant-specific key (HIPAA compliant)
	encryptedBase64, err := s.encryptSvc.Encrypt(base64.StdEncoding.EncodeToString(data), organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt file: %w", err)
	}

	encryptedBytes, _ := base64.StdEncoding.DecodeString(encryptedBase64)
	const nonceSize = 12

	attachment := &entity.Attachment{
		ID:            uuid.New(),
		NoteID:        noteID,
		FileName:      fileName,
		ContentType:   contentType,
		Size:          int64(len(data)),
		DataEncrypted: encryptedBytes,
		Nonce:         encryptedBytes[:nonceSize],
	}

	if err := s.repo.AddAttachment(attachment); err != nil {
		return nil, err
	}

	return s.mapAttachmentEntityToResponse(attachment), nil
}

func (s *clinicalNoteService) DownloadAttachment(
	ctx context.Context,
	attachmentID uuid.UUID,
	organizationID uuid.UUID,
) (string, []byte, string, error) {
	attachment, err := s.repo.GetAttachmentByID(attachmentID)
	if err != nil {
		return "", nil, "", err
	}

	// Verify organization via the note
	note, err := s.repo.FindByID(attachment.NoteID)
	if err != nil {
		return "", nil, "", err
	}

	if note.OrganizationID != organizationID {
		return "", nil, "", response.ErrNotFound
	}

	// Decrypt the file content with tenant-specific key (HIPAA compliant)
	encryptedBase64 := base64.StdEncoding.EncodeToString(attachment.DataEncrypted)
	decryptedBase64, err := s.encryptSvc.Decrypt(encryptedBase64, organizationID)
	if err != nil {
		return "", nil, "", fmt.Errorf("failed to decrypt file: %w", err)
	}

	decryptedBytes, err := base64.StdEncoding.DecodeString(decryptedBase64)
	if err != nil {
		return "", nil, "", fmt.Errorf("failed to decode decrypted file: %w", err)
	}

	return attachment.FileName, decryptedBytes, attachment.ContentType, nil
}

func (s *clinicalNoteService) GetOrganizationID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	return s.repo.GetOrganizationID(userID)
}

func (s *clinicalNoteService) encryptAddendum(a *entity.Addendum, organizationID uuid.UUID) error {
	// Use tenant-specific encryption key (HIPAA compliant)
	encryptedBase64, err := s.encryptSvc.Encrypt(a.Content, organizationID)
	if err != nil {
		return err
	}

	encryptedBytes, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return err
	}

	const nonceSize = 12
	if len(encryptedBytes) < nonceSize {
		return fmt.Errorf("encrypted data too short")
	}

	a.ContentEncrypted = encryptedBytes
	a.Nonce = encryptedBytes[:nonceSize]

	return nil
}

func (s *clinicalNoteService) decryptAddendum(a *entity.Addendum, organizationID uuid.UUID) error {
	if len(a.ContentEncrypted) == 0 {
		return nil
	}

	encryptedBase64 := base64.StdEncoding.EncodeToString(a.ContentEncrypted)
	// Use tenant-specific decryption key (HIPAA compliant)
	decryptedContent, err := s.encryptSvc.Decrypt(encryptedBase64, organizationID)
	if err != nil {
		return err
	}

	a.Content = decryptedContent
	return nil
}

func (s *clinicalNoteService) mapAddendumEntityToResponse(a *entity.Addendum) *dto.AddendumResponse {
	return &dto.AddendumResponse{
		ID:          a.ID,
		ClinicianID: a.ClinicianID,
		Content:     a.Content,
		SignedAt:    a.SignedAt,
	}
}

func (s *clinicalNoteService) encryptNote(n *entity.ClinicalNote, organizationID uuid.UUID) error {
	content := clinicalNoteContent{
		Subjective: n.Subjective,
		Objective:  n.Objective,
		Assessment: n.Assessment,
		Plan:       n.Plan,
	}

	jsonData, err := json.Marshal(content)
	if err != nil {
		return err
	}

	// Use tenant-specific encryption key (HIPAA compliant)
	encryptedBase64, err := s.encryptSvc.Encrypt(string(jsonData), organizationID)
	if err != nil {
		return err
	}

	encryptedBytes, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return err
	}

	// In GCM, nonce size is 12 bytes
	const nonceSize = 12
	if len(encryptedBytes) < nonceSize {
		return fmt.Errorf("encrypted data too short")
	}

	n.ContentEncrypted = encryptedBytes
	n.Nonce = encryptedBytes[:nonceSize]
	n.KeyID = "v1"

	return nil
}

func (s *clinicalNoteService) decryptNote(n *entity.ClinicalNote, organizationID uuid.UUID) error {
	if len(n.ContentEncrypted) > 0 {
		encryptedBase64 := base64.StdEncoding.EncodeToString(n.ContentEncrypted)
		// Use tenant-specific decryption key (HIPAA compliant)
		decryptedJSON, err := s.encryptSvc.Decrypt(encryptedBase64, organizationID)
		if err != nil {
			return err
		}

		var content clinicalNoteContent
		if err := json.Unmarshal([]byte(decryptedJSON), &content); err != nil {
			return err
		}

		n.Subjective = content.Subjective
		n.Objective = content.Objective
		n.Assessment = content.Assessment
		n.Plan = content.Plan
	}

	for i := range n.Addendums {
		if err := s.decryptAddendum(&n.Addendums[i], n.OrganizationID); err != nil {
			s.log.Error("Failed to decrypt addendum", zap.String("addendum_id", n.Addendums[i].ID.String()))
		}
	}

	return nil
}

func (s *clinicalNoteService) mapAttachmentEntityToResponse(a *entity.Attachment) *dto.AttachmentResponse {
	return &dto.AttachmentResponse{
		ID:          a.ID,
		FileName:    a.FileName,
		ContentType: a.ContentType,
		Size:        a.Size,
		CreatedAt:   a.CreatedAt,
	}
}

func (s *clinicalNoteService) mapEntityToResponse(n *entity.ClinicalNote) *dto.ClinicalNoteResponse {
	var addendums []dto.AddendumResponse
	for _, a := range n.Addendums {
		addendums = append(addendums, *s.mapAddendumEntityToResponse(&a))
	}

	var attachments []dto.AttachmentResponse
	for _, a := range n.Attachments {
		attachments = append(attachments, *s.mapAttachmentEntityToResponse(&a))
	}

	return &dto.ClinicalNoteResponse{
		ID:             n.ID,
		OrganizationID: n.OrganizationID,
		PatientID:      n.PatientID,
		ClinicianID:    n.ClinicianID,
		AppointmentID:  n.AppointmentID,
		NoteType:       n.NoteType,
		Subjective:     n.Subjective,
		Objective:      n.Objective,
		Assessment:     n.Assessment,
		Plan:           n.Plan,
		IsSigned:       n.IsSigned,
		SignedAt:       n.SignedAt,
		Addendums:      addendums,
		Attachments:    attachments,
		CreatedAt:      n.CreatedAt,
		UpdatedAt:      n.UpdatedAt,
	}
}
