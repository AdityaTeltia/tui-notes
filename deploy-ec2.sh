#!/bin/bash
# Quick deployment script for EC2
# Usage: ./deploy-ec2.sh

set -e

echo "🚀 SSH Notes EC2 Deployment Script"
echo "===================================="

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if running as root
if [ "$EUID" -eq 0 ]; then 
   echo -e "${RED}Please do not run as root${NC}"
   exit 1
fi

# Detect OS
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
else
    echo -e "${RED}Cannot detect OS${NC}"
    exit 1
fi

echo -e "${GREEN}Detected OS: $OS${NC}"

# Install dependencies
echo -e "${YELLOW}Installing dependencies...${NC}"
if [ "$OS" = "amzn" ] || [ "$OS" = "rhel" ]; then
    sudo dnf install -y golang git
elif [ "$OS" = "ubuntu" ] || [ "$OS" = "debian" ]; then
    sudo apt update
    sudo apt install -y golang-go git
else
    echo -e "${RED}Unsupported OS: $OS${NC}"
    exit 1
fi

# Build application
echo -e "${YELLOW}Building application...${NC}"
go mod download
go build -o ssh-notes-server

# Create directories
echo -e "${YELLOW}Creating directories...${NC}"
sudo mkdir -p /opt/ssh-notes/data
sudo mkdir -p /etc/ssh-notes
sudo mkdir -p /var/log/ssh-notes

# Set permissions
sudo chown -R $USER:$USER /opt/ssh-notes
sudo chown -R $USER:$USER /var/log/ssh-notes
chmod 700 /opt/ssh-notes/data

# Create configuration if it doesn't exist
if [ ! -f "/etc/ssh-notes/config.json" ]; then
    echo -e "${YELLOW}Creating configuration...${NC}"
    sudo cp config.json.example /etc/ssh-notes/config.json
    
    # Update config for production
    sudo sed -i 's/"port": "2222"/"port": "22"/' /etc/ssh-notes/config.json
    sudo sed -i 's|"data_dir": "./data"|"data_dir": "/opt/ssh-notes/data"|' /etc/ssh-notes/config.json
    sudo sed -i 's|"host_key": "host_key"|"host_key": "/opt/ssh-notes/host_key"|' /etc/ssh-notes/config.json
    sudo sed -i 's|"file": "./logs/ssh-notes.log"|"file": "/var/log/ssh-notes/ssh-notes.log"|' /etc/ssh-notes/config.json
fi

# Copy binary
echo -e "${YELLOW}Installing binary...${NC}"
sudo cp ssh-notes-server /usr/local/bin/ssh-notes-server
sudo chmod +x /usr/local/bin/ssh-notes-server

# Create systemd service
echo -e "${YELLOW}Creating systemd service...${NC}"
sudo tee /etc/systemd/system/ssh-notes.service > /dev/null <<EOF
[Unit]
Description=SSH Notes Server
After=network.target

[Service]
Type=simple
User=$USER
Group=$USER
WorkingDirectory=$(pwd)
ExecStart=/usr/local/bin/ssh-notes-server -config /etc/ssh-notes/config.json -port 22
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

# Security settings
NoNewPrivileges=true
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF

# Configure firewall
echo -e "${YELLOW}Configuring firewall...${NC}"
if [ "$OS" = "amzn" ] || [ "$OS" = "rhel" ]; then
    sudo systemctl enable firewalld 2>/dev/null || true
    sudo systemctl start firewalld 2>/dev/null || true
    sudo firewall-cmd --permanent --add-service=ssh 2>/dev/null || true
    sudo firewall-cmd --reload 2>/dev/null || true
elif [ "$OS" = "ubuntu" ] || [ "$OS" = "debian" ]; then
    sudo ufw allow 22/tcp 2>/dev/null || true
    sudo ufw --force enable 2>/dev/null || true
fi

# Enable and start service
echo -e "${YELLOW}Starting service...${NC}"
sudo systemctl daemon-reload
sudo systemctl enable ssh-notes
sudo systemctl start ssh-notes

# Check status
sleep 2
if sudo systemctl is-active --quiet ssh-notes; then
    echo -e "${GREEN}✓ Service is running!${NC}"
    sudo systemctl status ssh-notes --no-pager
else
    echo -e "${RED}✗ Service failed to start${NC}"
    sudo journalctl -u ssh-notes -n 20 --no-pager
    exit 1
fi

echo ""
echo -e "${GREEN}====================================${NC}"
echo -e "${GREEN}Deployment Complete!${NC}"
echo -e "${GREEN}====================================${NC}"
echo ""
echo "Service commands:"
echo "  sudo systemctl start ssh-notes"
echo "  sudo systemctl stop ssh-notes"
echo "  sudo systemctl restart ssh-notes"
echo "  sudo systemctl status ssh-notes"
echo ""
echo "View logs:"
echo "  sudo journalctl -u ssh-notes -f"
echo "  tail -f /var/log/ssh-notes/ssh-notes.log"
echo ""
echo -e "${YELLOW}⚠️  IMPORTANT: Make sure to:${NC}"
echo "  1. Configure DNS to point to this server's IP"
echo "  2. Update security group to allow SSH (port 22)"
echo "  3. Test connection: ssh notes.write"
echo ""

