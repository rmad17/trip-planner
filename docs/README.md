# Trip Planner API Documentation

This directory contains the auto-generated Swagger/OpenAPI documentation for the Trip Planner API.

## Files

- `docs.go` - Generated Go code for embedding Swagger docs
- `swagger.json` - OpenAPI specification in JSON format
- `swagger.yaml` - OpenAPI specification in YAML format

## Accessing the Documentation

When the application is running, you can access the interactive Swagger UI at:

```
http://localhost:8080/swagger/index.html
```

## API Endpoints

### Trips

- **POST `/api/v1/trips/create`** - Create a new trip
  - Creates a trip plan with automatic creation of default hop and stay
  - Requires authentication
  - Request body: `CreateTripRequest`

- **GET `/api/v1/trips`** - Get all trips with user information
  - Retrieves all trips along with associated user data
  - Requires authentication

## Authentication

All trip endpoints require Bearer token authentication. Include the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

## Example Request

### Create a Trip

```bash
curl -X POST "http://localhost:8080/api/v1/trips/create" \
  -H "Authorization: Bearer <your-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "place_name": "Trip to Paris",
    "start_date": "2024-06-01T00:00:00Z",
    "end_date": "2024-06-10T00:00:00Z",
    "min_days": 7,
    "travel_mode": "flight",
    "notes": "Romantic getaway",
    "hotels": ["Hotel de Paris", "Le Bristol"],
    "tags": ["romantic", "europe", "culture"]
  }'
```

## Models

### CreateTripRequest

- `place_name` (required): Name of the trip destination
- `start_date`: Start date of the trip (ISO 8601 format)
- `end_date`: End date of the trip (ISO 8601 format)
- `min_days`: Minimum number of days for the trip
- `travel_mode`: Mode of travel (flight, car, train, etc.)
- `notes`: Additional notes about the trip
- `hotels`: Array of preferred hotels
- `tags`: Array of tags associated with the trip

## Auto-generated Files

These files are auto-generated using `swag init -g app.go`. To regenerate:

```bash
go run github.com/swaggo/swag/cmd/swag@latest init -g app.go
```