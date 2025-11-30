# Digital Ocean Deployment Guide

Complete guide to deploy Trip Planner on Digital Ocean with Caddy reverse proxy.

## Architecture

- **Frontend**: Digital Ocean Spaces (https://trip-planner-fe.blr1.digitaloceanspaces.com)
- **Backend**: Docker container on Droplet with Caddy reverse proxy
- **Database**: PostgreSQL in Docker
- **Cache**: Redis (optional)

## Prerequisites

1. Digital Ocean Account
2. Domain name (optional but recommended)
3. Local machine with:
   - SSH access
   - Git
   - Docker (for local testing)

## Step 1: Create Digital Ocean Droplet

```bash
# Using doctl CLI (optional)
doctl compute droplet create trip-planner-api \
  --region blr1 \
  --size s-2vcpu-2gb \
  --image ubuntu-22-04-x64 \
  --ssh-keys your-ssh-key-id

# Or create via Digital Ocean Dashboard:
# - Region: Bangalore (BLR1) - same as your Spaces
# - Plan: Basic, 2GB RAM / 2 vCPUs
# - OS: Ubuntu 22.04 LTS
# - Add SSH key
```

## Step 2: Initial Droplet Setup

SSH into your droplet:
```bash
ssh root@your-droplet-ip
```

Run the setup script:
```bash
# Download and run setup script
curl -O https://raw.githubusercontent.com/yourusername/trip-planner/main/setup-droplet.sh
chmod +x setup-droplet.sh
./setup-droplet.sh
```

Or manually:
```bash
# Update system
apt update && apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# Install Docker Compose
apt install docker-compose -y

# Install Git
apt install git -y

# Setup firewall
ufw allow ssh
ufw allow 80/tcp
ufw allow 443/tcp
ufw allow 443/udp
ufw enable

# Create app directory
mkdir -p /opt/trip-planner
```

## Step 3: Upload Code to Droplet

### Option A: Using the upload script (from local machine)

```bash
# Edit the script with your droplet IP
nano upload-to-droplet.sh
# Set: DROPLET_IP="your.droplet.ip.here"

# Make executable and run
chmod +x upload-to-droplet.sh
./upload-to-droplet.sh
```

### Option B: Using Git (recommended for production)

```bash
# On the droplet
cd /opt/trip-planner
git clone https://github.com/yourusername/trip-planner.git .
# Or if already cloned:
git pull origin main
```

### Option C: Manual rsync

```bash
# From your local machine
rsync -avz --exclude '.git' --exclude 'node_modules' \
  ./ root@your-droplet-ip:/opt/trip-planner/
```

## Step 4: Configure Environment

On the droplet:

```bash
cd /opt/trip-planner

# Copy environment template
cp .env.example .env

# Edit with your values
nano .env
```

Required `.env` configuration:
```bash
# Database
DB_NAME=tripplanner
DB_USER=tripplanner_user
DB_PASSWORD=generate_secure_password_here

# JWT Secret (use: openssl rand -base64 32)
JWT_SECRET=your_jwt_secret_at_least_32_chars

# Mapbox Token
MAPBOX_TOKEN=your_mapbox_token

# Allowed Origins - IMPORTANT for CORS
ALLOWED_ORIGINS=https://trip-planner-fe.blr1.digitaloceanspaces.com,https://trip-planner-fe.blr1.cdn.digitaloceanspaces.com

# Redis Password (optional)
REDIS_PASSWORD=your_redis_password
```

## Step 5: Configure Caddy

Edit the Caddyfile:
```bash
nano Caddyfile
```

**With domain:**
```caddy
api.yourdomain.com {
    reverse_proxy api:8080 {
        header_down Access-Control-Allow-Origin "https://trip-planner-fe.blr1.digitaloceanspaces.com"
        header_down Access-Control-Allow-Credentials "true"
    }
}
```

**Without domain (using IP):**
```caddy
:80, :443 {
    reverse_proxy api:8080 {
        header_down Access-Control-Allow-Origin "https://trip-planner-fe.blr1.digitaloceanspaces.com"
        header_down Access-Control-Allow-Credentials "true"
    }
}
```

## Step 6: Deploy

```bash
# Make deploy script executable
chmod +x deploy.sh

# Run deployment
./deploy.sh
```

The script will:
1. Pull latest code (if using Git)
2. Build Docker images
3. Start containers
4. Run database migrations
5. Show service status

## Step 7: Verify Deployment

Check if services are running:
```bash
docker-compose ps
```

Check logs:
```bash
docker-compose logs -f
```

Test API health:
```bash
curl http://localhost:8080/health
# Should return: {"message":"Trip Planner API is running","status":"healthy"}
```

Test externally (with domain):
```bash
curl https://api.yourdomain.com/health
```

## Step 8: DNS Configuration (if using domain)

Add these DNS records:

| Type | Name | Value | TTL |
|------|------|-------|-----|
| A | api | your-droplet-ip | 3600 |

Caddy will automatically get SSL certificates from Let's Encrypt!

## Frontend Configuration

Update your frontend API URL in `trip-planner-fe/src/services/api.js`:

```javascript
// Change from:
const API_BASE_URL = 'http://localhost:8080';

// To:
const API_BASE_URL = 'https://api.yourdomain.com';
// Or if using IP:
const API_BASE_URL = 'http://your-droplet-ip';
```

Then rebuild and upload to Spaces:
```bash
cd trip-planner-fe
npm run build
aws s3 sync build/ s3://trip-planner-fe/ --endpoint-url=https://blr1.digitaloceanspaces.com
```

## Monitoring & Maintenance

### View logs
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f api
docker-compose logs -f caddy
```

### Restart services
```bash
# Restart all
docker-compose restart

# Restart specific service
docker-compose restart api
```

### Database backup
```bash
# Backup
docker-compose exec db pg_dump -U tripplanner_user tripplanner > backup_$(date +%Y%m%d).sql

# Restore
cat backup_20240101.sql | docker-compose exec -T db psql -U tripplanner_user tripplanner
```

### Update application
```bash
cd /opt/trip-planner
git pull origin main
./deploy.sh
```

## Troubleshooting

### CORS errors
1. Check ALLOWED_ORIGINS in .env
2. Verify Caddyfile has correct Access-Control headers
3. Restart Caddy: `docker-compose restart caddy`

### SSL certificate issues
```bash
# Check Caddy logs
docker-compose logs caddy

# Caddy auto-renews certificates, but you can force:
docker-compose exec caddy caddy reload
```

### Database connection errors
```bash
# Check if DB is running
docker-compose ps db

# Check DB logs
docker-compose logs db

# Connect to DB
docker-compose exec db psql -U tripplanner_user tripplanner
```

### Port conflicts
```bash
# Check what's using ports
sudo lsof -i :80
sudo lsof -i :443

# Stop conflicting services
sudo systemctl stop nginx  # If Nginx is running
```

## Security Best Practices

1. **Change default passwords** in .env
2. **Enable firewall**: Only allow ports 22, 80, 443
3. **Regular updates**:
   ```bash
   apt update && apt upgrade -y
   docker-compose pull
   ./deploy.sh
   ```
4. **Setup fail2ban** for SSH protection (included in setup script)
5. **Regular backups** of database and uploads
6. **Monitor logs** regularly

## Costs (Approximate)

- Droplet (2GB): $12/month
- Spaces (250GB): $5/month
- Domain: $10-15/year
- **Total**: ~$17/month

## Next Steps

- [ ] Setup monitoring (e.g., UptimeRobot, Datadog)
- [ ] Configure automated backups
- [ ] Setup CI/CD with GitHub Actions
- [ ] Add logging aggregation
- [ ] Setup SSL monitoring
