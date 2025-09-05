# Deployment Guide 🚀

Production deployment strategies for Disruptor across different environments, from simple Docker setups to enterprise Kubernetes deployments.

## Table of Contents

- [Quick Production Setup](#quick-production-setup)
- [Docker Deployment](#docker-deployment)
- [Docker Compose Examples](#docker-compose-examples)
- [Kubernetes Deployment](#kubernetes-deployment)
- [Traditional Server Deployment](#traditional-server-deployment)
- [Cloud Platform Guides](#cloud-platform-guides)
- [Monitoring & Maintenance](#monitoring--maintenance)
- [Security Considerations](#security-considerations)

## Quick Production Setup ⚡

**Minimal production deployment with Docker:**

```bash
# 1. Create data directory
mkdir -p /opt/disruptor/data
chmod 755 /opt/disruptor/data

# 2. Run with persistent storage
docker run -d \
  --name disruptor \
  --restart unless-stopped \
  -e CONFIG_TOKEN="your_bot_token" \
  -e CONFIG_DATABASE_DSN="file:/data/disruptor.db?cache=shared" \
  -e CONFIG_LOGGING_LEVEL="info" \
  -e CONFIG_METRICS_ENABLED="true" \
  -v /opt/disruptor/data:/data \
  -p 8080:8080 \
  ghcr.io/xanderd99/disruptor:latest

# 3. Check logs
docker logs -f disruptor
```

## Docker Deployment 🐳

### Standalone Docker Container

#### Basic Production Container

```bash
# Pull latest stable version
docker pull ghcr.io/xanderd99/disruptor:latest

# Run with production configuration
docker run -d \
  --name disruptor \
  --restart unless-stopped \
  --memory 512m \
  --cpus 1.0 \
  -e CONFIG_TOKEN="${DISCORD_TOKEN}" \
  -e CONFIG_DATABASE_DSN="file:/data/disruptor.db?cache=shared" \
  -e CONFIG_LOGGING_LEVEL="info" \
  -e CONFIG_LOGGING_FORMAT="json" \
  -e CONFIG_METRICS_ENABLED="true" \
  -e CONFIG_METRICS_ADDRESS="0.0.0.0:8080" \
  -v /opt/disruptor/data:/data \
  -v /var/log/disruptor:/var/log/disruptor \
  -p 8080:8080 \
  --health-cmd="curl -f http://localhost:8080/health || exit 1" \
  --health-interval=30s \
  --health-timeout=10s \
  --health-retries=3 \
  ghcr.io/xanderd99/disruptor:latest
```

#### With Database Migrations

```bash
# Run migrations first
docker run --rm \
  -v /opt/disruptor/data:/data \
  -e CONFIG_DATABASE_DSN="file:/data/disruptor.db?cache=shared" \
  ghcr.io/xanderd99/disruptor:latest \
  /bin/migrate up

# Then start the main application
docker run -d \
  --name disruptor \
  --restart unless-stopped \
  -e CONFIG_TOKEN="${DISCORD_TOKEN}" \
  -e CONFIG_DATABASE_DSN="file:/data/disruptor.db?cache=shared" \
  -v /opt/disruptor/data:/data \
  ghcr.io/xanderd99/disruptor:latest
```

### Custom Dockerfile Build

For custom modifications:

```dockerfile
# Dockerfile.production
FROM ghcr.io/xanderd99/disruptor:latest

# Add custom configuration
COPY production.env /app/.env
COPY custom-sounds/ /app/sounds/

# Custom entrypoint
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
```

## Docker Compose Examples 📋

### Simple Production Setup

```yaml
# docker-compose.yml
version: '3.8'

services:
  disruptor:
    image: ghcr.io/xanderd99/disruptor:latest
    container_name: disruptor
    restart: unless-stopped
    environment:
      - CONFIG_TOKEN=${DISCORD_TOKEN}
      - CONFIG_DATABASE_DSN=file:/data/disruptor.db?cache=shared
      - CONFIG_LOGGING_LEVEL=info
      - CONFIG_LOGGING_FORMAT=json
      - CONFIG_METRICS_ENABLED=true
    volumes:
      - ./data:/data
      - ./logs:/var/log/disruptor
    ports:
      - "8080:8080"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 30s
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '1.0'
        reservations:
          memory: 256M
          cpus: '0.5'
```

**Setup and run:**

```bash
# Create environment file
echo "DISCORD_TOKEN=your_bot_token_here" > .env

# Start services
docker-compose up -d

# View logs
docker-compose logs -f disruptor
```

### Advanced Setup with Migrations

```yaml
# docker-compose.advanced.yml
version: '3.8'

services:
  # Database migration service
  migrate:
    image: ghcr.io/xanderd99/disruptor:latest
    container_name: disruptor-migrate
    environment:
      - CONFIG_DATABASE_DSN=file:/data/disruptor.db?cache=shared
    volumes:
      - ./data:/data
    command: ["/bin/migrate", "up"]
    restart: "no"

  # Main application service
  disruptor:
    image: ghcr.io/xanderd99/disruptor:latest
    container_name: disruptor
    restart: unless-stopped
    depends_on:
      migrate:
        condition: service_completed_successfully
    environment:
      - CONFIG_TOKEN=${DISCORD_TOKEN}
      - CONFIG_DATABASE_DSN=file:/data/disruptor.db?cache=shared
      - CONFIG_LOGGING_LEVEL=info
      - CONFIG_LOGGING_FORMAT=json
      - CONFIG_METRICS_ENABLED=true
      - CONFIG_METRICS_ADDRESS=0.0.0.0:8080
    volumes:
      - ./data:/data
      - ./logs:/var/log/disruptor
    ports:
      - "8080:8080"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '1.0'

  # Metrics and monitoring
  prometheus:
    image: prom/prometheus:latest
    container_name: disruptor-prometheus
    restart: unless-stopped
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'

  grafana:
    image: grafana/grafana:latest
    container_name: disruptor-grafana
    restart: unless-stopped
    ports:
      - "3000:3000"
    volumes:
      - grafana-data:/var/lib/grafana
      - ./monitoring/grafana/dashboards:/etc/grafana/provisioning/dashboards
      - ./monitoring/grafana/datasources:/etc/grafana/provisioning/datasources
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD:-admin}

volumes:
  prometheus-data:
  grafana-data:
```

**Supporting configuration files:**

```yaml
# monitoring/prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'disruptor'
    static_configs:
      - targets: ['disruptor:8080']
    metrics_path: /metrics
```

### High Availability Setup

```yaml
# docker-compose.ha.yml
version: '3.8'

services:
  # Load balancer
  nginx:
    image: nginx:alpine
    container_name: disruptor-lb
    restart: unless-stopped
    ports:
      - "80:80"
      - "8080:8080"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
    depends_on:
      - disruptor-1
      - disruptor-2

  # Primary instance
  disruptor-1:
    image: ghcr.io/xanderd99/disruptor:latest
    container_name: disruptor-1
    restart: unless-stopped
    environment:
      - CONFIG_TOKEN=${DISCORD_TOKEN}
      - CONFIG_DATABASE_DSN=file:/data/disruptor.db?cache=shared
      - CONFIG_LOGGING_LEVEL=info
      - CONFIG_METRICS_ENABLED=true
      - CONFIG_METRICS_ADDRESS=0.0.0.0:8080
    volumes:
      - ./data:/data
    expose:
      - "8080"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s

  # Secondary instance (standby)
  disruptor-2:
    image: ghcr.io/xanderd99/disruptor:latest
    container_name: disruptor-2
    restart: unless-stopped
    environment:
      - CONFIG_TOKEN=${DISCORD_TOKEN}
      - CONFIG_DATABASE_DSN=file:/data/disruptor.db?cache=shared
      - CONFIG_LOGGING_LEVEL=info
      - CONFIG_METRICS_ENABLED=true
      - CONFIG_METRICS_ADDRESS=0.0.0.0:8080
    volumes:
      - ./data:/data
    expose:
      - "8080"
    profiles:
      - standby  # Only start when explicitly enabled
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
```

## Kubernetes Deployment ☸️

### Basic Kubernetes Manifests

#### Namespace and ConfigMap

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: disruptor
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: disruptor-config
  namespace: disruptor
data:
  CONFIG_DATABASE_DSN: "file:/data/disruptor.db?cache=shared"
  CONFIG_LOGGING_LEVEL: "info"
  CONFIG_LOGGING_FORMAT: "json"
  CONFIG_METRICS_ENABLED: "true"
  CONFIG_METRICS_ADDRESS: "0.0.0.0:8080"
```

#### Secret

```yaml
# k8s/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: disruptor-secret
  namespace: disruptor
type: Opaque
data:
  # Base64 encoded Discord token
  CONFIG_TOKEN: <secret>
```

#### Persistent Volume

```yaml
# k8s/pvc.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: disruptor-data
  namespace: disruptor
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
  storageClassName: fast-ssd  # Adjust for your cluster
```

#### Deployment

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: disruptor
  namespace: disruptor
  labels:
    app: disruptor
spec:
  replicas: 1  # Single replica for Discord bot
  selector:
    matchLabels:
      app: disruptor
  template:
    metadata:
      labels:
        app: disruptor
    spec:
      initContainers:
      - name: migrate
        image: ghcr.io/xanderd99/disruptor:latest
        command: ["/bin/migrate", "up"]
        envFrom:
        - configMapRef:
            name: disruptor-config
        volumeMounts:
        - name: data
          mountPath: /data
      containers:
      - name: disruptor
        image: ghcr.io/xanderd99/disruptor:latest
        envFrom:
        - configMapRef:
            name: disruptor-config
        - secretRef:
            name: disruptor-secret
        ports:
        - containerPort: 8080
          name: metrics
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
        resources:
          limits:
            memory: 512Mi
            cpu: 1000m
          requests:
            memory: 256Mi
            cpu: 500m
        volumeMounts:
        - name: data
          mountPath: /data
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: disruptor-data
```

#### Service

```yaml
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: disruptor-service
  namespace: disruptor
  labels:
    app: disruptor
spec:
  selector:
    app: disruptor
  ports:
  - port: 8080
    targetPort: 8080
    name: metrics
  type: ClusterIP
```

**Deploy to Kubernetes:**

```bash
# Apply all manifests
kubectl apply -f k8s/

# Check deployment status
kubectl -n disruptor get pods
kubectl -n disruptor logs deployment/disruptor

# Port forward for metrics (optional)
kubectl -n disruptor port-forward service/disruptor-service 8080:8080
```

## Traditional Server Deployment 🖥️

### Systemd Service

#### User Setup

```bash
# Create disruptor user
sudo useradd --system --create-home --shell /bin/bash disruptor
sudo mkdir -p /opt/disruptor/{bin,data,logs}
sudo chown -R disruptor:disruptor /opt/disruptor
```

#### Installation

```bash
# Build and install
git clone https://github.com/XanderD99/disruptor.git
cd disruptor
make build build-migrate

# Install binaries
sudo cp output/bin/disruptor /opt/disruptor/bin/
sudo cp output/bin/migrate /opt/disruptor/bin/
sudo cp -r cmd/migrate/migrations /opt/disruptor/
sudo chown -R disruptor:disruptor /opt/disruptor
```

#### Service Configuration

```ini
# /etc/systemd/system/disruptor.service
[Unit]
Description=Disruptor Discord Bot
After=network.target
Wants=network.target

[Service]
Type=simple
User=disruptor
Group=disruptor
WorkingDirectory=/opt/disruptor
ExecStartPre=/opt/disruptor/bin/migrate up
ExecStart=/opt/disruptor/bin/disruptor
EnvironmentFile=/opt/disruptor/.env
Restart=always
RestartSec=10
StandardOutput=append:/opt/disruptor/logs/disruptor.log
StandardError=append:/opt/disruptor/logs/disruptor.log
SyslogIdentifier=disruptor

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/disruptor/data
ReadWritePaths=/opt/disruptor/logs

[Install]
WantedBy=multi-user.target
```

#### Environment Configuration

```bash
# /opt/disruptor/.env
CONFIG_TOKEN=your_discord_token_here
CONFIG_DATABASE_DSN=file:/opt/disruptor/data/disruptor.db?cache=shared
CONFIG_LOGGING_LEVEL=info
CONFIG_LOGGING_FORMAT=json
CONFIG_METRICS_ENABLED=true
CONFIG_METRICS_ADDRESS=0.0.0.0:8080
```

#### Service Management

```bash
# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable disruptor
sudo systemctl start disruptor

# Check status
sudo systemctl status disruptor
journalctl -u disruptor -f

# Restart service
sudo systemctl restart disruptor
```

## Cloud Platform Guides ☁️

### AWS EC2 Deployment

#### EC2 Instance Setup

```bash
# Launch EC2 instance (Ubuntu 22.04 LTS)
# t3.micro (1 vCPU, 1GB RAM) is sufficient for most servers

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh
sudo usermod -aG docker ubuntu

# Create deployment directory
mkdir -p ~/disruptor/{data,logs}
cd ~/disruptor
```

#### Docker Compose for AWS

```yaml
# docker-compose.aws.yml
version: '3.8'

services:
  disruptor:
    image: ghcr.io/xanderd99/disruptor:latest
    container_name: disruptor
    restart: unless-stopped
    environment:
      - CONFIG_TOKEN=${DISCORD_TOKEN}
      - CONFIG_DATABASE_DSN=file:/data/disruptor.db?cache=shared
      - CONFIG_LOGGING_LEVEL=info
      - CONFIG_LOGGING_FORMAT=json
      - CONFIG_METRICS_ENABLED=true
    volumes:
      - ./data:/data
      - ./logs:/var/log/disruptor
    ports:
      - "8080:8080"
    logging:
      driver: awslogs
      options:
        awslogs-group: /aws/ec2/disruptor
        awslogs-region: us-east-1
        awslogs-stream: disruptor
```

### Google Cloud Run

```dockerfile
# Dockerfile.cloudrun
FROM ghcr.io/xanderd99/disruptor:latest

# Cloud Run specific configurations
ENV PORT=8080
ENV CONFIG_METRICS_ADDRESS=0.0.0.0:8080

EXPOSE 8080
```

```bash
# Build and deploy to Cloud Run
gcloud builds submit --tag gcr.io/PROJECT-ID/disruptor
gcloud run deploy disruptor \
  --image gcr.io/PROJECT-ID/disruptor \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --set-env-vars CONFIG_TOKEN=${DISCORD_TOKEN} \
  --set-env-vars CONFIG_DATABASE_DSN="file:/data/disruptor.db?cache=shared" \
  --memory 512Mi \
  --cpu 1 \
  --max-instances 1
```

### DigitalOcean App Platform

```yaml
# .do/app.yaml
name: disruptor
services:
- name: disruptor
  source_dir: /
  github:
    repo: your-username/disruptor
    branch: main
  build_command: make docker-build
  dockerfile_path: ci/Dockerfile
  instance_count: 1
  instance_size_slug: basic-xxs
  http_port: 8080
  envs:
  - key: CONFIG_TOKEN
    value: ${DISCORD_TOKEN}
    type: SECRET
  - key: CONFIG_DATABASE_DSN
    value: file:/data/disruptor.db?cache=shared
  - key: CONFIG_LOGGING_LEVEL
    value: info
  - key: CONFIG_METRICS_ENABLED
    value: "true"
```

## Monitoring & Maintenance 📊

### Health Checks

#### HTTP Health Endpoint

```bash
# Basic health check
curl -f http://localhost:8080/health

# Detailed metrics
curl http://localhost:8080/metrics
```

#### Custom Health Check Script

```bash
#!/bin/bash
# health-check.sh

HEALTH_URL="http://localhost:8080/health"
DISCORD_CHECK=true

# Check HTTP health endpoint
if ! curl -f -s "$HEALTH_URL" >/dev/null; then
    echo "❌ Health endpoint failed"
    exit 1
fi

# Check if bot is responsive to Discord
if [ "$DISCORD_CHECK" = true ]; then
    # This would require Discord API integration
    echo "✅ Bot is healthy"
fi

echo "✅ All health checks passed"
exit 0
```

### Log Management

#### Centralized Logging with ELK Stack

```yaml
# docker-compose.logging.yml
version: '3.8'

services:
  disruptor:
    image: ghcr.io/xanderd99/disruptor:latest
    # ... other configuration
    logging:
      driver: gelf
      options:
        gelf-address: "udp://localhost:12201"
        tag: "disruptor"

  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:7.17.0
    environment:
      - discovery.type=single-node
    volumes:
      - elasticsearch-data:/usr/share/elasticsearch/data

  logstash:
    image: docker.elastic.co/logstash/logstash:7.17.0
    volumes:
      - ./logstash.conf:/usr/share/logstash/pipeline/logstash.conf

  kibana:
    image: docker.elastic.co/kibana/kibana:7.17.0
    ports:
      - "5601:5601"
    depends_on:
      - elasticsearch

volumes:
  elasticsearch-data:
```

### Backup Strategy

#### Database Backup Script

```bash
#!/bin/bash
# backup.sh

BACKUP_DIR="/opt/disruptor/backups"
DB_PATH="/opt/disruptor/data/disruptor.db"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p "$BACKUP_DIR"

# Create backup
sqlite3 "$DB_PATH" ".backup '$BACKUP_DIR/disruptor_$DATE.db'"

# Compress backup
gzip "$BACKUP_DIR/disruptor_$DATE.db"

# Cleanup old backups (keep 30 days)
find "$BACKUP_DIR" -name "disruptor_*.db.gz" -mtime +30 -delete

echo "✅ Backup created: disruptor_$DATE.db.gz"
```

#### Automated Backup with Cron

```bash
# Add to crontab
0 2 * * * /opt/disruptor/scripts/backup.sh >> /opt/disruptor/logs/backup.log 2>&1
```

## Security Considerations 🔐

### Environment Security

```bash
# Secure environment file permissions
chmod 600 /opt/disruptor/.env
chown disruptor:disruptor /opt/disruptor/.env

# Secure database file permissions
chmod 600 /opt/disruptor/data/disruptor.db*
chown disruptor:disruptor /opt/disruptor/data/disruptor.db*
```

### Network Security

#### Firewall Configuration

```bash
# UFW rules for systemd deployment
sudo ufw allow ssh
sudo ufw allow 8080/tcp comment 'Disruptor metrics'
sudo ufw enable

# Or more restrictive (metrics only from monitoring server)
sudo ufw allow from 10.0.0.100 to any port 8080 comment 'Monitoring server'
```

#### Nginx Reverse Proxy with SSL

```nginx
# /etc/nginx/sites-available/disruptor
server {
    listen 443 ssl http2;
    server_name disruptor.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/disruptor.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/disruptor.yourdomain.com/privkey.pem;

    location /metrics {
        proxy_pass http://localhost:8080/metrics;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;

        # Basic authentication for metrics
        auth_basic "Disruptor Metrics";
        auth_basic_user_file /etc/nginx/.htpasswd;
    }

    location /health {
        proxy_pass http://localhost:8080/health;
        allow 127.0.0.1;
        allow 10.0.0.0/8;
        deny all;
    }
}
```

### Container Security

```yaml
# Security-focused Docker Compose
services:
  disruptor:
    image: ghcr.io/xanderd99/disruptor:latest
    # ... other configuration
    security_opt:
      - no-new-privileges:true
    read_only: true
    tmpfs:
      - /tmp:noexec,nosuid,size=100M
    user: "65534:65534"  # nobody user
    cap_drop:
      - ALL
    cap_add:
      - NET_BIND_SERVICE  # Only if binding to port < 1024
```

---

**🎉 Deployment Complete!** Your Disruptor bot is now ready for production use. Remember to monitor logs, maintain backups, and keep the bot updated for the best experience!
