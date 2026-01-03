package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"triplanner/core"
	"triplanner/travelknowledge"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// AIProvider interface for generating content
type AIProvider interface {
	GenerateContent(ctx context.Context, city travelknowledge.City) (*travelknowledge.GeneratedContent, error)
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
}

// ContentGenerator handles the generation and storage of travel documents
type ContentGenerator struct {
	db       *gorm.DB
	provider AIProvider
}

// NewContentGenerator creates a new content generator
func NewContentGenerator(db *gorm.DB, provider AIProvider) *ContentGenerator {
	return &ContentGenerator{
		db:       db,
		provider: provider,
	}
}

// GenerateForCity generates all travel content for a city
func (cg *ContentGenerator) GenerateForCity(ctx context.Context, city travelknowledge.City) error {
	log.Printf("📝 Generating travel content for %s, %s...\n", city.Name, city.Country)

	// Generate content using AI
	generatedContent, err := cg.provider.GenerateContent(ctx, city)
	if err != nil {
		return fmt.Errorf("failed to generate content: %w", err)
	}

	log.Printf("✅ Generated %d sections for %s\n", len(generatedContent.Sections), city.Name)

	// Process and store each section
	for i, section := range generatedContent.Sections {
		log.Printf("  [%d/%d] Processing: %s - %s\n", i+1, len(generatedContent.Sections), section.Type, section.Title)

		// Create chunks for large content
		chunks := chunkContent(section.Content, 1000)

		for chunkIdx, chunk := range chunks {
			// Generate embedding
			embedding, err := cg.provider.GenerateEmbedding(ctx, chunk)
			if err != nil {
				log.Printf("⚠️  Warning: Failed to generate embedding for chunk %d: %v\n", chunkIdx, err)
				continue
			}

			// Prepare metadata
			metadata := travelknowledge.TravelDocumentMetadata{
				Tags:     extractTags(section.Type),
				Language: "en",
				Keywords: extractKeywords(chunk),
				Extra: map[string]interface{}{
					"chunk_index":  chunkIdx,
					"total_chunks": len(chunks),
					"generated_at": time.Now().Format(time.RFC3339),
				},
			}

			if city.Coordinates != nil {
				metadata.Coordinates = city.Coordinates
			}

			metadataJSON, err := json.Marshal(metadata)
			if err != nil {
				log.Printf("⚠️  Warning: Failed to marshal metadata: %v\n", err)
				metadataJSON = []byte("{}")
			}

			// Create title with chunk info if multiple chunks
			title := section.Title
			if len(chunks) > 1 {
				title = fmt.Sprintf("%s (Part %d/%d)", section.Title, chunkIdx+1, len(chunks))
			}

			// Create document
			doc := travelknowledge.TravelDocument{
				ID:           uuid.New(),
				CityName:     city.Name,
				CityCountry:  city.Country,
				DocumentType: section.Type,
				Title:        title,
				Content:      chunk,
				Metadata:     metadataJSON,
				Embedding:    pgvector.NewVector(embedding),
			}

			// Save to database
			if err := cg.db.Create(&doc).Error; err != nil {
				log.Printf("❌ Failed to save document: %v\n", err)
				return err
			}

			log.Printf("    ✓ Saved chunk %d/%d with embedding\n", chunkIdx+1, len(chunks))
		}
	}

	log.Printf("🎉 Successfully generated and stored all content for %s!\n\n", city.Name)
	return nil
}

// chunkContent splits content into smaller chunks for embedding
func chunkContent(content string, maxChunkSize int) []string {
	if len(content) <= maxChunkSize {
		return []string{content}
	}

	var chunks []string
	paragraphs := strings.Split(content, "\n\n")

	currentChunk := ""
	for _, para := range paragraphs {
		// If adding this paragraph exceeds max size, save current chunk
		if len(currentChunk)+len(para)+2 > maxChunkSize && len(currentChunk) > 0 {
			chunks = append(chunks, strings.TrimSpace(currentChunk))
			currentChunk = para
		} else {
			if len(currentChunk) > 0 {
				currentChunk += "\n\n" + para
			} else {
				currentChunk = para
			}
		}
	}

	// Add the last chunk
	if len(currentChunk) > 0 {
		chunks = append(chunks, strings.TrimSpace(currentChunk))
	}

	return chunks
}

// extractTags extracts relevant tags based on document type
func extractTags(docType travelknowledge.DocumentType) []string {
	switch docType {
	case travelknowledge.DocumentTypeHistorical:
		return []string{"history", "culture", "heritage", "monuments"}
	case travelknowledge.DocumentTypeLogistics:
		return []string{"travel", "transportation", "accommodation", "planning"}
	case travelknowledge.DocumentTypeFood:
		return []string{"food", "cuisine", "restaurants", "dining"}
	case travelknowledge.DocumentTypeActivities:
		return []string{"activities", "attractions", "entertainment", "tourism"}
	case travelknowledge.DocumentTypeNearbyPlaces:
		return []string{"nearby", "excursions", "day-trips", "surrounding"}
	default:
		return []string{"general", "travel"}
	}
}

// extractKeywords extracts basic keywords from content
func extractKeywords(content string) []string {
	// Simple keyword extraction - in production, use NLP libraries
	words := strings.Fields(strings.ToLower(content))
	wordMap := make(map[string]int)

	for _, word := range words {
		// Remove common words and punctuation
		word = strings.Trim(word, ".,!?;:")
		if len(word) > 4 && !isStopWord(word) {
			wordMap[word]++
		}
	}

	// Get top 10 most frequent words
	keywords := []string{}
	for word, count := range wordMap {
		if count >= 2 && len(keywords) < 10 {
			keywords = append(keywords, word)
		}
	}

	return keywords
}

// isStopWord checks if a word is a common stop word
func isStopWord(word string) bool {
	stopWords := map[string]bool{
		"the": true, "and": true, "for": true, "that": true, "this": true,
		"with": true, "from": true, "have": true, "will": true, "your": true,
		"about": true, "which": true, "their": true, "there": true, "would": true,
	}
	return stopWords[word]
}

// GetInitialCities returns the list of cities to generate content for
func GetInitialCities() []travelknowledge.City {
	return []travelknowledge.City{
		{
			Name:    "Varanasi",
			Country: "India",
			State:   "Uttar Pradesh",
			Coordinates: &travelknowledge.GeoCoordinates{
				Latitude:  25.3176,
				Longitude: 82.9739,
			},
			Description: "Ancient spiritual city on the banks of the Ganges",
		},
		{
			Name:    "Ayodhya",
			Country: "India",
			State:   "Uttar Pradesh",
			Coordinates: &travelknowledge.GeoCoordinates{
				Latitude:  26.7922,
				Longitude: 82.1998,
			},
			Description: "Historic and religious city, birthplace of Lord Rama",
		},
		{
			Name:    "Vrindavan",
			Country: "India",
			State:   "Uttar Pradesh",
			Coordinates: &travelknowledge.GeoCoordinates{
				Latitude:  27.5819,
				Longitude: 77.6831,
			},
			Description: "Sacred town associated with Lord Krishna",
		},
		{
			Name:    "Ujjain",
			Country: "India",
			State:   "Madhya Pradesh",
			Coordinates: &travelknowledge.GeoCoordinates{
				Latitude:  23.1765,
				Longitude: 75.7885,
			},
			Description: "Ancient city, one of the four sites of Kumbh Mela",
		},
		{
			Name:    "Dwarka",
			Country: "India",
			State:   "Gujarat",
			Coordinates: &travelknowledge.GeoCoordinates{
				Latitude:  22.2442,
				Longitude: 68.9685,
			},
			Description: "Coastal city, one of the seven most ancient cities in India",
		},
		{
			Name:    "Asansol",
			Country: "India",
			State:   "West Bengal",
			Coordinates: &travelknowledge.GeoCoordinates{
				Latitude:  23.6739,
				Longitude: 86.9524,
			},
			Description: "Industrial city in eastern India",
		},
		{
			Name:    "Kolkata",
			Country: "India",
			State:   "West Bengal",
			Coordinates: &travelknowledge.GeoCoordinates{
				Latitude:  22.5726,
				Longitude: 88.3639,
			},
			Description: "Cultural capital of India, City of Joy",
		},
		{
			Name:    "New Delhi",
			Country: "India",
			State:   "Delhi",
			Coordinates: &travelknowledge.GeoCoordinates{
				Latitude:  28.6139,
				Longitude: 77.2090,
			},
			Description: "Capital of India, blend of historical monuments and modern infrastructure",
		},
	}
}

func main() {
	// Load environment variables
	core.LoadEnvs()

	// Initialize database
	core.ConnectDB()
	db := core.GetDB()
	if db == nil {
		log.Fatalf("❌ Failed to connect to database")
	}

	// Check if pgvector extension is enabled
	var extensionExists bool
	err := db.Raw("SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'vector')").Scan(&extensionExists).Error
	if err != nil {
		log.Fatalf("❌ Failed to check pgvector extension: %v", err)
	}

	if !extensionExists {
		log.Println("⚠️  pgvector extension not found. Attempting to enable...")
		if err := db.Exec("CREATE EXTENSION IF NOT EXISTS vector").Error; err != nil {
			log.Fatalf("❌ Failed to enable pgvector extension: %v\nPlease run: CREATE EXTENSION vector; manually", err)
		}
		log.Println("✅ pgvector extension enabled!")
	}

	// Auto-migrate the schema
	// if err := db.AutoMigrate(&travelknowledge.TravelDocument{}); err != nil {
	// 	log.Fatalf("❌ Failed to migrate database: %v", err)
	// }
	// log.Println("✅ Database schema ready!")

	// Determine which AI provider to use
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			log.Println("⚠️  No API key found. Set OPENAI_API_KEY or ANTHROPIC_API_KEY")
			log.Println("💡 Using mock provider for testing...")
		}
	}

	// Create AI provider
	var provider AIProvider
	if os.Getenv("USE_OLLAMA") == "true" {
		provider = NewOllamaProvider(os.Getenv("OLLAMA_MODEL"))
		log.Println("🦙 Using Ollama local LLM")
	} else if apiKey != "" && strings.HasPrefix(apiKey, "sk-") {
		// OpenAI key detected
		provider = NewOpenAIProvider(apiKey)
		log.Println("🤖 Using OpenAI provider")
	} else if apiKey != "" && strings.HasPrefix(apiKey, "sk-ant-") {
		// Anthropic key detected
		provider = NewClaudeProvider(apiKey)
		log.Println("🤖 Using Claude provider")
	} else {
		// Mock provider for testing
		provider = NewMockProvider()
		log.Println("🎭 Using mock provider (for testing)")
	}

	// Create content generator
	generator := NewContentGenerator(db, provider)

	// Get cities to process
	cities := GetInitialCities()
	log.Printf("\n🌍 Processing %d cities...\n\n", len(cities))

	// Process each city
	ctx := context.Background()
	successCount := 0
	failCount := 0

	for i, city := range cities {
		log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		log.Printf("City %d/%d: %s, %s\n", i+1, len(cities), city.Name, city.Country)
		log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

		if err := generator.GenerateForCity(ctx, city); err != nil {
			log.Printf("❌ Failed to process %s: %v\n\n", city.Name, err)
			failCount++
			continue
		}

		successCount++

		// Add a small delay between cities to avoid rate limiting
		if i < len(cities)-1 {
			log.Println("⏸️  Waiting 2 seconds before next city...")
			time.Sleep(2 * time.Second)
		}
	}

	// Print summary
	log.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	log.Printf("📊 Generation Summary\n")
	log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	log.Printf("✅ Successful: %d/%d cities\n", successCount, len(cities))
	if failCount > 0 {
		log.Printf("❌ Failed: %d/%d cities\n", failCount, len(cities))
	}
	log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	// Query and display statistics
	var totalDocs int64
	db.Model(&travelknowledge.TravelDocument{}).Count(&totalDocs)
	log.Printf("📚 Total documents in database: %d\n", totalDocs)

	// Documents by type
	var typeStats []struct {
		DocumentType string
		Count        int64
	}
	db.Model(&travelknowledge.TravelDocument{}).
		Select("document_type, COUNT(*) as count").
		Group("document_type").
		Scan(&typeStats)

	log.Println("\n📈 Documents by type:")
	for _, stat := range typeStats {
		log.Printf("  - %s: %d\n", stat.DocumentType, stat.Count)
	}

	log.Println("\n🎉 Travel document generation complete!")
}
