#!/bin/bash
# 手动扫描功能诊断脚本

echo "=== NOFX 手动扫描功能诊断 ==="
echo ""

# 1. 检查后端服务是否运行
echo "1. 检查后端服务状态..."
curl -s http://localhost:8888/api/health > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "   ✅ 后端服务正常运行"
else
    echo "   ❌ 后端服务未运行或无法访问"
    echo "   请先启动后端服务"
    exit 1
fi
echo ""

# 2. 检查前端服务
echo "2. 检查前端服务状态..."
curl -s http://localhost:3300 > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "   ✅ 前端服务正常运行"
else
    echo "   ❌ 前端服务未运行或无法访问"
fi
echo ""

# 3. 检查是否有交易员在运行
echo "3. 检查交易员状态..."
echo "   请在浏览器中:"
echo "   - 打开 http://localhost:3300/dashboard"
echo "   - 选择一个交易员"
echo "   - 确认交易员状态显示为\"运行中\""
echo ""

# 4. 检查AI模型配置
echo "4. 检查AI模型配置..."
echo "   请确认:"
echo "   - AI模型已配置并启用"
echo "   - Custom API URL 设置为: http://host.docker.internal:11434"
echo "   - 模型名称正确(如: deepseek-coder-v2:latest)"
echo ""

# 5. 测试Ollama连接
echo "5. 测试Docker容器访问Ollama..."
docker exec nofx-dev-backend wget -qO- --timeout=5 http://host.docker.internal:11434/api/tags > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "   ✅ Docker容器可以访问Ollama"
else
    echo "   ❌ Docker容器无法访问Ollama"
    echo "   请检查:"
    echo "   - Ollama服务是否运行 (g:\\Ollama\\ollama.exe serve)"
    echo "   - AI模型URL配置是否正确"
fi
echo ""

# 6. 查看后端日志
echo "6. 查看后端最近日志(最后20行)..."
docker logs nofx-dev-backend --tail 20
echo ""

echo "=== 诊断完成 ==="
echo ""
echo "如果手动扫描仍然无效，请:"
echo "1. 打开浏览器开发者工具 (F12)"
echo "2. 切换到Network标签"
echo "3. 点击手动扫盘按钮"
echo "4. 查看 /api/traders/XXX/execute-decision 请求"
echo "5. 检查请求状态码和响应内容"
echo ""
echo "常见问题:"
echo "- 401: 未登录或token过期"
echo "- 404: API路径错误或trader不存在"
echo "- 400: 交易员未运行"
echo "- 409: AI决策已在执行中"
echo "- 500: 服务器内部错误(查看后端日志)"
