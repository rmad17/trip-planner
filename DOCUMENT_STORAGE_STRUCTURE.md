# Document Storage Structure

## Overview

The Trip Planner application has a dedicated `documents` table that stores all file upload metadata with comprehensive tracking of storage providers, file paths, purposes, and user/trip mappings.

## Database Table: `documents`

### Schema

| Column | Type | Description | Your Requirement |
|--------|------|-------------|------------------|
| `id` | UUID | Primary key | - |
| `created_at` | timestamp | Record creation time | - |
| `updated_at` | timestamp | Last update time | - |
| **`storage_provider`** | varchar(50) | **Service name** (digitalocean, s3, gcs, azure, local, cloudflare) | ✅ **Service** |
| **`storage_path`** | text | **Full path/key in storage** for reading later | ✅ **File path** |
| **`category`** | varchar(50) | **Purpose/category** of the document | ✅ **Purpose** |
| **`name`** | text | **Display name** of the document | ✅ **Name** |
| **`original_name`** | text | **Original filename** when uploaded | ✅ **Name** |
| **`user_id`** | UUID | **Owner of the document** | ✅ **Mapped to user** |
| **`entity_type`** | text | **Type of entity** (trip_plan, trip_hop, stay, etc.) | ✅ **Mapped to trip** |
| **`entity_id`** | text | **ID of the trip/entity** | ✅ **Mapped to trip** |
| `file_size` | bigint | Size in bytes | - |
| `content_type` | text | MIME type | - |
| `description` | text | Optional description | - |
| `notes` | text | Optional user notes | - |
| `tags` | text (JSON) | Optional tags for organization | - |
| `uploaded_at` | timestamp | Upload timestamp | - |
| `expires_at` | timestamp | Optional expiration date | - |
| `is_public` | boolean | Public access flag | - |

## Supported Storage Providers

The system supports multiple cloud storage services:

### 1. DigitalOcean Spaces
- **Provider Code**: `digitalocean`
- **Status**: ✅ Fully implemented
- **Use Case**: Production-ready S3-compatible storage
- **Configuration**: See `DIGITALOCEAN_SPACES_SETUP.md`

### 2. AWS S3
- **Provider Code**: `s3`
- **Status**: ⚠️ Implementation pending
- **Use Case**: Amazon Web Services object storage
- **Note**: Interface defined, implementation needed

### 3. Google Cloud Storage (GCS)
- **Provider Code**: `gcs`
- **Status**: ⚠️ Implementation pending
- **Use Case**: Google Cloud Platform object storage
- **Note**: Interface defined, implementation needed

### 4. Azure Blob Storage
- **Provider Code**: `azure`
- **Status**: ⚠️ Implementation pending
- **Use Case**: Microsoft Azure object storage
- **Note**: Interface defined, implementation needed

### 5. Cloudflare R2
- **Provider Code**: `cloudflare`
- **Status**: ⚠️ Implementation pending
- **Use Case**: Cloudflare's S3-compatible storage
- **Note**: Interface defined, implementation needed

### 6. Local Storage
- **Provider Code**: `local`
- **Status**: ✅ Fully implemented
- **Use Case**: Development and testing
- **Storage Path**: `./uploads/documents/`

## Document Categories (Purpose)

The `category` field defines the purpose of each document:

| Category | Description | Example Use Cases |
|----------|-------------|-------------------|
| `tickets` | Travel tickets | Flight tickets, train tickets, bus tickets |
| `invoices` | Payment invoices | Hotel invoices, tour invoices |
| `identity_proofs` | Identity documents | Passport copies, ID cards, driver's licenses |
| `medical` | Medical documents | Vaccination certificates, prescriptions, insurance |
| `hotel_bookings` | Accommodation confirmations | Hotel confirmations, Airbnb bookings |
| `insurance` | Insurance policies | Travel insurance, health insurance |
| `visas` | Visa documents | Visa copies, visa applications |
| `receipts` | Purchase receipts | Expense receipts, payment proofs |
| `itineraries` | Trip itineraries | Day-by-day plans, tour schedules |
| `other` | Miscellaneous | Any other documents |

## User and Trip Mapping

### User Mapping
Every document is linked to the user who uploaded it:
```sql
user_id UUID NOT NULL
```

### Trip Mapping
Documents can be attached to various trip entities:

**Entity Types:**
- `trip_plan` - Main trip
- `trip_hop` - Individual destinations within a trip
- `stay` - Hotel/accommodation stays
- `activity` - Planned activities
- `expense` - Related expenses

**Example:**
```json
{
  "entity_type": "trip_plan",
  "entity_id": "123e4567-e89b-12d3-a456-426614174000",
  "user_id": "987e6543-e21b-12d3-a456-426614174000"
}
```

## Storage Path Structure

Files are organized with the following path structure:

### DigitalOcean Spaces / Cloud Storage
```
documents/{uuid}_{timestamp}{extension}
```

**Example:**
```
documents/a1b2c3d4-e5f6-7890-abcd-ef1234567890_1704067200.pdf
```

### Local Storage
```
uploads/documents/{uuid}_{timestamp}{extension}
```

**Example:**
```
uploads/documents/a1b2c3d4-e5f6-7890-abcd-ef1234567890_1704067200.pdf
```

## Document Lifecycle

### 1. Upload
```
POST /api/v1/trip/{trip_id}/documents
```

**What happens:**
1. File is uploaded via multipart form
2. Storage provider (from env) receives the file
3. Metadata is saved to `documents` table:
   - `storage_provider`: Active provider (e.g., "digitalocean")
   - `storage_path`: Full key/path in that provider
   - `category`: Purpose of the document
   - `name`: Display name
   - `user_id`: Uploader's ID
   - `entity_type` & `entity_id`: Linked trip/entity

### 2. Retrieve
```
GET /api/v1/documents/{document_id}
```

**Returns:**
```json
{
  "id": "uuid",
  "name": "Flight Ticket",
  "original_name": "ticket_123.pdf",
  "storage_provider": "digitalocean",
  "storage_path": "documents/abc123_1704067200.pdf",
  "category": "tickets",
  "user_id": "user-uuid",
  "entity_type": "trip_plan",
  "entity_id": "trip-uuid",
  "file_size": 2048576,
  "content_type": "application/pdf",
  "uploaded_at": "2024-01-01T12:00:00Z"
}
```

### 3. Download
```
GET /api/v1/documents/{document_id}/download
```

**What happens:**
1. Query `documents` table to get `storage_provider` and `storage_path`
2. Use appropriate storage provider to fetch file
3. Stream file to user

### 4. Delete
```
DELETE /api/v1/documents/{document_id}
```

**What happens:**
1. Query `documents` table to get file metadata
2. Delete file from storage provider using `storage_path`
3. Delete record from `documents` table

## Query Examples

### Get all documents for a user
```sql
SELECT * FROM documents WHERE user_id = 'user-uuid';
```

### Get all documents for a trip
```sql
SELECT * FROM documents
WHERE entity_type = 'trip_plan'
  AND entity_id = 'trip-uuid';
```

### Get documents by category
```sql
SELECT * FROM documents
WHERE category = 'tickets'
  AND user_id = 'user-uuid';
```

### Get documents by storage provider
```sql
SELECT * FROM documents
WHERE storage_provider = 'digitalocean';
```

### Get all DigitalOcean files for a specific trip
```sql
SELECT * FROM documents
WHERE storage_provider = 'digitalocean'
  AND entity_type = 'trip_plan'
  AND entity_id = 'trip-uuid';
```

## Document Sharing

A separate `document_shares` table handles sharing:

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID | Primary key |
| `document_id` | UUID | Reference to document |
| `shared_with` | UUID | User receiving access |
| `shared_by` | UUID | User granting access |
| `permission` | varchar(20) | Permission level (view, download) |
| `expires_at` | timestamp | Optional expiration |
| `is_active` | boolean | Active status |

## API Endpoints Summary

| Method | Endpoint | Purpose |
|--------|----------|---------|
| POST | `/api/v1/trip/{id}/documents` | Upload document to trip |
| GET | `/api/v1/trip/{id}/documents` | List all documents for trip |
| GET | `/api/v1/documents/{id}` | Get document metadata |
| PUT | `/api/v1/documents/{id}` | Update document metadata |
| DELETE | `/api/v1/documents/{id}` | Delete document |
| GET | `/api/v1/documents/{id}/download` | Download file |

## Code References

### Models
- **Document Model**: `documents/models.go:48-67`
- **Storage Providers**: `documents/models.go:29-36`
- **Document Categories**: `documents/models.go:13-24`

### Controllers
- **Upload Handler**: `documents/controllers.go:150-310`
- **Download Handler**: `documents/controllers.go:471-543`
- **Delete Handler**: `documents/controllers.go:425-473`

### Storage Layer
- **Storage Interface**: `storage/interfaces.go`
- **DigitalOcean Provider**: `storage/digitalocean.go`
- **Storage Manager**: `storage/config.go`
- **Storage Initialization**: `storage/init.go`

## Migration Status

The `documents` table is created via GORM auto-migration or Atlas migrations. All required columns are present in the current database schema.

To verify the table exists:
```bash
psql -U postgres -d trip -c "\d documents"
```

## Summary

✅ **All your requirements are met:**

1. **Service tracking**: `storage_provider` column stores the cloud service
2. **File path**: `storage_path` column stores the full path/key for retrieval
3. **Purpose**: `category` column defines the document purpose
4. **Name**: `name` and `original_name` columns
5. **User mapping**: `user_id` column links to the user
6. **Trip mapping**: `entity_type` and `entity_id` columns link to trips and related entities
7. **Separate table**: Dedicated `documents` table with all metadata

The system is production-ready with full support for DigitalOcean Spaces and a clean architecture for adding additional providers (AWS S3, GCP, Azure, Cloudflare R2).
