# NOFX 部署配置详细文档

## 部署概述

NOFX支持多种部署方式，包括本地开发部署、Docker容器化部署和云平台部署。文档涵盖从环境准备到生产部署的完整流程。

## 部署环境要求

### 硬件要求

#### 最低配置
- **CPU**: 4核 (Intel i5/AMD Ryzen 5或同等性能)
- **内存**: 8GB RAM
- **存储**: 50GB可用空间 (SSD推荐)
- **网络**: 100Mbps带宽

#### 推荐配置
- **CPU**: 8核 (Intel i7/AMD Ryzen 7或更高)
- **内存**: 16GB RAM
- **存储**: 200GB SSD存储
- **网络**: 1Gbps带宽

#### AI本地模型配置
- **运行Ollama模型**: 32GB+内存
- **GPU加速**: NVIDIA RTX 3070 8GB或更高
- **CUDA支持**: 11.8+

### 软件依赖

#### 操作系统
- **Linux**: Ubuntu 20.04+, CentOS 8+, Debian 11+
- **Windows**: Windows 10/11 (需要WSL2)
- **macOS**: 12.0+ (Intel/Apple Silicon)

#### 核心依赖
```bash
# 系统工具
git >= 2.30
curl >= 7.68
wget >= 1.20

# 开发环境
go >= 1.25.3
node >= 18.0
npm >= 8.0
python >= 3.8 (可选)

# 容器化
docker >= 20.10
docker-compose >= 2.0

# 数据库
sqlite3 >= 3.35 (内置)
postgresql >= 13.0 (生产环境)
```

## 本地开发部署

### 1. 环境准备

#### Windows环境 (推荐使用WSL2)
```bash
# 安装WSL2
wsl --install

# 在WSL2中安装依赖
sudo apt update
sudo apt install -y build-essential git curl wget
```

#### Linux/macOS环境
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install -y build-essential git curl wget

# CentOS/RHEL
sudo yum groupinstall -y "Development Tools"
sudo yum install -y git curl wget

# macOS (使用Homebrew)
brew install git curl wget
```

### 2. 安装Go环境

```bash
# 下载Go
wget https://go.dev/dl/go1.25.3.linux-amd64.tar.gz

# 解压到/usr/local
sudo tar -C /usr/local -xzf go1.25.3.linux-amd64.tar.gz

# 配置环境变量
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export GOROOT=/usr/local/go' >> ~/.bashrc
source ~/.bashrc

# 验证安装
go version
```

### 3. 安装前端依赖

```bash
# 安装Node.js (推荐使用NVM)
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
source ~/.bashrc
nvm install 18
nvm use 18

# 或直接安装
# Ubuntu/Debian
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs

# 验证安装
node --version
npm --version
```

### 4. 安装技术指标库

```bash
# Ubuntu/Debian
sudo apt-get install -y libta-lib0-dev

# CentOS/RHEL
sudo yum install -y ta-lib-devel

# macOS
brew install ta-lib

# Windows (通过MSYS2)
pacman -S mingw-w64-x86_64-ta-lib
```

### 5. 克隆项目代码

```bash
# 克隆仓库
git clone https://github.com/NoFxAiOS/nofx.git
cd nofx

# 检出稳定分支
git checkout main
```

### 6. 配置环境变量

```bash
# 复制配置模板
cp .env.example .env

# 编辑配置文件
nano .env
```

**关键配置项**:
```bash
# 服务器配置
NOFX_BACKEND_PORT=8888
NOFX_FRONTEND_PORT=3300

# 认证配置 (必须修改)
JWT_SECRET=your-32-char-random-string-here
DATA_ENCRYPTION_KEY=$(openssl rand -base64 32)
RSA_PRIVATE_KEY=$(openssl genrsa 2048 | sed ':a;N;$!ba;s/\n/\\n/g')

# 数据库配置
DB_TYPE=sqlite
DB_PATH=data/data.db

# AI模型配置
MODEL_MAX_TOKENS_DEEPSEEK=32768
DEEPSEEK_API_KEY=your-deepseek-api-key
```

### 7. 启动开发服务

#### 后端服务启动
```bash
# 安装Go依赖
go mod download

# 编译并运行
go build -o nofx main.go
./nofx

# 或直接运行
go run main.go
```

#### 前端服务启动
```bash
# 进入前端目录
cd web

# 安装前端依赖
npm install

# 启动开发服务器
npm run dev
```

#### 一键启动脚本
```bash
# Windows
01开发版后端8888.bat
02开发版前端3300.bat

# Linux/macOS
./start_dev_backend.sh
./start_dev_frontend.sh
```

## Docker容器化部署

### 1. Docker环境安装

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install -y docker.io docker-compose

# CentOS/RHEL
sudo yum install -y docker docker-compose
sudo systemctl start docker
sudo systemctl enable docker

# Windows/macOS
# 下载Docker Desktop: https://www.docker.com/products/docker-desktop
```

### 2. 基础Docker部署

#### 使用一键脚本
```bash
# Windows
start_nofx_dev_docker.bat

# Linux/macOS
curl -fsSL https://raw.githubusercontent.com/NoFxAiOS/nofx/main/install.sh | bash
```

#### 手动Docker部署
```bash
# 拉取镜像
docker pull nofx/nofx:latest

# 创建网络
docker network create nofx-network

# 启动服务
docker run -d \
  --name nofx-backend \
  --network nofx-network \
  -p 8888:8080 \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/.env:/app/.env \
  nofx/nofx:latest

# 启动前端
docker run -d \
  --name nofx-frontend \
  --network nofx-network \
  -p 3300:3000 \
  nofx/nofx-frontend:latest
```

### 3. Docker Compose部署

#### 开发环境部署
```yaml
# docker-compose.dev.yml
version: '3.8'

services:
  nofx-dev:
    build:
      context: .
      dockerfile: docker/Dockerfile.backend.dev
    ports:
      - "8888:8080"
      - "3300:3000"
    volumes:
      - ./data:/app/data
      - ./.env:/app/.env
    environment:
      - NOFX_BACKEND_PORT=8080
      - NOFX_FRONTEND_PORT=3000
    depends_on:
      - db
    networks:
      - nofx-network

  db:
    image: postgres:13
    environment:
      POSTGRES_DB: nofx
      POSTGRES_USER: nofx_user
      POSTGRES_PASSWORD: nofx_password
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - nofx-network

volumes:
  postgres_data:

networks:
  nofx-network:
    driver: bridge
```

#### 生产环境部署
```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf
      - ./nginx/ssl:/etc/nginx/ssl
    depends_on:
      - backend
      - frontend
    networks:
      - nofx-network

  backend:
    build:
      context: .
      dockerfile: docker/Dockerfile.backend
    environment:
      - DB_TYPE=postgres
      - DB_HOST=db
      - DB_PORT=5432
      - DB_USER=nofx_user
      - DB_PASSWORD=${DB_PASSWORD}
      - DB_NAME=nofx
    depends_on:
      - db
    networks:
      - nofx-network

  frontend:
    build:
      context: ./web
      dockerfile: ../docker/Dockerfile.frontend
    networks:
      - nofx-network

  db:
    image: postgres:13
    environment:
      POSTGRES_DB: nofx
      POSTGRES_USER: nofx_user
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - nofx-network

  redis:
    image: redis:alpine
    volumes:
      - redis_data:/data
    networks:
      - nofx-network

volumes:
  postgres_data:
  redis_data:

networks:
  nofx-network:
    driver: bridge
```

### 4. 热更新开发环境

```yaml
# docker-compose.dev.watch.yml
version: '3.8'

services:
  nofx-dev-watch:
    build:
      context: .
      dockerfile: docker/Dockerfile.backend.dev.watch
    ports:
      - "8888:8080"
      - "3300:3000"
    volumes:
      - .:/app
      - /app/web/node_modules
    environment:
      - NODE_ENV=development
      - WATCH=true
    networks:
      - nofx-network

  nofx-frontend-dev-watch:
    build:
      context: ./web
      dockerfile: ../docker/Dockerfile.frontend.dev.watch
    ports:
      - "3300:3000"
    volumes:
      - ./web:/app
      - /app/node_modules
    environment:
      - VITE_API_BASE_URL=http://localhost:8888
    networks:
      - nofx-network
```

## 云平台部署

### 1. Railway部署

#### 一键部署
[![Deploy on Railway](https://railway.com/button.svg)](https://railway.com/deploy/nofx?referralCode=nofx)

#### 手动配置
```bash
# 安装Railway CLI
npm install -g @railway/cli

# 登录
railway login

# 初始化项目
railway init

# 部署
railway up
```

**Railway环境变量**:
```bash
JWT_SECRET=your-secret-key
DATA_ENCRYPTION_KEY=your-encryption-key
DATABASE_URL=postgresql://user:pass@host:port/db
DEEPSEEK_API_KEY=your-api-key
```

### 2. Heroku部署

```bash
# 安装Heroku CLI
curl https://cli-assets.heroku.com/install.sh | sh

# 登录
heroku login

# 创建应用
heroku create nofx-app

# 设置环境变量
heroku config:set JWT_SECRET=your-secret-key
heroku config:set DATA_ENCRYPTION_KEY=your-encryption-key

# 部署
git push heroku main
```

### 3. AWS部署

#### 使用ECS
```bash
# 创建ECS集群
aws ecs create-cluster --cluster-name nofx-cluster

# 推送镜像到ECR
aws ecr create-repository --repository-name nofx/backend
docker tag nofx/nofx:latest <account>.dkr.ecr.<region>.amazonaws.com/nofx/backend:latest
docker push <account>.dkr.ecr.<region>.amazonaws.com/nofx/backend:latest
```

#### 使用EC2
```bash
# 启动EC2实例
aws ec2 run-instances \
  --image-id ami-0abcdef1234567890 \
  --count 1 \
  --instance-type t3.medium \
  --key-name my-key \
  --security-group-ids sg-0123456789abcdef0
```

### 4. Google Cloud部署

```bash
# 使用Cloud Run
gcloud run deploy nofx-backend \
  --image gcr.io/your-project/nofx:latest \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated

# 使用Compute Engine
gcloud compute instances create nofx-instance \
  --image-family ubuntu-2004-lts \
  --image-project ubuntu-os-cloud \
  --machine-type e2-medium
```

## 反向代理配置

### Nginx配置

```nginx
# nginx/nginx.conf
upstream nofx_backend {
    server backend:8080;
}

upstream nofx_frontend {
    server frontend:3000;
}

server {
    listen 80;
    server_name your-domain.com;
    
    # 重定向到HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;
    
    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;
    
    # 前端静态文件
    location / {
        proxy_pass http://nofx_frontend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    # API接口
    location /api/ {
        proxy_pass http://nofx_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket支持
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
    
    # 健康检查
    location /health {
        proxy_pass http://nofx_backend/health;
    }
}
```

### Apache配置

```apache
<VirtualHost *:80>
    ServerName your-domain.com
    Redirect permanent / https://your-domain.com/
</VirtualHost>

<VirtualHost *:443>
    ServerName your-domain.com
    
    SSLEngine on
    SSLCertificateFile /path/to/cert.pem
    SSLCertificateKeyFile /path/to/key.pem
    
    ProxyPreserveHost On
    
    # 前端代理
    ProxyPass / http://localhost:3300/
    ProxyPassReverse / http://localhost:3300/
    
    # API代理
    ProxyPass /api/ http://localhost:8888/api/
    ProxyPassReverse /api/ http://localhost:8888/api/
    
    # WebSocket支持
    RewriteEngine On
    RewriteCond %{HTTP:Upgrade} websocket [NC]
    RewriteCond %{HTTP:Connection} upgrade [NC]
    RewriteRule ^/?(.*) "ws://localhost:8888/$1" [P,L]
</VirtualHost>
```

## SSL证书配置

### Let's Encrypt免费证书

```bash
# 安装certbot
sudo apt install certbot python3-certbot-nginx

# 获取证书
sudo certbot --nginx -d your-domain.com

# 自动续期
sudo crontab -e
# 添加: 0 12 * * * /usr/bin/certbot renew --quiet
```

### 自签名证书 (开发环境)

```bash
# 生成私钥
openssl genrsa -out key.pem 2048

# 生成证书签名请求
openssl req -new -key key.pem -out csr.pem

# 生成自签名证书
openssl x509 -req -days 365 -in csr.pem -signkey key.pem -out cert.pem

# 清理
rm csr.pem
```

## 监控和日志

### 日志配置

```bash
# 创建日志目录
mkdir -p /var/log/nofx

# 配置日志轮转
cat > /etc/logrotate.d/nofx << EOF
/var/log/nofx/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 root root
}
EOF
```

### 监控配置

```yaml
# docker-compose.monitoring.yml
version: '3.8'

services:
  prometheus:
    image: prom/prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
    networks:
      - nofx-network

  grafana:
    image: grafana/grafana
    ports:
      - "3001:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    networks:
      - nofx-network

  node-exporter:
    image: prom/node-exporter
    ports:
      - "9100:9100"
    networks:
      - nofx-network
```

## 备份和恢复

### 自动备份脚本

```bash
#!/bin/bash
# backup.sh

BACKUP_DIR="/backup/nofx"
DATE=$(date +%Y%m%d_%H%M%S)

# 创建备份目录
mkdir -p ${BACKUP_DIR}/${DATE}

# 备份数据库
if [ "$DB_TYPE" = "sqlite" ]; then
    cp data/data.db ${BACKUP_DIR}/${DATE}/data.db
elif [ "$DB_TYPE" = "postgres" ]; then
    pg_dump -h $DB_HOST -U $DB_USER $DB_NAME > ${BACKUP_DIR}/${DATE}/dump.sql
fi

# 备份配置文件
cp .env ${BACKUP_DIR}/${DATE}/.env
cp -r data/browser_data ${BACKUP_DIR}/${DATE}/browser_data

# 创建压缩包
tar -czf ${BACKUP_DIR}/backup_${DATE}.tar.gz -C ${BACKUP_DIR}/${DATE} .

# 清理旧备份
find ${BACKUP_DIR} -name "backup_*.tar.gz" -mtime +30 -delete

# 上传到云存储 (可选)
# aws s3 cp ${BACKUP_DIR}/backup_${DATE}.tar.gz s3://your-bucket/backups/
```

### 恢复脚本

```bash
#!/bin/bash
# restore.sh

BACKUP_FILE=$1
TEMP_DIR="/tmp/nofx_restore"

# 解压备份文件
mkdir -p ${TEMP_DIR}
tar -xzf ${BACKUP_FILE} -C ${TEMP_DIR}

# 恢复数据库
if [ -f "${TEMP_DIR}/data.db" ]; then
    cp ${TEMP_DIR}/data.db data/data.db
elif [ -f "${TEMP_DIR}/dump.sql" ]; then
    psql -h $DB_HOST -U $DB_USER $DB_NAME < ${TEMP_DIR}/dump.sql
fi

# 恢复配置
cp ${TEMP_DIR}/.env .env
cp -r ${TEMP_DIR}/browser_data data/

# 清理临时文件
rm -rf ${TEMP_DIR}
```

## 故障排除

### 常见部署问题

#### 端口冲突
```bash
# 检查端口占用
netstat -tuln | grep :8888
lsof -i :8888

# 解决方案
# 1. 修改端口配置
# 2. 停止占用进程
sudo kill -9 $(lsof -t -i:8888)
```

#### 权限问题
```bash
# 修复文件权限
sudo chown -R $USER:$USER data/
sudo chmod -R 755 data/

# Docker权限
sudo usermod -aG docker $USER
newgrp docker
```

#### 磁盘空间不足
```bash
# 清理Docker资源
docker system prune -a
docker volume prune

# 清理日志文件
find /var/log/nofx -name "*.log" -mtime +7 -delete
```

### 性能调优

#### Docker资源限制
```yaml
# docker-compose.yml
services:
  nofx-backend:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 4G
        reservations:
          cpus: '1'
          memory: 2G
```

#### 系统优化
```bash
# 调整文件描述符限制
echo "* soft nofile 65536" >> /etc/security/limits.conf
echo "* hard nofile 65536" >> /etc/security/limits.conf

# 调整内核参数
echo 'net.core.somaxconn = 65535' >> /etc/sysctl.conf
echo 'net.ipv4.tcp_max_syn_backlog = 65535' >> /etc/sysctl.conf
sysctl -p
```

---
*部署文档版本: v1.0*
*最后更新: 2026年3月*