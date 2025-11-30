#!/bin/bash
set -e

# Change to the directory where this script is located
cd "$(dirname "$0")"

echo "🚀 Starting deployment..."

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if .env exists
if [ ! -f .env ]; then
    echo -e "${RED}❌ Error: .env file not found!${NC}"
    echo "Please create .env file from .env.example"
    exit 1
fi

# Load environment variables
source .env

echo -e "${YELLOW}📦 Pulling latest code...${NC}"
git pull origin main

echo -e "${YELLOW}🛑 Stopping existing containers...${NC}"
docker-compose down

echo -e "${YELLOW}🏗️  Building containers...${NC}"
docker-compose build --no-cache

echo -e "${YELLOW}🗄️  Running database migrations...${NC}"
docker-compose up migration

echo -e "${YELLOW}🚀 Starting containers...${NC}"
docker-compose up -d

echo -e "${YELLOW}⏳ Waiting for services to be healthy...${NC}"
sleep 10

# Check if services are running
if docker-compose ps | grep -q "Up"; then
    echo -e "${GREEN}✅ Services are running!${NC}"
else
    echo -e "${RED}❌ Some services failed to start${NC}"
    docker-compose logs --tail=50
    exit 1
fi

echo -e "${YELLOW}📊 Service Status:${NC}"
docker-compose ps

echo -e "${YELLOW}📝 Recent logs:${NC}"
docker-compose logs --tail=20

echo -e "${GREEN}✅ Deployment complete!${NC}"
echo -e "${GREEN}API is accessible at: https://api.yourdomain.com${NC}"
