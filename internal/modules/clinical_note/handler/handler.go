package handler

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/clinical_note/dto"
	"github.com/sahabatharianmu/OpenMind/internal/modules/clinical_note/service"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
)

type ClinicalNoteHandler struct {
	svc service.ClinicalNoteService
}

func NewClinicalNoteHandler(svc service.ClinicalNoteService) *ClinicalNoteHandler {
	return &ClinicalNoteHandler{svc: svc}
}

func (h *ClinicalNoteHandler) Create(_ context.Context, c *app.RequestContext) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	userID := userIDVal.(uuid.UUID)

	orgID, err := h.svc.GetOrganizationID(context.Background(), userID)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve organization")
		return
	}

	var req dto.CreateClinicalNoteRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	// Ensure clinician_id is set to the current user if not provided or valid?
	// For now, trust the input but maybe validate it belongs to org?
	// The requirement says "trust me bro" for now, so let's stick to simple.

	resp, err := h.svc.Create(context.Background(), req, orgID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Created(c, resp, "Clinical note created successfully")
}

func (h *ClinicalNoteHandler) List(_ context.Context, c *app.RequestContext) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	userID := userIDVal.(uuid.UUID)

	orgID, err := h.svc.GetOrganizationID(context.Background(), userID)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve organization")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	resp, total, err := h.svc.List(context.Background(), orgID, page, pageSize)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Clinical notes retrieved successfully", map[string]interface{}{
		"items":     resp,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}))
}

func (h *ClinicalNoteHandler) Get(_ context.Context, c *app.RequestContext) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	userID := userIDVal.(uuid.UUID)

	orgID, err := h.svc.GetOrganizationID(context.Background(), userID)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve organization")
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid clinical note ID", nil)
		return
	}

	resp, err := h.svc.Get(context.Background(), id, orgID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Clinical note retrieved successfully", resp))
}

func (h *ClinicalNoteHandler) Update(_ context.Context, c *app.RequestContext) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	userID := userIDVal.(uuid.UUID)

	orgID, err := h.svc.GetOrganizationID(context.Background(), userID)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve organization")
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid clinical note ID", nil)
		return
	}

	var req dto.UpdateClinicalNoteRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	resp, err := h.svc.Update(context.Background(), id, orgID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Clinical note updated successfully", resp))
}

func (h *ClinicalNoteHandler) Delete(_ context.Context, c *app.RequestContext) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	userID := userIDVal.(uuid.UUID)

	orgID, err := h.svc.GetOrganizationID(context.Background(), userID)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve organization")
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid clinical note ID", nil)
		return
	}

	if err := h.svc.Delete(context.Background(), id, orgID); err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusOK, response.Success("Clinical note deleted successfully", nil))
}

func (h *ClinicalNoteHandler) AddAddendum(_ context.Context, c *app.RequestContext) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	userID := userIDVal.(uuid.UUID)

	orgID, err := h.svc.GetOrganizationID(context.Background(), userID)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve organization")
		return
	}

	idStr := c.Param("id")
	noteID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid clinical note ID", nil)
		return
	}

	var req dto.AddAddendumRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	// Override clinician_id with current user for security
	req.ClinicianID = userID

	resp, err := h.svc.AddAddendum(context.Background(), noteID, orgID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusCreated, response.Success("Addendum added successfully", resp))
}

func (h *ClinicalNoteHandler) UploadAttachment(_ context.Context, c *app.RequestContext) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	userID := userIDVal.(uuid.UUID)

	orgID, err := h.svc.GetOrganizationID(context.Background(), userID)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve organization")
		return
	}

	idStr := c.Param("id")
	noteID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid clinical note ID", nil)
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "No file uploaded", nil)
		return
	}

	f, err := file.Open()
	if err != nil {
		response.InternalServerError(c, "Failed to open file")
		return
	}
	defer f.Close()

	// Read file data into memory
	data := make([]byte, file.Size)
	if _, err := f.Read(data); err != nil {
		response.InternalServerError(c, "Failed to read file")
		return
	}

	resp, err := h.svc.UploadAttachment(
		context.Background(),
		noteID,
		orgID,
		file.Filename,
		file.Header.Get("Content-Type"),
		data,
	)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(consts.StatusCreated, response.Success("Attachment uploaded successfully", resp))
}

func (h *ClinicalNoteHandler) DownloadAttachment(_ context.Context, c *app.RequestContext) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	userID := userIDVal.(uuid.UUID)

	orgID, err := h.svc.GetOrganizationID(context.Background(), userID)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve organization")
		return
	}

	attachmentIDStr := c.Param("attachment_id")
	attachmentID, err := uuid.Parse(attachmentIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid attachment ID", nil)
		return
	}

	fileName, data, contentType, err := h.svc.DownloadAttachment(context.Background(), attachmentID, orgID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Write(data)
}
