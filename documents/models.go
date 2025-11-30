package documents

import (
	"time"
	"triplanner/core"

	"github.com/google/uuid"
)

// DocumentCategory represents the category of a document
type DocumentCategory string

const (
	CategoryTickets        DocumentCategory = "tickets"
	CategoryInvoices       DocumentCategory = "invoices"
	CategoryIdentityProofs DocumentCategory = "identity_proofs"
	CategoryMedical        DocumentCategory = "medical"
	CategoryHotelBookings  DocumentCategory = "hotel_bookings"
	CategoryInsurance      DocumentCategory = "insurance"
	CategoryVisas          DocumentCategory = "visas"
	CategoryReceipts       DocumentCategory = "receipts"
	CategoryItineraries    DocumentCategory = "itineraries"
	CategoryOther          DocumentCategory = "other"
)

// StorageProvider represents the storage provider for documents
type StorageProvider string

const (
	StorageProviderDigitalOcean StorageProvider = "digitalocean"
	StorageProviderS3           StorageProvider = "s3"           // AWS S3
	StorageProviderGCS          StorageProvider = "gcs"          // Google Cloud Storage
	StorageProviderAzure        StorageProvider = "azure"        // Azure Blob Storage
	StorageProviderLocal        StorageProvider = "local"
	StorageProviderCloudflare   StorageProvider = "cloudflare"
)

// StoragePath represents the storage path pattern for organizing documents
type StoragePath string

const (
	StoragePathDocuments StoragePath = "documents"
	StoragePathImages    StoragePath = "images"
	StoragePathBackups   StoragePath = "backups"
	StoragePathTmp       StoragePath = "tmp"
	StoragePathUploads   StoragePath = "uploads"
)

// Document represents a document uploaded by users
type Document struct {
	core.BaseModel
	Name            string           `json:"name" gorm:"not null" example:"Flight Ticket" description:"Display name of the document"`
	OriginalName    string           `json:"original_name" gorm:"not null" example:"ticket_flight_123.pdf" description:"Original filename"`
	StorageProvider StorageProvider  `json:"storage_provider" gorm:"type:varchar(50);not null" example:"digitalocean" description:"Storage provider used (digitalocean, s3, etc.)"`
	StoragePath     string           `json:"storage_path" gorm:"not null" example:"documents/2024/01/15/abc123.pdf" description:"Path/key in storage provider"`
	FileSize        int64            `json:"file_size" example:"2048576" description:"File size in bytes"`
	ContentType     string           `json:"content_type" example:"application/pdf" description:"MIME type of the file"`
	Category        DocumentCategory `json:"category" gorm:"type:varchar(50);not null" example:"tickets" description:"Document category"`
	Description     *string          `json:"description" example:"Return flight ticket from NYC to Paris" description:"Optional description"`
	Notes           *string          `json:"notes" example:"Keep this handy at airport" description:"Optional user notes"`
	Tags            []string         `json:"tags" gorm:"serializer:json" example:"flight,business-class" description:"Optional tags for organization"`
	EntityType      *string          `json:"entity_type" example:"trip_plan" description:"Type of entity this document is attached to (trip_plan, trip_hop, stay, etc.)"`
	EntityID        *uuid.UUID       `json:"entity_id" example:"123e4567-e89b-12d3-a456-426614174000" description:"ID of the entity this document is attached to"`
	UserID          uuid.UUID        `json:"user_id" gorm:"type:uuid;not null" example:"123e4567-e89b-12d3-a456-426614174000" description:"ID of the user who uploaded the document"`
	UploadedAt      time.Time        `json:"uploaded_at" gorm:"not null" description:"Timestamp when document was uploaded"`
	ExpiresAt       *time.Time       `json:"expires_at" example:"2024-12-31T23:59:59Z" description:"Optional expiration date for documents like visas, insurance"`
	IsPublic        bool             `json:"is_public" gorm:"default:false" description:"Whether the document is publicly accessible"`
}

// DocumentShare represents sharing permissions for documents
type DocumentShare struct {
	core.BaseModel
	DocumentID uuid.UUID  `json:"document_id" gorm:"type:uuid;not null" description:"ID of the shared document"`
	SharedWith uuid.UUID  `json:"shared_with" gorm:"type:uuid;not null" description:"ID of user the document is shared with"`
	SharedBy   uuid.UUID  `json:"shared_by" gorm:"type:uuid;not null" description:"ID of user who shared the document"`
	Permission string     `json:"permission" gorm:"type:varchar(20);default:'view'" example:"view" description:"Permission level (view, download)"`
	ExpiresAt  *time.Time `json:"expires_at" description:"Optional expiration for the share"`
	IsActive   bool       `json:"is_active" gorm:"default:true" description:"Whether the share is currently active"`
}

// GetModels returns all models for Atlas/GORM
func GetModels() []interface{} {
	return []interface{}{
		&Document{},
		&DocumentShare{},
	}
}

// GetValidCategories returns all valid document categories
func GetValidCategories() []DocumentCategory {
	return []DocumentCategory{
		CategoryTickets,
		CategoryInvoices,
		CategoryIdentityProofs,
		CategoryMedical,
		CategoryHotelBookings,
		CategoryInsurance,
		CategoryVisas,
		CategoryReceipts,
		CategoryItineraries,
		CategoryOther,
	}
}

// IsValidCategory checks if a category is valid
func IsValidCategory(category string) bool {
	for _, validCategory := range GetValidCategories() {
		if string(validCategory) == category {
			return true
		}
	}
	return false
}

// GetValidStorageProviders returns all valid storage providers
func GetValidStorageProviders() []StorageProvider {
	return []StorageProvider{
		StorageProviderDigitalOcean,
		StorageProviderS3,
		StorageProviderGCS,
		StorageProviderAzure,
		StorageProviderLocal,
		StorageProviderCloudflare,
	}
}

// IsValidStorageProvider checks if a storage provider is valid
func IsValidStorageProvider(provider string) bool {
	for _, validProvider := range GetValidStorageProviders() {
		if string(validProvider) == provider {
			return true
		}
	}
	return false
}

// GetValidStoragePaths returns all valid storage paths
func GetValidStoragePaths() []StoragePath {
	return []StoragePath{
		StoragePathDocuments,
		StoragePathImages,
		StoragePathBackups,
		StoragePathTmp,
		StoragePathUploads,
	}
}

// IsValidStoragePath checks if a storage path is valid
func IsValidStoragePath(path string) bool {
	for _, validPath := range GetValidStoragePaths() {
		if string(validPath) == path {
			return true
		}
	}
	return false
}
