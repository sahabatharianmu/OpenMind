package router

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/sahabatharianmu/OpenMind/internal/core/middleware"
	appointmentHandler "github.com/sahabatharianmu/OpenMind/internal/modules/appointment/handler"
	clinicalNoteHandler "github.com/sahabatharianmu/OpenMind/internal/modules/clinical_note/handler"
	invoiceHandler "github.com/sahabatharianmu/OpenMind/internal/modules/invoice/handler"
	patientHandler "github.com/sahabatharianmu/OpenMind/internal/modules/patient/handler"
	"github.com/sahabatharianmu/OpenMind/internal/modules/user/handler"
)

func RegisterRoutes(h *server.Hertz, authHandler *handler.AuthHandler, patientHandler *patientHandler.PatientHandler, appointmentHandler *appointmentHandler.AppointmentHandler, clinicalNoteHandler *clinicalNoteHandler.ClinicalNoteHandler, invoiceHandler *invoiceHandler.InvoiceHandler, authMiddleware *middleware.AuthMiddleware) {
	api := h.Group("/api")
	v1 := api.Group("/v1")

	auth := v1.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}
	protected := v1.Group("/")
	protected.Use(authMiddleware.Middleware())
	{
		patients := protected.Group("/patients")
		{
			patients.POST("", patientHandler.Create)
			patients.GET("", patientHandler.List)
			patients.GET("/:id", patientHandler.Get)
			patients.PUT("/:id", patientHandler.Update)
			patients.DELETE("/:id", patientHandler.Delete)
		}

		appointments := protected.Group("/appointments")
		{
			appointments.POST("", appointmentHandler.Create)
			appointments.GET("", appointmentHandler.List)
			appointments.GET("/:id", appointmentHandler.Get)
			appointments.PUT("/:id", appointmentHandler.Update)
			appointments.DELETE("/:id", appointmentHandler.Delete)
		}

		clinicalNotes := protected.Group("/clinical-notes")
		{
			clinicalNotes.POST("", clinicalNoteHandler.Create)
			clinicalNotes.GET("", clinicalNoteHandler.List)
			clinicalNotes.GET("/:id", clinicalNoteHandler.Get)
			clinicalNotes.PUT("/:id", clinicalNoteHandler.Update)
			clinicalNotes.DELETE("/:id", clinicalNoteHandler.Delete)
		}

		invoices := protected.Group("/invoices")
		{
			invoices.POST("", invoiceHandler.Create)
			invoices.GET("", invoiceHandler.List)
			invoices.GET("/:id", invoiceHandler.Get)
			invoices.PUT("/:id", invoiceHandler.Update)
			invoices.DELETE("/:id", invoiceHandler.Delete)
		}
	}

	h.Static("/assets", "./web/dist")
	h.StaticFile("/favicon.ico", "./web/dist/favicon.ico")

	h.NoRoute(func(ctx context.Context, c *app.RequestContext) {
		path := string(c.Request.URI().Path())
		if len(path) >= 4 && path[:4] == "/api" {
			c.JSON(404, map[string]interface{}{"error": "Not Found"})
			return
		}
		c.File("./web/dist/index.html")
	})
}
