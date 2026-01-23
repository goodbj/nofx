#!/bin/bash
# 测试从 Docker 容器内访问宿主机的 Ollama

echo "=== Testing Ollama Connection from Docker Container ==="
echo ""

# 测试 1: Windows/Mac 方式
echo "1. Testing host.docker.internal:11434..."
docker exec nofx-trading-dev-watch wget -qO- --timeout=5 http://host.docker.internal:11434/api/tags 2>&1
if [ $? -eq 0 ]; then
    echo "✅ host.docker.internal:11434 is accessible"
else
    echo "❌ host.docker.internal:11434 failed"
fi
echo ""

# 测试 2: 网关 IP 方式
echo "2. Getting Docker gateway IP..."
GATEWAY_IP=$(docker inspect nofx-trading-dev-watch | grep -m1 '"Gateway"' | awk -F'"' '{print $4}')
echo "   Gateway IP: $GATEWAY_IP"

if [ ! -z "$GATEWAY_IP" ]; then
    echo "   Testing $GATEWAY_IP:11434..."
    docker exec nofx-trading-dev-watch wget -qO- --timeout=5 http://$GATEWAY_IP:11434/api/tags 2>&1
    if [ $? -eq 0 ]; then
        echo "✅ $GATEWAY_IP:11434 is accessible"
    else
        echo "❌ $GATEWAY_IP:11434 failed"
    fi
fi
echo ""

echo "=== Summary ==="
echo "If host.docker.internal works, use: http://host.docker.internal:11434"
echo "If gateway IP works, use: http://$GATEWAY_IP:11434"
