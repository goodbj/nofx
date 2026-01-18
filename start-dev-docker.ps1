#!/usr/bin/env pwsh
# NOFX Development Docker Deployment Script
# For development at E:\AI\nofx_Dev
# Backend runs on port 8888, Frontend on port 3300

Write-Host "🚀 Starting NOFX Development Environment" -ForegroundColor Green
Write-Host "📁 Path: E:\AI\nofx_Dev" -ForegroundColor Yellow
Write-Host "🔌 Backend: http://localhost:8888" -ForegroundColor Yellow
Write-Host "🌐 Frontend: http://localhost:3300" -ForegroundColor Yellow
Write-Host "💾 Database: E:\AI\nofx_Dev\data\data.db" -ForegroundColor Yellow
Write-Host ""

# Check if Docker is installed and running
try {
    $dockerVersion = docker --version 2>$null
    if (-not $dockerVersion) {
        Write-Host "❌ Docker is not installed or not in PATH" -ForegroundColor Red
        exit 1
    }
    Write-Host "✅ Docker is available: $dockerVersion" -ForegroundColor Green
    
    $composeVersion = docker compose version 2>$null
    if (-not $composeVersion) {
        Write-Host "❌ Docker Compose is not available" -ForegroundColor Red
        exit 1
    }
    Write-Host "✅ Docker Compose is available: $composeVersion" -ForegroundColor Green
} catch {
    Write-Host "❌ Error checking Docker: $_" -ForegroundColor Red
    exit 1
}

# Check if .env file exists
if (-not (Test-Path ".env")) {
    Write-Host "⚠️  .env file not found, using .env.example as template" -ForegroundColor Yellow
    if (Test-Path ".env.example") {
        Copy-Item ".env.example" ".env"
        Write-Host "📋 Copied .env.example to .env, please update with your settings" -ForegroundColor Yellow
    } else {
        Write-Host "❌ Neither .env nor .env.example found" -ForegroundColor Red
        exit 1
    }
}

Write-Host ""
Write-Host "🔧 Building and starting development containers..." -ForegroundColor Cyan

# Start the development environment with docker compose
try {
    # Build and start the services in detached mode
    docker compose -f docker-compose.dev.watch.yml up --build -d
    
    if ($LASTEXITCODE -ne 0) {
        throw "Docker compose command failed"
    }
    
    Write-Host "✅ Development environment started successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "📊 Service Status:" -ForegroundColor Cyan
    
    # Wait a moment for services to start
    Start-Sleep -Seconds 3
    
    # Show the running containers
    docker ps --filter "name=nofx-trading-dev-watch,nofx-frontend-dev-watch" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
    
    Write-Host ""
    Write-Host "🔗 Access your NOFX Development Environment:" -ForegroundColor Green
    Write-Host "   Backend API: http://localhost:8888" -ForegroundColor White
    Write-Host "   Health Check: http://localhost:8888/api/health" -ForegroundColor White
    Write-Host "   Frontend UI: http://localhost:3300" -ForegroundColor White
    Write-Host ""
    Write-Host "🔄 Hot reload is enabled - changes to source code will be reflected automatically" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "📝 To stop the development environment, run:" -ForegroundColor White
    Write-Host "   docker compose -f docker-compose.dev.watch.yml down" -ForegroundColor White
    
} catch {
    Write-Host "❌ Failed to start development environment: $_" -ForegroundColor Red
    exit 1
}