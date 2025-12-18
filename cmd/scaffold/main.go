package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func main() {
	moduleName := flag.String("name", "", "Name of the module to create")
	flag.Parse()

	if *moduleName == "" {
		log.Println("Please provide a module name using -name flag")
		os.Exit(1)
	}

	if err := run(*moduleName); err != nil {
		log.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func run(moduleName string) error {
	name := strings.ToLower(moduleName)
	basePath := filepath.Join("internal", "modules", name)

	dirs := []string{
		"dto",
		"entity",
		"handler",
		"repository",
		"service",
	}

	for _, dir := range dirs {
		path := filepath.Join(basePath, dir)
		// G301: Expect directory permissions to be 0750 or less
		if err := os.MkdirAll(path, 0750); err != nil {
			return fmt.Errorf("error creating directory %s: %w", path, err)
		}
	}

	// Generate boilerplate files
	if err := createFile(filepath.Join(basePath, "dto", "dto.go"), dtoTemplate, name); err != nil {
		return err
	}
	if err := createFile(filepath.Join(basePath, "entity", name+".go"), entityTemplate, name); err != nil {
		return err
	}
	if err := createFile(filepath.Join(basePath, "handler", "handler.go"), handlerTemplate, name); err != nil {
		return err
	}
	if err := createFile(filepath.Join(basePath, "repository", "repository.go"), repositoryTemplate, name); err != nil {
		return err
	}
	if err := createFile(filepath.Join(basePath, "service", "service.go"), serviceTemplate, name); err != nil {
		return err
	}

	log.Printf("Module %s created successfully at %s\n", name, basePath)
	return nil
}

func createFile(path string, tmpl string, name string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error creating file %s: %w", path, err)
	}
	defer f.Close()

	t := template.Must(template.New("file").Parse(tmpl))
	data := struct {
		Name      string
		TitleName string
	}{
		Name:      name,
		TitleName: cases.Title(language.English).String(name),
	}

	if err := t.Execute(f, data); err != nil {
		return fmt.Errorf("error writing to file %s: %w", path, err)
	}
	return nil
}

const dtoTemplate = `package dto

type Create{{.TitleName}}Request struct {
}

type {{.TitleName}}Response struct {
}
`

const entityTemplate = `package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type {{.TitleName}} struct {
	ID        uuid.UUID      ` + "`" + `gorm:"primaryKey" json:"id"` + "`" + `
	CreatedAt time.Time      ` + "`" + `json:"created_at"` + "`" + `
	UpdatedAt time.Time      ` + "`" + `json:"updated_at"` + "`" + `
	DeletedAt gorm.DeletedAt ` + "`" + `gorm:"index" json:"-"` + "`" + `
}
`

const handlerTemplate = `package handler

import (
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/sahabatharianmu/OpenMind/internal/modules/{{.Name}}/service"
)

type {{.TitleName}}Handler struct {
	svc service.{{.TitleName}}Service
}

func New{{.TitleName}}Handler(svc service.{{.TitleName}}Service) *{{.TitleName}}Handler {
	return &{{.TitleName}}Handler{svc: svc}
}
`

const repositoryTemplate = `package repository

import (
	"github.com/sahabatharianmu/OpenMind/internal/modules/{{.Name}}/entity"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"gorm.io/gorm"
)

type {{.TitleName}}Repository interface {
	Create(e *entity.{{.TitleName}}) error
}

type {{.Name}}Repository struct {
	db  *gorm.DB
	log logger.Logger
}

func New{{.TitleName}}Repository(db *gorm.DB, log logger.Logger) {{.TitleName}}Repository {
	return &{{.Name}}Repository{
		db:  db,
		log: log,
	}
}

func (r *{{.Name}}Repository) Create(e *entity.{{.TitleName}}) error {
	return r.db.Create(e).Error
}
`

const serviceTemplate = `package service

import (
	"github.com/sahabatharianmu/OpenMind/internal/modules/{{.Name}}/repository"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
)

type {{.TitleName}}Service interface {
}

type {{.Name}}Service struct {
	repo repository.{{.TitleName}}Repository
	log  logger.Logger
}

func New{{.TitleName}}Service(repo repository.{{.TitleName}}Repository, log logger.Logger) {{.TitleName}}Service {
	return &{{.Name}}Service{
		repo: repo,
		log:  log,
	}
}
`
