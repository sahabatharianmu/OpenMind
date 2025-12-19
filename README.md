# OpenMind 🧠

**Proprietary Clinic Management System (EMR) for Mental Health Professionals - Cloud Version**

[![CI](https://github.com/sahabatharianmu/OpenMind/actions/workflows/ci.yml/badge.svg)](https://github.com/sahabatharianmu/OpenMind/actions/workflows/ci.yml)

> **⚠️ NOTICE: This is PRIVATE, PROPRIETARY source code owned by Sahari (PT Sahabat Harianmu).**  
> This repository contains the cloud version of OpenMind Practice. Unauthorized access, copying, modification, or distribution is strictly prohibited.

OpenMind is a secure, sovereign platform designed to help therapists and clinics manage their practice without trading patient privacy for convenience. Built with a "Privacy First" architecture, all clinical notes are encrypted at the application layer using AES-256-GCM encryption.

**Copyright © 2025 Sahari (PT Sahabat Harianmu). All rights reserved.**

---

## ✨ Key Features

- **🔐 Privacy-First Clinical Notes**: AES-256-GCM encryption for all SOAP notes. Your data is encrypted _before_ it hits the database.
- **📅 Appointment Scheduling**: Drag-and-drop calendar for managing sessions.
- **👥 Patient Management**: Comprehensive patient profiles, history, and intake forms.
- **💰 Invoicing & Billing**: Generate invoices, track payments, and manage superbills.
- **⚡ Modern Performance**: Built with Go (Fiber/Hertz) and React for blazing fast interactions.
- **🐳 Cloud-Ready**: Containerized architecture for scalable cloud deployment.

## 🛠️ Tech Stack

- **Backend**: Go 1.25+ (Hertz Framework)
- **Frontend**: React 18, TypeScript, TailwindCSS, Vite
- **Database**: PostgreSQL 18+
- **Infrastructure**: Docker, Docker Compose

---

## 🚀 Cloud Deployment

This is the cloud version of OpenMind Practice, designed for production cloud environments.

### Prerequisites

- Docker & Docker Compose installed.

### Quick Start

1.  **Run with Docker Compose**:
    Create a `docker-compose.yml` file (or use the one in this repo):

    ```yaml
    version: "3.8"
    services:
      openmind:
        image: ghcr.io/sahabatharianmu/openmind:latest
        ports:
          - "8080:8080"
        environment:
          - OPENMIND_DATABASE_HOST=postgres
          - OPENMIND_DATABASE_USER=postgres
          - OPENMIND_DATABASE_PASSWORD=postgres
          - OPENMIND_DATABASE_DB_NAME=openmind
          - OPENMIND_SECURITY_JWT_SECRET_KEY=change-this-secret
        depends_on:
          - postgres

      postgres:
        image: postgres:18-alpine
        environment:
          - POSTGRES_USER=postgres
          - POSTGRES_PASSWORD=postgres
          - POSTGRES_DB=openmind
        volumes:
          - openmind_data:/var/lib/postgresql/data

    volumes:
      openmind_data:
    ```

2.  **Start the Server**:

    ```bash
    docker-compose up -d
    ```

3.  **Access the App**:
    Open your browser to `http://localhost:8080`.

---

## 💻 Development Setup

**For Authorized Developers Only**

1.  **Clone the Repository**:

    ```bash
    git clone <private-repo-url>
    cd OpenMind
    ```

2.  **Setup Environment**:

    ```bash
    cp .env.example .env
    ```

3.  **Start Services (DB)**:

    ```bash
    docker-compose up -d postgres
    ```

4.  **Run Backend**:

    ```bash
    go mod download
    go run cmd/server/main.go
    ```

5.  **Run Frontend**:
    ```bash
    cd web
    bun install
    bun run dev
    ```

---

## 📂 Project Structure

This project follows a **Modular Monolith** architecture.

```
openmind/
├── cmd/server/        # Entry Point
├── internal/
│   ├── modules/       # Domain Logic (Auth, Clinical, Finance)
│   ├── core/          # Shared Kernel (Router, DB, Middleware)
├── pkg/crypto/        # Encryption Engine (AES-256-GCM)
├── web/               # React Frontend (Vite)
└── deploy/            # Docker Configs
```

## 📄 License

This software is **PROPRIETARY** and **CONFIDENTIAL**. All rights reserved.

Copyright © 2025 **Sahari (PT Sahabat Harianmu)**, also known as **Sahabat Harianmu**.

This source code is private and proprietary. Unauthorized access, use, copying, modification, or distribution is strictly prohibited and may result in civil and criminal penalties.

For licensing inquiries, please contact:
- **Email**: contact@sahari.id
- **Website**: https://sahari.id

See the [LICENSE](LICENSE) file for complete terms and conditions.
