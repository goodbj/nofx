# NOFX 安全配置详细文档

## 安全概述

NOFX系统采用多层安全防护设计，涵盖数据传输加密、身份认证、访问控制、日志审计等全方位安全机制，确保系统和用户数据的安全性。

## 安全架构设计

### 安全层次模型
```
┌─────────────────────────────────────────────────────────────┐
│                    应用层安全                                │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │  认证授权   │  │  访问控制   │  │  业务逻辑   │         │
│  │   JWT/OAuth │  │   RBAC      │  │   验证      │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
├─────────────────────────────────────────────────────────────┤
│                    传输层安全                                │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   HTTPS     │  │  WebSocket  │  │   API限流   │         │
│  │   TLS 1.3   │  │   安全      │  │   防护      │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
├─────────────────────────────────────────────────────────────┤
│                    数据层安全                                │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │  数据加密   │  │  敏感信息   │  │  数据库     │         │
│  │  AES-256    │  │  脱敏      │  │  安全      │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
├─────────────────────────────────────────────────────────────┤
│                    系统层安全                                │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │  系统加固   │  │  日志审计   │  │  监控告警   │         │
│  │  防火墙     │  │  安全日志   │  │  入侵检测   │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────────────────────────────────────────────┘
```

## 认证与授权安全

### JWT令牌安全

#### 令牌生成配置
```go
// JWT配置
type JWTConfig struct {
    SecretKey      string        `json:"secret_key"`
    ExpirationDays int           `json:"expiration_days"`
    RefreshDays    int           `json:"refresh_days"`
    Issuer         string        `json:"issuer"`
    Algorithm      string        `json:"algorithm"` // HS256, RS256
}

// 推荐配置
JWTConfig{
    SecretKey:      "32-character-random-string-here", // 必须32位以上
    ExpirationDays: 7,   // 令牌有效期7天
    RefreshDays:    30,  // 刷新令牌30天
    Issuer:         "NOFX",
    Algorithm:      "HS256",
}
```

#### 令牌安全策略
```go
// 令牌生成
func GenerateToken(userID string, config JWTConfig) (string, error) {
    claims := jwt.MapClaims{
        "user_id": userID,
        "exp":     time.Now().Add(time.Duration(config.ExpirationDays) * 24 * time.Hour).Unix(),
        "iat":     time.Now().Unix(),
        "iss":     config.Issuer,
        "jti":     uuid.New().String(), // JWT ID防重放
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(config.SecretKey))
}

// 令牌验证
func ValidateToken(tokenString string, config JWTConfig) (*jwt.Token, error) {
    return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(config.SecretKey), nil
    })
}
```

### 多因子认证 (MFA)

#### OTP配置
```yaml
# MFA配置
mfa:
  enabled: true
  totp:
    issuer: "NOFX"
    algorithm: "SHA1"
    digits: 6
    period: 30
  backup_codes:
    count: 10
    length: 16
```

#### TOTP实现
```go
import "github.com/pquerna/otp/totp"

// 生成TOTP密钥
func GenerateTOTPSecret() (string, string, error) {
    key, err := totp.Generate(totp.GenerateOpts{
        Issuer:      "NOFX",
        AccountName: "user@example.com",
        Algorithm:   otp.AlgorithmSHA1,
        Digits:      otp.DigitsSix,
        Period:      30,
    })
    if err != nil {
        return "", "", err
    }
    return key.Secret(), key.String(), nil
}

// 验证TOTP
func ValidateTOTP(secret, code string) bool {
    return totp.Validate(code, secret)
}
```

### OAuth2集成

#### 支持的OAuth提供商
```yaml
oauth:
  google:
    client_id: "your-google-client-id"
    client_secret: "your-google-client-secret"
    redirect_url: "https://your-domain.com/auth/google/callback"
    scopes: ["email", "profile"]
  
  github:
    client_id: "your-github-client-id"
    client_secret: "your-github-client-secret"
    redirect_url: "https://your-domain.com/auth/github/callback"
    scopes: ["user:email"]
```

## 数据传输安全

### HTTPS/TLS配置

#### 服务器TLS配置
```go
// TLS配置
tlsConfig := &tls.Config{
    MinVersion: tls.VersionTLS13,
    CipherSuites: []uint16{
        tls.TLS_AES_128_GCM_SHA256,
        tls.TLS_AES_256_GCM_SHA384,
        tls.TLS_CHACHA20_POLY1305_SHA256,
    },
    PreferServerCipherSuites: true,
    CurvePreferences: []tls.CurveID{
        tls.X25519,
        tls.CurveP256,
    },
}

// HTTP服务器配置
server := &http.Server{
    Addr:      ":443",
    TLSConfig: tlsConfig,
    Handler:   router,
}
```

#### HSTS配置
```go
// HTTP严格传输安全
func hstsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Strict-Transport-Security", 
            "max-age=31536000; includeSubDomains; preload")
        c.Next()
    }
}
```

### API密钥传输加密

#### 浏览器端加密
```javascript
// Web Crypto API加密
async function encryptAPIKey(apiKey, publicKey) {
    const encoder = new TextEncoder();
    const data = encoder.encode(apiKey);
    
    const encrypted = await window.crypto.subtle.encrypt(
        {
            name: "RSA-OAEP"
        },
        publicKey,
        data
    );
    
    return btoa(String.fromCharCode(...new Uint8Array(encrypted)));
}

// 使用示例
const publicKey = await importPublicKey(serverPublicKey);
const encryptedKey = await encryptAPIKey("sk-xxx", publicKey);
```

#### 服务端解密
```go
// RSA解密
func decryptAPIKey(encryptedData, privateKey string) (string, error) {
    block, _ := pem.Decode([]byte(privateKey))
    if block == nil {
        return "", errors.New("failed to decode private key")
    }
    
    key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
    if err != nil {
        return "", err
    }
    
    decoded, err := base64.StdEncoding.DecodeString(encryptedData)
    if err != nil {
        return "", err
    }
    
    decrypted, err := rsa.DecryptOAEP(
        sha256.New(),
        rand.Reader,
        key,
        decoded,
        nil,
    )
    if err != nil {
        return "", err
    }
    
    return string(decrypted), nil
}
```

## 数据存储安全

### 数据库加密

#### 敏感字段加密
```go
// 加密服务
type CryptoService struct {
    aesKey []byte
    rsaKey *rsa.PrivateKey
}

// 加密字段类型
type EncryptedString string

func (es *EncryptedString) Scan(value interface{}) error {
    if value == nil {
        *es = ""
        return nil
    }
    
    bytes, ok := value.([]byte)
    if !ok {
        return errors.New("cannot scan into EncryptedString")
    }
    
    // 解密数据
    decrypted, err := cryptoService.Decrypt(bytes)
    if err != nil {
        return err
    }
    
    *es = EncryptedString(decrypted)
    return nil
}

func (es EncryptedString) Value() (driver.Value, error) {
    if string(es) == "" {
        return nil, nil
    }
    
    // 加密数据
    encrypted, err := cryptoService.Encrypt(string(es))
    if err != nil {
        return nil, err
    }
    
    return encrypted, nil
}
```

#### 数据库字段加密示例
```go
type Exchange struct {
    ID          string          `gorm:"primaryKey"`
    UserID      string          `gorm:"index"`
    Name        string          `gorm:"size:50"`
    APIKey      EncryptedString `gorm:"size:1000"`  // 加密存储
    APISecret   EncryptedString `gorm:"size:1000"`  // 加密存储
    Passphrase  EncryptedString `gorm:"size:500"`   // 加密存储
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### 数据脱敏

#### 日志脱敏
```go
// 敏感信息脱敏
func maskSensitiveInfo(data string) string {
    // API密钥脱敏
    re := regexp.MustCompile(`(sk-[a-zA-Z0-9]{32})`)
    data = re.ReplaceAllString(data, "sk-***")
    
    // 邮箱脱敏
    emailRe := regexp.MustCompile(`([a-zA-Z0-9._%+-]+)@([a-zA-Z0-9.-]+\.[a-zA-Z]{2,})`)
    data = emailRe.ReplaceAllString(data, "$1@***.$2")
    
    // 手机号脱敏
    phoneRe := regexp.MustCompile(`(\d{3})\d{4}(\d{4})`)
    data = phoneRe.ReplaceAllString(data, "$1****$2")
    
    return data
}
```

#### 响应数据脱敏
```go
// API响应脱敏中间件
func dataMaskingMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        
        // 对响应数据进行脱敏
        if c.Writer.Status() == 200 {
            body := c.GetString("response_body")
            if body != "" {
                maskedBody := maskSensitiveInfo(body)
                c.Set("response_body", maskedBody)
            }
        }
    }
}
```

## 访问控制安全

### 基于角色的访问控制 (RBAC)

#### 角色定义
```go
type Role string

const (
    RoleAdmin    Role = "admin"
    RoleUser     Role = "user"
    RoleGuest    Role = "guest"
    RoleAuditor  Role = "auditor"
)

type Permission string

const (
    PermReadTrader    Permission = "trader:read"
    PermWriteTrader   Permission = "trader:write"
    PermReadStrategy  Permission = "strategy:read"
    PermWriteStrategy Permission = "strategy:write"
    PermRunBacktest   Permission = "backtest:run"
    PermSystemAdmin   Permission = "system:admin"
)
```

#### 权限检查
```go
// 权限检查中间件
func requirePermission(permission Permission) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := getUserIDFromContext(c)
        userRole := getUserRoleFromContext(c)
        
        // 检查权限
        if !hasPermission(userRole, permission) {
            c.JSON(403, gin.H{
                "error": "FORBIDDEN",
                "message": "Insufficient permissions",
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// 权限映射
var rolePermissions = map[Role][]Permission{
    RoleAdmin: {
        PermReadTrader, PermWriteTrader,
        PermReadStrategy, PermWriteStrategy,
        PermRunBacktest, PermSystemAdmin,
    },
    RoleUser: {
        PermReadTrader, PermWriteTrader,
        PermReadStrategy, PermWriteStrategy,
        PermRunBacktest,
    },
    RoleGuest: {
        PermReadTrader, PermReadStrategy,
    },
}
```

### IP白名单控制

```go
// IP白名单配置
type SecurityConfig struct {
    IPWhitelist []string `json:"ip_whitelist"`
    IPBlacklist []string `json:"ip_blacklist"`
    RateLimit   int      `json:"rate_limit"` // 每分钟请求数
}

// IP检查中间件
func ipWhitelistMiddleware(config SecurityConfig) gin.HandlerFunc {
    return func(c *gin.Context) {
        clientIP := c.ClientIP()
        
        // 检查黑名单
        if isIPInList(clientIP, config.IPBlacklist) {
            c.JSON(403, gin.H{
                "error": "IP_BLOCKED",
                "message": "Your IP address is blocked",
            })
            c.Abort()
            return
        }
        
        // 检查白名单（如果配置了）
        if len(config.IPWhitelist) > 0 && !isIPInList(clientIP, config.IPWhitelist) {
            c.JSON(403, gin.H{
                "error": "IP_NOT_ALLOWED",
                "message": "Your IP address is not in whitelist",
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

## API安全防护

### 速率限制

```go
import "golang.org/x/time/rate"

// 令牌桶限流器
type RateLimiter struct {
    visitors map[string]*rate.Limiter
    mutex    sync.RWMutex
    r        rate.Limit // 每秒令牌数
    b        int        // 桶容量
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
    return &RateLimiter{
        visitors: make(map[string]*rate.Limiter),
        r:        r,
        b:        b,
    }
}

func (rl *RateLimiter) GetLimiter(ip string) *rate.Limiter {
    rl.mutex.Lock()
    defer rl.mutex.Unlock()
    
    limiter, exists := rl.visitors[ip]
    if !exists {
        limiter = rate.NewLimiter(rl.r, rl.b)
        rl.visitors[ip] = limiter
    }
    
    return limiter
}

// 限流中间件
func rateLimitMiddleware(rl *RateLimiter) gin.HandlerFunc {
    return func(c *gin.Context) {
        ip := c.ClientIP()
        limiter := rl.GetLimiter(ip)
        
        if !limiter.Allow() {
            c.JSON(429, gin.H{
                "error": "RATE_LIMIT_EXCEEDED",
                "message": "Too many requests",
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### 输入验证和清理

```go
// 输入验证
type TraderCreateRequest struct {
    Name       string `json:"name" binding:"required,min=1,max=100"`
    ExchangeID string `json:"exchange_id" binding:"required"`
    StrategyID string `json:"strategy_id" binding:"required"`
    Config     json.RawMessage `json:"config" binding:"required"`
}

// 自定义验证器
func validateJSONSchema(data []byte, schema string) error {
    // 使用JSON Schema验证
    schemaLoader := gojsonschema.NewStringLoader(schema)
    documentLoader := gojsonschema.NewBytesLoader(data)
    
    result, err := gojsonschema.Validate(schemaLoader, documentLoader)
    if err != nil {
        return err
    }
    
    if !result.Valid() {
        var errors []string
        for _, desc := range result.Errors() {
            errors = append(errors, desc.String())
        }
        return fmt.Errorf("validation failed: %s", strings.Join(errors, ", "))
    }
    
    return nil
}
```

### CORS安全配置

```go
// 安全的CORS配置
func secureCORS() gin.HandlerFunc {
    return cors.New(cors.Config{
        AllowOrigins: []string{
            "https://your-domain.com",
            "https://app.your-domain.com",
        },
        AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders: []string{
            "Origin", "Content-Type", "Accept", 
            "Authorization", "X-Requested-With",
        },
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    })
}
```

## 安全监控和审计

### 安全日志

```go
// 安全事件日志
type SecurityEvent struct {
    ID          string    `json:"id"`
    EventType   string    `json:"event_type"` // LOGIN_FAILED, AUTH_TOKEN_EXPIRED, etc.
    UserID      string    `json:"user_id"`
    IPAddress   string    `json:"ip_address"`
    UserAgent   string    `json:"user_agent"`
    Description string    `json:"description"`
    Severity    string    `json:"severity"` // LOW, MEDIUM, HIGH, CRITICAL
    Details     JSON      `json:"details"`
    Timestamp   time.Time `json:"timestamp"`
}

// 安全日志记录
func logSecurityEvent(event SecurityEvent) {
    // 记录到安全日志文件
    logFile, err := os.OpenFile("security.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil {
        log.Printf("Failed to open security log: %v", err)
        return
    }
    defer logFile.Close()
    
    event.Timestamp = time.Now()
    event.ID = uuid.New().String()
    
    jsonData, _ := json.Marshal(event)
    fmt.Fprintf(logFile, "%s\n", string(jsonData))
    
    // 发送告警（高严重性事件）
    if event.Severity == "HIGH" || event.Severity == "CRITICAL" {
        sendSecurityAlert(event)
    }
}
```

### 入侵检测

```go
// 异常行为检测
type IntrusionDetector struct {
    failedLoginAttempts map[string]int
    mutex               sync.RWMutex
    threshold           int
}

func NewIntrusionDetector(threshold int) *IntrusionDetector {
    return &IntrusionDetector{
        failedLoginAttempts: make(map[string]int),
        threshold:           threshold,
    }
}

func (id *IntrusionDetector) RecordFailedLogin(ip string) {
    id.mutex.Lock()
    defer id.mutex.Unlock()
    
    id.failedLoginAttempts[ip]++
    
    if id.failedLoginAttempts[ip] >= id.threshold {
        // 记录安全事件
        logSecurityEvent(SecurityEvent{
            EventType:   "BRUTE_FORCE_ATTEMPT",
            IPAddress:   ip,
            Severity:    "HIGH",
            Description: fmt.Sprintf("Failed login attempts exceeded threshold (%d)", id.threshold),
        })
        
        // 可以考虑临时封禁IP
        blockIP(ip, time.Hour)
    }
}
```

## 安全配置最佳实践

### 环境变量安全

```bash
# .env安全配置示例
# ===========================================
# 安全配置
# ===========================================

# JWT配置
JWT_SECRET=your-32-character-random-secret-key-here
JWT_EXPIRATION_DAYS=7

# 加密密钥
DATA_ENCRYPTION_KEY=$(openssl rand -base64 32)
RSA_PRIVATE_KEY=$(openssl genrsa 2048 | sed ':a;N;$!ba;s/\n/\\n/g')

# 数据库安全
DB_PASSWORD=your-secure-database-password
DB_SSLMODE=require

# API密钥安全
DEEPSEEK_API_KEY=sk-xxx
BINANCE_API_KEY=your-binance-api-key

# 安全选项
TRANSPORT_ENCRYPTION=true
SECURE_COOKIES=true
HSTS_ENABLED=true
```

### 防火墙配置

```bash
# UFW防火墙配置
sudo ufw default deny incoming
sudo ufw default allow outgoing

# 允许必要端口
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS
sudo ufw allow 8888/tcp  # NOFX后端API

# 启用防火墙
sudo ufw enable
```

### 系统加固

```bash
# 系统安全加固脚本
#!/bin/bash

# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装安全工具
sudo apt install -y fail2ban unattended-upgrades

# 配置自动安全更新
echo 'Unattended-Upgrade::Allowed-Origins {
    "${distro_id}:${distro_codename}";
    "${distro_id}:${distro_codename}-security";
};' | sudo tee /etc/apt/apt.conf.d/50unattended-upgrades

# 配置Fail2Ban
sudo cp /etc/fail2ban/jail.conf /etc/fail2ban/jail.local
sudo systemctl enable fail2ban
sudo systemctl start fail2ban

# 禁用不必要的服务
sudo systemctl disable bluetooth
sudo systemctl disable cups
```

## 安全测试和扫描

### 依赖安全扫描

```bash
# Go依赖安全扫描
go install github.com/securego/gosec/v2/cmd/gosec@latest
gosec ./...

# Node.js依赖安全扫描
npm audit
npm audit fix

# Docker镜像安全扫描
docker scan nofx/nofx:latest
```

### 渗透测试检查清单

```markdown
## 渗透测试检查清单

### 认证安全
- [ ] 弱密码测试
- [ ] 会话固定攻击
- [ ] JWT令牌劫持
- [ ] 令牌过期测试
- [ ] 多账户登录测试

### 授权安全
- [ ] 垂直权限提升
- [ ] 水平权限绕过
- [ ] IDOR漏洞测试
- [ ] 角色权限验证

### 输入验证
- [ ] SQL注入测试
- [ ] XSS攻击测试
- [ ] 命令注入测试
- [ ] 文件上传漏洞
- [ ] 路径遍历测试

### 业务逻辑
- [ ] 竞态条件测试
- [ ] 业务流程绕过
- [ ] 价格篡改测试
- [ ] 数量限制绕过

### 配置安全
- [ ] 敏感信息泄露
- [ ] 调试信息暴露
- [ ] 错误信息泄露
- [ ] 配置文件权限
```

## 应急响应计划

### 安全事件响应流程

```markdown
## 安全事件响应流程

### 1. 事件检测
- 监控系统告警
- 日志异常分析
- 用户报告

### 2. 事件评估
- 确定事件严重性
- 评估影响范围
- 确定响应优先级

### 3. 应急响应
- 隔离受影响系统
- 收集证据
- 阻止进一步损害

### 4. 修复恢复
- 修复安全漏洞
- 恢复系统功能
- 验证修复效果

### 5. 事后分析
- 事件根本原因分析
- 改进措施制定
- 文档更新
```

### 备份和恢复策略

```bash
# 安全备份脚本
#!/bin/bash

BACKUP_DIR="/secure/backup"
DATE=$(date +%Y%m%d_%H%M%S)

# 创建加密备份
tar -czf - /data | gpg --cipher-algo AES256 --compress-algo 1 --symmetric --output ${BACKUP_DIR}/backup_${DATE}.tar.gz.gpg

# 验证备份完整性
gpg --verify ${BACKUP_DIR}/backup_${DATE}.tar.gz.gpg

# 清理旧备份
find ${BACKUP_DIR} -name "backup_*.tar.gz.gpg" -mtime +30 -delete
```

---
*安全文档版本: v1.0*
*最后更新: 2026年3月*