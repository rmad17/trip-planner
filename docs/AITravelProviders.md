# AI Travel Providers — Phase 1 Design

## Overview

This document specifies the Phase 1 implementation for AI-orchestrated trip planning with real hotel offers, flight offers, and map routes. It extends the existing Gemini-backed `GenerateTrip` flow from a single prompt into a tool-using agent loop, while keeping the public API surface backwards-compatible with the existing React frontend.

The design follows **Pattern A — backend-orchestrated tools**: Gemini calls structured Go functions exposed as tools; orchestration, auth, caching, and rate-limiting all live in Go. No agent runtime (ADK / Vertex Agent Builder) is introduced in Phase 1.

## Goals

1. Replace AI-invented hotel/flight/route data with real provider data inside the existing `POST /trip/generate` response.
2. Expose the same provider data through dedicated `GET /search/*` endpoints for direct frontend use (Stays tab "search hotels" button, future flights tab, itinerary route overlay).
3. Use **only free or free-tier** providers for the v1 default configuration.
4. Make every provider swappable via a single env var, so paid providers (Duffel, Booking.com Demand, Google Routes paid tier) drop in without caller changes.
5. Preserve the existing FE contract — additive fields only on existing JSON responses.

## Non-Goals

- No SSE / WebSocket streaming. The FE has no `EventSource` and uses a synchronous request/response wizard.
- No booking flow. We surface offers and store the user's selection; we do not call provider booking APIs.
- No agent runtime. Tool orchestration stays in `gemini_provider.go`.
- No new top-level FE routes. New features slot into existing `Dashboard` and `TripDetails` tabs.
- No OpenAPI codegen on the FE (it does not have it set up).

## Provider Choices

| Domain   | Default (free)             | Why                                                      | Designed swap targets                  |
|----------|----------------------------|----------------------------------------------------------|----------------------------------------|
| Flights  | Amadeus Self-Service       | Real airline inventory; free test env + ~2k prod calls/mo; same OAuth as hotels | Duffel, Skyscanner, Kiwi               |
| Hotels   | Amadeus Self-Service       | Same key/auth as flights — one integration covers two domains | Booking.com Demand, LiteAPI, Expedia Rapid |
| Routes   | Google Routes API          | Aligns with existing Google stack (`places/googleapi.go`); future-compatible with Gemini grounded-with-Maps; ~10k free req/mo | Mapbox Directions (kept as fallback impl), OpenRouteService, OSRM |
| Geocode  | Mapbox Search (existing)   | Already used by FE place autocomplete                    | Google Places, Pelias                  |
| Soft info (visa/weather/events) | Gemini grounding with Google Search | Citations, freshness; ~1,500/day free in AI Studio | Tavily, Perplexity, Bing               |

**One Amadeus account** covers flights and hotels via shared OAuth2. **Two API keys total** for v1: Amadeus + Google Maps. Mapbox token remains for geocode and routing fallback.

## Package Layout

New packages mirror the existing `trips/llm_provider.go` factory pattern:

```
core/
  amadeus_client.go     # Shared OAuth2 token cache + HTTP client (used by hotels + flights)

hotels/
  provider.go           # Provider interface + canonical types
  amadeus.go            # AmadeusHotelProvider
  mock.go               # MockHotelProvider (fixtures, for offline dev + CI)
  factory.go            # NewProviderFactory(), reads HOTELS_PROVIDER
  api.go                # GET /search/hotels handler
  router.go

flights/
  provider.go
  amadeus.go            # AmadeusFlightProvider
  mock.go               # MockFlightProvider
  factory.go
  api.go                # GET /search/flights handler
  router.go

routes/
  provider.go
  google.go             # GoogleRoutesProvider (default)
  mapbox.go             # MapboxRouteProvider (fallback impl)
  mock.go               # MockRouteProvider
  factory.go
  api.go                # GET /routes handler
  router.go
```

`core/amadeus_client.go` is the rationale for sharing: same client_id/secret, same `/v1/security/oauth2/token` endpoint, same 30-min token TTL. Both hotel and flight providers borrow one cache.

## Provider Interfaces (the swap surface)

### Hotels

```go
// hotels/provider.go
type Provider interface {
    SearchHotels(ctx context.Context, q HotelSearchQuery) ([]HotelOffer, error)
    GetProviderName() string
}

type HotelSearchQuery struct {
    City             string
    CheckIn          string  // YYYY-MM-DD
    CheckOut         string  // YYYY-MM-DD
    Guests           int
    Currency         string
    MaxPricePerNight float64 // 0 = no cap
}

// HotelOffer field names mirror the existing `Stay` model on the FE
// so "Use this hotel" is a one-line spread into newStayData.
type HotelOffer struct {
    Provider       string          `json:"provider"`
    OfferID        string          `json:"offer_id"`
    Name           string          `json:"name"`
    Address        string          `json:"address"`
    CheckInDate    string          `json:"check_in_date"`
    CheckOutDate   string          `json:"check_out_date"`
    CostPerNight   float64         `json:"cost_per_night"`
    TotalCost      float64         `json:"total_cost"`
    Currency       string          `json:"currency"`
    Rating         float32         `json:"rating,omitempty"`
    Lat            float64         `json:"lat,omitempty"`
    Lng            float64         `json:"lng,omitempty"`
    DeepLink       string          `json:"deep_link,omitempty"`
    Raw            json.RawMessage `json:"-"` // provider-native blob, opaque
}
```

### Flights

```go
// flights/provider.go
type Provider interface {
    SearchFlights(ctx context.Context, q FlightSearchQuery) ([]FlightOffer, error)
    GetProviderName() string
}

type FlightSearchQuery struct {
    From, To  string  // IATA codes
    Depart    string  // YYYY-MM-DD
    Return    string  // YYYY-MM-DD, "" for one-way
    Pax       int
    Cabin     string  // economy|premium_economy|business|first
    Currency  string
}

type FlightOffer struct {
    Provider  string          `json:"provider"`
    OfferID   string          `json:"offer_id"`
    Price     float64         `json:"price"`
    Currency  string          `json:"currency"`
    Cabin     string          `json:"cabin"`
    Segments  []FlightSegment `json:"segments"`
    DeepLink  string          `json:"deep_link,omitempty"`
    Raw       json.RawMessage `json:"-"`
}

type FlightSegment struct {
    From, To       string  // IATA
    Depart, Arrive string  // RFC3339
    Carrier        string
    FlightNumber   string
    DurationMin    int
}
```

### Routes

```go
// routes/provider.go
type Provider interface {
    GetRoute(ctx context.Context, q RouteQuery) (*Route, error)
    GetProviderName() string
}

type RouteQuery struct {
    Stops []LngLat
    Mode  string // drive|walk|transit|cycle
}

type LngLat struct{ Lng, Lat float64 }

type Route struct {
    DistanceKm       float64         `json:"distance_km"`
    DurationMin      int             `json:"duration_min"`
    PolylineGeoJSON  json.RawMessage `json:"polyline_geojson"` // GeoJSON LineString — Mapbox GL renders directly
    Steps            []RouteStep     `json:"steps,omitempty"`
    Provider         string          `json:"provider"`
}

type RouteStep struct {
    Instruction string  `json:"instruction"`
    DistanceKm  float64 `json:"distance_km"`
    DurationMin int     `json:"duration_min"`
}
```

The `Raw` field on offers preserves provider-native data (offer IDs, fare rules, room types) that may be needed if/when a real booking flow is added. It is **not** marshaled in API responses but **is** persisted to JSONB columns.

## Factory Pattern

Identical shape across all three domains:

```go
// hotels/factory.go
type ProviderType string

const (
    ProviderAmadeus ProviderType = "amadeus"
    ProviderMock    ProviderType = "mock"
    // ProviderBooking, ProviderLiteAPI added later — no caller changes
)

type ProviderFactory struct{ defaultProvider ProviderType }

func NewProviderFactory() *ProviderFactory {
    p := os.Getenv("HOTELS_PROVIDER")
    if p == "" {
        p = string(ProviderAmadeus)
    }
    return &ProviderFactory{defaultProvider: ProviderType(p)}
}

func (f *ProviderFactory) GetProvider(t ProviderType) (Provider, error) { /* switch on t */ }
func (f *ProviderFactory) GetDefaultProvider() (Provider, error)        { return f.GetProvider(f.defaultProvider) }
```

Adding a new provider = one new file implementing `Provider`, one new case in the factory switch, one new env value. Zero changes to handlers or to the Gemini tool layer.

## HTTP Endpoints

All mounted under the existing `v1.Use(accounts.CheckAuth)` group. All return **top-level arrays** or a single object — no envelope — so they slot directly into FE `Promise.all` calls.

```
GET /api/v1/search/hotels
    ?city=Paris&check_in=2026-06-01&check_out=2026-06-05
    &guests=2&currency=EUR&max_price=200
    -> [HotelOffer]

GET /api/v1/search/flights
    ?from=DEL&to=CDG&depart=2026-06-01&return=2026-06-08
    &pax=2&cabin=economy&currency=EUR
    -> [FlightOffer]

GET /api/v1/routes
    ?stops=2.35,48.85;4.83,45.76&mode=drive
    -> Route
```

A 60-second in-memory LRU cache wraps each handler, keyed on the full canonicalized query string. Amadeus rate limits are low and identical queries are common during a single planning session.

## Gemini Tool Integration

Modify `trips/gemini_provider.go` (or extract into `trips/gemini_agent.go` if it grows past ~600 lines):

1. **Bump model** to `gemini-2.5-flash` (default) or `gemini-2.5-pro` (premium toggle via `GEMINI_MODEL` env).
2. **Switch to `response_schema`** on the final turn for strict JSON output. Delete the `cleanJSONResponse` regex hack.
3. **Declare 5 tools** that Gemini may call:

| Tool name           | Backed by                                        |
|---------------------|--------------------------------------------------|
| `search_flights`    | `flights.Provider.SearchFlights`                 |
| `search_hotels`     | `hotels.Provider.SearchHotels`                   |
| `get_route`         | `routes.Provider.GetRoute`                       |
| `geocode`           | existing Mapbox autocomplete (`places.MapboxApi`)|
| `web_search`        | Gemini grounding with Google Search              |

4. **Tool loop**: max 8 turns. Each tool result is appended to context. Final turn must produce schema-validated JSON. If the tool budget is exhausted before a final answer, the loop returns whatever last partial JSON was produced and logs a warning — never throws.

5. **Inject results** into the canonical `TripGenerationResponse`:
   - per-hop `route_to_next` from `get_route`
   - per-hop `suggested_hotels` (top 3) from `search_hotels`
   - per-hop `suggested_flights` (top 3, only if intercity) from `search_flights`

6. **Soft info** (visa, weather, events, "best time to visit") goes through `web_search` instead of being invented in the prompt. Citations, when present, are surfaced under a new `considerations_sources: [{title, url}]` array.

## Schema Additions

Pure additions to `trips/schema.go` types. The existing FE ignores unknown fields, so this ships without coordinated FE changes:

```go
type TripGenerationHop struct {
    // ... existing fields preserved ...
    RouteToNext      *RouteSummary  `json:"route_to_next,omitempty"`
    SuggestedHotels  []HotelOffer   `json:"suggested_hotels,omitempty"`
    SuggestedFlights []FlightOffer  `json:"suggested_flights,omitempty"`
}

type RouteSummary struct {
    DistanceKm      float64         `json:"distance_km"`
    DurationMin     int             `json:"duration_min"`
    Mode            string          `json:"mode"`
    PolylineGeoJSON json.RawMessage `json:"polyline_geojson"`
    Provider        string          `json:"provider"`
}

type TripGenerationResponse struct {
    // ... existing fields preserved ...
    ConsiderationsSources []SourceCitation `json:"considerations_sources,omitempty"`
}

type SourceCitation struct {
    Title string `json:"title"`
    URL   string `json:"url"`
}
```

`HotelOffer` / `FlightOffer` are imported from `hotels` / `flights` packages. To avoid the import cycle (`trips` already imports `core`; `hotels` and `flights` will also import `core`), the canonical types live in `hotels/` and `flights/` and the `trips` package imports them. There is no reverse import.

## Persistence

New nullable JSONB columns on `trip_hops` via Atlas migration:

```sql
ALTER TABLE trip_hops
  ADD COLUMN selected_flight  JSONB,
  ADD COLUMN selected_hotel   JSONB,
  ADD COLUMN route_to_next    JSONB;
```

Stored as raw JSON; not normalized. We do not query inside them in v1 — pure passthrough so the FE can render the chosen offer without re-hitting the provider after `POST /trip/generate/confirm`.

`CreateTripFromAIGeneration` (in `ai_controllers.go`) is updated to copy the three fields from each `TripGenerationHop` into the corresponding `TripHop` row. The existing transaction structure does not change.

## App Wiring

In `app.go`, after the existing trips routes:

```go
v1.Use(accounts.CheckAuth)
// ... existing ...
hotels.RouterGroupHotels(v1.Group("/search/hotels"))
flights.RouterGroupFlights(v1.Group("/search/flights"))
routes.RouterGroupRoutes(v1.Group("/routes"))
```

## Environment Variables

Additions to `.env_sample`:

```bash
# Amadeus (free self-service: flights + hotels)
AMADEUS_API_KEY=
AMADEUS_API_SECRET=
AMADEUS_ENV=test                # test | production

# Google Maps (Routes API; reuses existing GOOGLE_API_KEY if scoped to Routes)
GOOGLE_MAPS_API_KEY=

# Provider selection (defaults shown)
HOTELS_PROVIDER=amadeus         # amadeus | mock
FLIGHTS_PROVIDER=amadeus        # amadeus | mock
ROUTES_PROVIDER=google          # google | mapbox | mock

# Gemini tuning
GEMINI_MODEL=gemini-2.5-flash   # gemini-2.5-flash | gemini-2.5-pro
GEMINI_TOOL_BUDGET=8            # max tool-call turns per generate
GEMINI_GROUNDING_ENABLED=true   # toggle web_search tool
```

`MAPBOX_TOKEN` (existing as `core.SEARCH_API_KEY`) remains required for geocode and the Mapbox routes fallback.

## Build Order

Dependency-correct sequence:

1. `core/amadeus_client.go` — token cache + shared HTTP client.
2. `routes/` end-to-end — smallest external surface, easiest to test live (Google Routes is a single HTTP call, no OAuth dance).
3. `hotels/` end-to-end — exercises the Amadeus client.
4. `flights/` end-to-end — reuses Amadeus client.
5. Atlas migration adding the three JSONB columns to `trip_hops`.
6. `trips/schema.go` — additive type changes.
7. `trips/gemini_provider.go` — model bump, `response_schema`, tool registration, tool loop.
8. `trips/ai_controllers.go` — persist new JSONB fields in `CreateTripFromAIGeneration`.
9. `app.go` — register new route groups.
10. `.env_sample` — new variables documented.

Each step is independently testable. Steps 1–4 ship `mock` providers as the ship-blocker for the abstraction — flipping `HOTELS_PROVIDER=mock` must produce deterministic fixture data without any external network call.

## Acceptance Criteria

Phase 1 is done when:

1. `POST /trip/generate` for "Delhi → Paris, 5 days, $2000" returns a `TripGenerationResponse` where every hop carries non-empty `suggested_flights`, `suggested_hotels`, and `route_to_next`, and the existing FE renders the plan unchanged.
2. `POST /trip/generate/confirm` persists `selected_flight`, `selected_hotel`, and `route_to_next` JSONB on the resulting `TripHop` rows when those fields are populated in the request body.
3. `GET /api/v1/search/hotels`, `/search/flights`, `/routes` each return real data against the Amadeus free tier and Google Routes free quota.
4. Setting any of `HOTELS_PROVIDER=mock`, `FLIGHTS_PROVIDER=mock`, `ROUTES_PROVIDER=mock` makes the corresponding handler and Gemini tool return fixture data with no external network calls — verified by running the test suite with no API keys set.
5. The unit test for `trips/gemini_provider.go` exercises the tool loop against `mock` providers and asserts that the final response carries provider-sourced fields.

## Risks & Open Questions

- **IATA code resolution.** The FE wizard collects free-text "Delhi" / "Paris", not airport codes. Gemini will need to resolve city → primary IATA before calling `search_flights`. Either: rely on the model's own knowledge (cheap, occasionally wrong), or add a static city→IATA lookup table in `flights/iata.go`. Recommend the table; ~500 entries cover the long tail of the world's commercial airports.
- **Amadeus city codes for hotels** are not always IATA-aligned (some cities use distinct hotel codes). The Amadeus reference-data API can resolve them; for v1, pass the city name and let Amadeus's city-search endpoint do the lookup.
- **Tool-loop cost ceiling.** Eight tool turns × Gemini Flash is cheap, but a misbehaving model could call `search_hotels` repeatedly. The `GEMINI_TOOL_BUDGET` env caps it; per-user rate limiting on `POST /trip/generate` is out of scope for Phase 1 but worth tracking.
- **Currency normalization.** Amadeus returns offers in the requested currency when supported; for unsupported pairs it returns the property's native currency. The FE expects a single trip-level currency. Decision: pass through provider currency on each offer, do not convert server-side. The FE displays the offer currency next to the amount and converts only at display time using the existing `utils/currency.js` rates.
- **FE city-suggestions render bug** (renders objects as `[object Object]`) is pre-existing and out of scope here, but Phase 1 should not regress it. The existing `GetMultiCitySuggestions` response shape is preserved verbatim.

## Out of Scope (deferred to later phases)

- SSE / WebSocket streaming of tool progress.
- New flights tab in the FE.
- Booking handoff via provider booking APIs.
- ADK / Vertex Agent runtime.
- Per-user rate limiting on AI endpoints.
- Multi-provider blending (e.g., merging Amadeus + SerpAPI hotel results).
