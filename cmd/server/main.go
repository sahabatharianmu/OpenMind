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
	importHandler "github.com/sahabatharianmu/OpenMind/internal/modules/import/handler"
	importService "github.com/sahabatharianmu/OpenMind/internal/modules/import/service"
	invoiceHandler "github.com/sahabatharianmu/OpenMind/internal/modules/invoice/handler"
	invoiceRepository "github.com/sahabatharianmu/OpenMind/internal/modules/invoice/repository"
	invoiceService "github.com/sahabatharianmu/OpenMind/internal/modules/invoice/service"
	notificationHandler "github.com/sahabatharianmu/OpenMind/internal/modules/notification/handler"
	notificationRepository "github.com/sahabatharianmu/OpenMind/internal/modules/notification/repository"
	notificationService "github.com/sahabatharianmu/OpenMind/internal/modules/notification/service"
	organizationHandler "github.com/sahabatharianmu/OpenMind/internal/modules/organization/handler"
	organizationRepository "github.com/sahabatharianmu/OpenMind/internal/modules/organization/repository"
	organizationService "github.com/sahabatharianmu/OpenMind/internal/modules/organization/service"
	patientHandler "github.com/sahabatharianmu/OpenMind/internal/modules/patient/handler"
	patientRepository "github.com/sahabatharianmu/OpenMind/internal/modules/patient/repository"
	patientService "github.com/sahabatharianmu/OpenMind/internal/modules/patient/service"
	teamHandler "github.com/sahabatharianmu/OpenMind/internal/modules/team/handler"
	teamRepository "github.com/sahabatharianmu/OpenMind/internal/modules/team/repository"
	teamService "github.com/sahabatharianmu/OpenMind/internal/modules/team/service"
	tenantRepository "github.com/sahabatharianmu/OpenMind/internal/modules/tenant/repository"
	tenantService "github.com/sahabatharianmu/OpenMind/internal/modules/tenant/service"
	userHandler "github.com/sahabatharianmu/OpenMind/internal/modules/user/handler"
	userRepository "github.com/sahabatharianmu/OpenMind/internal/modules/user/repository"
	userService "github.com/sahabatharianmu/OpenMind/internal/modules/user/service"
	"github.com/sahabatharianmu/OpenMind/pkg/crypto"
	"github.com/sahabatharianmu/OpenMind/pkg/email"
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
	tenantRepo := tenantRepository.NewTenantRepository(db, appLogger)
	tenantKeyRepo := tenantRepository.NewTenantEncryptionKeyRepository(db, appLogger)
	teamInvitationRepo := teamRepository.NewTeamInvitationRepository(db, appLogger)
	patientHandoffRepo := patientRepository.NewPatientHandoffRepository(db, appLogger)
	notificationRepo := notificationRepository.NewNotificationRepository(db, appLogger)

	jwtService := security.NewJWTService(cfg)
	passwordService := crypto.NewPasswordService(cfg)
	encryptService := crypto.NewEncryptionService(cfg)
	emailService := email.NewEmailService(cfg, appLogger)

	// Set tenant key repository for encryption service (HIPAA compliant)
	encryptService.SetTenantKeyRepository(tenantKeyRepo)

	// Initialize tenant service first (needed for auth service)
	tenantSvc := tenantService.NewTenantService(tenantRepo, db, appLogger)

	// Set encryption service and key repository for tenant key generation (HIPAA compliant)
	tenantSvc.SetEncryptionService(encryptService)
	tenantSvc.SetKeyRepository(tenantKeyRepo)

	// Generate encryption keys for existing tenants (one-time operation)
	// This ensures all existing tenants have encryption keys for HIPAA compliance
	ctx := context.Background()
	if err := tenantSvc.GenerateKeysForExistingTenants(ctx); err != nil {
		appLogger.Warn("Failed to generate keys for existing tenants", zap.Error(err))
		// Don't fail startup, but log the warning
		// Keys will be generated on-demand when tenants are accessed
	}

	authService := userService.NewAuthService(userRepo, organizationRepo, jwtService, passwordService, tenantSvc, emailService, appLogger)
	userSvc := userService.NewUserService(userRepo, organizationRepo, appLogger)
	patientSvc := patientService.NewPatientService(patientRepo, patientHandoffRepo, userRepo, appLogger)
	appointmentSvc := service.NewAppointmentService(appointmentRepo, patientRepo, userRepo, appLogger)
	clinicalNoteSvc := clinicalNoteService.NewClinicalNoteService(clinicalNoteRepo, patientRepo, encryptService, appLogger)
	invoiceSvc := invoiceService.NewInvoiceService(
		invoiceRepo,
		organizationRepo,
		patientRepo,
		appointmentRepo,
		clinicalNoteRepo,
		appLogger,
	)
	auditLogSvc := auditLogService.NewAuditLogService(auditLogRepo, appLogger)
	organizationSvc := organizationService.NewOrganizationService(organizationRepo, userRepo, appLogger)
	exportSvc := exportService.NewExportService(
		organizationRepo,
		patientRepo,
		appointmentRepo,
		clinicalNoteSvc,
		invoiceRepo,
		auditLogSvc,
		appLogger,
	)
	importSvc := importService.NewImportService(
		patientRepo,
		clinicalNoteRepo,
		patientSvc,
		clinicalNoteSvc,
		encryptService,
		db,
		appLogger,
	)

	teamInvitationSvc := teamService.NewTeamInvitationService(
		teamInvitationRepo,
		organizationRepo,
		userRepo,
		passwordService,
		emailService,
		appLogger,
		cfg.Application.URL,
	)

	notificationSvc := notificationService.NewNotificationService(notificationRepo, appLogger)

	patientHandoffSvc := patientService.NewPatientHandoffService(
		patientHandoffRepo,
		patientRepo,
		patientSvc,
		userRepo,
		emailService,
		notificationRepo,
		appLogger,
	)

	authHandler := userHandler.NewAuthHandler(authService, cfg.Application.URL)
	userHdlr := userHandler.NewUserHandler(userSvc, authService)
	patientHdlr := patientHandler.NewPatientHandler(patientSvc)
	appointmentHdlr := handler.NewAppointmentHandler(appointmentSvc)
	clinicalNoteHdlr := clinicalNoteHandler.NewClinicalNoteHandler(clinicalNoteSvc)
	invoiceHdlr := invoiceHandler.NewInvoiceHandler(invoiceSvc)
	auditLogHdlr := auditLogHandler.NewAuditLogHandler(auditLogSvc)
	organizationHdlr := organizationHandler.NewOrganizationHandler(organizationSvc)
	exportHdlr := exportHandler.NewExportHandler(exportSvc)
	importHdlr := importHandler.NewImportHandler(importSvc)
	teamHdlr := teamHandler.NewTeamInvitationHandler(teamInvitationSvc)
	notificationHdlr := notificationHandler.NewNotificationHandler(notificationSvc)
	patientHandoffHdlr := patientHandler.NewPatientHandoffHandler(patientHandoffSvc, cfg.Application.URL)

	authMiddleware := middleware.NewAuthMiddleware(jwtService)
	auditMiddleware := middleware.NewAuditMiddleware(auditLogSvc)
	rbacMiddleware := middleware.NewRBACMiddleware()
	tenantMiddleware := middleware.TenantContextMiddleware(tenantSvc, organizationRepo, appLogger)

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
		patientHandoffHdlr,
		appointmentHdlr,
		clinicalNoteHdlr,
		invoiceHdlr,
		auditLogHdlr,
		organizationHdlr,
		exportHdlr,
		importHdlr,
		teamHdlr,
		notificationHdlr,
		authMiddleware,
		auditMiddleware,
		rbacMiddleware,
		tenantMiddleware,
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
