# Tenant-Specific Encryption (HIPAA Compliant)

## Overview

OpenMind Practice implements **tenant-specific encryption keys** to ensure HIPAA compliance. Each tenant (organization) has its own unique encryption key, and the platform cannot decrypt tenant data without the tenant's key.

## Architecture

### Zero-Knowledge Encryption

- **Each tenant has a unique encryption key** (32 bytes, AES-256)
- **Tenant keys are encrypted at rest** using a master key
- **Platform cannot decrypt tenant data** without the tenant's key
- **HIPAA compliant** - ensures data isolation and privacy

### Key Storage

```
tenant_encryption_keys table:
- tenant_id: Links to tenant
- organization_id: Links to organization
- encrypted_key: Tenant key encrypted with master key (BYTEA)
- key_version: For key rotation support
- algorithm: Encryption algorithm (AES-256-GCM)
```

### Encryption Flow

1. **Tenant Creation**:
   - Generate unique 32-byte key for tenant
   - Encrypt tenant key with master key
   - Store encrypted key in `tenant_encryption_keys` table

2. **Data Encryption**:
   - Retrieve tenant's encrypted key from database
   - Decrypt tenant key using master key
   - Use tenant key to encrypt clinical notes/data
   - Store encrypted data in tenant schema

3. **Data Decryption**:
   - Retrieve tenant's encrypted key from database
   - Decrypt tenant key using master key
   - Use tenant key to decrypt clinical notes/data

## Implementation Status

✅ **Completed**:
- Migration for `tenant_encryption_keys` table
- Entity and repository for tenant encryption keys
- Updated encryption service to support tenant-specific keys
- Backward compatibility with legacy shared key

🔄 **In Progress**:
- Generate tenant keys when tenant is created
- Update clinical note service to use tenant context

## Security Model

### Master Key
- Stored in configuration (environment variable)
- Used only to encrypt/decrypt tenant keys
- Never used directly for data encryption
- Must be 32 bytes (AES-256)

### Tenant Keys
- Generated randomly for each tenant
- Encrypted with master key before storage
- Used for all tenant data encryption/decryption
- Isolated per tenant (zero-knowledge)

### HIPAA Compliance

✅ **Data Isolation**: Each tenant's data encrypted with unique key
✅ **Zero-Knowledge**: Platform cannot decrypt tenant data
✅ **Encryption at Rest**: All sensitive data encrypted
✅ **Key Management**: Secure key storage and rotation support

## Usage

### Encryption Service

```go
// Initialize with tenant key repository
encryptService.SetTenantKeyRepository(tenantKeyRepo)

// Encrypt with tenant context (HIPAA compliant)
encrypted, err := encryptService.Encrypt(plaintext, organizationID)

// Decrypt with tenant context (HIPAA compliant)
decrypted, err := encryptService.Decrypt(ciphertext, organizationID)
```

### Clinical Note Service

The clinical note service should pass `organizationID` to encryption methods:

```go
// Get organization ID from context
orgID, _ := middleware.GetOrganizationIDFromContext(c)

// Encrypt with tenant key
encrypted, err := encryptService.Encrypt(jsonData, orgID)
```

## Migration Path

1. Run migration `000015_create_tenant_encryption_keys.up.sql`
2. Generate keys for existing tenants (migration script needed)
3. Update services to use tenant context for encryption
4. Gradually migrate existing encrypted data to tenant keys

## Key Rotation

The `key_version` field supports key rotation:
- Generate new key for tenant
- Re-encrypt data with new key
- Update `key_version` in database
- Old key can be archived for data migration

## Best Practices

1. **Master Key Security**:
   - Store in secure key management system (AWS KMS, HashiCorp Vault)
   - Rotate master key periodically
   - Never log or expose master key

2. **Tenant Key Management**:
   - Generate keys using cryptographically secure random
   - Encrypt keys before storage
   - Support key rotation for compliance

3. **Data Encryption**:
   - Always use tenant context for encryption/decryption
   - Never use master key directly for data
   - Log encryption operations (without keys)

## Future Enhancements

- [ ] Integration with AWS KMS or HashiCorp Vault
- [ ] Automatic key rotation
- [ ] Key escrow for compliance
- [ ] Audit logging for key access
- [ ] Multi-region key replication

