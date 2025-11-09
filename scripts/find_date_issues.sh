#!/bin/bash

# Script to find potential Date field assignment issues after migrating to core.Date

echo "=========================================="
echo "Searching for potential Date field issues"
echo "=========================================="
echo ""

# Find all Go files in trips directory (excluding tests)
for file in trips/*.go; do
    # Skip test files
    if [[ $file == *"_test.go" ]]; then
        continue
    fi

    # Check if file exists and is readable
    if [[ ! -f "$file" ]]; then
        continue
    fi

    echo "Checking $file..."

    # Look for patterns that might indicate Date assignment issues
    # Pattern 1: Date: followed by time. or a variable that's likely time.Time
    grep -n "Date:\s*\(time\.\|.*Date[^{]\|.*Parse\)" "$file" | while read -r line; do
        line_num=$(echo "$line" | cut -d: -f1)
        content=$(echo "$line" | cut -d: -f2-)

        # Check if it's not already using core.Date
        if ! echo "$content" | grep -q "core.Date"; then
            echo "  ⚠️  Line $line_num: Potential issue"
            echo "      $content"
            echo ""
        fi
    done

    # Pattern 2: tripDay.Date = (assignment after creation)
    grep -n "\.Date\s*=" "$file" | while read -r line; do
        line_num=$(echo "$line" | cut -d: -f1)
        content=$(echo "$line" | cut -d: -f2-)

        # Check if it's not already using core.Date
        if ! echo "$content" | grep -q "core.Date"; then
            echo "  ⚠️  Line $line_num: Potential assignment issue"
            echo "      $content"
            echo ""
        fi
    done
done

echo ""
echo "=========================================="
echo "How to fix:"
echo "=========================================="
echo "Replace: Date: dayDate"
echo "With:    Date: core.Date{Time: dayDate}"
echo ""
echo "Make sure to import: \"triplanner/core\""
echo ""
echo "See docs/DATE_MIGRATION_GUIDE.md for detailed examples"
echo "=========================================="
