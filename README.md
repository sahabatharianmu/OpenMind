# OpenMind
Provide mental health professionals with a secure, sovereign, and affordable platform to manage their practice without trading patient privacy for convenience.

# **📂 OpenMind Project Structure**

This directory tree illustrates the **Modular Monolith** architecture.

**Key Highlights:**

1. **pkg/crypto/**: This is the "Vault." It contains the AES-256-GCM logic. It is isolated so it can be audited easily without touching the rest of the app.  
2. **internal/modules/**: This is where the feature logic lives. auth, clinical, and finance are separate folders, enforcing clean boundaries.  
3. **web/**: The React Frontend lives inside the same repo (Monorepo), simplifying versioning.

## **🌳 Root Directory**

```
openmind/  
├── .github/  
│   └── workflows/  
│       └── ci-cd.yml          \# The GitHub Action we designed  
├── cmd/  
│   └── server/  
│       └── main.go            \# Entry Point: Wires up Modules \+ Starts Hertz  
├── config/  
│   └── config.yaml            \# Local dev config (GitIgnored in prod)  
├── deploy/  
│   ├── docker-compose.yml     \# Self-hosting setup  
│   └── Dockerfile             \# Multi-stage build  
├── internal/                  \# 🔒 Private Application Code  
│   ├── core/                  \# Shared Kernel  
│   │   ├── database/          \# GORM connection & migration runner  
│   │   ├── middleware/        \# Hertz Middleware (Auth, CORS, Logging)  
│   │   └── eventbus/          \# RabbitMQ integration  
│   └── modules/               \# 📦 The Modular Monolith Domains  
│       ├── auth/              \# Login, Session, RBAC  
│       ├── clinical/          \# Patients, SOAP Notes  
│       │   ├── dto/           \# JSON Request/Response structs  
│       │   ├── entity/        \# GORM Database Models  
│       │   ├── handler/       \# Hertz HTTP Controllers  
│       │   ├── service/       \# Business Logic (Calls Crypto)  
│       │   └── repository/    \# Database Queries  
│       └── finance/           \# Invoicing, Superbills  
├── pkg/                       \# 🔓 Public/Shared Libraries  
│   ├── crypto/                \# 🛡️ THE ENCRYPTION ENGINE  
│   │   ├── vault.go           \# Encrypt() / Decrypt() logic  
│   │   └── vault\_test.go      \# Security Unit Tests  
│   └── pdf/                   \# Maroto PDF Generator wrappers  
├── web/                       \# ⚛️ React Frontend  
│   ├── public/  
│   ├── src/  
│   │   ├── api/               \# Axios/Fetch wrappers  
│   │   ├── components/        \# Shared UI (Buttons, Layouts)  
│   │   ├── features/          \# Feature-based folder structure  
│   │   │   ├── auth/          \# Login Forms, Context  
│   │   │   ├── clinical/      \# Note Editor, Patient List  
│   │   │   └── finance/       \# Invoice Viewer  
│   │   ├── lib/               \# 3rd party setup (TanStack Query, Mantine)  
│   │   └── main.tsx  
│   ├── package.json  
│   └── vite.config.ts  
├── go.mod                     \# Go Dependencies  
├── go.sum  
└── Makefile                   \# Shortcuts (make run, make test)
```

## **🔍 Deep Dive: Where the "Magic" Happens**

### **1\. The Encryption Engine (pkg/crypto/vault.go)**

This package has **zero dependencies** on the rest of the app. It does one thing: mathematically secure data.

package crypto

// Vault handles the AES-GCM encryption  
type Vault interface {  
    Encrypt(plaintext \[\]byte) (ciphertext \[\]byte, nonce \[\]byte, keyID string, err error)  
    Decrypt(ciphertext \[\]byte, nonce \[\]byte, keyID string) (plaintext \[\]byte, err error)  
}

### **2\. The Clinical Service (internal/modules/clinical/service/note\_service.go)**

This is where we **use** the encryption. Notice how the Service layer calls the Vault before asking the Repository to save.

func (s \*NoteService) CreateNote(ctx context.Context, content string) error {  
    // 1\. Encrypt the sensitive content  
    encryptedData, nonce, keyID, err := s.vault.Encrypt(\[\]byte(content))  
    if err \!= nil {  
        return err  
    }

    // 2\. Prepare the entity  
    note := entity.ClinicalNote{  
        ContentEncrypted: encryptedData, // Blob  
        Nonce:            nonce,         // Blob  
        KeyID:            keyID,         // String  
        // ...  
    }

    // 3\. Save to DB (DB never sees plain text)  
    return s.repo.Create(ctx, \&note)  
}

### **3\. The React Feature Folder (web/src/features/clinical/)**

We organize frontend code by **Feature**, not by technical type. This scales better than putting everything in components/.

web/src/features/clinical/  
├── components/  
│   ├── NoteEditor.tsx         \# The Rich Text Editor  
│   ├── PatientCard.tsx        \# Display component  
│   └── SOAPTemplate.tsx       \# The Form Layout  
├── hooks/  
│   ├── usePatient.ts          \# TanStack Query (GET /api/patients)  
│   └── useSaveNote.ts         \# TanStack Query Mutation (POST /api/notes)  
└── routes/  
    └── ClinicalRoutes.tsx     \# Route definitions  
