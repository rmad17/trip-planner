package flights

import "strings"

// cityToIATA is a small lookup of common city names to their primary IATA code.
// Designed to cover the long tail of typical user inputs without a network call.
// Populated based on city name (lower-cased, ascii-folded by the lookup func).
var cityToIATA = map[string]string{
	// Europe
	"paris":      "CDG",
	"london":     "LHR",
	"madrid":     "MAD",
	"barcelona":  "BCN",
	"rome":       "FCO",
	"milan":      "MXP",
	"amsterdam":  "AMS",
	"berlin":     "BER",
	"munich":     "MUC",
	"frankfurt":  "FRA",
	"zurich":     "ZRH",
	"vienna":     "VIE",
	"prague":     "PRG",
	"lisbon":     "LIS",
	"dublin":     "DUB",
	"brussels":   "BRU",
	"copenhagen": "CPH",
	"oslo":       "OSL",
	"stockholm":  "ARN",
	"helsinki":   "HEL",
	"athens":     "ATH",
	"istanbul":   "IST",
	"moscow":     "SVO",

	// North America
	"new york":      "JFK",
	"newyork":       "JFK",
	"nyc":           "JFK",
	"los angeles":   "LAX",
	"san francisco": "SFO",
	"chicago":       "ORD",
	"miami":         "MIA",
	"boston":        "BOS",
	"seattle":       "SEA",
	"washington":    "IAD",
	"atlanta":       "ATL",
	"dallas":        "DFW",
	"houston":       "IAH",
	"toronto":       "YYZ",
	"vancouver":     "YVR",
	"montreal":      "YUL",
	"mexico city":   "MEX",

	// South America
	"sao paulo":      "GRU",
	"rio de janeiro": "GIG",
	"buenos aires":   "EZE",
	"santiago":       "SCL",
	"lima":           "LIM",
	"bogota":         "BOG",

	// Asia
	"delhi":     "DEL",
	"new delhi": "DEL",
	"mumbai":    "BOM",
	"bangalore": "BLR",
	"bengaluru": "BLR",
	"chennai":   "MAA",
	"kolkata":   "CCU",
	"hyderabad": "HYD",
	"goa":       "GOI",
	"jaipur":    "JAI",
	"kochi":     "COK",
	"colombo":   "CMB",
	"kathmandu": "KTM",
	"dhaka":     "DAC",

	"tokyo":     "HND",
	"osaka":     "KIX",
	"seoul":     "ICN",
	"beijing":   "PEK",
	"shanghai":  "PVG",
	"hong kong": "HKG",
	"taipei":    "TPE",
	"singapore": "SIN",
	"bangkok":   "BKK",
	"phuket":    "HKT",
	"bali":      "DPS",
	"denpasar":  "DPS",
	"jakarta":   "CGK",
	"manila":    "MNL",
	"hanoi":     "HAN",
	"ho chi minh city": "SGN",
	"saigon":           "SGN",
	"kuala lumpur":     "KUL",

	"dubai":     "DXB",
	"abu dhabi": "AUH",
	"doha":      "DOH",
	"riyadh":    "RUH",
	"jeddah":    "JED",
	"muscat":    "MCT",
	"tel aviv":  "TLV",

	// Africa
	"cairo":        "CAI",
	"johannesburg": "JNB",
	"cape town":    "CPT",
	"nairobi":      "NBO",
	"lagos":        "LOS",
	"casablanca":   "CMN",
	"addis ababa":  "ADD",

	// Oceania
	"sydney":    "SYD",
	"melbourne": "MEL",
	"brisbane":  "BNE",
	"perth":     "PER",
	"auckland":  "AKL",
}

// LookupIATA returns the primary IATA code for the given input.
// If input already looks like a 3-letter IATA code, it's returned upper-cased.
// Otherwise the lookup is case- and whitespace-insensitive.
func LookupIATA(input string) (string, bool) {
	s := strings.TrimSpace(input)
	if s == "" {
		return "", false
	}
	if len(s) == 3 {
		return strings.ToUpper(s), true
	}
	key := strings.ToLower(s)
	if code, ok := cityToIATA[key]; ok {
		return code, true
	}
	return "", false
}
