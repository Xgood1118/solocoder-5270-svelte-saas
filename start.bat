@echo off
echo ========================================
echo  SaaS Multi-Tenant System - Dev Server
echo ========================================
echo.

echo [1/2] Starting Go backend on :8080...
start "Backend" cmd /k "cd /d %~dp0backend && go run cmd/api/main.go"

timeout /t 2 /nobreak >nul

echo [2/2] Starting SvelteKit frontend on :5173...
start "Frontend" cmd /k "cd /d %~dp0frontend && npm run dev"

echo.
echo All services started!
echo - Backend API: http://localhost:8080
echo - Frontend:    http://localhost:5173
echo.
pause
