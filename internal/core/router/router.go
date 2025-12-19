package router

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/sahabatharianmu/OpenMind/internal/core/middleware"
	appointmentHandler "github.com/sahabatharianmu/OpenMind/internal/modules/appointment/handler"
	auditLogHandler "github.com/sahabatharianmu/OpenMind/internal/modules/audit_log/handler"
	clinicalNoteHandler "github.com/sahabatharianmu/OpenMind/internal/modules/clinical_note/handler"
	exportHandler "github.com/sahabatharianmu/OpenMind/internal/modules/export/handler"
	importHandler "github.com/sahabatharianmu/OpenMind/internal/modules/import/handler"
	invoiceHandler "github.com/sahabatharianmu/OpenMind/internal/modules/invoice/handler"
	organizationHandler "github.com/sahabatharianmu/OpenMind/internal/modules/organization/handler"
	patientHandler "github.com/sahabatharianmu/OpenMind/internal/modules/patient/handler"
	"github.com/sahabatharianmu/OpenMind/internal/modules/user/handler"
	"github.com/sahabatharianmu/OpenMind/pkg/constants"
)

func RegisterRoutes(
	h *server.Hertz,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	patientHandler *patientHandler.PatientHandler,
	appointmentHandler *appointmentHandler.AppointmentHandler,
	clinicalNoteHandler *clinicalNoteHandler.ClinicalNoteHandler,
	invoiceHandler *invoiceHandler.InvoiceHandler,
	auditLogHandler *auditLogHandler.AuditLogHandler,
	organizationHandler *organizationHandler.OrganizationHandler,
	exportHandler *exportHandler.ExportHandler,
	importHandler *importHandler.ImportHandler,
	authMiddleware *middleware.AuthMiddleware,
	auditMiddleware *middleware.AuditMiddleware,
	rbacMiddleware *middleware.RBACMiddleware,
	tenantMiddleware app.HandlerFunc,
) {
	api := h.Group("/api")
	v1 := api.Group("/v1")

	auth := v1.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}
	protected := v1.Group("/")
	protected.Use(authMiddleware.Middleware())
	protected.Use(tenantMiddleware) // Set tenant context after authentication
	protected.Use(auditMiddleware.Middleware()) // Add audit logging
	{
		// User routes
		users := protected.Group("/users")
		{
			users.GET("/me", userHandler.GetProfile)
			users.PUT("/me", userHandler.UpdateProfile)
		}

		protected.PUT("/auth/password", userHandler.ChangePassword)

		organizations := protected.Group("/organizations")
		{
			organizations.GET("/me", organizationHandler.GetMyOrganization)
			organizations.PUT("/me", rbacMiddleware.HasRole(constants.RoleAdmin), organizationHandler.UpdateOrganization)
		}

		protected.GET("/export", rbacMiddleware.HasRole(constants.RoleAdmin), exportHandler.ExportData)

		imports := protected.Group("/import")
		imports.Use(rbacMiddleware.HasRole(constants.RoleAdmin))
		{
			imports.GET("/template/:type", importHandler.DownloadTemplate)
			imports.POST("/preview", importHandler.PreviewImport)
			imports.POST("/execute", importHandler.ExecuteImport)
		}

		patients := protected.Group("/patients")
		{
			patients.POST("", rbacMiddleware.HasRole(constants.RoleClinician), patientHandler.Create)
			patients.GET("", patientHandler.List)
			patients.GET("/:id", patientHandler.Get)
			patients.PUT("/:id", rbacMiddleware.HasRole(constants.RoleClinician), patientHandler.Update)
			patients.DELETE("/:id", rbacMiddleware.HasRole(constants.RoleAdmin), patientHandler.Delete)
		}

		appointments := protected.Group("/appointments")
		{
			appointments.POST("", rbacMiddleware.HasRole(constants.RoleClinician), appointmentHandler.Create)
			appointments.GET("", appointmentHandler.List)
			appointments.GET("/:id", appointmentHandler.Get)
			appointments.PUT("/:id", rbacMiddleware.HasRole(constants.RoleClinician), appointmentHandler.Update)
			appointments.DELETE("/:id", rbacMiddleware.HasRole(constants.RoleClinician), appointmentHandler.Delete)
		}

		clinicalNotes := protected.Group("/clinical-notes")
		clinicalNotes.Use(rbacMiddleware.HasRole(constants.RoleClinician))
		{
			clinicalNotes.POST("", clinicalNoteHandler.Create)
			clinicalNotes.GET("", clinicalNoteHandler.List)
			clinicalNotes.GET("/:id", clinicalNoteHandler.Get)
			clinicalNotes.PUT("/:id", clinicalNoteHandler.Update)
			clinicalNotes.DELETE("/:id", clinicalNoteHandler.Delete)
			clinicalNotes.POST("/:id/addendums", clinicalNoteHandler.AddAddendum)
			clinicalNotes.POST("/:id/attachments", clinicalNoteHandler.UploadAttachment)
			clinicalNotes.GET("/attachments/:attachment_id", clinicalNoteHandler.DownloadAttachment)
		}

		invoices := protected.Group("/invoices")
		{
			invoices.POST("", rbacMiddleware.HasRole(constants.RoleAdmin), invoiceHandler.Create)
			invoices.GET("", invoiceHandler.List)
			invoices.GET("/:id", invoiceHandler.Get)
			invoices.PUT("/:id", rbacMiddleware.HasRole(constants.RoleAdmin), invoiceHandler.Update)
			invoices.DELETE("/:id", rbacMiddleware.HasRole(constants.RoleAdmin), invoiceHandler.Delete)
			invoices.GET("/:id/superbill", invoiceHandler.DownloadSuperbill)
		}

		auditLogs := protected.Group("/audit-logs")
		auditLogs.Use(rbacMiddleware.HasRole(constants.RoleAdmin))
		{
			auditLogs.GET("", auditLogHandler.List)
		}
	}

	h.Static("/assets", "./web/dist")
	
	h.GET("/SahariIcon.svg", func(ctx context.Context, c *app.RequestContext) {
		c.Header("Content-Type", "image/svg+xml")
		c.File("./web/dist/SahariIcon.svg")
	})
	
	h.StaticFile("/favicon.ico", "./web/dist/favicon.ico")
	h.StaticFile("/robots.txt", "./web/dist/robots.txt")
	h.StaticFile("/placeholder.svg", "./web/dist/placeholder.svg")

	h.NoRoute(func(ctx context.Context, c *app.RequestContext) {
		path := string(c.Request.URI().Path())
		if len(path) >= 4 && path[:4] == "/api" {
			c.JSON(404, map[string]interface{}{"error": "Not Found"})
			return
		}

		c.File("./web/dist/index.html")
	})
}
