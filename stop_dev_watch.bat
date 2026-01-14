@echo off
echo Stopping NOFX Development Environment...
echo.

REM Change to project directory
cd /d "e:\AI\nofx_Dev"

echo Stopping development environment...
docker-compose -f docker-compose.dev.watch.yml down

echo.
echo NOFX Development Environment has been stopped!
echo.
echo Press any key to exit...
pause >nul