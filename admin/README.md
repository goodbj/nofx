# NOFX 管理员系统

## 项目概述

NOFX 管理员系统是一个全面的管理面板，用于管理 NOFX 交易系统的各个方面。

## 目录结构

```
admin/
├── main.go                 # 主程序入口
├── go.mod                 # Go 模块定义
├── go.sum                 # Go 模块校验
├── .env.example           # 环境配置示例
├── Makefile               # 构建脚本
├── README.md              # 项目说明
├── test_admin_system.go   # 系统测试脚本
├── deploy_admin_system.bat # Windows 部署脚本
├── start_dev_environment.bat # Windows 开发环境启动脚本
├── config/
│   └── config.go          # 系统配置管理
├── models/
│   ├── db.go              # 数据库初始化
│   ├── admin_user.go      # 管理员用户数据模型
│   ├── audit_log.go       # 审计日志数据模型
│   ├── system_config.go   # 系统配置数据模型
│   └── ...                # 其他数据模型
├── auth/
│   ├── auth.go            # 认证服务
│   └── middleware.go      # 认证中间件
├── controllers/
│   ├── admin_controller.go      # 管理员控制器
│   ├── user_controller.go       # 用户管理控制器
│   ├── permission_controller.go # 权限管理控制器
│   ├── audit_controller.go      # 审计日志控制器
│   └── system_config_controller.go # 系统配置控制器
├── routes/
│   └── routes.go          # API 路由定义
├── services/
│   └── email_service.go   # 邮件服务
└── web/
    ├── package.json       # 前端项目配置
    ├── vite.config.ts     # Vite 配置
    ├── tsconfig.json      # TypeScript 配置
    ├── tsconfig.node.json # TypeScript 节点配置
    ├── index.html         # HTML 入口
    ├── src/
    │   ├── App.tsx        # 主应用组件
    │   ├── main.tsx       # 入口文件
    │   ├── index.css      # 全局样式
    │   ├── types/
    │   │   └── global.d.ts # 全局类型定义
    │   ├── utils/
    │   │   └── auth.ts    # 认证工具
    │   ├── services/
    │   │   └── api.ts     # API 服务
    │   └── pages/
    │       ├── Login.tsx        # 登录页面
    │       ├── Dashboard.tsx    # 仪表板页面
    │       ├── Users.tsx        # 用户管理页面
    │       ├── Permissions.tsx  # 权限管理页面
    │       ├── Monitoring.tsx   # 系统监控页面
    │       ├── SystemConfig.tsx # 系统配置页面
    │       ├── AuditLogs.tsx    # 审计日志页面
    │       └── Profile.tsx      # 个人资料页面
```

## 功能模块

1. **认证系统** - JWT 认证、会话管理、密码策略
2. **用户管理** - 用户 CRUD、状态管理、角色分配
3. **权限控制** - 基于角色的访问控制(RBAC)、细粒度权限管理
4. **系统监控** - 系统状态、性能指标、实时监控
5. **配置管理** - 系统参数配置、动态配置更新
6. **安全审计** - 操作日志、审计追踪、合规性报告
7. **仪表板** - 综合数据展示、统计图表、关键指标
8. **个人资料** - 用户个人信息管理、密码修改

## 技术栈

- 后端: Go, Gin, GORM, JWT
- 前端: React, TypeScript, Ant Design
- 数据库: PostgreSQL (或其他兼容 GORM 的数据库)
- 其他: Vite, Axios, React Router

## 快速开始

### 环境要求

- Go 1.19+
- Node.js 16+
- npm 或 yarn
- 数据库 (PostgreSQL, MySQL, SQLite)

### 开发环境启动

1. 克隆项目并进入 admin 目录
2. 复制 `.env.example` 为 `.env` 并填写配置
3. 运行开发环境启动脚本:
   ```bash
   # Windows
   start_dev_environment.bat
   ```

或者手动启动:

```bash
# 后端
go run main.go

# 前端 (在 web 目录下)
cd web
npm install
npm run dev
```

### 部署

运行部署脚本:
```bash
# Windows
deploy_admin_system.bat
```

或者手动部署:

```bash
# 设置环境变量
export DATABASE_URL="your_database_url"
export JWT_SECRET="your_jwt_secret"

# 构建后端
go build -o admin-server .

# 构建前端 (在 web 目录下)
cd web
npm install
npm run build

# 启动服务
./admin-server
```

### 访问系统

- 管理员登录: `http://localhost:3000/login`
- API 访问: `http://localhost:9000/api/`

默认管理员账号:
- 用户名: `admin`
- 密码: `admin123` (首次登录后请立即更改!)

### API 端点

- `POST /api/auth/login` - 用户登录
- `GET /api/dashboard` - 仪表板数据
- `GET /api/users` - 获取用户列表
- `POST /api/users` - 创建用户
- `PUT /api/users/:id` - 更新用户
- `DELETE /api/users/:id` - 删除用户
- `GET /api/permissions` - 获取权限列表
- `POST /api/permissions` - 授予权限
- `GET /api/monitoring/status` - 系统状态
- `GET /api/audit` - 审计日志
- `GET /api/system/config` - 系统配置

## 测试

运行系统测试:
```bash
go run test_admin_system.go
```

该脚本将验证所有系统组件是否正确配置和工作。