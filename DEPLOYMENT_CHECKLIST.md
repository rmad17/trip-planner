# Deployment Checklist

Quick reference guide for deploying Trip Planner to Digital Ocean.

## Pre-Deployment

### Backend (trip-planner)
- [ ] CORS middleware added to app.go
- [ ] Health endpoint `/health` available
- [ ] Environment variables configured in `.env`
- [ ] Docker and docker-compose files ready
- [ ] Caddyfile configured with domain/IP

### Frontend (trip-planner-fe)
- [ ] API URL updated in `.env.production` or `api.js`
- [ ] Build succeeds: `npm run build`
- [ ] AWS CLI configured for Digital Ocean Spaces
- [ ] Spaces created: `trip-planner-fe` in BLR1 region

## Backend Deployment Steps

### 1. Create & Setup Droplet
```bash
# Create droplet (via DO dashboard or CLI)
# Size: 2GB RAM minimum
# Region: BLR1 (Bangalore)
# OS: Ubuntu 22.04

# SSH into droplet
ssh root@your-droplet-ip

# Run setup
curl -O https://your-repo/setup-droplet.sh
chmod +x setup-droplet.sh
./setup-droplet.sh
```

### 2. Upload Code
```bash
# From local machine
cd trip-planner
nano upload-to-droplet.sh  # Set DROPLET_IP
./upload-to-droplet.sh

# Or use Git on droplet
ssh root@droplet-ip
cd /opt/trip-planner
git clone https://github.com/yourusername/trip-planner.git .
```

### 3. Configure Environment
```bash
# On droplet
cd /opt/trip-planner
cp .env.example .env
nano .env

# Required settings:
# - DB_PASSWORD
# - JWT_SECRET
# - MAPBOX_TOKEN
# - ALLOWED_ORIGINS=https://trip-planner-fe.blr1.digitaloceanspaces.com,https://trip-planner-fe.blr1.cdn.digitaloceanspaces.com
```

### 4. Configure Caddy
```bash
nano Caddyfile

# With domain:
api.yourdomain.com {
    reverse_proxy api:8080 {
        header_down Access-Control-Allow-Origin "https://trip-planner-fe.blr1.digitaloceanspaces.com"
        header_down Access-Control-Allow-Credentials "true"
    }
}

# Without domain (development):
:80, :443 {
    reverse_proxy api:8080
}
```

### 5. Deploy
```bash
chmod +x deploy.sh
./deploy.sh

# Verify
docker-compose ps
docker-compose logs -f
curl http://localhost:8080/health
```

### 6. DNS (if using domain)
```
Type: A
Name: api
Value: your-droplet-ip
TTL: 3600
```

## Frontend Deployment Steps

### 1. Configure API URL
```bash
cd trip-planner-fe

# Create .env.production
echo "REACT_APP_API_URL=https://api.yourdomain.com" > .env.production
# Or use IP: http://your-droplet-ip
```

### 2. Build
```bash
npm install
npm run build
```

### 3. Deploy to Spaces
```bash
# Make sure AWS CLI is configured
aws configure --profile digitalocean

# Deploy
chmod +x deploy-to-spaces.sh
./deploy-to-spaces.sh
```

### 4. Configure Space
In Digital Ocean Dashboard:
- Navigate to Spaces → trip-planner-fe → Settings
- Enable Static Website Hosting
  - Index: `index.html`
  - Error: `index.html`
- Enable CDN
- Note your URLs:
  - Direct: https://trip-planner-fe.blr1.digitaloceanspaces.com
  - CDN: https://trip-planner-fe.blr1.cdn.digitaloceanspaces.com

## Verification

### Backend
- [ ] API accessible: `curl https://api.yourdomain.com/health`
- [ ] Returns: `{"status":"healthy","message":"Trip Planner API is running"}`
- [ ] SSL certificate active (if using domain)
- [ ] CORS headers present in response

### Frontend
- [ ] App loads: https://trip-planner-fe.blr1.cdn.digitaloceanspaces.com
- [ ] No console errors
- [ ] API calls work (try login)
- [ ] No CORS errors
- [ ] Images/assets load correctly

### Test CORS
```bash
curl -H "Origin: https://trip-planner-fe.blr1.digitaloceanspaces.com" \
     -H "Access-Control-Request-Method: POST" \
     -H "Access-Control-Request-Headers: Content-Type,Authorization" \
     -X OPTIONS \
     https://api.yourdomain.com/api/v1/auth/login -v

# Should see:
# Access-Control-Allow-Origin: https://trip-planner-fe.blr1.digitaloceanspaces.com
# Access-Control-Allow-Credentials: true
```

## Post-Deployment

### Security
- [ ] Change all default passwords
- [ ] Firewall enabled (only 22, 80, 443)
- [ ] fail2ban running
- [ ] SSL certificates valid

### Monitoring
- [ ] Setup uptime monitoring
- [ ] Configure error tracking
- [ ] Enable log aggregation
- [ ] Database backups scheduled

### Documentation
- [ ] Update README with production URLs
- [ ] Document environment variables
- [ ] Create runbook for common issues

## Quick Commands Reference

### Backend (on droplet)
```bash
# View logs
docker-compose logs -f

# Restart
docker-compose restart

# Update and redeploy
git pull origin main
./deploy.sh

# Database backup
docker-compose exec db pg_dump -U tripplanner_user tripplanner > backup.sql

# Access database
docker-compose exec db psql -U tripplanner_user tripplanner
```

### Frontend (local)
```bash
# Build and deploy
npm run build
./deploy-to-spaces.sh

# Clear CDN cache (in DO dashboard)
# Spaces → trip-planner-fe → CDN → Purge Cache
```

## Troubleshooting

### CORS Issues
1. Check backend `.env` → ALLOWED_ORIGINS
2. Verify Caddyfile headers
3. Restart: `docker-compose restart caddy`
4. Test with curl command above

### API Not Accessible
1. Check firewall: `ufw status`
2. Check containers: `docker-compose ps`
3. Check logs: `docker-compose logs caddy api`
4. Verify DNS: `dig api.yourdomain.com`

### Frontend Shows Old Version
1. Rebuild: `npm run build`
2. Redeploy: `./deploy-to-spaces.sh`
3. Purge CDN cache in DO dashboard
4. Hard refresh browser

### Database Connection Error
1. Check DB running: `docker-compose ps db`
2. Check credentials in `.env`
3. Restart: `docker-compose restart db api`

## Rollback

### Backend
```bash
cd /opt/trip-planner
git log  # Find previous commit
git checkout <commit-hash>
./deploy.sh
```

### Frontend
```bash
git checkout <commit-hash>
npm run build
./deploy-to-spaces.sh
```

## Costs Summary

- Droplet (2GB): **$12/month**
- Spaces (250GB): **$5/month**
- Bandwidth: Usually free tier
- Domain: **~$1/month**
- **Total: ~$18/month**

## Support

- Digital Ocean Docs: https://docs.digitalocean.com
- Caddy Docs: https://caddyserver.com/docs
- Docker Docs: https://docs.docker.com

---

**Last Updated**: $(date)
**Deployed By**: Your Name
**Environment**: Production
