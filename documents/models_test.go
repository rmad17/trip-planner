package documents

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupDocumentsTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(&Document{}, &DocumentShare{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func StringPtr(s string) *string {
	return &s
}

func TimePtr(t time.Time) *time.Time {
	return &t
}

func UUIDPtr(u uuid.UUID) *uuid.UUID {
	return &u
}

func TestDocumentCategory_Constants(t *testing.T) {
	tests := []struct {
		name     string
		category DocumentCategory
		expected string
	}{
		{"Tickets", CategoryTickets, "tickets"},
		{"Invoices", CategoryInvoices, "invoices"},
		{"Identity Proofs", CategoryIdentityProofs, "identity_proofs"},
		{"Medical", CategoryMedical, "medical"},
		{"Hotel Bookings", CategoryHotelBookings, "hotel_bookings"},
		{"Insurance", CategoryInsurance, "insurance"},
		{"Visas", CategoryVisas, "visas"},
		{"Receipts", CategoryReceipts, "receipts"},
		{"Itineraries", CategoryItineraries, "itineraries"},
		{"Other", CategoryOther, "other"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.category))
		})
	}
}

func TestStorageProvider_Constants(t *testing.T) {
	tests := []struct {
		name     string
		provider StorageProvider
		expected string
	}{
		{"DigitalOcean", StorageProviderDigitalOcean, "digitalocean"},
		{"S3", StorageProviderS3, "s3"},
		{"GCS", StorageProviderGCS, "gcs"},
		{"Local", StorageProviderLocal, "local"},
		{"Cloudflare", StorageProviderCloudflare, "cloudflare"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.provider))
		})
	}
}

func TestStoragePath_Constants(t *testing.T) {
	tests := []struct {
		name     string
		path     StoragePath
		expected string
	}{
		{"Documents", StoragePathDocuments, "documents"},
		{"Images", StoragePathImages, "images"},
		{"Backups", StoragePathBackups, "backups"},
		{"Tmp", StoragePathTmp, "tmp"},
		{"Uploads", StoragePathUploads, "uploads"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.path))
		})
	}
}

func TestIsValidCategory(t *testing.T) {
	tests := []struct {
		name     string
		category string
		expected bool
	}{
		{"Valid - tickets", "tickets", true},
		{"Valid - invoices", "invoices", true},
		{"Valid - insurance", "insurance", true},
		{"Invalid - random", "random", false},
		{"Invalid - empty", "", false},
		{"Invalid - uppercase", "TICKETS", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidCategory(tt.category)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidStorageProvider(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		expected bool
	}{
		{"Valid - digitalocean", "digitalocean", true},
		{"Valid - s3", "s3", true},
		{"Valid - local", "local", true},
		{"Invalid - dropbox", "dropbox", false},
		{"Invalid - empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidStorageProvider(tt.provider)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidStoragePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"Valid - documents", "documents", true},
		{"Valid - images", "images", true},
		{"Valid - backups", "backups", true},
		{"Invalid - videos", "videos", false},
		{"Invalid - empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidStoragePath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetValidCategories(t *testing.T) {
	categories := GetValidCategories()

	t.Run("Returns all categories", func(t *testing.T) {
		assert.Len(t, categories, 10)
	})

	t.Run("Contains expected categories", func(t *testing.T) {
		assert.Contains(t, categories, CategoryTickets)
		assert.Contains(t, categories, CategoryInvoices)
		assert.Contains(t, categories, CategoryInsurance)
		assert.Contains(t, categories, CategoryVisas)
	})
}

func TestGetValidStorageProviders(t *testing.T) {
	providers := GetValidStorageProviders()

	t.Run("Returns all providers", func(t *testing.T) {
		assert.Len(t, providers, 5)
	})

	t.Run("Contains expected providers", func(t *testing.T) {
		assert.Contains(t, providers, StorageProviderDigitalOcean)
		assert.Contains(t, providers, StorageProviderS3)
		assert.Contains(t, providers, StorageProviderGCS)
		assert.Contains(t, providers, StorageProviderLocal)
		assert.Contains(t, providers, StorageProviderCloudflare)
	})
}

func TestGetValidStoragePaths(t *testing.T) {
	paths := GetValidStoragePaths()

	t.Run("Returns all paths", func(t *testing.T) {
		assert.Len(t, paths, 5)
	})

	t.Run("Contains expected paths", func(t *testing.T) {
		assert.Contains(t, paths, StoragePathDocuments)
		assert.Contains(t, paths, StoragePathImages)
		assert.Contains(t, paths, StoragePathBackups)
		assert.Contains(t, paths, StoragePathTmp)
		assert.Contains(t, paths, StoragePathUploads)
	})
}

func TestDocument_Model(t *testing.T) {
	db := setupDocumentsTestDB(t)

	t.Run("Create document with required fields", func(t *testing.T) {
		userID := uuid.New()

		doc := Document{
			Name:            "Flight Ticket",
			OriginalName:    "ticket_123.pdf",
			StorageProvider: StorageProviderDigitalOcean,
			StoragePath:     "documents/2024/01/ticket_123.pdf",
			FileSize:        2048576,
			ContentType:     "application/pdf",
			Category:        CategoryTickets,
			UserID:          userID,
			UploadedAt:      time.Now(),
			IsPublic:        false,
		}

		result := db.Create(&doc)
		assert.NoError(t, result.Error)
		assert.NotEqual(t, uuid.Nil, doc.ID)
		assert.Equal(t, "Flight Ticket", doc.Name)
		assert.Equal(t, CategoryTickets, doc.Category)
	})

	t.Run("Create document with all optional fields", func(t *testing.T) {
		userID := uuid.New()
		tripPlanID := uuid.New()
		expiryDate := time.Now().AddDate(1, 0, 0)

		doc := Document{
			Name:            "Travel Insurance",
			OriginalName:    "insurance_policy.pdf",
			StorageProvider: StorageProviderS3,
			StoragePath:     "documents/insurance/policy_123.pdf",
			FileSize:        1024000,
			ContentType:     "application/pdf",
			Category:        CategoryInsurance,
			Description:     StringPtr("Comprehensive travel insurance"),
			Notes:           StringPtr("Valid for 1 year"),
			Tags:            []string{"insurance", "important"},
			EntityType:      StringPtr("trip_plan"),
			EntityID:        UUIDPtr(tripPlanID),
			UserID:          userID,
			UploadedAt:      time.Now(),
			ExpiresAt:       TimePtr(expiryDate),
			IsPublic:        false,
		}

		result := db.Create(&doc)
		assert.NoError(t, result.Error)
		assert.NotNil(t, doc.Description)
		assert.Equal(t, "Comprehensive travel insurance", *doc.Description)
		assert.NotNil(t, doc.Notes)
		assert.Equal(t, 2, len(doc.Tags))
		assert.NotNil(t, doc.ExpiresAt)
	})

	t.Run("Create document with different categories", func(t *testing.T) {
		userID := uuid.New()
		categories := []DocumentCategory{
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

		for _, category := range categories {
			doc := Document{
				Name:            "Test Document",
				OriginalName:    "test.pdf",
				StorageProvider: StorageProviderLocal,
				StoragePath:     "documents/test.pdf",
				FileSize:        1024,
				ContentType:     "application/pdf",
				Category:        category,
				UserID:          userID,
				UploadedAt:      time.Now(),
			}

			result := db.Create(&doc)
			assert.NoError(t, result.Error)
			assert.Equal(t, category, doc.Category)
		}
	})

	t.Run("Create document with different storage providers", func(t *testing.T) {
		userID := uuid.New()
		providers := []StorageProvider{
			StorageProviderDigitalOcean,
			StorageProviderS3,
			StorageProviderGCS,
			StorageProviderLocal,
			StorageProviderCloudflare,
		}

		for _, provider := range providers {
			doc := Document{
				Name:            "Test Document",
				OriginalName:    "test.pdf",
				StorageProvider: provider,
				StoragePath:     "documents/test.pdf",
				FileSize:        1024,
				ContentType:     "application/pdf",
				Category:        CategoryOther,
				UserID:          userID,
				UploadedAt:      time.Now(),
			}

			result := db.Create(&doc)
			assert.NoError(t, result.Error)
			assert.Equal(t, provider, doc.StorageProvider)
		}
	})

	t.Run("Create document with tags", func(t *testing.T) {
		userID := uuid.New()

		doc := Document{
			Name:            "Tagged Document",
			OriginalName:    "tagged.pdf",
			StorageProvider: StorageProviderLocal,
			StoragePath:     "documents/tagged.pdf",
			FileSize:        1024,
			ContentType:     "application/pdf",
			Category:        CategoryOther,
			Tags:            []string{"important", "urgent", "review"},
			UserID:          userID,
			UploadedAt:      time.Now(),
		}

		result := db.Create(&doc)
		assert.NoError(t, result.Error)
		assert.Equal(t, 3, len(doc.Tags))
		assert.Contains(t, doc.Tags, "important")
		assert.Contains(t, doc.Tags, "urgent")
	})

	t.Run("Query documents by user", func(t *testing.T) {
		userID := uuid.New()

		for i := 0; i < 3; i++ {
			doc := Document{
				Name:            "User Document",
				OriginalName:    "doc.pdf",
				StorageProvider: StorageProviderLocal,
				StoragePath:     "documents/doc.pdf",
				FileSize:        1024,
				ContentType:     "application/pdf",
				Category:        CategoryOther,
				UserID:          userID,
				UploadedAt:      time.Now(),
			}
			db.Create(&doc)
		}

		var docs []Document
		result := db.Where("user_id = ?", userID).Find(&docs)
		assert.NoError(t, result.Error)
		assert.GreaterOrEqual(t, len(docs), 3)
	})

	t.Run("Query documents by category", func(t *testing.T) {
		userID := uuid.New()

		doc := Document{
			Name:            "Visa Document",
			OriginalName:    "visa.pdf",
			StorageProvider: StorageProviderLocal,
			StoragePath:     "documents/visa.pdf",
			FileSize:        1024,
			ContentType:     "application/pdf",
			Category:        CategoryVisas,
			UserID:          userID,
			UploadedAt:      time.Now(),
		}
		db.Create(&doc)

		var docs []Document
		result := db.Where("category = ?", CategoryVisas).Find(&docs)
		assert.NoError(t, result.Error)
		assert.GreaterOrEqual(t, len(docs), 1)
	})

	t.Run("Public vs private documents", func(t *testing.T) {
		userID := uuid.New()

		publicDoc := Document{
			Name:            "Public Document",
			OriginalName:    "public.pdf",
			StorageProvider: StorageProviderLocal,
			StoragePath:     "documents/public.pdf",
			FileSize:        1024,
			ContentType:     "application/pdf",
			Category:        CategoryOther,
			UserID:          userID,
			UploadedAt:      time.Now(),
			IsPublic:        true,
		}
		db.Create(&publicDoc)

		privateDoc := Document{
			Name:            "Private Document",
			OriginalName:    "private.pdf",
			StorageProvider: StorageProviderLocal,
			StoragePath:     "documents/private.pdf",
			FileSize:        1024,
			ContentType:     "application/pdf",
			Category:        CategoryOther,
			UserID:          userID,
			UploadedAt:      time.Now(),
			IsPublic:        false,
		}
		db.Create(&privateDoc)

		var publicDocs []Document
		db.Where("is_public = ?", true).Find(&publicDocs)
		assert.GreaterOrEqual(t, len(publicDocs), 1)
	})
}

func TestDocumentShare_Model(t *testing.T) {
	db := setupDocumentsTestDB(t)

	t.Run("Create document share", func(t *testing.T) {
		documentID := uuid.New()
		sharedWith := uuid.New()
		sharedBy := uuid.New()

		share := DocumentShare{
			DocumentID: documentID,
			SharedWith: sharedWith,
			SharedBy:   sharedBy,
			Permission: "view",
			IsActive:   true,
		}

		result := db.Create(&share)
		assert.NoError(t, result.Error)
		assert.NotEqual(t, uuid.Nil, share.ID)
		assert.Equal(t, "view", share.Permission)
		assert.True(t, share.IsActive)
	})

	t.Run("Create document share with expiration", func(t *testing.T) {
		documentID := uuid.New()
		sharedWith := uuid.New()
		sharedBy := uuid.New()
		expiryDate := time.Now().AddDate(0, 1, 0)

		share := DocumentShare{
			DocumentID: documentID,
			SharedWith: sharedWith,
			SharedBy:   sharedBy,
			Permission: "download",
			ExpiresAt:  TimePtr(expiryDate),
			IsActive:   true,
		}

		result := db.Create(&share)
		assert.NoError(t, result.Error)
		assert.NotNil(t, share.ExpiresAt)
		assert.Equal(t, "download", share.Permission)
	})

	t.Run("Deactivate document share", func(t *testing.T) {
		documentID := uuid.New()
		sharedWith := uuid.New()
		sharedBy := uuid.New()

		share := DocumentShare{
			DocumentID: documentID,
			SharedWith: sharedWith,
			SharedBy:   sharedBy,
			Permission: "view",
			IsActive:   true,
		}
		db.Create(&share)

		// Deactivate share
		share.IsActive = false
		db.Save(&share)

		var updated DocumentShare
		db.First(&updated, share.ID)
		assert.False(t, updated.IsActive)
	})

	t.Run("Query shares by document", func(t *testing.T) {
		documentID := uuid.New()
		sharedBy := uuid.New()

		for i := 0; i < 3; i++ {
			share := DocumentShare{
				DocumentID: documentID,
				SharedWith: uuid.New(),
				SharedBy:   sharedBy,
				Permission: "view",
				IsActive:   true,
			}
			db.Create(&share)
		}

		var shares []DocumentShare
		result := db.Where("document_id = ?", documentID).Find(&shares)
		assert.NoError(t, result.Error)
		assert.Equal(t, 3, len(shares))
	})

	t.Run("Query shares by user", func(t *testing.T) {
		userID := uuid.New()
		sharedBy := uuid.New()

		for i := 0; i < 2; i++ {
			share := DocumentShare{
				DocumentID: uuid.New(),
				SharedWith: userID,
				SharedBy:   sharedBy,
				Permission: "view",
				IsActive:   true,
			}
			db.Create(&share)
		}

		var shares []DocumentShare
		result := db.Where("shared_with = ?", userID).Find(&shares)
		assert.NoError(t, result.Error)
		assert.GreaterOrEqual(t, len(shares), 2)
	})
}

func TestGetModels_Documents(t *testing.T) {
	models := GetModels()

	t.Run("Returns correct number of models", func(t *testing.T) {
		assert.Len(t, models, 2)
	})

	t.Run("Contains Document model", func(t *testing.T) {
		found := false
		for _, model := range models {
			if _, ok := model.(*Document); ok {
				found = true
				break
			}
		}
		assert.True(t, found, "Document model should be in returned models")
	})

	t.Run("Contains DocumentShare model", func(t *testing.T) {
		found := false
		for _, model := range models {
			if _, ok := model.(*DocumentShare); ok {
				found = true
				break
			}
		}
		assert.True(t, found, "DocumentShare model should be in returned models")
	})
}
