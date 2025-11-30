#!/bin/bash
# Initial setup script for Digital Ocean Droplet
# Run this on the droplet after first login

set -e

echo "🔧 Setting up Digital Ocean Droplet for Trip Planner..."

# Update system
echo "📦 Updating system packages..."
apt update && apt upgrade -y

# Install Docker
echo "🐳 Installing Docker..."
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh
rm get-docker.sh

# Install Docker Compose
echo "🐳 Installing Docker Compose..."
apt install docker-compose -y

# Install Git
echo "📚 Installing Git..."
apt install git -y

# Install additional tools
echo "🛠️  Installing additional tools..."
apt install htop curl wget vim ufw fail2ban -y

# Setup firewall
echo "🔥 Configuring firewall..."
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
ufw allow 80/tcp
ufw allow 443/tcp
ufw allow 443/udp  # HTTP/3
echo "y" | ufw enable

# Setup fail2ban
echo "🛡️  Setting up fail2ban..."
systemctl enable fail2ban
systemctl start fail2ban

# Create application directory
echo "📁 Creating application directory..."
mkdir -p /opt/trip-planner
cd /opt/trip-planner

# Setup automatic security updates
echo "🔒 Setting up automatic security updates..."
apt install unattended-upgrades -y
dpkg-reconfigure -plow unattended-upgrades

# Create non-root user for deployment (optional but recommended)
echo "👤 Creating deployment user..."
if ! id -u deploy > /dev/null 2>&1; then
    useradd -m -s /bin/bash deploy
    usermod -aG docker deploy
    echo "User 'deploy' created and added to docker group"
fi

echo "✅ Droplet setup complete!"
echo ""
echo "Next steps:"
echo "1. Upload your code to /opt/trip-planner"
echo "2. Create .env file with your configuration"
echo "3. Update Caddyfile with your domain"
echo "4. Run: chmod +x deploy.sh && ./deploy.sh"
