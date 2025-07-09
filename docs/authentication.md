# IUDX Sandbox Backend - Authentication Guide

## Overview

The IUDX Sandbox Backend implements a **dual authentication system** designed for different types of API consumers:

1. **JWT + RBAC Authentication** - For user-facing APIs (frontend applications)
2. **Static API Key Authentication** - For service-to-service APIs (data onboarding)

---

## 🔐 **JWT Authentication (User-Facing APIs)**

### Used For:
- All notebook operations (`/api/notebooks/*`)
- Data reading operations (`/api/datasets`, `/api/resources/*`, `/api/referenceresources/*`)
- Metadata operations (`/api/tags`, `/api/domains`)

### Headers Required:
```http
Authorization: Bearer <jwt_token>
```

### Roles Supported:
- **`admin`** - Full system access
- **`data-provider`** - Can read all data, manage own resources
- **`consumer`** - Can read data, manage own notebooks

### Example Usage:
```bash
curl -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..." \
     -X GET https://api.example.com/api/datasets
```

---

## 🔑 **Static API Key Authentication (Service-to-Service APIs)**

### Used For:
- **Data onboarding endpoints** (not used by frontend):
  - `POST /api/dataset` - Onboard new dataset
  - `POST /api/resource` - Onboard new resource  
  - `POST /api/referenceresource` - Onboard new reference resource

### Supported Header Formats:

#### Option 1: X-API-Key Header (Recommended)
```http
X-API-Key: your_static_api_key_here
```

#### Option 2: Authorization Header with ApiKey Scheme
```http
Authorization: ApiKey your_static_api_key_here
```

#### Option 3: API-Key Header (Alternative)
```http
API-Key: your_static_api_key_here
```

### Example Usage:
```bash
# Onboard a new dataset
curl -H "X-API-Key: your_static_api_key_here" \
     -H "Content-Type: application/json" \
     -X POST https://api.example.com/api/dataset \
     -d '{"name": "Sample Dataset", "description": "..."}'

# Onboard a new resource
curl -H "Authorization: ApiKey your_static_api_key_here" \
     -H "Content-Type: application/json" \
     -X POST https://api.example.com/api/resource \
     -d '{"name": "Sample Resource", "dataset": "..."}'
```

---

## ⚙️ **Configuration**

### Required Environment Variables:

```bash
# JWT Authentication
KEYCLOAK_PUBLIC_KEY=your_keycloak_public_key_here

# Static API Key Authentication  
STATIC_API_KEY=your_static_api_key_here_minimum_32_characters_required
```

### Generate Secure API Key:
```bash
# Generate a 32-byte random key (64 hex characters)
openssl rand -hex 32

# Example output:
# 4a5e8f2b9c1d3e7f8a2b4c6d8e9f1a3b5c7d9e1f2a4b6c8d0e2f4a6b8c0d2e4f
```

---

## 🛡️ **Security Features**

### JWT Authentication:
- ✅ **RSA signature verification** using Keycloak public key
- ✅ **Token expiration validation** 
- ✅ **Role-based access control (RBAC)**
- ✅ **Issuer and audience validation**
- ✅ **Structured error responses**

### Static API Key Authentication:
- ✅ **Environment variable configuration** (no hardcoded keys)
- ✅ **Multiple header format support** for flexibility
- ✅ **Request logging** for audit purposes
- ✅ **Constant-time comparison** to prevent timing attacks
- ✅ **Remote IP logging** for security monitoring

---

## 📊 **API Endpoint Authentication Matrix**

| Endpoint | Method | Authentication | Required Role | Purpose |
|----------|--------|---------------|---------------|---------|
| `/health/*` | GET | None | - | Health monitoring |
| `/api/notebooks/*` | All | JWT | User | Notebook management |
| `/api/datasets` | GET | JWT | Consumer+ | List datasets |
| `/api/dataset/:id` | GET | JWT | Consumer+ | Get specific dataset |
| `/api/dataset` | POST | **Static API Key** | - | **Onboard dataset** |
| `/api/resources/:id` | GET | JWT | Consumer+ | List resources |
| `/api/resource` | POST | **Static API Key** | - | **Onboard resource** |
| `/api/referenceresources/:id` | GET | JWT | Consumer+ | List reference resources |
| `/api/referenceresource` | POST | **Static API Key** | - | **Onboard reference resource** |
| `/api/tags` | GET | JWT | User | List available tags |
| `/api/domains` | GET | JWT | User | List available domains |

---

## 🚨 **Error Responses**

### JWT Authentication Errors:
```json
{
  "error": {
    "code": "authentication_required",
    "message": "Valid authentication token is required"
  },
  "status": "error"
}
```

```json
{
  "error": {
    "code": "insufficient_privileges", 
    "message": "Access denied. Required roles: [consumer]"
  },
  "status": "error"
}
```

### Static API Key Errors:
```json
{
  "error": {
    "code": "api_key_required",
    "message": "Static API key is required for this endpoint"
  },
  "status": "error"
}
```

```json
{
  "error": {
    "code": "invalid_api_key",
    "message": "Invalid API key provided"
  },
  "status": "error"
}
```

---

## 🔄 **Migration from Previous System**

### What Changed:
- **Before**: All onboard endpoints required JWT + data-provider role
- **After**: Onboard endpoints use static API key (service-to-service)
- **Benefit**: Simplified integration for data ingestion services

### Backward Compatibility:
- ✅ All existing JWT-protected endpoints unchanged
- ✅ User-facing APIs continue to work with JWT tokens
- ✅ Only onboard endpoints switched to static API key

---

## 📋 **Best Practices**

### For Static API Key:
1. **Generate strong keys** (minimum 32 characters, preferably 64+)
2. **Rotate keys regularly** (quarterly recommended)
3. **Use environment variables** - never hardcode keys
4. **Monitor usage** - check logs for unauthorized attempts
5. **Restrict network access** - use firewalls/VPNs when possible

### For JWT Tokens:
1. **Implement token refresh** for long-running applications
2. **Use HTTPS only** - never send tokens over HTTP
3. **Store securely** - use secure storage mechanisms
4. **Implement logout** - invalidate tokens when needed

---

## 🔧 **Development & Testing**

### Local Development Setup:
```bash
# Set environment variables
export STATIC_API_KEY="dev_key_for_testing_only_not_for_production"
export KEYCLOAK_PUBLIC_KEY="your_dev_keycloak_key"

# Test static API key endpoint
curl -H "X-API-Key: dev_key_for_testing_only_not_for_production" \
     -H "Content-Type: application/json" \
     -X POST http://localhost:8080/api/dataset \
     -d '{"test": "data"}'
```

### Production Deployment:
```bash
# Use secure secret management
export STATIC_API_KEY="${SECRET_MANAGER_API_KEY}"
export KEYCLOAK_PUBLIC_KEY="${SECRET_MANAGER_JWT_KEY}"
```

---

*This dual authentication system provides both security and flexibility, allowing secure user access through JWT tokens while enabling efficient service-to-service integration through static API keys.* 