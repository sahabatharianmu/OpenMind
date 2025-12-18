package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/clinical_note/entity"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ClinicalNoteRepository interface {
	Create(note *entity.ClinicalNote) error
	Update(note *entity.ClinicalNote) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*entity.ClinicalNote, error)
	List(organizationID uuid.UUID, limit, offset int) ([]entity.ClinicalNote, int64, error)
	GetOrganizationID(userID uuid.UUID) (uuid.UUID, error)
}

type clinicalNoteRepository struct {
	db  *gorm.DB
	log logger.Logger
}

func NewClinicalNoteRepository(db *gorm.DB, log logger.Logger) ClinicalNoteRepository {
	return &clinicalNoteRepository{
		db:  db,
		log: log,
	}
}

func (r *clinicalNoteRepository) Create(note *entity.ClinicalNote) error {
	if err := r.db.Create(note).Error; err != nil {
		r.log.Error("Failed to create clinical note", zap.Error(err))
		return err
	}
	return nil
}

func (r *clinicalNoteRepository) Update(note *entity.ClinicalNote) error {
	if err := r.db.Save(note).Error; err != nil {
		r.log.Error("Failed to update clinical note", zap.Error(err), zap.String("id", note.ID.String()))
		return err
	}
	return nil
}

func (r *clinicalNoteRepository) Delete(id uuid.UUID) error {
	if err := r.db.Delete(&entity.ClinicalNote{}, "id = ?", id).Error; err != nil {
		r.log.Error("Failed to delete clinical note", zap.Error(err), zap.String("id", id.String()))
		return err
	}
	return nil
}

func (r *clinicalNoteRepository) FindByID(id uuid.UUID) (*entity.ClinicalNote, error) {
	var note entity.ClinicalNote
	if err := r.db.First(&note, "id = ?", id).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			r.log.Error("Failed to find clinical note", zap.Error(err), zap.String("id", id.String()))
		}
		return nil, err
	}
	return &note, nil
}

func (r *clinicalNoteRepository) List(
	organizationID uuid.UUID,
	limit, offset int,
) ([]entity.ClinicalNote, int64, error) {
	var notes []entity.ClinicalNote
	var total int64

	query := r.db.Model(&entity.ClinicalNote{}).Where("organization_id = ?", organizationID)

	if err := query.Count(&total).Error; err != nil {
		r.log.Error("Failed to count clinical notes", zap.Error(err))
		return nil, 0, err
	}

	if err := query.Limit(limit).Offset(offset).Order("created_at desc").Find(&notes).Error; err != nil {
		r.log.Error("Failed to list clinical notes", zap.Error(err))
		return nil, 0, err
	}

	return notes, total, nil
}

func (r *clinicalNoteRepository) GetOrganizationID(userID uuid.UUID) (uuid.UUID, error) {
	var orgIDStr string
	if err := r.db.Table("organization_members").Select("organization_id").Where("user_id = ?", userID).Limit(1).Scan(&orgIDStr).Error; err != nil {
		r.log.Error("Failed to get organization ID", zap.Error(err), zap.String("user_id", userID.String()))
		return uuid.Nil, err
	}
	if orgIDStr == "" {
		return uuid.Nil, errors.New("organization not found for user")
	}
	return uuid.Parse(orgIDStr)
}
