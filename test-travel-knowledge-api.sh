#!/bin/bash

# Test script for Travel Knowledge API
# Make sure the server is running on localhost:8080

BASE_URL="http://localhost:8080/api/v1/travel-knowledge"

echo "========================================="
echo "Testing Travel Knowledge API"
echo "========================================="
echo ""

# Test 1: Get available categories
echo "1. Getting available categories..."
curl -s "$BASE_URL/categories" | jq '.'
echo ""
echo ""

# Test 2: List all cities
echo "2. Listing all cities..."
curl -s "$BASE_URL/cities" | jq '.'
echo ""
echo ""

# Test 3: Get city info for Ayodhya
echo "3. Getting city info for Ayodhya..."
curl -s "$BASE_URL/city/Ayodhya/info" | jq '.city_name, .city_country, .total_docs'
echo ""
echo ""

# Test 4: Get documents for a city (all types)
echo "4. Getting all documents for Dwarka..."
curl -s "$BASE_URL/city?city=Dwarka" | jq '.city_name, .total_count, .categories | keys'
echo ""
echo ""

# Test 5: Get documents for a city (filtered by type)
echo "5. Getting food documents for Ayodhya..."
curl -s "$BASE_URL/city?city=Ayodhya&type=food" | jq '.city_name, .total_count, .categories.food[0].title'
echo ""
echo ""

# Test 6: Get all documents of a specific type
echo "6. Getting all historical documents..."
curl -s "$BASE_URL/category/historical" | jq '.category, .count'
echo ""
echo ""

# Test 7: Get knowledge base statistics
echo "7. Getting knowledge base statistics..."
curl -s "$BASE_URL/stats" | jq '.'
echo ""
echo ""

echo "========================================="
echo "All tests completed!"
echo "========================================="
