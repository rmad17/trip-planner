package travelknowledge

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

// DocumentType represents the category of travel document
type DocumentType string

const (
	DocumentTypeHistorical   DocumentType = "historical"
	DocumentTypeLogistics    DocumentType = "logistics"
	DocumentTypeFood         DocumentType = "food"
	DocumentTypeActivities   DocumentType = "activities"
	DocumentTypeNearbyPlaces DocumentType = "nearby_places"
)

// TravelDocument represents a knowledge base entry for travel information
type TravelDocument struct {
	ID           uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CityName     string          `gorm:"type:varchar(255);not null;index:idx_travel_docs_city" json:"city_name"`
	CityCountry  string          `gorm:"type:varchar(255);not null" json:"city_country"`
	DocumentType DocumentType    `gorm:"type:varchar(100);not null;index:idx_travel_docs_type" json:"document_type"`
	Title        string          `gorm:"type:varchar(500);not null" json:"title"`
	Content      string          `gorm:"type:text;not null" json:"content"`
	Metadata     json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	Embedding    pgvector.Vector `gorm:"type:vector(1536)" json:"-"`
	CreatedAt    time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name for GORM
func (TravelDocument) TableName() string {
	return "travel_documents"
}

// TravelDocumentMetadata represents the structured metadata
type TravelDocumentMetadata struct {
	Tags        []string           `json:"tags,omitempty"`
	Sources     []string           `json:"sources,omitempty"`
	Coordinates *GeoCoordinates    `json:"coordinates,omitempty"`
	Season      string             `json:"season,omitempty"`
	PriceRange  string             `json:"price_range,omitempty"`
	Duration    string             `json:"duration,omitempty"`
	BestTime    string             `json:"best_time,omitempty"`
	Language    string             `json:"language,omitempty"`
	Keywords    []string           `json:"keywords,omitempty"`
	RelatedDocs []uuid.UUID        `json:"related_docs,omitempty"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
}

// GeoCoordinates represents latitude and longitude
type GeoCoordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Value implements driver.Valuer for database storage
func (m TravelDocumentMetadata) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan implements sql.Scanner for database retrieval
func (m *TravelDocumentMetadata) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}
	return json.Unmarshal(bytes, m)
}

// CityTravelInfo aggregates all travel information for a city
type CityTravelInfo struct {
	CityName     string                        `json:"city_name"`
	CityCountry  string                        `json:"city_country"`
	Documents    map[DocumentType][]TravelDocument `json:"documents"`
	TotalDocs    int                           `json:"total_docs"`
	LastUpdated  time.Time                     `json:"last_updated"`
}

// SearchResult represents a semantic search result with similarity score
type SearchResult struct {
	Document   TravelDocument `json:"document"`
	Similarity float64        `json:"similarity"`
	Rank       int            `json:"rank"`
}

// SearchQuery represents a semantic search request
type SearchQuery struct {
	Query          string         `json:"query"`
	CityName       string         `json:"city_name,omitempty"`
	DocumentTypes  []DocumentType `json:"document_types,omitempty"`
	Limit          int            `json:"limit,omitempty"`
	MinSimilarity  float64        `json:"min_similarity,omitempty"`
	IncludeContent bool           `json:"include_content"`
}

// ContentChunk represents a chunk of content for embedding
type ContentChunk struct {
	Text      string            `json:"text"`
	Metadata  map[string]string `json:"metadata"`
	ChunkSize int               `json:"chunk_size"`
}

// GeneratedContent represents AI-generated travel content before storage
type GeneratedContent struct {
	CityName    string                 `json:"city_name"`
	CityCountry string                 `json:"city_country"`
	Sections    []ContentSection       `json:"sections"`
	Metadata    TravelDocumentMetadata `json:"metadata"`
}

// ContentSection represents a section of generated content
type ContentSection struct {
	Type    DocumentType `json:"type"`
	Title   string       `json:"title"`
	Content string       `json:"content"`
}

// City represents a city for which to generate travel content
type City struct {
	Name        string  `json:"name"`
	Country     string  `json:"country"`
	State       string  `json:"state,omitempty"`
	Coordinates *GeoCoordinates `json:"coordinates,omitempty"`
	Population  int     `json:"population,omitempty"`
	Description string  `json:"description,omitempty"`
}
