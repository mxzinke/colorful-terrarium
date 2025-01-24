#!/bin/bash

# Exit on error
set -e

echo "Starting server setup..."

# Update system
apt-get update
apt-get upgrade -y

# Install required packages
apt-get install -y \
    ca-certificates \
    curl \
    gnupg \
    lsb-release

# Install Docker using the official repository
mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg

echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null

apt-get update
apt-get install -y docker-ce docker-ce-cli containerd.io

# Configure containerd to use systemd as cgroup driver
mkdir -p /etc/containerd
containerd config default | tee /etc/containerd/config.toml
sed -i 's/SystemdCgroup = false/SystemdCgroup = true/' /etc/containerd/config.toml
systemctl restart containerd

# Configure IPv6 iptables
ip6tables -F
ip6tables -X
ip6tables -P INPUT DROP
ip6tables -P FORWARD DROP
ip6tables -P OUTPUT ACCEPT

# Allow loopback interface
ip6tables -A INPUT -i lo -j ACCEPT

# Allow ICMPv6 for proper IPv6 functionality
ip6tables -A INPUT -p ipv6-icmp -j ACCEPT

# Allow established and related connections
ip6tables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT

# Allow SSH and HTTP
ip6tables -A INPUT -p tcp --dport 22 -j ACCEPT
ip6tables -A INPUT -p tcp --dport 80 -j ACCEPT

# Allow Neighbor Discovery Protocol (NDP)
ip6tables -A INPUT -p icmpv6 --icmpv6-type router-advertisement -j ACCEPT
ip6tables -A INPUT -p icmpv6 --icmpv6-type router-solicitation -j ACCEPT
ip6tables -A INPUT -p icmpv6 --icmpv6-type neighbour-advertisement -j ACCEPT
ip6tables -A INPUT -p icmpv6 --icmpv6-type neighbour-solicitation -j ACCEPT

# Save iptables rules
apt-get install -y iptables-persistent
ip6tables-save > /etc/iptables/rules.v6

# Install and configure Nginx
apt-get install -y nginx

# Create Nginx cache directory
mkdir -p /var/cache/nginx
chown -R www-data:www-data /var/cache/nginx

# Configure Nginx
cat > /etc/nginx/conf.d/proxy-cache.conf << 'EOL'
proxy_cache_path /var/cache/nginx levels=1:2 keys_zone=my_cache:10m max_size=30g inactive=30d use_temp_path=off;

map "$http_origin" $cors {
    default '';
    "~^http://localhost(:[0-9]+)?$" "$http_origin";
    "https://mapstudio.ai" "$http_origin";
}

server {
    listen 80;
    listen [::]:80;
    server_name _;

    location / {
        if ($request_method = 'OPTIONS') {
            add_header 'Access-Control-Allow-Origin' $cors;
            add_header 'Access-Control-Allow-Credentials' 'false' always;
            add_header 'Access-Control-Allow-Methods' 'GET, OPTIONS' always;
            add_header 'Access-Control-Allow-Headers' 'DNT,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range' always;
            add_header 'Access-Control-Max-Age' 1728000;
            add_header 'Content-Type' 'text/plain; charset=utf-8';
            add_header 'Content-Length' 0;
            return 204;
        }

        if ($request_method = 'GET') {
            add_header 'Access-Control-Allow-Origin' $cors;
            add_header 'Access-Control-Allow-Credentials' 'false' always;
            add_header 'Access-Control-Allow-Methods' 'GET, OPTIONS' always;
            add_header 'Access-Control-Allow-Headers' 'DNT,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range' always;
            add_header 'Access-Control-Expose-Headers' 'Content-Encoding,Content-Length,Content-Range,X-Cache-Status' always;
        }

        proxy_cache my_cache;
        proxy_cache_use_stale error timeout http_500 http_502 http_503 http_504 updating;
        proxy_cache_valid 200 202 30d;
        proxy_cache_valid any 1m;
        proxy_cache_min_uses 1;
        proxy_cache_background_update on;
        proxy_cache_lock on;
        proxy_cache_lock_age 60s;
        proxy_cache_revalidate on;
        proxy_cache_bypass $http_cache_control;
        proxy_cache_key "$host$request_uri";
        add_header X-Cache-Status $upstream_cache_status;

        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
EOL

# Remove default Nginx config
rm /etc/nginx/sites-enabled/default

# Test and reload Nginx
nginx -t
systemctl reload nginx

# Proxy for ghcr.io
echo "2a01:4f8:c010:d56::6 ghcr.io" >> /etc/hosts

# Pull and run Docker container
docker pull ghcr.io/mxzinke/colorful-terrarium:latest

# Create Docker service file for automatic restart
cat > /etc/systemd/system/colorful-terrarium.service << 'EOL'
[Unit]
Description=Colorful Terrarium Container
Requires=docker.service
After=docker.service

[Service]
Restart=always
ExecStart=/usr/bin/docker run --rm --net=host --dns 2a01:4f8:c2c:123f::1 --dns 2a00:1098:2b::1 --name colorful-terrarium ghcr.io/mxzinke/colorful-terrarium:latest
ExecStop=/usr/bin/docker stop colorful-terrarium

[Install]
WantedBy=multi-user.target
EOL

# Enable and start the service
systemctl daemon-reload
systemctl enable colorful-terrarium
systemctl start colorful-terrarium

echo "Setup completed successfully!"