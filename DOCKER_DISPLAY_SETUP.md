# Docker Display Setup for nofx-dev-web

This guide explains how to set up and run the development version with browser popup functionality.

## Overview

The `nofx-dev-web` configuration includes:
- Frontend: nofx-dev-frontend-display (port 3300)
- Backend: nofx-dev-backend-display (port 8888)
- Feature: Chrome non-headless mode (will show popup windows) for testing browser automation
- Directory: E:\AI\nofx_dev
- Database: E:\AI\nofx_dev\data\data.db (read/write)

## Prerequisites

Before running the display-enabled version, ensure you have:

1. Docker Desktop installed and running
2. Sufficient permissions to run Docker containers
3. At least 4GB free disk space
4. Stable internet connection

## Setup Instructions

### Windows

1. Open Command Prompt or PowerShell as Administrator
2. Navigate to your project directory:
   ```cmd
   cd E:\AI\nofx_Dev
   ```
3. Run the display-enabled version:
   ```cmd
   docker-compose -f docker-compose.dev.display.yml up --build
   ```

Or simply run the batch script:
```cmd
start_dev_display.bat
```

### Linux/macOS

1. Open Terminal
2. Navigate to your project directory:
   ```bash
   cd /path/to/nofx_Dev
   ```
3. Run the display-enabled version:
   ```bash
   docker-compose -f docker-compose.dev.display.yml up --build
   ```

## Accessing the Services

Once the containers are running, you can access:

- Frontend: http://localhost:3300
- Backend API: http://localhost:8888
- Backend Swagger UI (if available): http://localhost:8888/swagger

## Features

- **Browser Automation**: The backend runs with Chrome in non-headless mode, allowing you to see the browser windows during AI interactions
- **Real-time Development**: Both frontend and backend support hot-reloading for rapid development
- **Database Access**: Direct access to the shared data.db file
- **Logging**: Enhanced logging with debug level for troubleshooting

## Stopping the Services

To stop the services, press `Ctrl+C` in the terminal where they are running.

To completely remove the containers:
```bash
docker-compose -f docker-compose.dev.display.yml down
```

## Troubleshooting

### Common Issues

1. **Permission Denied**: Ensure you're running Docker with sufficient privileges
2. **Port Already in Use**: Make sure ports 3300 and 8888 are not being used by other applications
3. **Docker Not Running**: Verify that Docker Desktop is running
4. **Chrome Not Found**: The Docker container should have Chromium installed, but if you encounter issues, verify the Dockerfile

### Debugging Tips

- Check the logs for specific error messages
- Ensure your host machine has enough resources (RAM, CPU)
- If using Windows, ensure WSL2 is properly configured for Docker Desktop

## Security Note

This configuration is intended for development purposes only. Do not use in production environments as it runs browsers in visible mode which may expose sensitive information.

## Environment Variables

The display-enabled version uses the following key environment variables:

- `GUARDIAN_DISPLAY_ENABLED=true` - Enables browser windows to be visible
- `DOCKER_ENV=true` - Indicates the application is running in Docker
- `CHROME_BIN=/usr/bin/chromium-browser` - Points to the Chromium binary in the container
- `LOG_LEVEL=debug` - Provides detailed logging for development