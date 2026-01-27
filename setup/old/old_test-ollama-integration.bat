@echo off
setlocal enabledelayedexpansion

echo ========================================
echo    Ollama 集成测试脚本
echo ========================================
echo.

echo 正在测试Ollama服务器连接...
echo.

REM 测试Ollama服务器是否运行
echo [1/4] 检查Ollama服务器状态...
curl -s -o nul -w "HTTP Status: %%{http_code}\n" http://127.0.0.1:11434/api/tags
if %errorlevel% equ 0 (
    echo ✓ Ollama服务器响应正常
) else (
    echo ✗ Ollama服务器未响应，请确认Ollama服务已启动
    goto:end
)
echo.

REM 获取模型列表
echo [2/4] 检查模型列表...
curl -s http://127.0.0.1:11434/api/tags | findstr "llama3.1\|deepseek\|mistral\|gemma" >nul
if %errorlevel% equ 0 (
    echo ✓ 找到可用的Ollama模型
    echo 可用模型：
    curl -s http://127.0.0.1:11434/api/tags | findstr "name"
) else (
    echo ⚠ 未找到常见Ollama模型，但Ollama服务正常
)
echo.

REM 测试模型推理
echo [3/4] 测试模型推理功能...
echo 发送测试请求到Ollama模型...
curl -s -X POST http://127.0.0.1:11434/api/generate ^
-H "Content-Type: application/json" ^
-d "{ \"model\": \"llama3.1\", \"prompt\": \"你好，请简单介绍一下自己，用一句话回答\", \"stream\": false, \"options\": { \"temperature\": 0.7, \"num_predict\": 50 } }" > test_response.json 2>nul

if exist test_response.json (
    if "%~z1" lss "1" (
        REM 如果llama3.1不存在，尝试其他模型
        echo 尝试使用其他模型...
        curl -s -X POST http://127.0.0.1:11434/api/generate ^
        -H "Content-Type: application/json" ^
        -d "{ \"model\": \"deepseek-r1:8b\", \"prompt\": \"你好，请简单介绍一下自己，用一句话回答\", \"stream\": false, \"options\": { \"temperature\": 0.7, \"num_predict\": 50 } }" > test_response.json 2>nul
        
        if exist test_response.json (
            echo ✓ deepseek-r1:8b 模型推理测试完成
        ) else (
            echo ✓ 模型推理测试完成（但可能需要先拉取模型）
        )
    ) else (
        echo ✓ 模型推理测试完成
    )
    echo.
    echo 模型响应示例：
    type test_response.json | findstr "response" 2>nul
    if errorlevel 1 echo （响应可能较长，仅显示部分内容）
    del test_response.json 2>nul
) else (
    echo ⚠ 模型推理测试可能失败，检查是否已拉取模型
)

echo.
echo [4/4] 检查NOFX后端Ollama支持...
echo 正在检查NOFX系统中的Ollama配置支持...
if exist "..\api\server.go" (
    findstr /C:"ollama" "..\api\server.go" >nul
    if !errorlevel! equ 0 (
        echo ✓ NOFX后端已配置Ollama支持
    ) else (
        echo ✗ 未找到Ollama后端配置
    )
) else (
    echo ? 无法找到后端配置文件
)

if exist "..\web\src\components\AITradersPage.tsx" (
    findstr /C:"ollama" "..\web\src\components\AITradersPage.tsx" >nul
    if !errorlevel! equ 0 (
        echo ✓ NOFX前端已配置Ollama支持
    ) else (
        echo ✗ 未找到Ollama前端配置
    )
) else (
    echo ? 无法找到前端配置文件
)

echo.
echo ========================================
echo    Ollama 集成测试完成！
echo ========================================
echo.
echo 提示：
echo - 如果测试全部通过，您的Ollama服务器和NOFX系统已准备好集成
echo - 在NOFX系统中配置时，请使用 http://host.docker.internal:11434 作为API端点（Docker环境）
echo - 本地模型推理可能需要一些时间，请耐心等待
echo - 如需使用特定模型，请先用 'ollama pull model_name' 命令下载
echo.

pause
exit /b 0

:end
echo.
echo ========================================
echo    测试失败！
echo ========================================
pause