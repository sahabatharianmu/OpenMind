# OpenMind 🧠

**Open Source Clinic Management System (EMR) for Mental Health Professionals.**

[![CI](https://github.com/sahabatharianmu/OpenMind/actions/workflows/ci.yml/badge.svg)](https://github.com/sahabatharianmu/OpenMind/actions/workflows/ci.yml)
[![Docker Image Version (latest semver)](https://img.shields.io/docker/v/sahabatharianmu/openmind?label=docker)](https://hub.docker.com/r/sahabatharianmu/openmind)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)

OpenMind is a secure, sovereign, and affordable platform designed to help therapists and clinics manage their practice without trading patient privacy for convenience. Built with a "Privacy First" architecture, all clinical notes are encrypted at the application layer.

---

## ✨ Key Features

- **🔐 Privacy-First Clinical Notes**: AES-256-GCM encryption for all SOAP notes. Your data is encrypted _before_ it hits the database.
- **📅 Appointment Scheduling**: Drag-and-drop calendar for managing sessions.
- **busts Patient Management**: Comprehensive patient profiles, history, and intake forms.
- **💰 Invoicing & Billing**: Generate invoices, track payments, and manage superbills.
- **⚡ Modern Performance**: Built with Go (Fiber/Hertz) and React for blazing fast interactions.
- **🐳 Self-Hostable**: Single Docker container for easy deployment anywhere.

## 🛠️ Tech Stack

- **Backend**: Go 1.25+ (Hertz Framework)
- **Frontend**: React 18, TypeScript, TailwindCSS, Vite
- **Database**: PostgreSQL 18+
- **Infrastructure**: Docker, Docker Compose

---

## 🚀 Getting Started (Self-Hosting)

You can run OpenMind on any server with Docker installed (VPS, Raspberry Pi, Home Lab).

### Prerequisites

- Docker & Docker Compose installed.

### Quick Start

1.  **Run with Docker Compose**:
    Create a `docker-compose.yml` file (or use the one in this repo):

    ```yaml
    version: "3.8"
    services:
      openmind:
        image: sahabatharianmu/openmind:latest
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

## 💻 Local Development

If you want to contribute or modify the code:

1.  **Clone the Repo**:

    ```bash
    git clone https://github.com/sahabatharianmu/OpenMind.git
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

## 🤝 Contributing

Contributions are welcome! Please check out the [Issues](https://github.com/sahabatharianmu/OpenMind/issues) tab.

## 📄 License

This project is licensed under the **GNU Affero General Public License v3.0 (AGPL-3.0)** - see the [LICENSE](LICENSE) file for details.
