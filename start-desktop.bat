@echo off
title ClaimPilot Desktop Launcher
echo ========================================================
echo        ClaimPilot Masaustu Uygulamasi Baslatiliyor
echo ========================================================
echo.

:: 1. Ollama Docker Servislerini Baslat (Eger Docker calisiyorsa)
echo [1/3] Ollama AI Model Servisleri Kontrol Ediliyor...
cd backend\deployments
docker-compose up -d ollama-analyst ollama-verifier >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo [OK] Ollama AI Model Servisleri (Port 11434 & 11435) Aktif!
) else (
    echo [BILGI] Docker veya Ollama arka planda bulunamadi (Standart modda devam ediliyor).
)
cd ..\..

:: 2. Frontend Bagimliliklarini ve Electron Masaustu Uygulamasini Baslat
echo.
echo [2/3] Frontend ve Electron Masaustu Hazirlaniyor...
cd frontend

if not exist node_modules (
    echo [BILGI] Ilk kurulum tespit edildi. npm paketleri yukleniyor (bu islem 1-2 dk surebilir)...
    call npm install
)

echo.
echo ========================================================
echo  [BILGI] Diger PC'lerden Baglanti Adresi (Ayni Wi-Fi/LAN):
echo  http://<YEREL_IP_ADRESINIZ>:5173  (Web Arayuzu)
echo  http://<YEREL_IP_ADRESINIZ>:8080  (Go Backend API)
echo ========================================================
echo.
echo [3/3] ClaimPilot Masaustu Penceresi Aciliyor...
call npm run electron:dev

pause
