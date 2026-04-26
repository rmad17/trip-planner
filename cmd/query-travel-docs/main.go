package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"triplanner/core"
	"triplanner/travelknowledge"
)

func main() {
	// Define command-line flags
	city := flag.String("city", "", "City name to query")
	docType := flag.String("type", "", "Document type (historical, logistics, food, activities, nearby_places)")
	listCities := flag.Bool("list", false, "List all cities in the knowledge base")
	stats := flag.Bool("stats", false, "Show database statistics")

	flag.Parse()

	// Load environment variables
	core.LoadEnvs()

	// Initialize database
	core.ConnectDB()
	db := core.GetDB()
	if db == nil {
		log.Fatalf("❌ Failed to connect to database")
	}

	// Create service
	service := travelknowledge.NewService(db)
	ctx := context.Background()

	// Handle different commands
	switch {
	case *stats:
		showStats(ctx, service)
	case *listCities:
		listAllCities(ctx, service)
	case *city != "":
		if *docType != "" {
			showCityDocsByType(ctx, service, *city, travelknowledge.DocumentType(*docType))
		} else {
			showCityInfo(ctx, service, *city)
		}
	default:
		fmt.Println("Travel Knowledge Query Tool")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  query-travel-docs -list                    # List all cities")
		fmt.Println("  query-travel-docs -stats                   # Show statistics")
		fmt.Println("  query-travel-docs -city \"Varanasi\"         # Get all info for a city")
		fmt.Println("  query-travel-docs -city \"Varanasi\" -type food  # Get specific type")
		fmt.Println()
		fmt.Println("Document types: historical, logistics, food, activities, nearby_places")
	}
}

func showStats(ctx context.Context, service *travelknowledge.Service) {
	stats, err := service.GetDocumentStats(ctx)
	if err != nil {
		log.Fatalf("❌ Error getting stats: %v", err)
	}

	fmt.Println("📊 Travel Knowledge Database Statistics")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Total Documents: %v\n", stats["total_documents"])
	fmt.Printf("Unique Cities: %v\n", stats["unique_cities"])
	fmt.Println()
	fmt.Println("Documents by Type:")

	if byType, ok := stats["by_type"].([]struct {
		DocumentType string
		Count        int64
	}); ok {
		for _, t := range byType {
			fmt.Printf("  - %-20s: %d\n", t.DocumentType, t.Count)
		}
	}
}

func listAllCities(ctx context.Context, service *travelknowledge.Service) {
	cities, err := service.ListCities(ctx)
	if err != nil {
		log.Fatalf("❌ Error listing cities: %v", err)
	}

	fmt.Println("🌍 Cities in Knowledge Base")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	for i, city := range cities {
		fmt.Printf("%d. %-20s (%s) - %d documents\n",
			i+1, city.CityName, city.CityCountry, city.DocCount)
	}
}

func showCityInfo(ctx context.Context, service *travelknowledge.Service, cityName string) {
	info, err := service.GetCityInfo(ctx, cityName)
	if err != nil {
		log.Fatalf("❌ Error getting city info: %v", err)
	}

	fmt.Printf("📍 %s, %s\n", info.CityName, info.CityCountry)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Total Documents: %d\n", info.TotalDocs)
	fmt.Printf("Last Updated: %s\n", info.LastUpdated.Format("2006-01-02 15:04:05"))
	fmt.Println()

	for docType, docs := range info.Documents {
		fmt.Printf("\n%s (%d documents):\n", docType, len(docs))
		fmt.Println(string(make([]byte, 50, 50)))

		for i, doc := range docs {
			if i >= 3 {
				fmt.Printf("  ... and %d more\n", len(docs)-3)
				break
			}
			fmt.Printf("  %d. %s\n", i+1, doc.Title)
			preview := doc.Content
			if len(preview) > 200 {
				preview = preview[:200] + "..."
			}
			fmt.Printf("     %s\n\n", preview)
		}
	}
}

func showCityDocsByType(ctx context.Context, service *travelknowledge.Service, cityName string, docType travelknowledge.DocumentType) {
	docs, err := service.GetCityDocumentsByType(ctx, cityName, docType)
	if err != nil {
		log.Fatalf("❌ Error getting documents: %v", err)
	}

	if len(docs) == 0 {
		fmt.Printf("No %s documents found for %s\n", docType, cityName)
		return
	}

	fmt.Printf("📍 %s - %s (%d documents)\n", cityName, docType, len(docs))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	for i, doc := range docs {
		fmt.Printf("\n%d. %s\n", i+1, doc.Title)
		fmt.Println(string(make([]byte, 70, 70)))
		fmt.Println(doc.Content)
		fmt.Println()

		// Print metadata if present
		if len(doc.Metadata) > 2 {
			var metadata travelknowledge.TravelDocumentMetadata
			if err := json.Unmarshal(doc.Metadata, &metadata); err == nil {
				if len(metadata.Tags) > 0 {
					fmt.Printf("Tags: %v\n", metadata.Tags)
				}
				if len(metadata.Keywords) > 0 {
					fmt.Printf("Keywords: %v\n", metadata.Keywords)
				}
			}
		}

		fmt.Println()
	}

	// Export option
	fmt.Printf("\n💾 Export to file? (y/N): ")
	var response string
	fmt.Scanln(&response)
	if response == "y" || response == "Y" {
		filename := fmt.Sprintf("%s_%s.json", cityName, docType)
		data, _ := json.MarshalIndent(docs, "", "  ")
		os.WriteFile(filename, data, 0644)
		fmt.Printf("✅ Exported to %s\n", filename)
	}
}
