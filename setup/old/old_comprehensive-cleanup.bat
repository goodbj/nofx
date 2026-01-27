@echo off
setlocal enabledelayedexpansion

echo ========================================
echo    NOFX ȫ�������ű�
echo ========================================
echo.

echo ����ִ��ȫ������...
echo.

REM ����Go��������
echo [1/5] ����Go��������...
go clean -cache -modcache
if %errorlevel% equ 0 (
    echo Go���������������
) else (
    echo ����: Go������������δִ��
)
echo.

REM ����Docker��������
echo [2/5] ����Docker��������...
docker builder prune -f
if %errorlevel% equ 0 (
    echo Docker���������������
) else (
    echo ����: Docker������������δִ��
)
echo.

REM ����Dockerϵͳ����
echo [3/5] ����Dockerϵͳ����...
docker system prune -f
if %errorlevel% equ 0 (
    echo Dockerϵͳ�����������
) else (
    echo ����: Dockerϵͳ������������δִ��
)
echo.

REM ����δʹ�õ�Docker����
echo [4/5] ����δʹ�õ�Docker����...
docker image prune -f
if %errorlevel% equ 0 (
    echo δʹ�õ�Docker�����������
) else (
    echo ����: Docker������������δִ��
)
echo.

REM ������ʱ�ļ�
echo [5/5] ������ʱ�ļ�...
del /q /f /s "%TEMP%\\*nofx*"
del /q /f /s "%TMP%\\*nofx*"
if %errorlevel% equ 0 (
    echo ��ʱ�ļ��������
) else (
    echo ����: ��ʱ�ļ���������δִ��
)
echo.

echo ========================================
echo    ȫ��������ɣ�
echo ========================================
echo.

echo ��������Ŀ����:
echo   - Go���������ģ�黺��
echo   - Docker���������ϵͳ����
echo   - δʹ�õ�Docker����
echo   - ��ʱ�ļ��е�����ļ�
echo.
echo ��ʾ: ��Щ�ļ�ͨ�����԰�ȫɾ�����´�ʹ��ʱ����������
echo.

pause
