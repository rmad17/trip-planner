#!/bin/bash
# Script to upload code to Digital Ocean Droplet
# Run this from your local machine

set -e

# Configuration
DROPLET_IP="your.droplet.ip.here"  # Replace with your droplet IP
DROPLET_USER="root"  # or "deploy" if using non-root user
APP_DIR="/opt/trip-planner"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${YELLOW}📤 Uploading code to Digital Ocean Droplet...${NC}"

# Check if droplet IP is set
if [ "$DROPLET_IP" = "your.droplet.ip.here" ]; then
    echo -e "${RED}❌ Please set your DROPLET_IP in the script!${NC}"
    exit 1
fi

# Create rsync exclude file
cat > .rsync-exclude <<EOF
.git/
.env
node_modules/
build/
dist/
uploads/
logs/
*.log
.DS_Store
.vscode/
.idea/
__pycache__/
*.pyc
.pytest_cache/
coverage/
.coverage
EOF

echo -e "${YELLOW}📦 Syncing files via rsync...${NC}"

# Upload code using rsync (faster and more efficient than scp)
rsync -avz --delete \
    --exclude-from='.rsync-exclude' \
    --progress \
    ./ ${DROPLET_USER}@${DROPLET_IP}:${APP_DIR}/

# Clean up
rm .rsync-exclude

echo -e "${GREEN}✅ Code uploaded successfully!${NC}"
echo -e "${YELLOW}🔧 Now SSH into your droplet and run:${NC}"
echo -e "   ssh ${DROPLET_USER}@${DROPLET_IP}"
echo -e "   cd ${APP_DIR}"
echo -e "   cp .env.example .env  # Then edit .env with your values"
echo -e "   nano Caddyfile        # Update with your domain"
echo -e "   chmod +x deploy.sh"
echo -e "   ./deploy.sh"
