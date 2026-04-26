# Travel Knowledge API Documentation

The Travel Knowledge API provides access to AI-generated travel content with categories (tags) for semantic search and retrieval.

## Base URL

```
/api/v1/travel-knowledge
```

**Note:** These endpoints are **public** and do not require authentication.

## Available Categories (Tags)

The system organizes travel documents into the following categories:

- **historical** - Historical & cultural information
- **logistics** - Travel planning & logistics
- **food** - Cuisine & dining recommendations
- **activities** - Attractions & things to do
- **nearby_places** - Nearby destinations & day trips

## Endpoints

### 1. Get Available Categories

Returns all available document categories/types that can be used as filters.

**Endpoint:** `GET /categories`

**Response:**
```json
{
  "categories": [
    {
      "type": "historical",
      "description": "Historical & cultural information"
    },
    {
      "type": "logistics",
      "description": "Travel planning & logistics"
    },
    {
      "type": "food",
      "description": "Cuisine & dining recommendations"
    },
    {
      "type": "activities",
      "description": "Attractions & things to do"
    },
    {
      "type": "nearby_places",
      "description": "Nearby destinations & day trips"
    }
  ]
}
```

---

### 2. List All Cities

Returns all cities in the knowledge base with document counts.

**Endpoint:** `GET /cities`

**Response:**
```json
{
  "count": 8,
  "cities": [
    {
      "CityName": "Ayodhya",
      "CityCountry": "India",
      "DocCount": 10
    },
    {
      "CityName": "Dwarka",
      "CityCountry": "India",
      "DocCount": 10
    }
  ]
}
```

---

### 3. Get City Documents

Retrieves all travel documents for a specific city, grouped by categories/tags.

**Endpoint:** `GET /city`

**Query Parameters:**
- `city` (required) - City name (e.g., "Ayodhya")
- `type` (optional) - Filter by document type (e.g., "food", "historical")

**Example Requests:**

Get all documents for a city:
```bash
GET /city?city=Ayodhya
```

Get specific category for a city:
```bash
GET /city?city=Ayodhya&type=food
```

**Response:**
```json
{
  "city_name": "Ayodhya",
  "city_country": "India",
  "total_count": 10,
  "categories": {
    "historical": [
      {
        "id": "uuid",
        "city_name": "Ayodhya",
        "city_country": "India",
        "document_type": "historical",
        "title": "Historical Significance of Ayodhya",
        "content": "Ayodhya is one of the seven sacred cities...",
        "metadata": {
          "tags": ["religious", "ancient", "heritage"],
          "keywords": ["Ram Mandir", "history", "culture"]
        },
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ],
    "food": [...],
    "activities": [...],
    "logistics": [...],
    "nearby_places": [...]
  }
}
```

---

### 4. Get City Info

Returns comprehensive aggregated information for a city.

**Endpoint:** `GET /city/:city/info`

**Path Parameters:**
- `city` (required) - City name

**Example Request:**
```bash
GET /city/Varanasi/info
```

**Response:**
```json
{
  "city_name": "Varanasi",
  "city_country": "India",
  "documents": {
    "historical": [...],
    "logistics": [...],
    "food": [...],
    "activities": [...],
    "nearby_places": [...]
  },
  "total_docs": 10,
  "last_updated": "2024-01-01T00:00:00Z"
}
```

---

### 5. Get Documents by Category

Retrieves all documents of a specific category/type across all cities.

**Endpoint:** `GET /category/:type`

**Path Parameters:**
- `type` (required) - Document type (historical, logistics, food, activities, nearby_places)

**Example Request:**
```bash
GET /category/food
```

**Response:**
```json
{
  "category": "food",
  "count": 16,
  "documents": [
    {
      "id": "uuid",
      "city_name": "Ayodhya",
      "city_country": "India",
      "document_type": "food",
      "title": "Local Cuisine of Ayodhya",
      "content": "...",
      "metadata": {...},
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

---

### 6. Get Knowledge Base Statistics

Returns statistics about the travel knowledge base.

**Endpoint:** `GET /stats`

**Response:**
```json
{
  "total_documents": 80,
  "unique_cities": 8,
  "by_type": [
    {
      "DocumentType": "historical",
      "Count": 16
    },
    {
      "DocumentType": "food",
      "Count": 16
    },
    {
      "DocumentType": "activities",
      "Count": 16
    },
    {
      "DocumentType": "logistics",
      "Count": 16
    },
    {
      "DocumentType": "nearby_places",
      "Count": 16
    }
  ]
}
```

---

## Error Responses

All endpoints return appropriate HTTP status codes:

### 400 Bad Request
```json
{
  "error": "city parameter is required"
}
```

### 404 Not Found
```json
{
  "error": "no documents found for this city"
}
```

### 500 Internal Server Error
```json
{
  "error": "error message details"
}
```

---

## Usage Examples

### Example 1: Build a City Guide Page

```javascript
// Fetch all travel information for a city
const response = await fetch('/api/v1/travel-knowledge/city?city=Ayodhya');
const data = await response.json();

// Access documents by category
const historicalInfo = data.categories.historical;
const foodRecommendations = data.categories.food;
const activities = data.categories.activities;
```

### Example 2: Display Food Recommendations

```javascript
// Get only food-related documents
const response = await fetch('/api/v1/travel-knowledge/city?city=Varanasi&type=food');
const data = await response.json();

// data.categories.food contains all food-related documents
data.categories.food.forEach(doc => {
  console.log(doc.title);
  console.log(doc.content);
  console.log('Tags:', doc.metadata.tags);
});
```

### Example 3: Browse by Category

```javascript
// Get all food documents across all cities
const response = await fetch('/api/v1/travel-knowledge/category/food');
const data = await response.json();

// Group by city or display all
data.documents.forEach(doc => {
  console.log(`${doc.city_name}: ${doc.title}`);
});
```

### Example 4: City Selection Dropdown

```javascript
// Get list of available cities
const response = await fetch('/api/v1/travel-knowledge/cities');
const data = await response.json();

// Build dropdown
data.cities.forEach(city => {
  const option = `${city.CityName}, ${city.CityCountry} (${city.DocCount} docs)`;
  // Add to dropdown
});
```

---

## Testing

Use the provided test script to verify all endpoints:

```bash
./test-travel-knowledge-api.sh
```

Or test individual endpoints with curl:

```bash
# Get categories
curl http://localhost:8080/api/v1/travel-knowledge/categories | jq

# Get city documents
curl "http://localhost:8080/api/v1/travel-knowledge/city?city=Ayodhya" | jq

# Get specific category
curl "http://localhost:8080/api/v1/travel-knowledge/city?city=Ayodhya&type=food" | jq

# Get statistics
curl http://localhost:8080/api/v1/travel-knowledge/stats | jq
```

---

## Data Model

### TravelDocument Structure

```go
type TravelDocument struct {
    ID           uuid.UUID       // Unique document ID
    CityName     string          // City name
    CityCountry  string          // Country name
    DocumentType DocumentType    // Category/tag (historical, logistics, etc.)
    Title        string          // Document title
    Content      string          // Full document content
    Metadata     json.RawMessage // Structured metadata (tags, keywords, etc.)
    Embedding    vector          // Vector embedding for semantic search (not returned in API)
    CreatedAt    time.Time       // Creation timestamp
    UpdatedAt    time.Time       // Last update timestamp
}
```

### Metadata Structure

The `metadata` field contains structured information:

```json
{
  "tags": ["tag1", "tag2"],
  "sources": ["source1"],
  "keywords": ["keyword1", "keyword2"],
  "coordinates": {
    "latitude": 26.8467,
    "longitude": 80.9462
  },
  "season": "winter",
  "best_time": "October to March",
  "language": "Hindi"
}
```

---

## Integration with Trip Planning

These APIs are designed to integrate seamlessly with your trip planning application:

1. **City Selection**: Use `/cities` to show available destinations
2. **Trip Preparation**: Fetch city info with `/city/:city/info` when user selects a destination
3. **Contextual Information**: Display relevant categories based on trip phase:
   - Planning phase → show `logistics`
   - Sightseeing → show `activities` and `historical`
   - Dining → show `food`
   - Day trips → show `nearby_places`

---

## Future Enhancements

The API is designed to support future features:

- **Semantic Search**: Vector-based search using embeddings
- **RAG Integration**: Use documents as context for AI-generated responses
- **Personalization**: Filter by user preferences, season, budget, etc.
- **Multi-language**: Support for multiple languages

---

## Support

For issues or questions:
- Check the main README: `TRAVEL_KNOWLEDGE_RAG.md`
- Review the service implementation: `travelknowledge/service.go`
- Check controller logic: `travelknowledge/controllers.go`
