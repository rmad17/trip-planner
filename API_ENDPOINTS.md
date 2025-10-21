# Trip Planner API Endpoints - Complete CRUD Operations

This document provides a comprehensive list of all available API endpoints for the Trip Planner system with full CRUD (Create, Read, Update, Delete) functionality.

## Base URL
```
http://localhost:8080/api/v1
```

## Authentication
All endpoints require JWT Bearer token authentication except for auth endpoints.

**Header**: `Authorization: Bearer <token>`

---

## 🔐 Authentication Endpoints

### User Authentication
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/register` | Register new user |
| POST | `/auth/login` | Login user |
| POST | `/auth/logout` | Logout user |
| POST | `/auth/refresh` | Refresh token |

### Google OAuth
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/auth/google` | Google OAuth login |
| GET | `/auth/google/callback` | Google OAuth callback |

### User Profile
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/user/profile` | Get user profile |
| PUT | `/user/profile` | Update user profile |

---

## 🗺️ Trip Management Endpoints

### Trip Plans (Main Trip Entity)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/trip-plans` | Get all trip plans for user |
| GET | `/trip-plans/{id}` | Get specific trip plan |
| POST | `/trip-plans` | Create new trip plan |
| PUT | `/trip-plans/{id}` | Update trip plan |
| DELETE | `/trip-plans/{id}` | Delete trip plan |

**Query Parameters for GET /trip-plans:**
- `limit` (int): Number of records (default: 50)
- `offset` (int): Records to skip (default: 0)

### Trip Hops (Destinations within trips)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/trip-plans/{trip_plan_id}/hops` | Get hops for a trip |
| POST | `/trip-plans/{trip_plan_id}/hops` | Create new hop |
| PUT | `/trip-hops/{id}` | Update specific hop |
| DELETE | `/trip-hops/{id}` | Delete hop |

### Trip Days (Daily itineraries)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/trip-plans/{trip_plan_id}/days` | Get days for a trip |
| GET | `/trip-days/{id}` | Get specific trip day |
| POST | `/trip-plans/{trip_plan_id}/days` | Create new day |
| PUT | `/trip-days/{id}` | Update trip day |
| DELETE | `/trip-days/{id}` | Delete trip day |

### Activities (Events within days)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/trip-days/{trip_day_id}/activities` | Get activities for a day |
| POST | `/trip-days/{trip_day_id}/activities` | Create new activity |
| PUT | `/activities/{id}` | Update activity |
| DELETE | `/activities/{id}` | Delete activity |

### Travellers (Trip participants)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/trip-plans/{trip_plan_id}/travellers` | Get travellers for a trip |
| GET | `/travellers/{id}` | Get specific traveller |
| POST | `/trip-plans/{trip_plan_id}/travellers` | Add traveller to trip |
| POST | `/trip-plans/{trip_plan_id}/travellers/invite` | Invite traveller via email |
| PUT | `/travellers/{id}` | Update traveller |
| DELETE | `/travellers/{id}` | Remove traveller (soft delete) |

---

## 💰 Expense Management Endpoints

### Expenses
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/trip/{id}/expenses` | Get expenses for a trip |
| GET | `/expenses/{id}` | Get specific expense |
| POST | `/trip/{id}/expenses` | Create new expense |
| PUT | `/expenses/{id}` | Update expense |
| DELETE | `/expenses/{id}` | Delete expense |

**Query Parameters for GET expenses:**
- `category` (string): Filter by expense category
- `traveller` (uuid): Filter by traveller who paid
- `limit` (int): Number of records (default: 50)
- `offset` (int): Records to skip (default: 0)

### Expense Splits
| Method | Endpoint | Description |
|--------|----------|-------------|
| PUT | `/expense-splits/{id}` | Update expense split |
| POST | `/expense-splits/{id}/mark-paid` | Mark split as paid |

### Expense Settlements
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/trip/{id}/settlements` | Get settlements for a trip |
| POST | `/trip/{id}/settlements` | Create new settlement |

### Expense Summary
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/trip/{id}/expense-summary` | Get comprehensive expense summary |

---

## 📄 Document Management Endpoints

### Documents
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/trip/{id}/documents` | Get documents for a trip |
| POST | `/trip/{id}/documents` | Upload new document |
| GET | `/documents/{id}` | Get specific document |
| PUT | `/documents/{id}` | Update document |
| DELETE | `/documents/{id}` | Delete document |
| GET | `/documents/{id}/download` | Download document file |

**Query Parameters for GET documents:**
- `category` (string): Filter by document category
- `entity_type` (string): Filter by entity type
- `limit` (int): Number of records (default: 50)
- `offset` (int): Records to skip (default: 0)

**Document Categories:**
- `tickets`, `invoices`, `identity_proofs`, `medical`, `hotel_bookings`, `insurance`, `visas`, `receipts`, `itineraries`, `other`

---

## 🌍 Places API (Existing)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/places/search` | Search places using Google/Mapbox |

---

## 📊 Admin Panel Access

### GoAdmin Interface
- **URL**: `http://localhost:8080/admin`
- **Username**: `admin`
- **Password**: `admin`

### Admin Table Management
- **Trip Plans**: `/admin/info/trip_plans`
- **Users**: `/admin/info/users`
- **Travellers**: `/admin/info/travellers`
- **Expenses**: `/admin/info/expenses`
- **Documents**: `/admin/info/documents`

---

## 📝 Request/Response Examples

### Create Trip Plan
```json
POST /api/v1/trip-plans
{
  "name": "European Adventure",
  "description": "A 2-week journey through Europe",
  "start_date": "2024-06-01T00:00:00Z",
  "end_date": "2024-06-14T23:59:59Z",
  "budget": 5000.00,
  "currency": "EUR",
  "trip_type": "leisure",
  "status": "planning"
}
```

### Create Trip Hop
```json
POST /api/v1/trip-plans/{trip_plan_id}/hops
{
  "name": "Paris Visit",
  "description": "3 days in the City of Light",
  "city": "Paris",
  "country": "France",
  "start_date": "2024-06-01T00:00:00Z",
  "end_date": "2024-06-04T00:00:00Z",
  "estimated_budget": 1200.00
}
```

### Create Trip Day
```json
POST /api/v1/trip-plans/{trip_plan_id}/days
{
  "date": "2024-06-01",
  "day_number": 1,
  "title": "Arrival in Paris",
  "day_type": "travel",
  "estimated_budget": 150.00,
  "notes": "Arrive at CDG airport, check into hotel"
}
```

### Create Activity
```json
POST /api/v1/trip-days/{trip_day_id}/activities
{
  "name": "Visit Eiffel Tower",
  "description": "Take elevator to the top",
  "activity_type": "sightseeing",
  "start_time": "2024-06-01T10:00:00Z",
  "end_time": "2024-06-01T12:00:00Z",
  "location": "Champ de Mars, Paris",
  "estimated_cost": 25.50,
  "priority": 1
}
```

### Add Traveller
```json
POST /api/v1/trip-plans/{trip_plan_id}/travellers
{
  "first_name": "Jane",
  "last_name": "Smith",
  "email": "jane@example.com",
  "phone": "+1-555-0123",
  "role": "participant",
  "dietary_restrictions": "Vegetarian"
}
```

### Create Expense
```json
POST /api/v1/trip/{id}/expenses
{
  "title": "Hotel Booking",
  "description": "3 nights at Hotel de Paris",
  "amount": 450.00,
  "currency": "EUR",
  "category": "accommodation",
  "date": "2024-06-01T15:00:00Z",
  "location": "Paris",
  "vendor": "Hotel de Paris",
  "payment_method": "card",
  "split_method": "equal",
  "paid_by": "traveller_id_here"
}
```

### Upload Document
```bash
POST /api/v1/trip/{id}/documents
Content-Type: multipart/form-data

Form fields:
- file: (binary file)
- name: "Flight Ticket"
- category: "tickets"
- description: "Return flight ticket from NYC to Paris"
- notes: "Keep this handy at airport"
- tags: "flight,business-class"
- expires_at: "2024-12-31T23:59:59Z"
- is_public: false
```

### Expense Summary Response
```json
GET /api/v1/trip/{id}/expense-summary
{
  "summary": {
    "trip_plan": "trip_plan_id",
    "total_expenses": 2450.75,
    "currency": "EUR",
    "expense_count": 15,
    "category_totals": {
      "accommodation": 800.00,
      "food": 650.25,
      "transportation": 500.50,
      "activities": 400.00,
      "miscellaneous": 100.00
    },
    "traveller_totals": {
      "traveller_id_1": {
        "traveller_name": "John Doe",
        "total_paid": 1200.00,
        "total_owed": 1225.38,
        "balance": -25.38
      }
    },
    "pending_settlements": [
      {
        "from_traveller": "traveller_id_1",
        "from_traveller_name": "John Doe",
        "to_traveller": "traveller_id_2", 
        "to_traveller_name": "Jane Smith",
        "amount": 25.38,
        "currency": "EUR"
      }
    ]
  }
}
```

---

## 🔍 Error Responses

### Standard Error Format
```json
{
  "error": "Error message describing what went wrong"
}
```

### Common HTTP Status Codes
- `200` - Success
- `201` - Created
- `204` - No Content (successful deletion)
- `400` - Bad Request (validation errors)
- `401` - Unauthorized (authentication required)
- `404` - Not Found
- `500` - Internal Server Error

---

## 📋 Data Models Overview

### Key Entities
1. **TripPlan** - Main trip container
2. **TripHop** - Destinations within a trip
3. **TripDay** - Daily itineraries 
4. **Activity** - Specific events/tasks
5. **Traveller** - Trip participants
6. **Expense** - Trip expenses with splitting
7. **ExpenseSplit** - Individual expense shares
8. **ExpenseSettlement** - Payments between travellers
9. **Document** - File uploads with metadata
10. **DocumentShare** - Document sharing permissions

### Supported Enums
- **Currency**: USD, EUR, GBP, INR, CAD, AUD, JPY, OTHER
- **TripDayType**: travel, explore, relax, business, adventure, cultural
- **ActivityType**: transport, sightseeing, dining, shopping, entertainment, adventure, cultural, business, personal, other
- **ExpenseCategory**: accommodation, transportation, food, activities, shopping_gifts, insurance, visas_fees, medical, communication, miscellaneous, other
- **SplitMethod**: equal, exact, percentage, shares, paid_by
- **PaymentMethod**: cash, card, digital_pay, bank_transfer, cheque, other
- **DocumentCategory**: tickets, invoices, identity_proofs, medical, hotel_bookings, insurance, visas, receipts, itineraries, other
- **StorageProvider**: digitalocean, s3, gcs, local, cloudflare

---

## 🚀 Getting Started

1. **Authentication**: Obtain JWT token via `/auth/login`
2. **Create Trip**: POST to `/trip-plans`
3. **Add Destinations**: POST to `/trip-plans/{id}/hops`
4. **Plan Daily Activities**: POST to `/trip-plans/{id}/days` and `/trip-days/{id}/activities`
5. **Invite Travellers**: POST to `/trip-plans/{id}/travellers`
6. **Track Expenses**: POST to `/trip/{id}/expenses`
7. **Upload Documents**: POST to `/trip/{id}/documents`
8. **Manage via Admin**: Access `http://localhost:8080/admin`

## 🔄 Legacy Compatibility

The original endpoints are maintained for backward compatibility:
- `POST /trips/create` - Still available
- `GET /trips` - Still available

All new functionality uses the comprehensive CRUD endpoints documented above.

---

This API provides complete trip planning functionality with expense management, traveller coordination, and administrative oversight through both REST API and web interface.