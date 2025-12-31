package travelknowledge

import (
	"context"
	"fmt"

	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// Service provides methods for querying travel documents
type Service struct {
	db *gorm.DB
}

// NewService creates a new travel knowledge service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// GetCityDocuments retrieves all documents for a specific city
func (s *Service) GetCityDocuments(ctx context.Context, cityName string) ([]TravelDocument, error) {
	var documents []TravelDocument
	err := s.db.WithContext(ctx).
		Where("city_name = ?", cityName).
		Order("document_type, created_at DESC").
		Find(&documents).Error
	return documents, err
}

// GetDocumentsByType retrieves all documents of a specific type
func (s *Service) GetDocumentsByType(ctx context.Context, docType DocumentType) ([]TravelDocument, error) {
	var documents []TravelDocument
	err := s.db.WithContext(ctx).
		Where("document_type = ?", docType).
		Order("city_name, created_at DESC").
		Find(&documents).Error
	return documents, err
}

// GetCityDocumentsByType retrieves documents for a city filtered by type
func (s *Service) GetCityDocumentsByType(ctx context.Context, cityName string, docType DocumentType) ([]TravelDocument, error) {
	var documents []TravelDocument
	err := s.db.WithContext(ctx).
		Where("city_name = ? AND document_type = ?", cityName, docType).
		Order("created_at DESC").
		Find(&documents).Error
	return documents, err
}

// SemanticSearch performs vector similarity search
func (s *Service) SemanticSearch(ctx context.Context, query SearchQuery, queryEmbedding []float32) ([]SearchResult, error) {
	if len(queryEmbedding) == 0 {
		return nil, fmt.Errorf("query embedding is required")
	}

	// Set defaults
	if query.Limit <= 0 {
		query.Limit = 10
	}
	if query.MinSimilarity <= 0 {
		query.MinSimilarity = 0.5
	}

	// Build the base query
	baseQuery := s.db.WithContext(ctx).Model(&TravelDocument{})

	// Apply filters
	if query.CityName != "" {
		baseQuery = baseQuery.Where("city_name = ?", query.CityName)
	}

	if len(query.DocumentTypes) > 0 {
		baseQuery = baseQuery.Where("document_type IN ?", query.DocumentTypes)
	}

	// Perform vector similarity search using cosine distance
	// In pgvector: 1 - cosine_distance = cosine_similarity
	vector := pgvector.NewVector(queryEmbedding)

	var results []struct {
		TravelDocument
		Distance float64 `gorm:"column:distance"`
	}

	err := baseQuery.
		Select("*, embedding <=> ? as distance", vector).
		Where("1 - (embedding <=> ?) >= ?", vector, query.MinSimilarity).
		Order("distance ASC").
		Limit(query.Limit).
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	// Convert to SearchResult
	searchResults := make([]SearchResult, len(results))
	for i, r := range results {
		searchResults[i] = SearchResult{
			Document:   r.TravelDocument,
			Similarity: 1 - r.Distance, // Convert distance to similarity
			Rank:       i + 1,
		}
	}

	return searchResults, nil
}

// GetCityInfo retrieves aggregated information for a city
func (s *Service) GetCityInfo(ctx context.Context, cityName string) (*CityTravelInfo, error) {
	var documents []TravelDocument
	err := s.db.WithContext(ctx).
		Where("city_name = ?", cityName).
		Order("document_type, created_at DESC").
		Find(&documents).Error

	if err != nil {
		return nil, err
	}

	if len(documents) == 0 {
		return nil, fmt.Errorf("no documents found for city: %s", cityName)
	}

	// Group documents by type
	docsByType := make(map[DocumentType][]TravelDocument)
	var country string
	var lastUpdated string

	for _, doc := range documents {
		docsByType[doc.DocumentType] = append(docsByType[doc.DocumentType], doc)
		if country == "" {
			country = doc.CityCountry
		}
		if doc.UpdatedAt.String() > lastUpdated {
			lastUpdated = doc.UpdatedAt.String()
		}
	}

	return &CityTravelInfo{
		CityName:    cityName,
		CityCountry: country,
		Documents:   docsByType,
		TotalDocs:   len(documents),
		LastUpdated: documents[0].UpdatedAt,
	}, nil
}

// ListCities returns all cities in the knowledge base
func (s *Service) ListCities(ctx context.Context) ([]struct {
	CityName    string
	CityCountry string
	DocCount    int64
}, error) {
	var cities []struct {
		CityName    string
		CityCountry string
		DocCount    int64
	}

	err := s.db.WithContext(ctx).
		Model(&TravelDocument{}).
		Select("city_name, city_country, COUNT(*) as doc_count").
		Group("city_name, city_country").
		Order("city_name").
		Find(&cities).Error

	return cities, err
}

// GetDocumentStats returns statistics about the document collection
func (s *Service) GetDocumentStats(ctx context.Context) (map[string]interface{}, error) {
	var totalDocs int64
	var cities int64
	var typeStats []struct {
		DocumentType string
		Count        int64
	}

	// Total documents
	if err := s.db.WithContext(ctx).Model(&TravelDocument{}).Count(&totalDocs).Error; err != nil {
		return nil, err
	}

	// Unique cities
	if err := s.db.WithContext(ctx).
		Model(&TravelDocument{}).
		Distinct("city_name").
		Count(&cities).Error; err != nil {
		return nil, err
	}

	// Documents by type
	if err := s.db.WithContext(ctx).
		Model(&TravelDocument{}).
		Select("document_type, COUNT(*) as count").
		Group("document_type").
		Find(&typeStats).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_documents": totalDocs,
		"unique_cities":   cities,
		"by_type":         typeStats,
	}, nil
}

// HybridSearch combines text filtering with semantic search
func (s *Service) HybridSearch(ctx context.Context, textQuery string, queryEmbedding []float32, filters map[string]interface{}, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	// Build query with text and vector search
	query := s.db.WithContext(ctx).Model(&TravelDocument{})

	// Apply text search if provided
	if textQuery != "" {
		query = query.Where("to_tsvector('english', content || ' ' || title) @@ plainto_tsquery('english', ?)", textQuery)
	}

	// Apply additional filters
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}

	// Add vector similarity if embedding provided
	if len(queryEmbedding) > 0 {
		vector := pgvector.NewVector(queryEmbedding)

		var results []struct {
			TravelDocument
			Distance float64 `gorm:"column:distance"`
		}

		err := query.
			Select("*, embedding <=> ? as distance", vector).
			Order("distance ASC").
			Limit(limit).
			Find(&results).Error

		if err != nil {
			return nil, err
		}

		searchResults := make([]SearchResult, len(results))
		for i, r := range results {
			searchResults[i] = SearchResult{
				Document:   r.TravelDocument,
				Similarity: 1 - r.Distance,
				Rank:       i + 1,
			}
		}

		return searchResults, nil
	}

	// Fallback to text-only search
	var documents []TravelDocument
	err := query.Limit(limit).Find(&documents).Error
	if err != nil {
		return nil, err
	}

	searchResults := make([]SearchResult, len(documents))
	for i, doc := range documents {
		searchResults[i] = SearchResult{
			Document:   doc,
			Similarity: 1.0, // Text match assumed perfect
			Rank:       i + 1,
		}
	}

	return searchResults, nil
}
