# File Path Structure

## Overview

All uploaded files are stored using a structured, hierarchical path format that organizes files by user, trip, and category. This makes files easy to locate, manage, and debug.

## Path Format

```
{username}/{trip_name}/{trip_id}/{category}/{file_name}
```

### Components

1. **username**: The sanitized username of the file uploader
2. **trip_name**: The sanitized name of the trip
3. **trip_id**: The UUID of the trip (not sanitized, kept as-is)
4. **category**: The document category (tickets, invoices, etc.)
5. **file_name**: The sanitized name provided by the user with file extension

## Path Examples

### Example 1: Flight Ticket
```
User: john.doe@example.com (username: "johndoe")
Trip: "Summer Vacation 2024"
Trip ID: 123e4567-e89b-12d3-a456-426614174000
Category: tickets
File Name: "Flight to Paris"
Original Filename: ticket_ABC123.pdf

Resulting Path:
johndoe/summer-vacation-2024/123e4567-e89b-12d3-a456-426614174000/tickets/flight-to-paris.pdf
```

### Example 2: Hotel Invoice
```
User: jane_smith
Trip: "Business Trip NYC"
Trip ID: 987e6543-e21b-12d3-a456-426614174000
Category: invoices
File Name: "Hilton Invoice"
Original Filename: invoice-2024-01.pdf

Resulting Path:
jane_smith/business-trip-nyc/987e6543-e21b-12d3-a456-426614174000/invoices/hilton-invoice.pdf
```

### Example 3: Passport Copy
```
User: alice-cooper
Trip: "Around the World!"
Trip ID: a1b2c3d4-e5f6-7890-abcd-ef1234567890
Category: identity_proofs
File Name: "Passport"
Original Filename: scan_0001.jpg

Resulting Path:
alice-cooper/around-the-world/a1b2c3d4-e5f6-7890-abcd-ef1234567890/identity_proofs/passport.jpg
```

## Sanitization Rules

All path components (except trip_id) are sanitized for filesystem safety:

### Applied Transformations

1. **Convert to lowercase**
   - Input: `"Summer Vacation"`
   - Output: `"summer vacation"`

2. **Replace spaces with hyphens**
   - Input: `"summer vacation"`
   - Output: `"summer-vacation"`

3. **Remove special characters**
   - Keep only: `a-z`, `0-9`, `-`, `_`
   - Input: `"Trip! @2024#"`
   - Output: `"trip-2024"`

4. **Remove consecutive hyphens**
   - Input: `"trip---2024"`
   - Output: `"trip-2024"`

5. **Trim hyphens from edges**
   - Input: `"-my-trip-"`
   - Output: `"my-trip"`

6. **Handle empty results**
   - If sanitization results in empty string, use default:
     - For trip name: `"untitled-trip"`
     - For file name: `"untitled"`

### Examples

| Original | Sanitized |
|----------|-----------|
| `John Doe` | `john-doe` |
| `Trip to Paris!` | `trip-to-paris` |
| `Summer_2024` | `summer_2024` |
| `Business Trip (NYC)` | `business-trip-nyc` |
| `@#$%` | `untitled` |
| `My---Trip` | `my-trip` |

## File Extension Handling

The system automatically handles file extensions:

1. **Extension from original file** is preserved
2. **Added automatically** if not present in user-provided name
3. **Not duplicated** if already present

### Examples

| User Name | Original File | Final Filename |
|-----------|---------------|----------------|
| `"Flight Ticket"` | `ticket.pdf` | `flight-ticket.pdf` |
| `"Invoice.pdf"` | `inv_001.pdf` | `invoice.pdf` |
| `"Passport"` | `scan.jpg` | `passport.jpg` |

## Full Path Examples by Category

### Tickets
```
johndoe/paris-trip/123e4567-e89b-12d3-a456-426614174000/tickets/flight-ticket.pdf
johndoe/paris-trip/123e4567-e89b-12d3-a456-426614174000/tickets/train-pass.pdf
```

### Invoices
```
johndoe/paris-trip/123e4567-e89b-12d3-a456-426614174000/invoices/hotel-receipt.pdf
johndoe/paris-trip/123e4567-e89b-12d3-a456-426614174000/invoices/restaurant-bill.jpg
```

### Identity Proofs
```
johndoe/paris-trip/123e4567-e89b-12d3-a456-426614174000/identity_proofs/passport.jpg
johndoe/paris-trip/123e4567-e89b-12d3-a456-426614174000/identity_proofs/drivers-license.jpg
```

### Medical Documents
```
johndoe/paris-trip/123e4567-e89b-12d3-a456-426614174000/medical/vaccination-card.pdf
johndoe/paris-trip/123e4567-e89b-12d3-a456-426614174000/medical/travel-insurance.pdf
```

### Hotel Bookings
```
johndoe/paris-trip/123e4567-e89b-12d3-a456-426614174000/hotel_bookings/hilton-confirmation.pdf
johndoe/paris-trip/123e4567-e89b-12d3-a456-426614174000/hotel_bookings/airbnb-booking.pdf
```

## Benefits of This Structure

### 1. Organization
- Files grouped by user
- Further grouped by trip
- Categorized by purpose

### 2. Easy Navigation
Navigate through storage hierarchically:
```
my-bucket/
├── johndoe/
│   ├── paris-trip/
│   │   └── 123e4567-.../
│   │       ├── tickets/
│   │       ├── invoices/
│   │       └── hotel_bookings/
│   └── london-trip/
│       └── 987e6543-.../
└── janedoe/
    └── tokyo-trip/
```

### 3. Debugging
Easily identify file owner and trip from path alone

### 4. Backup & Migration
Simple to:
- Backup all files for a specific user
- Export all files for a specific trip
- Filter by category

### 5. Access Control
Path structure makes it easy to implement:
- User-based access (all files under `{username}/`)
- Trip-based access (all files under `{username}/{trip_name}/{trip_id}/`)
- Category-based filtering

## Storage Provider Support

This path structure works with all supported storage providers:

### DigitalOcean Spaces
```
https://my-bucket.nyc3.digitaloceanspaces.com/johndoe/paris-trip/123e4567-.../tickets/flight-ticket.pdf
```

### AWS S3
```
s3://my-bucket/johndoe/paris-trip/123e4567-.../tickets/flight-ticket.pdf
```

### Google Cloud Storage
```
gs://my-bucket/johndoe/paris-trip/123e4567-.../tickets/flight-ticket.pdf
```

### Azure Blob Storage
```
https://myaccount.blob.core.windows.net/my-bucket/johndoe/paris-trip/123e4567-.../tickets/flight-ticket.pdf
```

### Local Storage
```
./uploads/johndoe/paris-trip/123e4567-.../tickets/flight-ticket.pdf
```

## Implementation Details

### Code Location
- **Path Building**: `documents/controllers.go:259-279`
- **Sanitization Function**: `documents/controllers.go:559-585`

### Database Storage
The full path is stored in the `documents` table:

```sql
SELECT storage_provider, storage_path FROM documents;
```

| storage_provider | storage_path |
|------------------|--------------|
| digitalocean | johndoe/paris-trip/123e.../tickets/flight.pdf |
| local | johndoe/london-trip/987e.../invoices/hotel.pdf |

## API Usage

### Upload with Structured Path
```bash
curl -X POST http://localhost:8080/api/v1/trip/{trip_id}/documents \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "file=@flight_ticket.pdf" \
  -F "name=Flight to Paris" \
  -F "category=tickets"
```

**Generated Path:**
```
{username}/paris-vacation/123e4567-.../tickets/flight-to-paris.pdf
```

### Download Using Stored Path
The system automatically retrieves files using the stored `storage_path`:

```bash
curl http://localhost:8080/api/v1/documents/{document_id}/download \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Path Uniqueness

### Handling Duplicate Names

If multiple files with the same name are uploaded to the same category:

**First Upload:**
```
name: "Flight Ticket"
path: johndoe/paris-trip/123e.../tickets/flight-ticket.pdf
```

**Second Upload (same name):**
The second upload will **overwrite** the first file in cloud storage but create a new database record.

### Recommendation

To avoid overwrites, users should:
1. Use unique names: "Flight Ticket 1", "Flight Ticket 2"
2. Include dates: "Flight Ticket 2024-01-15"
3. Include identifiers: "Flight Ticket AA123"

### Future Enhancement

Consider adding timestamp or UUID suffix for automatic uniqueness:
```
flight-ticket-1704067200.pdf
flight-ticket-a1b2c3d4.pdf
```

## Migration from Old Path Format

### Old Format
```
documents/{uuid}_{timestamp}.pdf
```

### New Format
```
{username}/{trip_name}/{trip_id}/{category}/{file_name}.pdf
```

### Backward Compatibility
The system supports both formats:
- New uploads use the structured format
- Old files remain accessible with their original paths
- Download/delete operations work with both formats

## Troubleshooting

### Path Too Long
Some storage systems have path length limits:
- **DigitalOcean Spaces**: 1024 characters (max key length)
- **AWS S3**: 1024 characters
- **Local Filesystem**: 255 characters per component, 4096 total (Linux)

**Solution:** Keep names concise or truncate during sanitization

### Special Characters in Names
If a trip or file name contains only special characters:
```
Trip Name: "@#$%^&*"
Sanitized: "untitled-trip"
```

### Missing Trip Name
If trip has no name (`NULL`):
```
Default: "untitled-trip"
Path: johndoe/untitled-trip/123e.../tickets/flight.pdf
```
