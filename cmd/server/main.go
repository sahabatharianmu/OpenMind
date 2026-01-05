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
	paymentHandler "github.com/sahabatharianmu/OpenMind/internal/modules/payment/handler"
	paymentRepository "github.com/sahabatharianmu/OpenMind/internal/modules/payment/repository"
	paymentService "github.com/sahabatharianmu/OpenMind/internal/modules/payment/service"
	subscriptionHandler "github.com/sahabatharianmu/OpenMind/internal/modules/subscription/handler"
	subscriptionRepository "github.com/sahabatharianmu/OpenMind/internal/modules/subscription/repository"
	subscriptionService "github.com/sahabatharianmu/OpenMind/internal/modules/subscription/service"
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
	"github.com/sahabatharianmu/OpenMind/pkg/midtrans"
	"github.com/sahabatharianmu/OpenMind/pkg/payment"
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

	ctx := context.Background()
	if err := database.RunMigrationsForAllTenants(ctx, db, tenantRepo, appLogger); err != nil {
		appLogger.Warn("Some tenant migrations failed", zap.Error(err))
	}
	tenantKeyRepo := tenantRepository.NewTenantEncryptionKeyRepository(db, appLogger)
	teamInvitationRepo := teamRepository.NewTeamInvitationRepository(db, appLogger)
	patientHandoffRepo := patientRepository.NewPatientHandoffRepository(db, appLogger)
	notificationRepo := notificationRepository.NewNotificationRepository(db, appLogger)

	jwtService := security.NewJWTService(cfg)
	passwordService := crypto.NewPasswordService(cfg)
	encryptService := crypto.NewEncryptionService(cfg)
	emailService := email.NewEmailService(cfg, appLogger)

	encryptService.SetTenantKeyRepository(tenantKeyRepo)

	tenantSvc := tenantService.NewTenantService(tenantRepo, db, appLogger)
	tenantSvc.SetEncryptionService(encryptService)
	tenantSvc.SetKeyRepository(tenantKeyRepo)

	authService := userService.NewAuthService(userRepo, organizationRepo, jwtService, passwordService, tenantSvc, emailService, appLogger)
	userSvc := userService.NewUserService(userRepo, organizationRepo, appLogger)

	usageSvc := subscriptionService.NewUsageService(patientRepo, organizationRepo, appLogger)
	gatingSvc := subscriptionService.NewFeatureGatingService(organizationRepo, usageSvc, appLogger, cfg.Application.URL)
	planRepo := subscriptionRepository.NewPlanRepository(db, appLogger)
	planSvc := subscriptionService.NewPlanService(planRepo, appLogger)

	patientSvc := patientService.NewPatientService(patientRepo, patientHandoffRepo, userRepo, gatingSvc, appLogger)
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

	organizationSvc := organizationService.NewOrganizationService(organizationRepo, userRepo, usageSvc, gatingSvc, appLogger)
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
		gatingSvc,
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

	paymentProviderManager, err := payment.NewPaymentProviderManager(&cfg.Payment, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize payment provider manager", zap.Error(err))
	}

	paymentMethodRepo := paymentRepository.NewPaymentMethodRepository(db, appLogger)
	paymentMethodSvc := paymentService.NewPaymentMethodService(
		paymentMethodRepo,
		paymentProviderManager,
		encryptService,
		cfg,
		appLogger,
	)

	midtransService, err := midtrans.NewMidtransService(&cfg.Payment.Midtrans, appLogger)
	if err != nil {
		appLogger.Warn("Failed to initialize Midtrans service", zap.Error(err))
		midtransService = nil
	}

	paymentTransactionRepo := paymentRepository.NewPaymentTransactionRepository(db, appLogger)
	var paymentTransactionSvc paymentService.PaymentTransactionService
	if midtransService != nil {
		paymentTransactionSvc = paymentService.NewPaymentTransactionService(
			paymentTransactionRepo,
			organizationRepo,
			userRepo,
			emailService,
			midtransService,
			tenantSvc,
			db,
			appLogger,
			cfg.Application.URL,
		)
	}

	authHandler := userHandler.NewAuthHandler(authService, cfg.Application.URL)
	userHdlr := userHandler.NewUserHandler(userSvc, authService)
	patientHdlr := patientHandler.NewPatientHandler(patientSvc)
	appointmentHdlr := handler.NewAppointmentHandler(appointmentSvc)
	clinicalNoteHdlr := clinicalNoteHandler.NewClinicalNoteHandler(clinicalNoteSvc)
	invoiceHdlr := invoiceHandler.NewInvoiceHandler(invoiceSvc)
	auditLogHdlr := auditLogHandler.NewAuditLogHandler(auditLogSvc)
	organizationHdlr := organizationHandler.NewOrganizationHandler(organizationSvc)
	paymentHdlr := paymentHandler.NewPaymentMethodHandler(paymentMethodSvc, organizationSvc)
	var paymentTransactionHdlr *paymentHandler.PaymentTransactionHandler
	if paymentTransactionSvc != nil {
		paymentTransactionHdlr = paymentHandler.NewPaymentTransactionHandler(paymentTransactionSvc)
	}
	exportHdlr := exportHandler.NewExportHandler(exportSvc)
	importHdlr := importHandler.NewImportHandler(importSvc)
	teamHdlr := teamHandler.NewTeamInvitationHandler(teamInvitationSvc)
	notificationHdlr := notificationHandler.NewNotificationHandler(notificationSvc)
	patientHandoffHdlr := patientHandler.NewPatientHandoffHandler(patientHandoffSvc, cfg.Application.URL)
	adminPlanHdlr := subscriptionHandler.NewAdminPlanHandler(planSvc)
	publicPlanHdlr := subscriptionHandler.NewPublicPlanHandler(planSvc)

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
		paymentHdlr,
		paymentTransactionHdlr,
		exportHdlr,
		importHdlr,
		teamHdlr,
		notificationHdlr,
		adminPlanHdlr,
		publicPlanHdlr,
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
