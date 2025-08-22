# Deployment Guide 🚀

This guide covers various deployment scenarios for Disruptor, from simple server setups to enterprise-grade deployments.

## Table of Contents

- [Deployment Options Overview](#deployment-options-overview)
- [Docker Deployment](#docker-deployment)
- [Systemd Service (Linux)](#systemd-service-linux)
- [Cloud Deployments](#cloud-deployments)
- [Container Orchestration](#container-orchestration)
- [Monitoring and Maintenance](#monitoring-and-maintenance)
- [Security Considerations](#security-considerations)

---

## Deployment Options Overview

| Method | Complexity | Scalability | Use Case |
|--------|------------|-------------|----------|
| **Binary + Systemd** | Low | Low | Single server, VPS |
| **Docker** | Medium | Medium | Development, testing |
| **Docker Compose** | Medium | Medium | Small production |
| **Kubernetes** | High | High | Enterprise, multi-server |
| **Cloud Services** | Medium | High | Managed deployments |

---

## Docker Deployment

### Single Container

#### Basic Docker Run
```bash
# Build image
docker build -t disruptor:latest -f ci/Dockerfile .

# Run container
docker run -d \
  --name disruptor \
  --restart unless-stopped \
  -e CONFIG_TOKEN="your_discord_bot_token" \
  -e CONFIG_DATABASE_DSN="file:/data/disruptor.db?cache=shared" \
  -v /host/data:/data \
  -p 9090:9090 \
  disruptor:latest
```

#### With Environment File
```bash
# Create environment file
cat > .env << EOF
CONFIG_TOKEN=your_discord_bot_token
CONFIG_DATABASE_DSN=file:/data/disruptor.db?cache=shared
CONFIG_LOGGING_LEVEL=info
CONFIG_LOGGING_PRETTY=false
CONFIG_METRICS_PORT=9090
EOF

# Run with env file
docker run -d \
  --name disruptor \
  --restart unless-stopped \
  --env-file .env \
  -v /host/data:/data \
  -p 9090:9090 \
  disruptor:latest
```

### Docker Compose

#### Basic Setup
```yaml
# docker-compose.yml
version: '3.8'

services:
  disruptor:
    build:
      context: .
      dockerfile: ci/Dockerfile
    container_name: disruptor
    restart: unless-stopped
    environment:
      - CONFIG_TOKEN=${DISCORD_TOKEN}
      - CONFIG_DATABASE_DSN=file:/data/disruptor.db?cache=shared
      - CONFIG_LOGGING_LEVEL=info
      - CONFIG_LOGGING_PRETTY=false
    volumes:
      - ./data:/data
      - ./logs:/logs
    ports:
      - "9090:9090"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9090/metrics"]
      interval: 30s
      timeout: 10s
      retries: 3
```

```bash
# Create .env file
echo "DISCORD_TOKEN=your_discord_bot_token" > .env

# Deploy
docker-compose up -d

# View logs
docker-compose logs -f disruptor

# Update
docker-compose pull && docker-compose up -d
```

#### Production Setup with Monitoring
```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  disruptor:
    build:
      context: .
      dockerfile: ci/Dockerfile
    container_name: disruptor
    restart: unless-stopped
    environment:
      - CONFIG_TOKEN=${DISCORD_TOKEN}
      - CONFIG_DATABASE_DSN=file:/data/disruptor.db?cache=shared
      - CONFIG_LOGGING_LEVEL=info
      - CONFIG_LOGGING_PRETTY=false
      - CONFIG_METRICS_PORT=9090
    volumes:
      - disruptor-data:/data
      - disruptor-logs:/logs
    ports:
      - "9090:9090"
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:9090/metrics"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

  prometheus:
    image: prom/prometheus:latest
    container_name: prometheus
    restart: unless-stopped
    ports:
      - "9091:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'

  grafana:
    image: grafana/grafana:latest
    container_name: grafana
    restart: unless-stopped
    ports:
      - "3000:3000"
    volumes:
      - grafana-data:/var/lib/grafana
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin

volumes:
  disruptor-data:
  disruptor-logs:
  prometheus-data:
  grafana-data:
```

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'disruptor'
    static_configs:
      - targets: ['disruptor:9090']
```

---

## Systemd Service (Linux)

### Installation Setup

#### 1. Create Service User
```bash
# Create dedicated user
sudo useradd -r -s /bin/false -d /opt/disruptor disruptor

# Create directories
sudo mkdir -p /opt/disruptor/{bin,data,logs}
sudo chown -R disruptor:disruptor /opt/disruptor
```

#### 2. Install Binary
```bash
# Copy binary
sudo cp output/bin/disruptor /opt/disruptor/bin/
sudo chown disruptor:disruptor /opt/disruptor/bin/disruptor
sudo chmod +x /opt/disruptor/bin/disruptor
```

#### 3. Create Configuration
```bash
# Create environment file
sudo tee /opt/disruptor/.env > /dev/null << EOF
CONFIG_TOKEN=your_discord_bot_token
CONFIG_DATABASE_DSN=file:/opt/disruptor/data/disruptor.db?cache=shared
CONFIG_LOGGING_LEVEL=info
CONFIG_LOGGING_PRETTY=false
CONFIG_METRICS_PORT=9090
EOF

# Secure configuration
sudo chown disruptor:disruptor /opt/disruptor/.env
sudo chmod 600 /opt/disruptor/.env
```

#### 4. Create Systemd Service
```bash
sudo tee /etc/systemd/system/disruptor.service > /dev/null << EOF
[Unit]
Description=Disruptor Discord Bot
Documentation=https://github.com/XanderD99/disruptor
After=network.target
Wants=network.target

[Service]
Type=simple
User=disruptor
Group=disruptor
WorkingDirectory=/opt/disruptor
ExecStart=/opt/disruptor/bin/disruptor
EnvironmentFile=/opt/disruptor/.env

# Restart policy
Restart=always
RestartSec=10
StartLimitInterval=60
StartLimitBurst=3

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/disruptor/data /opt/disruptor/logs

# Resource limits
MemoryMax=512M
MemoryHigh=256M

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=disruptor

[Install]
WantedBy=multi-user.target
EOF
```

#### 5. Enable and Start Service
```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable service
sudo systemctl enable disruptor

# Start service
sudo systemctl start disruptor

# Check status
sudo systemctl status disruptor

# View logs
sudo journalctl -u disruptor -f
```

### Service Management

#### Basic Commands
```bash
# Start service
sudo systemctl start disruptor

# Stop service
sudo systemctl stop disruptor

# Restart service
sudo systemctl restart disruptor

# Check status
sudo systemctl status disruptor

# View logs
sudo journalctl -u disruptor -n 50

# Follow logs
sudo journalctl -u disruptor -f
```

#### Configuration Updates
```bash
# Edit configuration
sudo nano /opt/disruptor/.env

# Restart service to apply changes
sudo systemctl restart disruptor

# Verify changes
sudo systemctl status disruptor
```

---

## Cloud Deployments

### AWS EC2

#### 1. Launch Instance
```bash
# Launch Ubuntu 20.04 LTS instance
# Security Group: Allow inbound on port 9090 (metrics)
# Key pair: Your SSH key
```

#### 2. Setup Script
```bash
#!/bin/bash
# save as setup-aws.sh

# Update system
sudo apt update && sudo apt upgrade -y

# Install dependencies
sudo apt install -y git make pkg-config libopus-dev golang-go

# Clone and build
git clone https://github.com/XanderD99/disruptor.git
cd disruptor
go mod download
make build

# Create directories
sudo mkdir -p /opt/disruptor/{bin,data,logs}
sudo useradd -r -s /bin/false -d /opt/disruptor disruptor

# Install
sudo cp output/bin/disruptor /opt/disruptor/bin/
sudo chown -R disruptor:disruptor /opt/disruptor
sudo chmod +x /opt/disruptor/bin/disruptor

# Create configuration
sudo tee /opt/disruptor/.env > /dev/null << EOF
CONFIG_TOKEN=$DISCORD_TOKEN
CONFIG_DATABASE_DSN=file:/opt/disruptor/data/disruptor.db?cache=shared
CONFIG_LOGGING_LEVEL=info
CONFIG_METRICS_PORT=9090
EOF

sudo chown disruptor:disruptor /opt/disruptor/.env
sudo chmod 600 /opt/disruptor/.env

# Create systemd service (use service file from above)
# ... systemd service creation ...

# Start service
sudo systemctl daemon-reload
sudo systemctl enable disruptor
sudo systemctl start disruptor

echo "✅ Disruptor deployed successfully!"
echo "View logs: sudo journalctl -u disruptor -f"
echo "Metrics: http://$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4):9090/metrics"
```

### Google Cloud Platform

#### Cloud Run Deployment
```yaml
# cloudrun.yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: disruptor
  annotations:
    run.googleapis.com/ingress: all
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/maxScale: "1"
        run.googleapis.com/memory: "512Mi"
        run.googleapis.com/cpu: "1000m"
    spec:
      containers:
      - image: gcr.io/PROJECT_ID/disruptor:latest
        env:
        - name: CONFIG_TOKEN
          valueFrom:
            secretKeyRef:
              name: discord-token
              key: token
        - name: CONFIG_DATABASE_DSN
          value: "file:/data/disruptor.db?cache=shared"
        - name: CONFIG_LOGGING_LEVEL
          value: "info"
        ports:
        - containerPort: 9090
        volumeMounts:
        - name: data
          mountPath: /data
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: disruptor-data
```

```bash
# Build and deploy
gcloud builds submit --tag gcr.io/PROJECT_ID/disruptor
gcloud run deploy disruptor --image gcr.io/PROJECT_ID/disruptor:latest --platform managed
```

### Azure Container Instances

```bash
# Create resource group
az group create --name disruptor-rg --location eastus

# Create container instance
az container create \
  --resource-group disruptor-rg \
  --name disruptor \
  --image your-registry/disruptor:latest \
  --environment-variables \
    CONFIG_TOKEN=your_discord_bot_token \
    CONFIG_DATABASE_DSN="file:/data/disruptor.db?cache=shared" \
    CONFIG_LOGGING_LEVEL=info \
  --ports 9090 \
  --restart-policy Always \
  --memory 0.5 \
  --cpu 0.5
```

---

## Container Orchestration

### Kubernetes

#### Namespace and ConfigMap
```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: disruptor

---
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: disruptor-config
  namespace: disruptor
data:
  CONFIG_DATABASE_DSN: "file:/data/disruptor.db?cache=shared"
  CONFIG_LOGGING_LEVEL: "info"
  CONFIG_LOGGING_PRETTY: "false"
  CONFIG_METRICS_PORT: "9090"
```

#### Secret
```yaml
# secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: disruptor-secret
  namespace: disruptor
type: Opaque
stringData:
  CONFIG_TOKEN: "your_discord_bot_token"
```

#### Deployment
```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: disruptor
  namespace: disruptor
  labels:
    app: disruptor
spec:
  replicas: 1  # Discord bots should have exactly 1 replica
  selector:
    matchLabels:
      app: disruptor
  template:
    metadata:
      labels:
        app: disruptor
    spec:
      containers:
      - name: disruptor
        image: your-registry/disruptor:latest
        ports:
        - containerPort: 9090
        envFrom:
        - configMapRef:
            name: disruptor-config
        - secretRef:
            name: disruptor-secret
        volumeMounts:
        - name: data
          mountPath: /data
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /metrics
            port: 9090
          initialDelaySeconds: 30
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /metrics
            port: 9090
          initialDelaySeconds: 5
          periodSeconds: 10
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: disruptor-data

---
# pvc.yaml
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

---
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: disruptor-metrics
  namespace: disruptor
  labels:
    app: disruptor
spec:
  ports:
  - port: 9090
    targetPort: 9090
    name: metrics
  selector:
    app: disruptor
```

#### Deploy to Kubernetes
```bash
# Apply manifests
kubectl apply -f namespace.yaml
kubectl apply -f configmap.yaml
kubectl apply -f secret.yaml
kubectl apply -f deployment.yaml

# Check deployment
kubectl get pods -n disruptor
kubectl logs -f deployment/disruptor -n disruptor

# View metrics
kubectl port-forward service/disruptor-metrics 9090:9090 -n disruptor
```

---

## Monitoring and Maintenance

### Health Checks

#### HTTP Health Check
```bash
# Check metrics endpoint
curl -f http://localhost:9090/metrics
echo $?  # Should return 0 if healthy
```

#### Service Health Check
```bash
# Systemd
sudo systemctl is-active disruptor

# Docker
docker health check disruptor

# Kubernetes
kubectl get pods -n disruptor
```

### Log Management

#### Systemd Logs
```bash
# View recent logs
sudo journalctl -u disruptor -n 100

# Follow logs
sudo journalctl -u disruptor -f

# Filter by time
sudo journalctl -u disruptor --since "1 hour ago"

# Search logs
sudo journalctl -u disruptor | grep ERROR
```

#### Docker Logs
```bash
# View logs
docker logs disruptor

# Follow logs
docker logs -f disruptor

# Limit log size
docker logs --tail 100 disruptor
```

### Database Maintenance

#### Backup Database
```bash
# Systemd deployment
sudo -u disruptor cp /opt/disruptor/data/disruptor.db /opt/disruptor/data/backup-$(date +%Y%m%d_%H%M%S).db

# Docker deployment
docker exec disruptor cp /data/disruptor.db /data/backup-$(date +%Y%m%d_%H%M%S).db
```

#### Database Cleanup
```bash
# SQLite vacuum (compact database)
sqlite3 /path/to/disruptor.db "VACUUM;"

# Check database size
ls -lh /path/to/disruptor.db
```

### Updates and Rollbacks

#### Binary Deployment Update
```bash
# Build new version
make build

# Stop service
sudo systemctl stop disruptor

# Backup old binary
sudo cp /opt/disruptor/bin/disruptor /opt/disruptor/bin/disruptor.backup

# Install new binary
sudo cp output/bin/disruptor /opt/disruptor/bin/

# Start service
sudo systemctl start disruptor

# Check status
sudo systemctl status disruptor
```

#### Docker Deployment Update
```bash
# Pull new image
docker pull your-registry/disruptor:latest

# Stop and remove old container
docker stop disruptor
docker rm disruptor

# Run new container
docker run -d --name disruptor --env-file .env -v /host/data:/data your-registry/disruptor:latest

# Or with docker-compose
docker-compose pull
docker-compose up -d
```

#### Rollback
```bash
# Systemd rollback
sudo systemctl stop disruptor
sudo cp /opt/disruptor/bin/disruptor.backup /opt/disruptor/bin/disruptor
sudo systemctl start disruptor

# Docker rollback
docker stop disruptor
docker rm disruptor
docker run -d --name disruptor --env-file .env -v /host/data:/data your-registry/disruptor:previous-tag
```

---

## Security Considerations

### Network Security
- **Firewall**: Only expose necessary ports (9090 for metrics)
- **TLS**: Use HTTPS for metrics if exposed publicly
- **VPN**: Consider VPN access for administration

### Access Control
- **Service User**: Run as dedicated non-root user
- **File Permissions**: Restrict access to configuration and data files
- **Container Security**: Use non-root containers

### Secret Management
- **Environment Variables**: Use for token storage
- **Secret Stores**: Consider vault solutions for production
- **Rotation**: Regular token rotation

### Monitoring
- **Log Analysis**: Monitor for unusual activity
- **Metrics**: Track resource usage and performance
- **Alerts**: Set up alerts for service failures

---

## Troubleshooting Deployments

### Common Issues

#### Service Won't Start
```bash
# Check service status
sudo systemctl status disruptor

# Check logs
sudo journalctl -u disruptor -n 50

# Check configuration
sudo cat /opt/disruptor/.env

# Test binary manually
sudo -u disruptor /opt/disruptor/bin/disruptor
```

#### Container Issues
```bash
# Check container status
docker ps -a

# Check logs
docker logs disruptor

# Check environment
docker exec disruptor env

# Test inside container
docker exec -it disruptor /bin/sh
```

#### Permission Problems
```bash
# Fix ownership
sudo chown -R disruptor:disruptor /opt/disruptor

# Fix permissions
sudo chmod 600 /opt/disruptor/.env
sudo chmod +x /opt/disruptor/bin/disruptor
```

---

## Performance Optimization

### Resource Limits
```bash
# Systemd limits
[Service]
MemoryMax=512M
MemoryHigh=256M
```

### Database Optimization
```bash
CONFIG_DATABASE_DSN="file:/data/disruptor.db?cache=shared&_journal_mode=WAL&_synchronous=NORMAL"
```

### Go Runtime Tuning
```bash
export GOGC=100
export GOMAXPROCS=2
```

---

**Deployment complete!** 🎉 Your Disruptor bot is now running in production.