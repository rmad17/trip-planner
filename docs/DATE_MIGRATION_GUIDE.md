# Date Field Migration Guide

## Problem

The `TripDay.Date` field has been changed from `time.Time` to `core.Date` to support multiple date formats (both date-only and RFC3339).

## Error Message

```
cannot use dayDate (variable of struct type time.Time) as core.Date value in struct literal
```

## Solution

When creating a `TripDay` struct, wrap any `time.Time` value in `core.Date{}`:

### Before (Error)

```go
tripDay := TripDay{
    Date: dayDate,  // dayDate is time.Time - THIS CAUSES ERROR
    DayNumber: 1,
    DayType: TripDayTypeExplore,
}
```

### After (Fixed)

```go
tripDay := TripDay{
    Date: core.Date{Time: dayDate},  // Wrap time.Time in core.Date
    DayNumber: 1,
    DayType: TripDayTypeExplore,
}
```

## Common Patterns to Fix

### Pattern 1: Direct Assignment in Struct Literal

**Before:**
```go
tripDay := TripDay{
    Date: time.Date(2025, 11, 28, 0, 0, 0, 0, time.UTC),
    // other fields...
}
```

**After:**
```go
tripDay := TripDay{
    Date: core.Date{Time: time.Date(2025, 11, 28, 0, 0, 0, 0, time.UTC)},
    // other fields...
}
```

### Pattern 2: Assignment from Variable

**Before:**
```go
dayDate := time.Parse("2006-01-02", dateStr)
tripDay := TripDay{
    Date: dayDate,
    // other fields...
}
```

**After:**
```go
dayDate, _ := time.Parse("2006-01-02", dateStr)
tripDay := TripDay{
    Date: core.Date{Time: dayDate},
    // other fields...
}
```

### Pattern 3: Assignment from Function Return

**Before:**
```go
tripDay := TripDay{
    Date: time.Now(),
    // other fields...
}
```

**After:**
```go
tripDay := TripDay{
    Date: core.Date{Time: time.Now()},
    // other fields...
}
```

### Pattern 4: Assignment After Creation

**Before:**
```go
var tripDay TripDay
tripDay.Date = time.Now()  // ERROR
```

**After:**
```go
var tripDay TripDay
tripDay.Date = core.Date{Time: time.Now()}  // CORRECT
```

### Pattern 5: Parsing Date String

**Before:**
```go
parsedDate, _ := time.Parse("2006-01-02", "2025-11-28")
tripDay.Date = parsedDate  // ERROR
```

**After:**
```go
parsedDate, _ := time.Parse("2006-01-02", "2025-11-28")
tripDay.Date = core.Date{Time: parsedDate}  // CORRECT
```

## For ai_controllers.go Specifically

Based on the error at line 248, you likely have code like:

**Current (line 248):**
```go
tripDay := TripDay{
    Date: dayDate,  // ← ERROR HERE
    // ... other fields
}
```

**Fix:**
```go
tripDay := TripDay{
    Date: core.Date{Time: dayDate},  // ← FIXED
    // ... other fields
}
```

## Quick Fix Script

If you have multiple files to fix, you can use this search pattern:

```bash
# Find all potential issues (this is a heuristic, review each match)
grep -rn "Date:\s*\(time\.\|.*Date\)" trips/ --include="*.go" | grep -v "core.Date"
```

## Important Notes

1. **JSON Unmarshaling**: When using `c.BindJSON(&tripDay)`, the conversion happens automatically - no changes needed!

2. **Reading from Database**: GORM will automatically scan `time.Time` values into `core.Date` - no changes needed!

3. **Only Manual Construction**: You only need to wrap when manually creating `TripDay` structs in Go code.

4. **Import Required**: Make sure to import `triplanner/core` at the top of your file:
   ```go
   import (
       // other imports...
       "triplanner/core"
   )
   ```

## Example Complete Fix

**Before:**
```go
package trips

import (
    "time"
)

func CreateDaysFromDates(dates []string) []TripDay {
    var days []TripDay
    for i, dateStr := range dates {
        parsedDate, _ := time.Parse("2006-01-02", dateStr)
        day := TripDay{
            Date: parsedDate,  // ERROR
            DayNumber: i + 1,
            DayType: TripDayTypeExplore,
        }
        days = append(days, day)
    }
    return days
}
```

**After:**
```go
package trips

import (
    "time"
    "triplanner/core"  // ADD THIS IMPORT
)

func CreateDaysFromDates(dates []string) []TripDay {
    var days []TripDay
    for i, dateStr := range dates {
        parsedDate, _ := time.Parse("2006-01-02", dateStr)
        day := TripDay{
            Date: core.Date{Time: parsedDate},  // FIXED
            DayNumber: i + 1,
            DayType: TripDayTypeExplore,
        }
        days = append(days, day)
    }
    return days
}
```

## Testing Your Fix

After making changes, run:

```bash
# Test compilation
go build ./...

# Run tests
go test ./trips -v
```

## Why This Change Was Made

The `core.Date` type was introduced to:
- Accept both date-only format (`"2025-11-28"`) and RFC3339 format (`"2025-11-28T15:04:05Z"`)
- Serialize dates consistently as date-only in JSON responses
- Maintain compatibility with database operations
- Fix the original error: `parsing time "2025-11-28" as "2006-01-02T15:04:05Z07:00"`

## Need Help?

If you have many occurrences to fix, share the file content and I can provide specific fixes for each location.
