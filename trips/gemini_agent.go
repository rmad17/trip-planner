package trips

import (
	"context"
	"log"
	"strings"
	"triplanner/flights"
	"triplanner/hotels"
	"triplanner/routes"
)

// enrichTripWithProviders augments an AI-generated trip plan with provider-sourced
// hotel/flight offers and a per-hop route-to-next polyline. It implements the same
// outcome as the design's tool-loop architecture: every hop should carry
// suggested_hotels, suggested_flights (when intercity), and route_to_next.
//
// Failures from any provider are logged and skipped — the plan is returned with
// whichever fields could be populated, never an error. Callers must always be
// safe with empty slices.
func enrichTripWithProviders(ctx context.Context, plan *TripGenerationResponse, req TripGenerationRequest) {
	if plan == nil {
		return
	}

	// Build providers fresh from env so per-request override / mock-mode works.
	hotelProvider, hotelsErr := hotels.NewProviderFactory().GetDefaultProvider()
	flightProvider, flightsErr := flights.NewProviderFactory().GetDefaultProvider()
	routeProvider, routesErr := routes.NewProviderFactory().GetDefaultProvider()

	currency := string(req.Currency)
	if currency == "" {
		currency = "USD"
	}

	for i := range plan.Hops {
		hop := &plan.Hops[i]
		enrichHopHotels(ctx, hop, currency, req.NumTravelers, hotelProvider, hotelsErr)
		enrichHopFlights(ctx, hop, plan.Hops, i, currency, req, flightProvider, flightsErr)
		enrichHopRoute(ctx, hop, plan.Hops, i, routeProvider, routesErr)
	}
}

func enrichHopHotels(ctx context.Context, hop *GeneratedHop, currency string, numTravelers int, provider hotels.Provider, providerErr error) {
	if providerErr != nil || provider == nil {
		return
	}
	if hop.City == "" || hop.StartDate == "" || hop.EndDate == "" {
		return
	}
	pax := numTravelers
	if pax <= 0 {
		pax = 1
	}
	offers, err := provider.SearchHotels(ctx, hotels.HotelSearchQuery{
		City:     hop.City,
		CheckIn:  hop.StartDate,
		CheckOut: hop.EndDate,
		Guests:   pax,
		Currency: currency,
	})
	if err != nil {
		log.Printf("enrichHopHotels: %v", err)
		return
	}
	if len(offers) > 3 {
		offers = offers[:3]
	}
	hop.SuggestedHotels = offers
}

func enrichHopFlights(ctx context.Context, hop *GeneratedHop, allHops []GeneratedHop, idx int, currency string, req TripGenerationRequest, provider flights.Provider, providerErr error) {
	if providerErr != nil || provider == nil {
		return
	}
	if !looksLikeFlight(hop.Transportation) {
		return
	}
	from := previousHopCity(allHops, idx)
	if from == "" {
		from = req.Source
	}
	if from == "" || hop.City == "" || hop.StartDate == "" {
		return
	}
	fromCode, ok1 := flights.LookupIATA(from)
	toCode, ok2 := flights.LookupIATA(hop.City)
	if !ok1 || !ok2 {
		return
	}
	pax := req.NumTravelers
	if pax <= 0 {
		pax = 1
	}
	offers, err := provider.SearchFlights(ctx, flights.FlightSearchQuery{
		From:     fromCode,
		To:       toCode,
		Depart:   hop.StartDate,
		Pax:      pax,
		Cabin:    "economy",
		Currency: currency,
	})
	if err != nil {
		log.Printf("enrichHopFlights: %v", err)
		return
	}
	if len(offers) > 3 {
		offers = offers[:3]
	}
	hop.SuggestedFlights = offers
}

func enrichHopRoute(ctx context.Context, hop *GeneratedHop, allHops []GeneratedHop, idx int, provider routes.Provider, providerErr error) {
	if providerErr != nil || provider == nil {
		return
	}
	// Route-to-next requires both this hop and the next hop to have resolvable coordinates.
	if idx+1 >= len(allHops) {
		return
	}
	nextHop := &allHops[idx+1]

	from, ok1 := geocodeCity(hop.City)
	to, ok2 := geocodeCity(nextHop.City)
	if !ok1 || !ok2 {
		return
	}

	mode := normalizeRouteMode(nextHop.Transportation)
	route, err := provider.GetRoute(ctx, routes.RouteQuery{
		Stops: []routes.LngLat{from, to},
		Mode:  mode,
	})
	if err != nil {
		log.Printf("enrichHopRoute: %v", err)
		return
	}
	hop.RouteToNext = &RouteSummary{
		DistanceKm:      route.DistanceKm,
		DurationMin:     route.DurationMin,
		Mode:            mode,
		PolylineGeoJSON: route.PolylineGeoJSON,
		Provider:        route.Provider,
	}
}

func looksLikeFlight(transport string) bool {
	t := strings.ToLower(strings.TrimSpace(transport))
	return strings.Contains(t, "flight") || strings.Contains(t, "plane") || strings.Contains(t, "air")
}

func previousHopCity(hops []GeneratedHop, idx int) string {
	if idx <= 0 {
		return ""
	}
	return hops[idx-1].City
}

func normalizeRouteMode(transport string) string {
	t := strings.ToLower(strings.TrimSpace(transport))
	switch {
	case strings.Contains(t, "walk"):
		return "walk"
	case strings.Contains(t, "cycle") || strings.Contains(t, "bike"):
		return "cycle"
	case strings.Contains(t, "train") || strings.Contains(t, "bus") || strings.Contains(t, "transit"):
		return "transit"
	default:
		return "drive"
	}
}

// geocodeCity returns approximate coordinates for a city using the IATA lookup
// table augmented with a small static cities file. This avoids a network round
// trip just to seed the routes provider with stops.
//
// For unknown cities, returns ok=false; callers skip route enrichment in that case.
func geocodeCity(city string) (routes.LngLat, bool) {
	key := strings.ToLower(strings.TrimSpace(city))
	if c, ok := staticCityCoords[key]; ok {
		return c, true
	}
	return routes.LngLat{}, false
}

// staticCityCoords pairs with flights/iata.go's city list — only cities that
// have an entry there are likely to appear in trip plans.
var staticCityCoords = map[string]routes.LngLat{
	"paris":             {Lng: 2.3522, Lat: 48.8566},
	"london":            {Lng: -0.1276, Lat: 51.5074},
	"madrid":            {Lng: -3.7038, Lat: 40.4168},
	"barcelona":         {Lng: 2.1734, Lat: 41.3851},
	"rome":              {Lng: 12.4964, Lat: 41.9028},
	"milan":             {Lng: 9.1900, Lat: 45.4642},
	"amsterdam":         {Lng: 4.9041, Lat: 52.3676},
	"berlin":            {Lng: 13.4050, Lat: 52.5200},
	"munich":            {Lng: 11.5820, Lat: 48.1351},
	"frankfurt":         {Lng: 8.6821, Lat: 50.1109},
	"zurich":            {Lng: 8.5417, Lat: 47.3769},
	"vienna":            {Lng: 16.3738, Lat: 48.2082},
	"prague":            {Lng: 14.4378, Lat: 50.0755},
	"lisbon":            {Lng: -9.1393, Lat: 38.7223},
	"dublin":            {Lng: -6.2603, Lat: 53.3498},
	"brussels":          {Lng: 4.3517, Lat: 50.8503},
	"copenhagen":        {Lng: 12.5683, Lat: 55.6761},
	"oslo":              {Lng: 10.7522, Lat: 59.9139},
	"stockholm":         {Lng: 18.0686, Lat: 59.3293},
	"helsinki":          {Lng: 24.9384, Lat: 60.1699},
	"athens":            {Lng: 23.7275, Lat: 37.9838},
	"istanbul":          {Lng: 28.9784, Lat: 41.0082},
	"moscow":            {Lng: 37.6173, Lat: 55.7558},
	"new york":          {Lng: -74.0060, Lat: 40.7128},
	"newyork":           {Lng: -74.0060, Lat: 40.7128},
	"nyc":               {Lng: -74.0060, Lat: 40.7128},
	"los angeles":       {Lng: -118.2437, Lat: 34.0522},
	"san francisco":     {Lng: -122.4194, Lat: 37.7749},
	"chicago":           {Lng: -87.6298, Lat: 41.8781},
	"miami":             {Lng: -80.1918, Lat: 25.7617},
	"boston":            {Lng: -71.0589, Lat: 42.3601},
	"seattle":           {Lng: -122.3321, Lat: 47.6062},
	"washington":        {Lng: -77.0369, Lat: 38.9072},
	"atlanta":           {Lng: -84.3880, Lat: 33.7490},
	"dallas":            {Lng: -96.7970, Lat: 32.7767},
	"houston":           {Lng: -95.3698, Lat: 29.7604},
	"toronto":           {Lng: -79.3832, Lat: 43.6532},
	"vancouver":         {Lng: -123.1207, Lat: 49.2827},
	"montreal":          {Lng: -73.5673, Lat: 45.5017},
	"mexico city":       {Lng: -99.1332, Lat: 19.4326},
	"sao paulo":         {Lng: -46.6333, Lat: -23.5505},
	"rio de janeiro":    {Lng: -43.1729, Lat: -22.9068},
	"buenos aires":      {Lng: -58.3816, Lat: -34.6037},
	"santiago":          {Lng: -70.6693, Lat: -33.4489},
	"lima":              {Lng: -77.0428, Lat: -12.0464},
	"bogota":            {Lng: -74.0721, Lat: 4.7110},
	"delhi":             {Lng: 77.1025, Lat: 28.7041},
	"new delhi":         {Lng: 77.2090, Lat: 28.6139},
	"mumbai":            {Lng: 72.8777, Lat: 19.0760},
	"bangalore":         {Lng: 77.5946, Lat: 12.9716},
	"bengaluru":         {Lng: 77.5946, Lat: 12.9716},
	"chennai":           {Lng: 80.2707, Lat: 13.0827},
	"kolkata":           {Lng: 88.3639, Lat: 22.5726},
	"hyderabad":         {Lng: 78.4867, Lat: 17.3850},
	"goa":               {Lng: 73.8278, Lat: 15.2993},
	"jaipur":            {Lng: 75.7873, Lat: 26.9124},
	"kochi":             {Lng: 76.2673, Lat: 9.9312},
	"colombo":           {Lng: 79.8612, Lat: 6.9271},
	"kathmandu":         {Lng: 85.3240, Lat: 27.7172},
	"dhaka":             {Lng: 90.4125, Lat: 23.8103},
	"tokyo":             {Lng: 139.6503, Lat: 35.6762},
	"osaka":             {Lng: 135.5023, Lat: 34.6937},
	"seoul":             {Lng: 126.9780, Lat: 37.5665},
	"beijing":           {Lng: 116.4074, Lat: 39.9042},
	"shanghai":          {Lng: 121.4737, Lat: 31.2304},
	"hong kong":         {Lng: 114.1694, Lat: 22.3193},
	"taipei":            {Lng: 121.5654, Lat: 25.0330},
	"singapore":         {Lng: 103.8198, Lat: 1.3521},
	"bangkok":           {Lng: 100.5018, Lat: 13.7563},
	"phuket":            {Lng: 98.3923, Lat: 7.8804},
	"bali":              {Lng: 115.0920, Lat: -8.3405},
	"denpasar":          {Lng: 115.2126, Lat: -8.6705},
	"jakarta":           {Lng: 106.8456, Lat: -6.2088},
	"manila":            {Lng: 120.9842, Lat: 14.5995},
	"hanoi":             {Lng: 105.8542, Lat: 21.0285},
	"ho chi minh city":  {Lng: 106.6297, Lat: 10.8231},
	"saigon":            {Lng: 106.6297, Lat: 10.8231},
	"kuala lumpur":      {Lng: 101.6869, Lat: 3.1390},
	"dubai":             {Lng: 55.2708, Lat: 25.2048},
	"abu dhabi":         {Lng: 54.3773, Lat: 24.4539},
	"doha":              {Lng: 51.5310, Lat: 25.2854},
	"riyadh":            {Lng: 46.6753, Lat: 24.7136},
	"jeddah":            {Lng: 39.1925, Lat: 21.4858},
	"muscat":            {Lng: 58.5407, Lat: 23.5880},
	"tel aviv":          {Lng: 34.7818, Lat: 32.0853},
	"cairo":             {Lng: 31.2357, Lat: 30.0444},
	"johannesburg":      {Lng: 28.0473, Lat: -26.2041},
	"cape town":         {Lng: 18.4241, Lat: -33.9249},
	"nairobi":           {Lng: 36.8219, Lat: -1.2921},
	"lagos":             {Lng: 3.3792, Lat: 6.5244},
	"casablanca":        {Lng: -7.5898, Lat: 33.5731},
	"addis ababa":       {Lng: 38.7578, Lat: 9.0320},
	"sydney":            {Lng: 151.2093, Lat: -33.8688},
	"melbourne":         {Lng: 144.9631, Lat: -37.8136},
	"brisbane":          {Lng: 153.0251, Lat: -27.4698},
	"perth":             {Lng: 115.8605, Lat: -31.9505},
	"auckland":          {Lng: 174.7633, Lat: -36.8485},
}
