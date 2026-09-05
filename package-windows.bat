@echo off
title ClaimPilot Windows Exe Package Builder
echo ========================================================
echo        ClaimPilot Windows Executable (.exe) Uretiliyor
echo ========================================================
echo.

cd frontend

if not exist node_modules (
    echo [BILGI] npm paketleri yukleniyor...
    call npm install
)

echo.
echo [BILGI] Masaustu .exe Paketi Derleniyor (Vite + Electron Builder)...
call npm run package:win

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ========================================================
    echo  [BASARILI] Masaustu .exe paketi olusturuldu!
    echo  Konum: frontend\release\
    echo ========================================================
) else (
    echo.
    echo [HATA] Paketleme sirasinda bir sorun olustu.
)

pause
