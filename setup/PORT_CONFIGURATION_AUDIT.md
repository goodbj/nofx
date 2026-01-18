# NOFX 端口配置审计报告

## 审计日期
2026年1月18日

## 审计目的
验证 NOFX 项目的开发版和稳定版端口设置是否符合规范要求

## 规范要求
- 稳定版使用 8080/3000 端口
- 开发版使用 8888/3300 端口
- 容器名称独立，避免冲突
- 数据持久化分离
- 两套系统完全独立运行

## 审计结果

### 1. README.md 文档检查
✅ **符合规范**
- 稳定版端口配置：后端8080，前端3000
- 开发版端口配置：后端8888，前端3300
- 容器名称独立：稳定版(nofx-trading, nofx-frontend)，开发版(nofx-trading-dev-watch, nofx-frontend-dev-watch)
- 数据库路径分离：稳定版(E:\AI\nofx\data\data.db)，开发版(E:\AI\nofx_Dev\data\data.db)

### 2. docker-compose.dev.watch.yml 检查
✅ **符合规范**
- 使用环境变量：${NOFX_BACKEND_PORT:-8888} 和 ${NOFX_FRONTEND_PORT:-3300}
- 默认端口：后端8888，前端3300（通过3300:5173映射）
- 容器名称：nofx-trading-dev-watch, nofx-frontend-dev-watch
- 数据卷挂载：../data:/app/data（对应E:\AI\nofx_Dev\data）

### 3. docker-compose.stable.yml 检查
⚠️ **部分符合规范** - 存在潜在问题
- 使用环境变量：${NOFX_BACKEND_PORT:-8888} 和 ${NOFX_FRONTEND_PORT:-3300}
- **问题**：默认值设置为8888/3300，但应为8080/3000
- 端口映射：${NOFX_BACKEND_PORT:-8888}:8088（后端），${NOFX_FRONTEND_PORT:-3300}:80（前端）
- **问题**：默认值不符合稳定版规范
- 容器名称：nofx-trading, nofx-frontend
- 数据卷挂载：../data:/app/data

### 4. install-dev.bat 检查
✅ **符合规范**
- 服务信息显示：前端3300，后端8888
- 数据库路径：E:\AI\nofx_Dev\data\data.db

### 5. install-prod.bat 检查
⚠️ **部分符合规范** - 存在问题
- **问题**：数据库路径显示错误，显示为 E:\AI\nofx_Dev\data\data.db 而不是 E:\AI\nofx\data\data.db
- 应该根据当前目录来确定数据库路径

### 6. system-status.bat 检查
⚠️ **部分符合规范** - 存在问题
- **问题**：服务访问地址显示不正确
- 稳定版前端: http://localhost:3000 ✅
- 稳定版后端: http://localhost:8080 ✅
- 开发版前端: http://localhost:3300 ✅
- 开发版后端: http://localhost:8888 ✅
- 但是文档中提到的端口配置与实际docker-compose文件中的配置需要一致

### 7. deploy-dev.bat 检查
✅ **符合规范**
- 访问地址显示：前端3300，后端8888

### 8. deploy-stable.bat 检查
⚠️ **部分符合规范** - 存在问题
- **问题**：显示前端访问地址为3000，后端为8080，但实际docker-compose.stable.yml中默认值是8888/3300

## 发现的主要问题

### 问题1：docker-compose.stable.yml 默认值错误
**位置**：docker-compose.stable.yml
**问题**：稳定版配置文件中的环境变量默认值设为开发版端口
```
ports:
  - "${NOFX_BACKEND_PORT:-8888}:8088"  # 应该是8080
ports:
  - "${NOFX_FRONTEND_PORT:-3300}:80"    # 应该是3000
```

### 问题2：install-prod.bat 数据库路径错误
**位置**：install-prod.bat
**问题**：显示的数据库路径与目录不符

### 问题3：deploy-stable.bat 描述与实际不符
**位置**：deploy-stable.bat
**问题**：脚本显示的端口与docker-compose文件中的默认值不一致

## 修正建议

### 修正 docker-compose.stable.yml
```yaml
services:
  nofx:
    ports:
      - "${NOFX_BACKEND_PORT:-8080}:8088"  # 修正默认值为8080
    environment:
      - NOFX_BACKEND_PORT=${NOFX_BACKEND_PORT:-8080}
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:${NOFX_BACKEND_PORT:-8080}/api/health"]

  nofx-frontend:
    ports:
      - "${NOFX_FRONTEND_PORT:-3000}:80"  # 修正默认值为3000
    environment:
      - NOFX_BACKEND_PORT=${NOFX_BACKEND_PORT:-8080}
```

### 修正 install-prod.bat
调整数据库路径描述以反映实际部署目录

### 修正 deploy-stable.bat
确保显示的端口信息与实际配置一致

## 总结
- 总体设计思路正确：支持开发版和稳定版独立运行
- 端口分离：8080/3000 (稳定版) vs 8888/3300 (开发版) ✅
- 容器名称独立 ✅
- 数据持久化分离 ✅
- 主要问题是稳定版配置文件中的默认值设置错误，可能导致稳定版使用错误的端口

## 建议
1. 立即修正 docker-compose.stable.yml 中的默认端口值
2. 验证所有脚本中的端口描述与实际配置一致
3. 建议创建 .env.stable 和 .env.dev 文件作为端口配置模板