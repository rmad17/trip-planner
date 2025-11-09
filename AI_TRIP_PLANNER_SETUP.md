# AI-Powered Trip Planner Setup Guide

## Overview

This feature adds an intelligent trip planning system that uses **Claude 3.5 Haiku** to generate complete trip itineraries based on user preferences. The system supports multi-city trips, provides AI-powered route suggestions, and automatically creates a comprehensive trip plan with hops, daily itineraries, and activities.

## Features

### 1. AI Trip Generation
- **Intelligent Route Planning**: Claude analyzes the best travel routes between cities
- **Travel Mode Recommendation**: Automatically suggests optimal transportation (flight, train, car, bus, or mixed)
- **Multi-City Support**: Plan trips with multiple destinations
- **Day-by-Day Itinerary**: Generates detailed daily plans with activities, timing, and costs
- **Budget Breakdown**: Provides cost estimates by category (accommodation, food, activities, etc.)
- **Places of Interest**: Suggests top attractions, restaurants, and activities for each destination
- **Travel Tips**: Offers practical advice and best time to visit information

### 2. Multi-City Suggestions
- **Smart Route Optimization**: AI suggests additional cities that fit logically into your route
- **Time-Aware Planning**: Considers travel duration and optimal days per city
- **Interest-Based Recommendations**: Tailors suggestions based on travel preferences

### 3. User-Friendly Interface
- **Interactive Modal**: Clean, step-by-step trip creation process
- **Place Autocomplete**: Search locations with existing PlaceSearchInput component
- **Preference Tags**: Select travel interests (adventure, culture, food, etc.)
- **Pace Options**: Choose between relaxed, moderate, or fast-paced travel
- **Live Preview**: Review AI-generated plan before saving
- **Loading States**: Progress indicators during AI generation

## Architecture

### Backend Components

#### 1. Claude Service (`/trips/claude_service.go`)
- **Model**: Claude 3.5 Haiku (`claude-3-5-haiku-20241022`)
- **Primary Functions**:
  - `GenerateTrip()`: Creates comprehensive trip plans
  - `SuggestMultiCityRoute()`: Provides additional city recommendations
- **Structured Output**: Returns JSON-formatted trip data ready for database insertion

#### 2. API Controllers (`/trips/ai_controllers.go`)
Three new endpoints:

1. **POST `/api/v1/trip/generate`**
   - Generates AI-powered trip plan
   - Returns preview without saving to database
   - Request body:
     ```json
     {
       "source": "New York",
       "destinations": ["Paris", "Rome"],
       "start_date": "2025-06-01T00:00:00Z",
       "end_date": "2025-06-15T00:00:00Z",
       "num_travelers": 2,
       "budget": 5000,
       "currency": "USD",
       "trip_preferences": ["culture", "food"],
       "pace_preference": "moderate"
     }
     ```

2. **POST `/api/v1/trip/generate/confirm`**
   - Saves AI-generated plan to database
   - Creates TripPlan, TripHops, TripDays, and Activities in single transaction
   - Returns trip ID for navigation

3. **GET `/api/v1/trip/suggest-cities`**
   - Query parameters: source, destination, duration, preferences
   - Returns list of recommended cities with reasoning

#### 3. Router Updates (`/trips/routers.go`)
Routes added to `RouterGroupTripPlans()` function

### Frontend Components

#### 1. AITripModal Component (`/src/components/AITripModal.js`)
**Features**:
- Three-step wizard: Input → Preview → Confirmation
- Multi-destination management (add/remove cities)
- AI suggestions integration
- Budget and currency selection
- Travel preference tags
- Pace preference (relaxed/moderate/fast)
- Loading states with progress messages
- Error handling

**State Management**:
- Form data for user inputs
- Generated plan preview
- Loading and error states
- Multi-city suggestions

#### 2. Dashboard Integration (`/src/pages/Dashboard.js`)
**Updates**:
- New "AI Trip Planner" button in hero section
- Modal state management
- Trip creation callback handler
- Navigation to created trip

#### 3. API Service (`/src/services/api.js`)
New endpoints in `tripAPI`:
- `generateTrip(generationData)`
- `confirmAITrip(tripPlan)`
- `getMultiCitySuggestions(source, destination, duration, preferences)`

## Setup Instructions

### Backend Setup

#### 1. Environment Variables
Add to your `.env` file:
```bash
ANTHROPIC_API_KEY=your_anthropic_api_key_here
```

Get your API key from: https://console.anthropic.com/

#### 2. Install/Verify Dependencies
The service uses standard Go libraries already in the project. Verify:
```bash
cd /home/sourav/projects/trip-planner
go mod tidy
```

#### 3. Start the Backend
```bash
go run app.go
```

The API will be available at `http://localhost:8080`

### Frontend Setup

#### 1. Install Dependencies (if needed)
```bash
cd /home/sourav/projects/trip-planner-fe
npm install
```

#### 2. Start the Frontend
```bash
npm start
```

The app will be available at `http://localhost:3000`

## Usage Guide

### Creating an AI-Powered Trip

1. **Login to the Dashboard**
   - Navigate to `http://localhost:3000`
   - Login with your credentials

2. **Open AI Trip Planner**
   - Click the "AI Trip Planner" button in the hero section
   - The modal will open with the input form

3. **Enter Trip Details**
   - **Starting From**: Enter your departure city (uses autocomplete)
   - **Destinations**: Add one or more destination cities
   - **Dates**: Select start and end dates
   - **Travelers**: Number of people traveling
   - **Budget**: Optional budget per person
   - **Preferences**: Select travel interests (adventure, culture, food, etc.)
   - **Pace**: Choose travel pace (relaxed, moderate, fast)

4. **Get Multi-City Suggestions (Optional)**
   - Click "Get AI Suggestions for Multi-City Route"
   - Review suggested cities that fit your route
   - Manually add interesting suggestions to destinations

5. **Generate Trip**
   - Click "Generate Trip with AI"
   - Wait for Claude to analyze and create your itinerary
   - This typically takes 10-30 seconds

6. **Review Preview**
   - See the complete trip plan:
     - Trip name and description
     - Recommended travel mode
     - Budget breakdown
     - Route with all stops
     - Daily itinerary with activities
     - Travel tips
   - Click "Back to Edit" to modify inputs
   - Click "Confirm & Save Trip" to create

7. **Access Your Trip**
   - After confirmation, you'll be redirected to the trip details page
   - All hops, days, and activities are pre-populated
   - You can edit any details as needed

## Data Flow

```
User Input (Modal)
    ↓
Frontend API Call (generateTrip)
    ↓
Backend Controller (GenerateTripWithAI)
    ↓
Claude Service (GenerateTrip)
    ↓
Claude API (Haiku 3.5)
    ↓
Structured JSON Response
    ↓
Preview in Modal
    ↓
User Confirms
    ↓
Frontend API Call (confirmAITrip)
    ↓
Backend Controller (CreateTripFromAIGeneration)
    ↓
Database Transaction
    ↓
Create: TripPlan → TripHops → TripDays → Activities
    ↓
Return Trip ID
    ↓
Navigate to Trip Details
```

## API Examples

### Generate Trip Request
```bash
curl -X POST http://localhost:8080/api/v1/trip/generate \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "source": "San Francisco",
    "destinations": ["Tokyo", "Kyoto"],
    "start_date": "2025-09-01T00:00:00Z",
    "end_date": "2025-09-14T00:00:00Z",
    "num_travelers": 2,
    "budget": 8000,
    "currency": "USD",
    "trip_preferences": ["culture", "food", "photography"],
    "pace_preference": "moderate"
  }'
```

### Get Multi-City Suggestions
```bash
curl -X GET "http://localhost:8080/api/v1/trip/suggest-cities?source=Paris&destination=Rome&duration=10&preferences=culture,food" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Troubleshooting

### Issue: "ANTHROPIC_API_KEY not configured"
**Solution**: Ensure you've added the API key to your `.env` file and restarted the backend server.

### Issue: "Failed to generate trip"
**Possible causes**:
1. Invalid API key
2. Network connectivity issues
3. Rate limiting from Anthropic API
4. Invalid date format

**Solution**: Check backend logs for detailed error messages.

### Issue: Modal doesn't show generated plan
**Solution**: Check browser console for errors. Ensure API response matches expected structure.

### Issue: Trip creation fails after confirmation
**Solution**:
- Check database connection
- Verify all required fields are present
- Check backend logs for transaction errors

## File Structure

### Backend Files
```
/home/sourav/projects/trip-planner/
├── trips/
│   ├── claude_service.go      # Claude API integration
│   ├── ai_controllers.go      # API endpoints for AI features
│   ├── routers.go             # Route definitions (updated)
│   ├── models.go              # Existing trip models
│   └── crud_controllers.go    # Existing CRUD operations
```

### Frontend Files
```
/home/sourav/projects/trip-planner-fe/
├── src/
│   ├── components/
│   │   └── AITripModal.js     # AI trip planning modal
│   ├── pages/
│   │   └── Dashboard.js       # Dashboard with AI button (updated)
│   └── services/
│       └── api.js             # API endpoints (updated)
```

## Technical Details

### Claude Prompt Engineering
The system uses a carefully crafted prompt that:
- Provides clear trip requirements
- Requests structured JSON output
- Specifies exact data format matching backend models
- Includes budget considerations
- Accounts for travel preferences and pace
- Requests practical travel tips

### Database Transaction
The `CreateTripFromAIGeneration` function uses a database transaction to ensure:
- Atomic creation of all trip components
- Rollback on any error
- Proper foreign key relationships
- Linked list structure for hop sequence

### Error Handling
- Frontend: User-friendly error messages in the modal
- Backend: Detailed error logs with request/response data
- API: Proper HTTP status codes (400, 500, etc.)

## Future Enhancements

Potential improvements:
1. **Image Integration**: Fetch destination images for preview
2. **Real-time Pricing**: Integration with flight/hotel APIs for live pricing
3. **Weather Data**: Include weather forecasts in itinerary
4. **Collaborative Planning**: Multiple users can edit AI-generated plans
5. **Export Options**: PDF/calendar export of itinerary
6. **Preference Learning**: Remember user preferences for future trips
7. **Alternative Routes**: Generate multiple route options to choose from
8. **Budget Optimization**: AI suggests ways to reduce costs

## API Rate Limits

Claude API rate limits (as of January 2025):
- Haiku 3.5: Check your Anthropic console for tier-specific limits
- Recommended: Implement caching for frequently requested routes
- Consider adding rate limiting middleware for production

## Security Considerations

1. **API Key Protection**: Never commit API keys to version control
2. **Authentication**: All endpoints require valid JWT token
3. **Input Validation**: Backend validates all user inputs
4. **SQL Injection**: Using GORM prevents SQL injection
5. **XSS Protection**: React automatically escapes output

## Support

For issues or questions:
1. Check this documentation
2. Review backend logs: `app.log` (if configured)
3. Check browser console for frontend errors
4. Review Anthropic API status: https://status.anthropic.com/

## License

This feature is part of the Trip Planner application.
