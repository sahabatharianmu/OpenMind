package main

import (
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/sahabatharianmu/OpenMind/config"
	"github.com/sahabatharianmu/OpenMind/internal/core/database"
	"github.com/sahabatharianmu/OpenMind/internal/core/middleware"
	"github.com/sahabatharianmu/OpenMind/internal/core/router"
	"github.com/sahabatharianmu/OpenMind/internal/modules/appointment/handler"
	"github.com/sahabatharianmu/OpenMind/internal/modules/appointment/repository"
	"github.com/sahabatharianmu/OpenMind/internal/modules/appointment/service"
	auditLogHandler "github.com/sahabatharianmu/OpenMind/internal/modules/audit_log/handler"
	auditLogRepository "github.com/sahabatharianmu/OpenMind/internal/modules/audit_log/repository"
	auditLogService "github.com/sahabatharianmu/OpenMind/internal/modules/audit_log/service"
	clinicalNoteHandler "github.com/sahabatharianmu/OpenMind/internal/modules/clinical_note/handler"
	clinicalNoteRepository "github.com/sahabatharianmu/OpenMind/internal/modules/clinical_note/repository"
	clinicalNoteService "github.com/sahabatharianmu/OpenMind/internal/modules/clinical_note/service"
	exportHandler "github.com/sahabatharianmu/OpenMind/internal/modules/export/handler"
	exportService "github.com/sahabatharianmu/OpenMind/internal/modules/export/service"
	invoiceHandler "github.com/sahabatharianmu/OpenMind/internal/modules/invoice/handler"
	invoiceRepository "github.com/sahabatharianmu/OpenMind/internal/modules/invoice/repository"
	invoiceService "github.com/sahabatharianmu/OpenMind/internal/modules/invoice/service"
	organizationHandler "github.com/sahabatharianmu/OpenMind/internal/modules/organization/handler"
	organizationRepository "github.com/sahabatharianmu/OpenMind/internal/modules/organization/repository"
	organizationService "github.com/sahabatharianmu/OpenMind/internal/modules/organization/service"
	patientHandler "github.com/sahabatharianmu/OpenMind/internal/modules/patient/handler"
	patientRepository "github.com/sahabatharianmu/OpenMind/internal/modules/patient/repository"
	patientService "github.com/sahabatharianmu/OpenMind/internal/modules/patient/service"
	userHandler "github.com/sahabatharianmu/OpenMind/internal/modules/user/handler"
	userRepository "github.com/sahabatharianmu/OpenMind/internal/modules/user/repository"
	userService "github.com/sahabatharianmu/OpenMind/internal/modules/user/service"
	"github.com/sahabatharianmu/OpenMind/pkg/crypto"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"github.com/sahabatharianmu/OpenMind/pkg/security"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	appLogger := logger.NewLogger(cfg.Server.Mode)
	appLogger.Info(
		"Starting OpenMind server",
		zap.String("version", cfg.Application.Version),
		zap.String("environment", cfg.Application.Environment),
	)

	database.InitDB(cfg, appLogger)
	db := database.GetDB()

	if err := database.RunMigrations(db, appLogger); err != nil {
		appLogger.Fatal("Failed to run database migrations", zap.Error(err))
	}

	userRepo := userRepository.NewUserRepository(db, appLogger)
	patientRepo := patientRepository.NewPatientRepository(db, appLogger)
	appointmentRepo := repository.NewAppointmentRepository(db, appLogger)
	clinicalNoteRepo := clinicalNoteRepository.NewClinicalNoteRepository(db, appLogger)
	invoiceRepo := invoiceRepository.NewInvoiceRepository(db, appLogger)
	auditLogRepo := auditLogRepository.NewAuditLogRepository(db, appLogger)
	organizationRepo := organizationRepository.NewOrganizationRepository(db, appLogger)

	jwtService := security.NewJWTService(cfg)
	passwordService := crypto.NewPasswordService(cfg)
	encryptService := crypto.NewEncryptionService(cfg)

	authService := userService.NewAuthService(userRepo, jwtService, passwordService, appLogger)
	userSvc := userService.NewUserService(userRepo, appLogger)
	patientSvc := patientService.NewPatientService(patientRepo, appLogger)
	appointmentSvc := service.NewAppointmentService(appointmentRepo, appLogger)
	clinicalNoteSvc := clinicalNoteService.NewClinicalNoteService(clinicalNoteRepo, encryptService, appLogger)
	invoiceSvc := invoiceService.NewInvoiceService(
		invoiceRepo,
		organizationRepo,
		patientRepo,
		appointmentRepo,
		clinicalNoteRepo,
		appLogger,
	)
	auditLogSvc := auditLogService.NewAuditLogService(auditLogRepo, appLogger)
	organizationSvc := organizationService.NewOrganizationService(organizationRepo, appLogger)
	exportSvc := exportService.NewExportService(
		organizationRepo,
		patientRepo,
		appointmentRepo,
		clinicalNoteSvc,
		invoiceRepo,
		auditLogSvc,
		appLogger,
	)

	authHandler := userHandler.NewAuthHandler(authService)
	userHdlr := userHandler.NewUserHandler(userSvc, authService)
	patientHdlr := patientHandler.NewPatientHandler(patientSvc)
	appointmentHdlr := handler.NewAppointmentHandler(appointmentSvc)
	clinicalNoteHdlr := clinicalNoteHandler.NewClinicalNoteHandler(clinicalNoteSvc)
	invoiceHdlr := invoiceHandler.NewInvoiceHandler(invoiceSvc)
	auditLogHdlr := auditLogHandler.NewAuditLogHandler(auditLogSvc)
	organizationHdlr := organizationHandler.NewOrganizationHandler(organizationSvc)
	exportHdlr := exportHandler.NewExportHandler(exportSvc)

	authMiddleware := middleware.NewAuthMiddleware(jwtService)
	auditMiddleware := middleware.NewAuditMiddleware(auditLogSvc)
	rbacMiddleware := middleware.NewRBACMiddleware()

	h := server.New(
		server.WithHostPorts(fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)),
		server.WithReadTimeout(cfg.Server.ReadTimeout),
		server.WithWriteTimeout(cfg.Server.WriteTimeout),
		server.WithIdleTimeout(cfg.Server.IdleTimeout),
		server.WithExitWaitTime(cfg.Server.ExitTimeout),
	)

	router.RegisterRoutes(
		h,
		authHandler,
		userHdlr,
		patientHdlr,
		appointmentHdlr,
		clinicalNoteHdlr,
		invoiceHdlr,
		auditLogHdlr,
		organizationHdlr,
		exportHdlr,
		authMiddleware,
		auditMiddleware,
		rbacMiddleware,
	)

	h.OnShutdown = append(h.OnShutdown, func(_ context.Context) {
		appLogger.Info("Shutting down server gracefully...")

		// TODO: Add other cleanup logic here (e.g., closing Database connections, Redis, etc.)

		appLogger.Info("Server resources released")
	})

	appLogger.Info("Server starting", zap.String("host", cfg.Server.Host), zap.Int("port", cfg.Server.Port))
	h.Spin()

	appLogger.Info("Server exited")
}
