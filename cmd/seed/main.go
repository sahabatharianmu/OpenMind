package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/sahabatharianmu/OpenMind/config"
	"github.com/sahabatharianmu/OpenMind/internal/core/database"
	"github.com/sahabatharianmu/OpenMind/internal/modules/user/entity"
	"github.com/sahabatharianmu/OpenMind/pkg/crypto"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	email := flag.String("email", "", "Admin email address (required)")
	password := flag.String("password", "", "Admin password (required)")
	name := flag.String("name", "System Admin", "Admin full name")
	flag.Parse()

	if *email == "" || *password == "" {
		fmt.Println("Usage: go run ./cmd/seed --email admin@example.com --password YourPass123! [--name \"Admin Name\"]")
		os.Exit(1)
	}

	// Load config and init logger
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}
	appLogger := logger.NewLogger(cfg.Application.Environment)
	appLogger.Info("Starting admin seed...")

	// Connect to database
	database.InitDB(cfg, appLogger)
	db := database.GetDB()

	// Run public migrations first
	if err := database.RunPublicMigrations(db, appLogger); err != nil {
		appLogger.Fatal("Failed to run migrations", zap.Error(err))
	}

	// Check if an admin already exists
	var count int64
	db.Model(&entity.User{}).Where("system_role = ?", "admin").Count(&count)
	if count > 0 {
		appLogger.Info("Admin user already exists, skipping seed")
		fmt.Println("✓ Admin user already exists — nothing to do.")
		return
	}

	// Hash the password
	passwordService := crypto.NewPasswordService(cfg)
	if err := passwordService.ValidatePassword(*password); err != nil {
		appLogger.Fatal("Password validation failed", zap.Error(err))
	}

	hashedPassword, err := passwordService.HashPassword(*password)
	if err != nil {
		appLogger.Fatal("Failed to hash password", zap.Error(err))
	}

	// Create the admin user
	admin := &entity.User{
		ID:           uuid.New(),
		Email:        *email,
		PasswordHash: hashedPassword,
		FullName:     *name,
		SystemRole:   "admin",
	}

	if err := db.Create(admin).Error; err != nil {
		appLogger.Fatal("Failed to create admin user", zap.Error(err))
	}

	appLogger.Info("Admin user created successfully",
		zap.String("email", admin.Email),
		zap.String("id", admin.ID.String()))

	fmt.Printf("✓ Admin user created: %s (%s)\n", admin.Email, admin.ID.String())
}
