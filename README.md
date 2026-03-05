# 📖 VoicePaper - 双语阅读与听写应用

一个现代化的双语阅读和听写应用，支持音频同步文本高亮、文章管理、单词学习等功能。文章、音频和时间轴数据存储在阿里云OSS，通过MySQL数据库管理。

## ✨ 功能特性

### 🎵 音频播放控制
- 播放/暂停
- 前进/后退 10秒
- 进度条拖动
- 时间显示
- 支持OSS私有Bucket访问（自动生成签名URL）
- 

### 📝 文章阅读
- Markdown格式文章渲染
- 美观的排版样式
- 响应式设计
- 文章分类管理
- 每日文章功能

### 🎯 音频文本同步
- 根据音频时间自动高亮对应文本
- 平滑滚动到当前位置
- 精确的时间轴匹配（毫秒级）

### 🎨 阅读主题切换
- 支持浅色/深色两种阅读模式
- 点击左上角 "VoicePaper" logo 即可切换主题
- 主题偏好自动保存到 localStorage
- 平滑的主题切换动画
- 个人中心也提供主题切换选项

### 🔐 用户认证
- 邮箱验证码登录
- GitHub OAuth登录（已实现，待配置）
- Google OAuth登录（已实现，待配置）
- JWT Token认证
- 用户会话管理

### 📚 学习功能
- 单词学习（支持音标、释义、难度等级）
- 句子听写（规划中）
- 单词听写（规划中）
- 学习记录追踪（规划中）

### ⌨️ 键盘快捷键
- `空格键`: 播放/暂停
- `←`: 后退10秒
- `→`: 前进10秒

### 🎨 主题切换
- **切换方式**: 点击左上角 "VoicePaper" logo
- **支持模式**: 浅色模式（默认）、深色模式
- **自动保存**: 主题偏好自动保存，刷新后保持
- **个人中心**: 也可在个人中心设置中切换主题

## 🏗️ 技术栈

### 后端
- **语言**: Go 1.21+
- **Web框架**: Gin
- **ORM**: GORM
- **数据库**: MySQL 8.0+
- **存储**: 阿里云OSS（支持本地存储降级）
- **语音合成**: MiniMax API

### 前端
- **框架**: React 18+
- **语言**: TypeScript
- **构建工具**: Vite
- **样式**: Tailwind CSS
- **状态管理**: Zustand
- **Markdown渲染**: react-markdown

## 📁 项目结构

```
VoicePaper/
├── backend/                    # 后端服务
│   ├── cmd/
│   │   └── server/            # 主服务入口
│   ├── config/                # 配置文件
│   │   ├── config.yaml.example
│   │   └── config.go
│   ├── internal/
│   │   ├── api/v1/           # API路由层
│   │   ├── model/            # 数据模型
│   │   │   ├── article.go
│   │   │   ├── category.go
│   │   │   ├── word.go
│   │   │   └── dictation.go
│   │   ├── repository/       # 数据访问层
│   │   ├── service/          # 业务逻辑层
│   │   └── storage/          # 存储抽象层（OSS/本地）
│   ├── pkg/
│   │   └── minimax/          # MiniMax API客户端
│   ├── docs/                 # 文档
│   │   ├── CREATE_TABLES_MYSQL.sql
│   │   └── MIGRATION_TO_DATABASE.md
│   └── go.mod
│
├── frontend/                  # 前端应用
│   ├── src/
│   │   ├── components/       # React组件
│   │   │   ├── AudioPlayer.tsx
│   │   │   └── ArticleViewer.tsx
│   │   ├── hooks/            # 自定义Hooks
│   │   │   └── useAudioHighlight.ts
│   │   ├── services/         # API服务
│   │   │   └── api.ts
│   │   ├── store/            # 状态管理
│   │   │   └── useStore.ts
│   │   ├── types/            # TypeScript类型定义
│   │   └── App.tsx
│   ├── package.json
│   └── vite.config.ts
│
└── README.md
```

## 🗄️ 数据库设计

### 核心表结构

- **vp_articles** - 文章表
  - 存储文章基本信息（标题、上线状态、分类等）
  - 关联OSS存储的音频、时间轴、文章内容URL

- **vp_categories** - 分类表
  - 文章分类管理

- **vp_sentences** - 句子表
  - 文章中的句子（用于听写练习）

- **vp_words** - 单词表
  - 文章中的单词（音标、释义、难度等级）

- **vp_dictation_records** - 听写记录表（规划中）
  - 用户听写练习记录

- **vp_users** - 用户表
  - 用户基本信息（邮箱、手机号、OAuth绑定等）
  - 支持邮箱验证码登录和OAuth登录

- **vp_verification_codes** - 验证码表
  - 存储邮箱/手机验证码
  - 支持防刷机制（5分钟内最多3次）

- **vp_oauth_bindings** - OAuth绑定表
  - GitHub/Google OAuth账号绑定

- **vp_user_sessions** - 用户会话表
  - JWT Token会话管理

详细数据库设计请参考：`backend/docs/CREATE_TABLES_MYSQL.sql` 和 `backend/docs/USER_AUTH_SCHEMA.sql`

## 🚀 快速开始

### 前置要求

- Go 1.21+
- Node.js 18+
- MySQL 8.0+
- 阿里云OSS账号（可选，也可使用本地存储）

### 1. 克隆项目

```bash
git clone <repository-url>
cd VoicePaper
```

### 2. 配置后端

#### 2.1 复制配置文件

```bash
cd backend
cp config/config.yaml.example config/config.yaml
```

#### 2.2 编辑配置文件

编辑 `backend/config/config.yaml`，配置以下内容：

```yaml
# 数据库配置
database:
  host: "your-mysql-host"
  port: 3306
  database: "voice_paper"
  username: "your-username"
  password: "your-password"
  charset: "utf8mb4"

# OSS配置（可选，如果使用本地存储可跳过）
storage:
  type: "oss"  # 或 "local"
  oss:
    endpoint: "oss-cn-chengdu.aliyuncs.com"
    access_key_id: "your-access-key-id"
    access_key_secret: "your-access-key-secret"
    bucket: "your-bucket-name"
    region: "cn-chengdu"
    use_https: true

# MiniMax API配置
minimax:
  api_key: "your-minimax-api-key"

# 认证配置
auth:
  # JWT配置
  jwt:
    secret: "your-jwt-secret-key"  # 生产环境必须修改为强随机字符串
    expiration: 168  # 过期时间（小时），默认7天
  
  # 邮箱验证码配置（SMTP）
  email:
    smtp_host: "smtp.qq.com"  # SMTP服务器地址
    smtp_port: 587  # SMTP端口（587=STARTTLS, 465=SSL）
    smtp_user: "your_email@qq.com"  # SMTP用户名（邮箱地址）
    smtp_password: "your_email_auth_code"  # SMTP密码（邮箱授权码，不是登录密码）
    from_name: "VoicePaper"  # 发件人名称
    from_email: "your_email@qq.com"  # 发件人邮箱
  
  # OAuth配置（可选）
  oauth:
    github:
      client_id: "your_github_client_id"
      client_secret: "your_github_client_secret"
      redirect_uri: "http://localhost:8080/api/v1/auth/github/callback"
    google:
      client_id: "your_google_client_id"
      client_secret: "your_google_client_secret"
      redirect_uri: "http://localhost:8080/api/v1/auth/google/callback"

# 服务配置
service:
  port: ":8080"
```

#### 2.3 初始化数据库

项目提供了一键初始化脚本，可自动创建数据库、导入表结构和示例数据。

**方法 1：使用初始化脚本（推荐）**

```bash
cd backend
chmod +x init_db.sh
./init_db.sh
```
*脚本会提示输入数据库密码（或自动读取环境变量），并自动导入表结构和演示用的示例文章。*

**方法 2：手动导入**

如果你更喜欢手动操作，可以依次执行以下 SQL 文件：

```bash
# 1. 创建表结构
mysql -u your-username -p your-database < backend/database/schema.sql

# 2. 导入种子数据（包含示例文章和分类）
mysql -u your-username -p your-database < backend/database/seed.sql
```

### 3. 启动后端服务

```bash
cd backend
go mod download
go run cmd/server/main.go
```

后端服务将在 `http://localhost:8080` 启动。

### 4. 配置前端

#### 4.1 安装依赖

```bash
cd frontend
npm install
```

#### 4.2 启动开发服务器

```bash
npm run dev
```

前端应用将在 `http://localhost:5173` 启动。

### 5. 访问应用

打开浏览器访问：`http://localhost:5173`

## 📡 API接口

### 文章相关

#### 获取文章列表
```http
GET /api/v1/articles
```

响应示例：
```json
[
  {
    "id": 1,
    "title": "文章标题",
    "online": "1",
    "category_id": 1,
    "publish_date": "2025-12-03",
    "is_daily": true,
    "audio_url": "https://...",
    "timeline_url": "https://...",
    "article_url": "https://...",
    "created_at": "2025-12-03T13:45:42+08:00"
  }
]
```

#### 获取文章详情
```http
GET /api/v1/articles/:id
```

响应包含：
- 文章基本信息
- 文章内容（从OSS加载的Markdown）
- 音频URL（自动生成签名URL，24小时有效）
- 时间轴URL
- 关联的句子和单词

#### 获取时间轴数据
```http
GET /api/v1/articles/:id/timeline
```

返回时间轴JSON数据，用于音频文本同步。

#### 创建文章
```http
POST /api/v1/articles
Content-Type: application/json

{
  "title": "文章标题",
  "content": "文章内容（Markdown）"
}
```

### 认证相关

#### 发送邮箱验证码
```http
POST /api/v1/auth/email/send
Content-Type: application/json

{
  "email": "user@example.com",
  "purpose": "login"  // login | register | reset_password
}
```

#### 邮箱验证码登录
```http
POST /api/v1/auth/email/login
Content-Type: application/json

{
  "email": "user@example.com",
  "code": "123456"
}
```

响应示例：
```json
{
  "message": "登录成功",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "nickname": null,
    "avatar": null
  }
}
```

#### 获取当前用户信息
```http
GET /api/v1/auth/me
Authorization: Bearer <token>
```

#### GitHub OAuth登录
```http
GET /api/v1/auth/github
```
重定向到GitHub授权页面，授权后回调到前端。

#### Google OAuth登录
```http
GET /api/v1/auth/google
```
重定向到Google授权页面，授权后回调到前端。

## 🔐 OSS存储配置

### 为什么需要OSS签名URL？

如果OSS Bucket设置为私有，直接访问会返回403错误。应用会自动为音频URL生成签名URL，有效期24小时。

### 配置OSS

1. 在阿里云OSS控制台创建Bucket
2. 设置Bucket为私有（推荐）
3. 在配置文件中填入AccessKey信息
4. 上传音频、文章和时间轴文件到OSS
5. 在数据库中记录对应的URL

详细配置指南请参考：`backend/docs/OSS_SETUP_GUIDE.md`

## 🎨 界面预览

- **顶部导航**: 应用Logo（点击可切换主题）和文章统计
- **文章选择器**: 浮动侧边栏按钮，点击展开文章列表
- **文章阅读区**: Markdown渲染，支持音频同步高亮
- **底部播放器**: 悬浮音频播放控制面板
- **主题切换**: 点击左上角 "VoicePaper" logo 切换浅色/深色模式

## 🔧 开发指南

### 后端开发

#### 架构分层

- **API层** (`internal/api/v1/`): 仅处理HTTP请求解析、验证和响应格式化
- **Service层** (`internal/service/`): 包含所有业务逻辑
- **Repository层** (`internal/repository/`): 仅处理数据库操作

#### 添加新功能

1. 在 `internal/model/` 中定义数据模型
2. 在 `internal/repository/` 中添加数据访问方法
3. 在 `internal/service/` 中实现业务逻辑
4. 在 `internal/api/v1/` 中添加API路由

### 前端开发

#### 组件结构

- **Smart Components**: 包含逻辑和状态（如 `App.tsx`）
- **Dumb Components**: 仅负责UI展示（如 `ArticleViewer.tsx`）
- **Custom Hooks**: 抽象逻辑（如 `useAudioHighlight`）

#### 状态管理

使用Zustand进行全局状态管理，定义在 `src/store/useStore.ts`。

## 📝 数据迁移

从本地文件迁移到数据库的详细说明请参考：`backend/docs/MIGRATION_TO_DATABASE.md`

## 🐛 常见问题

### 1. 音频无法播放

**原因**: OSS Bucket是私有的，需要签名URL

**解决**: 
- 确保后端正确配置了OSS AccessKey
- 检查后端日志，确认签名URL生成成功
- 查看浏览器控制台，检查是否有CORS错误

### 2. 数据库连接失败

**检查**:
- MySQL服务是否运行
- 配置文件中的数据库连接信息是否正确
- 数据库用户是否有足够权限

### 3. 前端无法加载数据

**检查**:
- 后端服务是否正常运行（`http://localhost:8080`）
- 浏览器控制台是否有错误信息
- API接口是否返回正确数据

### 4. 文章内容不显示

**检查**:
- 数据库中 `article_url` 字段是否正确
- OSS文件是否存在且可访问
- 后端日志是否有加载错误

### 5. 邮箱验证码发送失败

**可能原因**:
- SMTP配置错误（服务器、端口、授权码）
- QQ邮箱授权码过期或未开启SMTP服务
- 防刷限制（5分钟内最多3次）

**解决方法**:
- 检查 `config.yaml` 中的SMTP配置
- 登录QQ邮箱，重新生成授权码
- 确认SMTP服务已开启（设置 → 账户 → POP3/IMAP/SMTP服务）
- 等待5分钟后重试，或使用临时清理接口：`POST /api/v1/auth/email/clear`

### 6. 登录失败

**检查**:
- 验证码是否正确（6位数字）
- 验证码是否已过期（5分钟有效）
- 验证码是否已被使用（每个验证码只能使用一次）
- 查看后端日志：`tail -50 /tmp/voicepaper_backend.log`

## 🛠️ 部署

### 生产环境配置

1. **后端部署**
   ```bash
   cd backend
   go build -o server cmd/server/main.go
   ./server
   ```

2. **前端构建**
   ```bash
   cd frontend
   npm run build
   # 将 dist/ 目录部署到静态文件服务器
   ```

3. **Nginx配置示例**

   参考 `deploy/nginx.conf` 配置反向代理。

### 环境变量

建议使用环境变量管理敏感配置：

```bash
export MYSQL_PASSWORD="your-password"
export OSS_ACCESS_KEY_SECRET="your-secret"
export MINIMAX_API_KEY="your-api-key"
```

## 📄 许可证

MIT License

## 🤝 贡献

欢迎提交Issue和Pull Request！

## 📚 相关文档

- [数据库设计文档](backend/docs/DATABASE_SCHEMA_DESIGN.md)
- [OSS配置指南](backend/docs/OSS_SETUP_GUIDE.md)
- [数据迁移说明](backend/docs/MIGRATION_TO_DATABASE.md)
- [用户认证系统设计](backend/docs/USER_AUTH_SCHEMA.sql)
- [邮箱SMTP配置指南](backend/docs/EMAIL_SMTP_SETUP.md)
- [OAuth配置指南](backend/docs/OAUTH_SETUP.md)
- [本地测试指南](backend/docs/LOCAL_TESTING_GUIDE.md)
- [API接口文档](backend/docs/)

---

**享受双语阅读和学习的乐趣! 📖🎵**
