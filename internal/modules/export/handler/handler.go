package handler

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/internal/modules/export/service"
	"github.com/sahabatharianmu/OpenMind/pkg/response"
)

type ExportHandler struct {
	svc service.ExportService
}

func NewExportHandler(svc service.ExportService) *ExportHandler {
	return &ExportHandler{svc: svc}
}

func (h *ExportHandler) ExportData(_ context.Context, c *app.RequestContext) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	userID := userIDVal.(uuid.UUID)

	// Get all data files
	files, err := h.svc.ExportAllData(userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	// Create ZIP archive in memory
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	for filename, content := range files {
		writer, err := zipWriter.Create(filename)
		if err != nil {
			response.InternalServerError(c, "Failed to create ZIP file")
			return
		}
		_, _ = writer.Write(content)
	}

	if err := zipWriter.Close(); err != nil {
		response.InternalServerError(c, "Failed to finalize ZIP file")
		return
	}

	// Set headers for file download
	c.Response.Header.Set("Content-Type", "application/zip")
	c.Response.Header.Set(
		"Content-Disposition",
		fmt.Sprintf("attachment; filename=openmind-export-%s.zip", userID.String()[:8]),
	)
	c.Response.SetStatusCode(consts.StatusOK)

	// Write ZIP to response
	c.Response.SetBody(buf.Bytes())
}
