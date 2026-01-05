package router

import (
	"context"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/sahabatharianmu/OpenMind/internal/core/middleware"
	appointmentHandler "github.com/sahabatharianmu/OpenMind/internal/modules/appointment/handler"
	auditLogHandler "github.com/sahabatharianmu/OpenMind/internal/modules/audit_log/handler"
	clinicalNoteHandler "github.com/sahabatharianmu/OpenMind/internal/modules/clinical_note/handler"
	exportHandler "github.com/sahabatharianmu/OpenMind/internal/modules/export/handler"
	importHandler "github.com/sahabatharianmu/OpenMind/internal/modules/import/handler"
	invoiceHandler "github.com/sahabatharianmu/OpenMind/internal/modules/invoice/handler"
	notificationHandler "github.com/sahabatharianmu/OpenMind/internal/modules/notification/handler"
	organizationHandler "github.com/sahabatharianmu/OpenMind/internal/modules/organization/handler"
	patientHandler "github.com/sahabatharianmu/OpenMind/internal/modules/patient/handler"
	paymentHandler "github.com/sahabatharianmu/OpenMind/internal/modules/payment/handler"
	paymentTransactionHandler "github.com/sahabatharianmu/OpenMind/internal/modules/payment/handler"
	subscriptionHandler "github.com/sahabatharianmu/OpenMind/internal/modules/subscription/handler"
	teamHandler "github.com/sahabatharianmu/OpenMind/internal/modules/team/handler"
	"github.com/sahabatharianmu/OpenMind/internal/modules/user/handler"
	"github.com/sahabatharianmu/OpenMind/pkg/constants"
)

func RegisterRoutes(
	h *server.Hertz,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	patientHandler *patientHandler.PatientHandler,
	patientHandoffHandler *patientHandler.PatientHandoffHandler,
	appointmentHandler *appointmentHandler.AppointmentHandler,
	clinicalNoteHandler *clinicalNoteHandler.ClinicalNoteHandler,
	invoiceHandler *invoiceHandler.InvoiceHandler,
	auditLogHandler *auditLogHandler.AuditLogHandler,
	organizationHandler *organizationHandler.OrganizationHandler,
	paymentHandler *paymentHandler.PaymentMethodHandler,
	paymentTransactionHandler *paymentTransactionHandler.PaymentTransactionHandler,
	exportHandler *exportHandler.ExportHandler,
	importHandler *importHandler.ImportHandler,
	teamHandler *teamHandler.TeamInvitationHandler,
	notificationHandler *notificationHandler.NotificationHandler,
	adminPlanHandler *subscriptionHandler.AdminPlanHandler,
	publicPlanHandler *subscriptionHandler.PublicPlanHandler,
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

	plans := v1.Group("/plans")
	{
		plans.GET("", publicPlanHandler.ListActivePlans)
	}

	admin := v1.Group("/admin")
	admin.Use(authMiddleware.Middleware())
	admin.Use(middleware.RequireSystemRole("admin"))
	{
		plans := admin.Group("/plans")
		{
			plans.POST("", adminPlanHandler.CreatePlan)
			plans.GET("", adminPlanHandler.ListPlans)
			plans.GET("/:id", adminPlanHandler.GetPlan)
			plans.PUT("/:id", adminPlanHandler.UpdatePlan)
		}
	}

	team := v1.Group("/team")
	{
		team.GET("/invitations/:token", teamHandler.GetInvitation)
		team.POST("/invitations/register", teamHandler.RegisterAndAcceptInvitation)
	}
	protected := v1.Group("/")
	protected.Use(authMiddleware.Middleware())
	protected.Use(tenantMiddleware)
	protected.Use(auditMiddleware.Middleware())
	{
		users := protected.Group("/users")
		{
			users.GET("/me", userHandler.GetProfile)
			users.PUT("/me", userHandler.UpdateProfile)
		}

		protected.PUT("/auth/password", userHandler.ChangePassword)

		organizations := protected.Group("/organizations")
		{
			organizations.GET("/me", organizationHandler.GetMyOrganization)
			organizations.PUT(
				"/me",
				rbacMiddleware.HasRole(constants.RoleAdmin, constants.RoleOwner),
				organizationHandler.UpdateOrganization,
			)
			organizations.GET(
				"/me/members",
				rbacMiddleware.HasRole(constants.RoleAdmin, constants.RoleOwner),
				organizationHandler.ListTeamMembers,
			)
			organizations.PUT(
				"/me/members/:user_id/role",
				rbacMiddleware.HasRole(constants.RoleAdmin, constants.RoleOwner),
				organizationHandler.UpdateMemberRole,
			)
			organizations.DELETE(
				"/me/members/:user_id",
				rbacMiddleware.HasRole(constants.RoleAdmin, constants.RoleOwner),
				organizationHandler.RemoveMember,
			)
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
			patients.POST("/:id/assign", patientHandler.AssignClinician)
			patients.DELETE("/:id/assign/:clinician_id", patientHandler.UnassignClinician)
			patients.GET("/:id/assignments", patientHandler.GetAssignedClinicians)
			patients.POST(
				"/:id/handoff",
				rbacMiddleware.HasRole(constants.RoleClinician, constants.RoleAdmin, constants.RoleOwner),
				patientHandoffHandler.RequestHandoff,
			)
			patients.GET("/:id/handoffs", patientHandoffHandler.ListHandoffs)
		}

		handoffs := protected.Group("/patients/handoffs")
		{
			handoffs.GET("/pending", patientHandoffHandler.ListPendingHandoffs)
			handoffs.GET("/:id", patientHandoffHandler.GetHandoff)
			handoffs.POST(
				"/:id/approve",
				rbacMiddleware.HasRole(constants.RoleClinician, constants.RoleAdmin, constants.RoleOwner),
				patientHandoffHandler.ApproveHandoff,
			)
			handoffs.POST(
				"/:id/reject",
				rbacMiddleware.HasRole(constants.RoleClinician, constants.RoleAdmin, constants.RoleOwner),
				patientHandoffHandler.RejectHandoff,
			)
			handoffs.POST(
				"/:id/cancel",
				rbacMiddleware.HasRole(constants.RoleClinician, constants.RoleAdmin, constants.RoleOwner),
				patientHandoffHandler.CancelHandoff,
			)
		}

		notifications := protected.Group("/notifications")
		{
			notifications.GET("", notificationHandler.GetNotifications)
			notifications.PUT("/:id/read", notificationHandler.MarkAsRead)
			notifications.PUT("/read-all", notificationHandler.MarkAllAsRead)
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

		// Team management routes (admin/owner only)
		team := protected.Group("/team")
		team.Use(rbacMiddleware.HasRole(constants.RoleAdmin, constants.RoleOwner))
		{
			team.POST("/invitations", teamHandler.SendInvitation)
			team.GET("/invitations", teamHandler.ListInvitations)
			team.DELETE("/invitations/:id", teamHandler.CancelInvitation)
			team.POST("/invitations/:id/resend", teamHandler.ResendInvitation)
		}

		// Payment method routes (admin/owner only)
		paymentMethods := protected.Group("/payment-methods")
		paymentMethods.Use(rbacMiddleware.HasRole(constants.RoleAdmin, constants.RoleOwner))
		{
			paymentMethods.POST("", paymentHandler.CreatePaymentMethod)
			paymentMethods.GET("", paymentHandler.ListPaymentMethods)
			paymentMethods.GET("/:id", paymentHandler.GetPaymentMethod)
			paymentMethods.DELETE("/:id", paymentHandler.DeletePaymentMethod)
			paymentMethods.PUT("/:id/default", paymentHandler.SetDefaultPaymentMethod)
		}

		// Payment transaction routes (owner/admin only)
		// Only register if payment transaction handler is initialized (Midtrans configured)
		if paymentTransactionHandler != nil {
			payments := protected.Group("/payments")
			payments.Use(rbacMiddleware.HasRole(constants.RoleAdmin, constants.RoleOwner))
			{
				// QRIS payment routes
				payments.POST("/qris/create", paymentTransactionHandler.CreateQRISPayment)
				payments.GET("/qris/status/:id", paymentTransactionHandler.CheckPaymentStatus)
			}
		}
	}

	if paymentTransactionHandler != nil {
		webhooks := v1.Group("/webhooks")
		{
			midtrans := webhooks.Group("/midtrans")
			// QRIS payment webhook
			midtrans.POST("/v1.0/qr/qr-mpm-notify", paymentTransactionHandler.HandleQRISWebhook)
			// TODO: Add handlers for other webhook types when implemented
			// midtrans.POST("/v1.0/debit/notify", paymentTransactionHandler.HandleDebitWebhook)
			// midtrans.POST("/v1.0/transfer-va/payment", paymentTransactionHandler.HandleVAWebhook)
		}
	}

	h.Static("/assets", "./web/dist")

	h.NoRoute(func(ctx context.Context, c *app.RequestContext) {
		path := string(c.Request.URI().Path())

		if strings.HasPrefix(path, "/api") {
			c.JSON(404, map[string]interface{}{"error": "Not Found"})
			return
		}

		publicPath := filepath.Join("./web/public", path)
		if info, err := os.Stat(publicPath); err == nil && !info.IsDir() {
			ext := filepath.Ext(publicPath)
			contentType := mime.TypeByExtension(ext)

			if ext == ".svg" && contentType == "" {
				contentType = "image/svg+xml"
			}

			if contentType != "" {
				c.Header("Content-Type", contentType)
			}

			c.File(publicPath)
			return
		}

		c.File("./web/dist/index.html")
	})
}
