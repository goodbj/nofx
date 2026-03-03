# NOFX

**NOFX** 是一个开源的基于人工智能的多资产量化交易平台，支持加密货币、美股、外汇和贵金属等多市场自动化交易。该系统采用微服务架构设计，提供策略配置、回测验证、AI协作决策和实时交易执行等核心功能。

## 概述

NOFX平台旨在为个人投资者和机构用户提供专业级的量化交易解决方案。系统通过集成多种AI大语言模型，结合技术分析指标和风险管理机制，实现智能化的交易决策和执行。

## 技术架构

### 系统设计
NOFX采用前后端分离的微服务架构：

**后端服务**（Go语言）
- Web框架：Gin v1.11.0
- 数据库：SQLite（默认）/ PostgreSQL（生产环境）
- AI集成：支持DeepSeek、Qwen、OpenAI、Claude等多模型
- 交易所API：Binance、Bybit、OKX、Hyperliquid等10+平台

**前端界面**（React）
- 框架：React 18 + TypeScript
- 构建工具：Vite 6.4
- UI库：Ant Design + TailwindCSS
- 状态管理：Zustand
- 图表库：Lightweight Charts

### 核心组件

#### 1. 策略引擎
策略工作室提供可视化的交易策略配置界面，支持：
- 多种币种选择模式（静态列表、AI500池、OI排名）
- 技术指标配置（EMA、MACD、RSI、ATR等）
- 风险控制参数设置
- AI提示词构建和测试

#### 2. 回测系统
历史数据回测引擎具备以下特性：
- 多品种、多时间周期回测
- 实时进度监控和可视化
- 性能指标计算（夏普比率、最大回撤、胜率等）
- 断点续传和结果分析

#### 3. AI辩论竞技场
创新的多AI协作决策机制：
- 5种AI角色（看涨者、看跌者、分析师、逆向思维者、风险管理员）
- 多轮辩论和投票共识
- 决策过程透明化
- 自动执行共识交易

#### 4. 实时交易执行
- 支持10+主流交易所
- 统一交易接口抽象
- 精度管理和错误处理
- 仓位同步和风险管理

## 功能特性

### 交易支持
- **加密货币**：比特币、以太坊等主流币种
- **美股市场**：苹果、特斯拉、英伟达等股票
- **外汇市场**：EUR/USD、GBP/USD等货币对
- **贵金属**：黄金、白银等商品交易

### AI模型集成
- DeepSeek（成本优化首选）
- 通义千问（Qwen）
- OpenAI GPT系列
- Anthropic Claude
- Google Gemini
- 本地Ollama模型

### 风险管理
- 杠杆控制和仓位限制
- 止损止盈机制
- 资金使用率监控
- 异常交易检测

## 部署方式

### 开发环境
```bash
# Docker快速启动
start_nofx_dev_docker.bat

# 显示模式启动（带浏览器）
start_nofx_dev_display.bat
```

### 生产部署
```bash
# 一键安装脚本
curl -fsSL https://raw.githubusercontent.com/NoFxAiOS/nofx/main/install.sh | bash

# Docker Compose部署
docker compose -f docker-compose.prod.yml up -d
```

## 系统要求

### 硬件配置
- **最低配置**：4核CPU、8GB内存、50GB存储
- **推荐配置**：8核CPU、16GB内存、200GB SSD存储
- **AI模型**：如运行本地模型，建议32GB+内存

### 软件依赖
- Go 1.25.3或更高版本
- Node.js 18+（前端开发）
- Docker 20+（容器化部署）
- TA-Lib技术指标库

## 安全特性

### 数据保护
- AES-256数据加密存储
- RSA非对称加密传输
- 敏感信息浏览器端加密
- 数据库访问控制

### 认证授权
- JWT Token身份验证
- 基于角色的访问控制
- API密钥安全存储
- 会话管理机制

## 社区和发展

### 开源许可
NOFX采用GNU Affero General Public License v3.0（AGPL-3.0）开源许可证。

### 贡献机制
- 贡献者空投计划
- GitHub Issues跟踪
- Telegram开发者社区
- 代码审查和合并流程

### 版本发展
- **v1.0**：基础交易功能和AI集成
- **v1.1**：回测系统和策略工作室
- **v1.2**：AI辩论竞技场
- **v1.3**：多市场支持扩展

## 参考资料

### 技术文档
- [官方文档](https://github.com/NoFxAiOS/nofx/tree/main/docs)
- [架构设计文档](./SYSTEM_ARCHITECTURE.md)
- [API参考文档](./api/README.md)
- [部署指南](./docs/getting-started/README.md)

### 相关链接
- **GitHub仓库**：https://github.com/NoFxAiOS/nofx
- **官方Twitter**：@nofx_official
- **开发者社区**：https://t.me/nofx_dev_community
- **问题反馈**：https://github.com/NoFxAiOS/nofx/issues

### 同类项目对比
| 项目 | NOFX | 其他量化平台 |
|------|------|-------------|
| AI集成 | ✅ 多模型支持 | ⚠️ 单模型为主 |
| 开源性 | ✅ 完全开源 | ❌ 闭源/部分开源 |
| 多市场 | ✅ 4大市场 | ⚠️ 主要支持加密货币 |
| 易用性 | ✅ 可视化配置 | ⚠️ 需要编程基础 |

## 外部链接

- [GitHub项目主页](https://github.com/NoFxAiOS/nofx)
- [Docker Hub镜像](https://hub.docker.com/r/nofx/nofx)
- [技术博客](https://nofx-official.medium.com)
- [开发者文档](https://docs.nofx.ai)

---
**免责声明**：本系统为实验性项目，AI自动交易存在显著风险，建议仅用于学习研究或小资金测试。投资有风险，入市需谨慎。