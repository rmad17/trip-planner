package trips

import (
	"time"
	"triplanner/core"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// TripDayType represents the type of trip day
type TripDayType string

const (
	TripDayTypeTravel     TripDayType = "travel"     // Day primarily for traveling between locations
	TripDayTypeExplore    TripDayType = "explore"    // Day for exploring/sightseeing
	TripDayTypeRelax      TripDayType = "relax"      // Rest/leisure day
	TripDayTypeBusiness   TripDayType = "business"   // Business activities
	TripDayTypeAdventure  TripDayType = "adventure"  // Adventure activities
	TripDayTypeCultural   TripDayType = "cultural"   // Cultural experiences
)

// ActivityType represents the type of activity
type ActivityType string

const (
	ActivityTypeTransport    ActivityType = "transport"    // Transportation between places
	ActivityTypeSightseeing  ActivityType = "sightseeing"  // Tourist attractions
	ActivityTypeDining       ActivityType = "dining"       // Restaurants, cafes
	ActivityTypeShopping     ActivityType = "shopping"     // Shopping activities
	ActivityTypeEntertainment ActivityType = "entertainment" // Shows, movies, etc.
	ActivityTypeAdventure    ActivityType = "adventure"    // Adventure sports, hiking
	ActivityTypeCultural     ActivityType = "cultural"     // Museums, historical sites
	ActivityTypeBusiness     ActivityType = "business"     // Work-related activities
	ActivityTypePersonal     ActivityType = "personal"     // Personal time, rest
	ActivityTypeOther        ActivityType = "other"        // Other activities
)

// Currency represents supported currencies
type Currency string

const (
	CurrencyINR   Currency = "INR" // Indian Rupee
	CurrencyUSD   Currency = "USD" // US Dollar
	CurrencyGBP   Currency = "GBP" // British Pound
	CurrencyEUR   Currency = "EUR" // Euro
	CurrencyCAD   Currency = "CAD" // Canadian Dollar
	CurrencyAUD   Currency = "AUD" // Australian Dollar
	CurrencyJPY   Currency = "JPY" // Japanese Yen
	CurrencyOther Currency = "OTHER" // Other currencies
)

// TripPlan represents a trip plan in the system
type TripPlan struct {
	core.BaseModel
	Name            *string        `json:"name" example:"Trip to Paris" description:"Name of the trip"`
	Description     *string        `json:"description" example:"A romantic 10-day getaway to France" description:"Detailed description of the trip"`
	StartDate       *time.Time     `json:"start_date" example:"2024-06-01T00:00:00Z" description:"Start date of the trip"`
	EndDate         *time.Time     `json:"end_date" example:"2024-06-10T00:00:00Z" description:"End date of the trip"`
	MinDays         *int8          `json:"min_days" example:"7" description:"Minimum number of days for the trip"`
	MaxDays         *int8          `json:"max_days" example:"14" description:"Maximum number of days for the trip"`
	TravelMode      *string        `json:"travel_mode" example:"flight" description:"Primary mode of travel"`
	TripType        *string        `json:"trip_type" example:"leisure" description:"Type of trip (leisure, business, adventure, family, etc.)"`
	Budget          *float64       `json:"budget" example:"5000.00" description:"Total planned budget for the trip"`
	ActualSpent     *float64       `json:"actual_spent" example:"4750.25" description:"Total amount actually spent"`
	Currency        Currency       `json:"currency" gorm:"type:varchar(10);default:'USD'" example:"EUR" description:"Currency for budget and expenses"`
	Status          *string        `json:"status" example:"planning" description:"Status (planning, confirmed, in_progress, completed, cancelled)"`
	IsPublic        *bool          `json:"is_public" example:"false" description:"Whether the trip plan is publicly visible"`
	ShareCode       *string        `json:"share_code" example:"PARIS2024ABC" description:"Shareable code for the trip"`
	Notes           *string        `json:"notes" example:"Romantic getaway" description:"Additional notes"`
	Hotels          pq.StringArray `json:"hotels" gorm:"type:text[]" swaggertype:"array,string" example:"Hotel de Paris,Le Bristol" description:"List of preferred hotels"`
	Tags            pq.StringArray `json:"tags" gorm:"type:text[]" swaggertype:"array,string" example:"romantic,europe" description:"Trip tags"`
	Participants    pq.StringArray `json:"participants" gorm:"type:text[]" swaggertype:"array,string" example:"john@example.com,jane@example.com" description:"Email addresses of trip participants"`
	UserID          uuid.UUID      `json:"user_id" gorm:"type:uuid;not null" example:"123e4567-e89b-12d3-a456-426614174000" description:"ID of the user who created the trip"`
	TripHops        []TripHop      `json:"trip_hops,omitempty" gorm:"foreignKey:TripPlan" description:"Trip hops in this plan"`
	TripDays        []TripDay      `json:"trip_days,omitempty" gorm:"foreignKey:TripPlan" description:"Trip days in this plan"`
	Travellers      []Traveller    `json:"travellers,omitempty" gorm:"foreignKey:TripPlan" description:"Travellers in this trip"`
}

// TripHop represents a hop/leg in a trip itinerary
type TripHop struct {
	core.BaseModel
	Name            *string        `json:"name" example:"Paris Visit" description:"Name of the hop"`
	Description     *string        `json:"description" example:"3 days exploring the City of Light" description:"Detailed description of the hop"`
	City            *string        `json:"city" example:"Paris" description:"City name"`
	Country         *string        `json:"country" example:"France" description:"Country name"`
	Region          *string        `json:"region" example:"Île-de-France" description:"Region/state name"`
	MapSource       *string        `json:"map_source" example:"google" description:"Map service used (google, mapbox, etc.)"`
	PlaceID         *string        `json:"place_id" example:"ChIJD7fiBh9u5kcRYJSMaMOCCwQ" description:"Place ID from map service"`
	Latitude        *float64       `json:"latitude" example:"48.8566" description:"Latitude coordinate"`
	Longitude       *float64       `json:"longitude" example:"2.3522" description:"Longitude coordinate"`
	StartDate       *time.Time     `json:"start_date" example:"2024-06-01T00:00:00Z" description:"Start date of the hop"`
	EndDate         *time.Time     `json:"end_date" example:"2024-06-03T00:00:00Z" description:"End date of the hop"`
	EstimatedBudget *float64       `json:"estimated_budget" example:"800.00" description:"Estimated budget for this hop"`
	ActualSpent     *float64       `json:"actual_spent" example:"750.50" description:"Actual amount spent for this hop"`
	Transportation  *string        `json:"transportation" example:"train" description:"How to reach this hop (flight, train, car, etc.)"`
	Notes           *string        `json:"notes" example:"Visit Eiffel Tower" description:"Notes for this hop"`
	POIs            pq.StringArray `json:"pois" gorm:"type:text[]" swaggertype:"array,string" example:"Eiffel Tower,Louvre Museum" description:"Points of interest"`
	Restaurants     pq.StringArray `json:"restaurants" gorm:"type:text[]" swaggertype:"array,string" example:"Le Jules Verne,L'As du Fallafel" description:"Recommended restaurants"`
	Activities      pq.StringArray `json:"activities" gorm:"type:text[]" swaggertype:"array,string" example:"Seine River Cruise,Walking Tour" description:"Planned activities"`
	HopOrder        *int           `json:"hop_order" example:"1" description:"Order of this hop in the trip sequence"`
	PreviousHop     *uuid.UUID     `json:"previous_hop" gorm:"type:uuid" example:"123e4567-e89b-12d3-a456-426614174000" description:"ID of previous hop in sequence"`
	NextHop         *uuid.UUID     `json:"next_hop" gorm:"type:uuid" example:"123e4567-e89b-12d3-a456-426614174000" description:"ID of next hop in sequence"`
	TripPlan        uuid.UUID      `json:"trip_plan" gorm:"type:uuid;not null" example:"123e4567-e89b-12d3-a456-426614174000" description:"ID of the parent trip plan"`
	Stays           []Stay         `json:"stays,omitempty" gorm:"foreignKey:TripHop" description:"Accommodations for this hop"`
	TripDays        []TripDay      `json:"trip_days,omitempty" gorm:"foreignKey:FromTripHop" description:"Trip days starting from this hop"`
}

// Stay represents accommodation details for a trip hop
type Stay struct {
	core.BaseModel
	GoogleLocation *string    `json:"google_location" example:"ChIJD7fiBh9u5kcRYJSMaMOCCwQ" description:"Google Maps location identifier"`
	MapboxLocation *string    `json:"mapbox_location" example:"paris.hotel.123" description:"Mapbox location identifier"`
	StayType       *string    `json:"stay_type" example:"hotel" description:"Type of accommodation (hotel, airbnb, hostel, etc.)"`
	StayNotes      *string    `json:"stay_notes" example:"Near Eiffel Tower" description:"Notes about the accommodation"`
	StartDate      *time.Time `json:"start_date" example:"2024-06-01T00:00:00Z" description:"Check-in date"`
	EndDate        *time.Time `json:"end_date" example:"2024-06-03T00:00:00Z" description:"Check-out date"`
	IsPrepaid      *bool      `json:"is_prepaid" example:"true" description:"Whether the stay is prepaid"`
	PaymentMode    *string    `json:"payment_mode" example:"credit_card" description:"Payment method used"`
	TripHop        uuid.UUID  `json:"trip_hop" gorm:"type:uuid;not null" example:"123e4567-e89b-12d3-a456-426614174000" description:"ID of the associated trip hop"`
}

// TripDay represents a specific day within a trip, can span across trip hops
type TripDay struct {
	core.BaseModel
	Date            time.Time     `json:"date" gorm:"not null" example:"2024-06-01" description:"The specific date of this day"`
	DayNumber       int           `json:"day_number" gorm:"not null" example:"1" description:"Sequential day number in the trip (1-based)"`
	Title           *string       `json:"title" example:"Exploring Paris" description:"Title/theme for the day"`
	DayType         TripDayType   `json:"day_type" gorm:"type:varchar(20);not null" example:"explore" description:"Type of day (travel, explore, relax, etc.)"`
	Notes           *string       `json:"notes" example:"Start early to avoid crowds" description:"General notes for the day"`
	StartLocation   *string       `json:"start_location" example:"Hotel de Paris" description:"Where the day starts"`
	EndLocation     *string       `json:"end_location" example:"Eiffel Tower" description:"Where the day ends"`
	EstimatedBudget *float64      `json:"estimated_budget" example:"150.50" description:"Estimated budget for the day"`
	ActualBudget    *float64      `json:"actual_budget" example:"175.25" description:"Actual money spent"`
	Weather         *string       `json:"weather" example:"Sunny, 22°C" description:"Weather conditions/forecast"`
	TripPlan        uuid.UUID     `json:"trip_plan" gorm:"type:uuid;not null" example:"123e4567-e89b-12d3-a456-426614174000" description:"ID of the parent trip plan"`
	FromTripHop     *uuid.UUID    `json:"from_trip_hop" gorm:"type:uuid" example:"123e4567-e89b-12d3-a456-426614174000" description:"Starting trip hop for the day"`
	ToTripHop       *uuid.UUID    `json:"to_trip_hop" gorm:"type:uuid" example:"123e4567-e89b-12d3-a456-426614174000" description:"Ending trip hop for the day"`
	Activities      []Activity    `json:"activities" gorm:"foreignKey:TripDay" description:"Activities planned for this day"`
}

// Activity represents a specific activity/event planned for a day
type Activity struct {
	core.BaseModel
	Name           string       `json:"name" gorm:"not null" example:"Visit Eiffel Tower" description:"Name of the activity"`
	Description    *string      `json:"description" example:"Take elevator to the top, enjoy views" description:"Detailed description"`
	ActivityType   ActivityType `json:"activity_type" gorm:"type:varchar(20);not null" example:"sightseeing" description:"Type of activity"`
	StartTime      *time.Time   `json:"start_time" example:"2024-06-01T10:00:00Z" description:"Planned start time"`
	EndTime        *time.Time   `json:"end_time" example:"2024-06-01T12:00:00Z" description:"Planned end time"`
	Duration       *int         `json:"duration" example:"120" description:"Duration in minutes"`
	Location       *string      `json:"location" example:"Champ de Mars, 5 Avenue Anatole France" description:"Location/address"`
	MapSource      *string      `json:"map_source" example:"google" description:"Map service used"`
	PlaceID        *string      `json:"place_id" example:"ChIJLU7jMh9u5kcR4PcOOO6p3I0" description:"Place ID from map service"`
	EstimatedCost  *float64     `json:"estimated_cost" example:"25.50" description:"Estimated cost"`
	ActualCost     *float64     `json:"actual_cost" example:"30.00" description:"Actual cost incurred"`
	Priority       *int8        `json:"priority" example:"1" description:"Priority level (1=highest, 5=lowest)"`
	Status         *string      `json:"status" example:"planned" description:"Status (planned, confirmed, completed, cancelled)"`
	BookingRef     *string      `json:"booking_ref" example:"EIF123456" description:"Booking reference if applicable"`
	ContactInfo    *string      `json:"contact_info" example:"+33 1 44 11 23 23" description:"Contact information"`
	Notes          *string      `json:"notes" example:"Book tickets in advance to avoid queue" description:"Additional notes"`
	Tags           pq.StringArray `json:"tags" gorm:"type:text[]" swaggertype:"array,string" example:"must-see,photo-op" description:"Activity tags"`
	TripDay        uuid.UUID    `json:"trip_day" gorm:"type:uuid;not null" example:"123e4567-e89b-12d3-a456-426614174000" description:"ID of the parent trip day"`
	TripHop        *uuid.UUID   `json:"trip_hop" gorm:"type:uuid" example:"123e4567-e89b-12d3-a456-426614174000" description:"Associated trip hop if applicable"`
}

// Traveller represents a person participating in a trip
type Traveller struct {
	core.BaseModel
	FirstName       string         `json:"first_name" gorm:"not null" example:"John" description:"First name of the traveller"`
	LastName        string         `json:"last_name" gorm:"not null" example:"Doe" description:"Last name of the traveller"`
	Email           *string        `json:"email" example:"john.doe@example.com" description:"Email address"`
	Phone           *string        `json:"phone" example:"+1-555-0123" description:"Phone number with country code"`
	DateOfBirth     *time.Time     `json:"date_of_birth" example:"1990-01-15T00:00:00Z" description:"Date of birth"`
	Nationality     *string        `json:"nationality" example:"US" description:"Nationality (ISO country code)"`
	PassportNumber  *string        `json:"passport_number" example:"A12345678" description:"Passport number"`
	PassportExpiry  *time.Time     `json:"passport_expiry" example:"2030-01-15T00:00:00Z" description:"Passport expiry date"`
	EmergencyContact *string       `json:"emergency_contact" example:"Jane Doe: +1-555-9876" description:"Emergency contact information"`
	DietaryRestrictions *string    `json:"dietary_restrictions" example:"Vegetarian, No nuts" description:"Dietary restrictions or preferences"`
	MedicalNotes    *string        `json:"medical_notes" example:"Allergic to shellfish" description:"Important medical information"`
	Role            *string        `json:"role" example:"organizer" description:"Role in trip (organizer, participant, etc.)"`
	IsActive        bool           `json:"is_active" gorm:"default:true" description:"Whether the traveller is active in the trip"`
	JoinedAt        time.Time      `json:"joined_at" gorm:"not null" description:"When the traveller joined the trip"`
	Notes           *string        `json:"notes" example:"Prefers window seats" description:"Additional notes about the traveller"`
	TripPlan        uuid.UUID      `json:"trip_plan" gorm:"type:uuid;not null" example:"123e4567-e89b-12d3-a456-426614174000" description:"ID of the associated trip plan"`
	UserID          *uuid.UUID     `json:"user_id" gorm:"type:uuid" example:"123e4567-e89b-12d3-a456-426614174000" description:"ID of the associated user account (if registered)"`
}

// GetValidTripDayTypes returns all valid trip day types
func GetValidTripDayTypes() []TripDayType {
	return []TripDayType{
		TripDayTypeTravel,
		TripDayTypeExplore,
		TripDayTypeRelax,
		TripDayTypeBusiness,
		TripDayTypeAdventure,
		TripDayTypeCultural,
	}
}

// IsValidTripDayType checks if a trip day type is valid
func IsValidTripDayType(dayType string) bool {
	for _, validType := range GetValidTripDayTypes() {
		if string(validType) == dayType {
			return true
		}
	}
	return false
}

// GetValidActivityTypes returns all valid activity types
func GetValidActivityTypes() []ActivityType {
	return []ActivityType{
		ActivityTypeTransport,
		ActivityTypeSightseeing,
		ActivityTypeDining,
		ActivityTypeShopping,
		ActivityTypeEntertainment,
		ActivityTypeAdventure,
		ActivityTypeCultural,
		ActivityTypeBusiness,
		ActivityTypePersonal,
		ActivityTypeOther,
	}
}

// IsValidActivityType checks if an activity type is valid
func IsValidActivityType(activityType string) bool {
	for _, validType := range GetValidActivityTypes() {
		if string(validType) == activityType {
			return true
		}
	}
	return false
}

// GetValidCurrencies returns all valid currencies
func GetValidCurrencies() []Currency {
	return []Currency{
		CurrencyINR,
		CurrencyUSD,
		CurrencyGBP,
		CurrencyEUR,
		CurrencyCAD,
		CurrencyAUD,
		CurrencyJPY,
		CurrencyOther,
	}
}

// IsValidCurrency checks if a currency is valid
func IsValidCurrency(currency string) bool {
	for _, validCurrency := range GetValidCurrencies() {
		if string(validCurrency) == currency {
			return true
		}
	}
	return false
}

// Add method to get models for Atlas
func GetModels() []interface{} {
	return []interface{}{
		&TripPlan{},
		&TripHop{},
		&Stay{},
		&TripDay{},
		&Activity{},
		&Traveller{},
	}
}
