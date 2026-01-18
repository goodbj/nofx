@echo off
setlocal enabledelayedexpansion

echo ============================================
echo NOFX 前端功能验证脚本
echo ============================================
echo.
echo 正在验证前端服务...
echo.

REM 检查后端健康状态
echo 1. 检查后端健康状态...
powershell -Command "try { $response = Invoke-RestMethod -Uri http://localhost:8888/api/health; Write-Host '   后端健康检查: OK' -ForegroundColor Green; Write-Host '   状态: ' $response.status; } catch { Write-Host '   后端健康检查: FAILED' -ForegroundColor Red; }"
echo.

REM 提示用户手动验证前端页面
echo 2. 前端页面验证:
echo    请打开浏览器访问: http://localhost:3300
echo.
echo    验证步骤:
echo    a) 确认页面正常加载，无 JavaScript 错误
echo    b) 检查浏览器控制台是否有错误信息
echo    c) 尝试登录系统（如果需要）
echo    d) 访问不同功能页面，确认数据正常加载
echo    e) 检查图表、K线等组件是否正常渲染
echo.
echo 3. 功能验证:
echo    a) 测试添加交易对功能
echo    b) 检查市场数据是否实时更新
echo    c) 验证策略配置功能
echo    d) 确认订单管理功能正常
echo.
echo 4. 错误处理验证:
echo    a) 断开网络连接，检查错误提示
echo    b) 重新连接网络，确认自动恢复
echo.
echo ============================================
echo 验证完成，请按照以上提示进行手动验证
echo ============================================

pause