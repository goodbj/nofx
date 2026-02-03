import subprocess
import sys
import time

def run_command(cmd, description):
    """执行命令并返回结果"""
    print(f"\n📍 {description}")
    print(f"执行命令: {cmd}")
    try:
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True, timeout=30)
        if result.returncode == 0:
            print("✅ 执行成功")
            if result.stdout.strip():
                print(f"输出: {result.stdout.strip()}")
            return True, result.stdout
        else:
            print("❌ 执行失败")
            if result.stderr.strip():
                print(f"错误: {result.stderr.strip()}")
            return False, result.stderr
    except subprocess.TimeoutExpired:
        print("⏰ 命令执行超时")
        return False, "timeout"
    except Exception as e:
        print(f"💥 执行异常: {e}")
        return False, str(e)

def check_proxy_service():
    """检查代理服务状态"""
    print("=" * 50)
    print("🔍 代理服务环境状态检查")
    print("=" * 50)
    
    # 1. 检查端口占用
    success, output = run_command("netstat -an | findstr :8081", "检查8081端口占用情况")
    
    # 2. 检查进程
    success, output = run_command("tasklist | findstr binance-proxy", "查找binance-proxy进程")
    
    # 3. 测试HTTP连接
    success, output = run_command("curl -v http://localhost:8081/health 2>&1", "测试代理服务健康检查端点")
    
    # 4. 检查Docker容器（如果适用）
    success, output = run_command("docker ps | findstr binance-proxy", "检查Docker中的代理服务容器")
    
    print("\n" + "=" * 50)
    print("📊 检查完成")
    print("=" * 50)

if __name__ == "__main__":
    check_proxy_service()