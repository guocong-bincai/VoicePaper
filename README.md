# 📖 VoicePaper - 音频同步阅读器

一个现代化的音频同步文本阅读器,支持音频播放时自动高亮对应的文本内容。

## ✨ 功能特性

- 🎵 **音频播放控制**
  - 播放/暂停
  - 前进/后退 10秒
  - 进度条拖动
  - 时间显示

- 📝 **Markdown文本展示**
  - 支持Markdown格式渲染
  - 美观的排版样式
  - 响应式设计

- 🎯 **音频文本同步**
  - 根据音频时间自动高亮对应文本
  - 平滑滚动到当前位置
  - 精确的时间轴匹配

- ⌨️ **键盘快捷键**
  - `空格键`: 播放/暂停
  - `←`: 后退10秒
  - `→`: 前进10秒

## 📁 项目结构

项目已经分离为前后端结构，便于维护和部署：

```
VoicePaper/
├── backend/                # 后端目录 (Go语音合成服务)
│   ├── main.go            # Go主程序 - MiniMax语音合成API调用
│   └── go.mod             # Go模块依赖管理
│
├── frontend/              # 前端目录 (Web音频播放器)
│   ├── index.html         # 主页面
│   ├── style.css          # 样式文件
│   └── app.js             # 核心逻辑 - 音频同步与文本高亮
│
├── data/                  # 数据目录 (音频和文本资源)
│   ├── 1.md               # Markdown文本内容
│   ├── output.mp3         # 生成的音频文件
│   ├── output.tar         # 下载的音频压缩包
│   └── 1925118537643336511_202511251800_337964855271998_337969297908224/
│       ├── content-...titles  # 时间轴数据(JSON格式)
│       └── content-...extra   # 附加数据
│
├── README.md              # 项目文档
└── run.log                # 运行日志
```

## 🚀 快速开始

### 方式一：使用前端播放器（推荐）

如果已经有音频文件，可以直接启动前端播放器：

#### 1. 启动本地服务器

由于浏览器的安全限制,需要使用本地服务器运行:

**使用Python(推荐)**
```bash
# 在项目根目录运行
python3 -m http.server 8000
```

**使用Node.js**
```bash
npx http-server -p 8000
```

**使用VS Code**
- 安装 "Live Server" 插件
- 右键 `frontend/index.html` -> "Open with Live Server"

#### 2. 访问应用

打开浏览器访问: `http://localhost:8000/frontend/`

### 方式二：从文本生成音频

如果需要从Markdown文本生成新的音频：

#### 1. 准备文本

将你的文本内容放入 `data/1.md` 文件

#### 2. 运行后端服务

```bash
cd backend
go run main.go
```

程序会：
1. 读取 `../data/1.md` 文件
2. 调用 MiniMax API 生成语音
3. 下载并解压音频到 `../data/output.mp3`

#### 3. 启动前端

按照"方式一"的步骤启动前端播放器

## 📁 详细说明

### 后端 (backend/)

**main.go** - 语音合成服务
- 调用 MiniMax API 进行文本转语音
- 异步任务处理和状态查询
- 自动下载和解压音频文件
- 支持命令行参数指定文本

**go.mod** - Go模块管理
- 项目依赖配置
- Go版本: 1.21.1

### 前端 (frontend/)

**index.html** - 页面结构
- 响应式布局设计
- 音频播放控制面板
- 文章阅读区域

**style.css** - UI样式
- 现代化杂志风格设计
- 高亮动画效果
- 毛玻璃播放器

**app.js** - 核心逻辑
- 音频播放控制
- 文本同步高亮算法
- 时间轴数据处理

### 数据 (data/)

**1.md** - 文本内容
- Markdown格式
- 支持标题、段落、列表等

**output.mp3** - 音频文件
- 由后端生成
- 格式: MP3, 32kHz, 128kbps

**.titles文件** - 时间轴数据
- JSON格式数组
- 记录每个文本片段的时间范围

### 时间轴数据格式

```json
[
  {
    "text": "文本内容",
    "time_begin": 0,        // 开始时间(毫秒)
    "time_end": 10081,      // 结束时间(毫秒)
    "text_begin": 0,        // 文本起始位置
    "text_end": 51          // 文本结束位置
  }
]
```

## 🎨 界面预览

- **顶部**: 渐变色标题栏
- **中间**: 音频播放器控制面板
- **底部**: 文本内容展示区(带滚动和高亮)

## 🔧 技术栈

- **原生JavaScript**: 核心逻辑
- **marked.js**: Markdown渲染
- **HTML5 Audio API**: 音频控制
- **CSS3**: 现代化UI设计

## 🌟 特色功能

### 1. 智能文本高亮

系统会根据音频播放进度,自动在文本中高亮当前正在朗读的内容,并平滑滚动到可视区域。

### 2. 精确时间同步

通过 `.titles` 时间轴文件,实现毫秒级的音频与文本同步。

### 3. 美观的UI设计

- 渐变色主题
- 圆角卡片设计
- 平滑动画效果
- 响应式布局

## 📱 浏览器支持

- ✅ Chrome/Edge (推荐)
- ✅ Firefox
- ✅ Safari
- ✅ 移动端浏览器

## 🐛 常见问题

### 1. 页面打不开或文件加载失败

**原因**: 浏览器的CORS安全策略限制

**解决**: 必须使用本地服务器运行,不能直接双击打开HTML文件

### 2. 音频无法播放

**检查**:
- 音频文件路径是否正确
- 音频格式是否支持(MP3/WAV/OGG)
- 浏览器是否支持该音频格式

### 3. 文本不高亮

**检查**:
- `.titles` 时间轴文件是否存在
- 时间轴数据格式是否正确
- 打开浏览器控制台查看错误信息

## 🛠️ 自定义配置

### 修改音频文件

在 `frontend/index.html` 中修改:
```html
<audio id="audioPlayer" src="../data/你的音频.mp3"></audio>
```

### 修改Markdown文件

1. 将新的文本文件放入 `data/` 目录
2. 在 `frontend/app.js` 的 `loadMarkdownContent()` 方法中修改:
```javascript
const response = await fetch('../data/你的文件.md');
```

### 修改时间轴文件

在 `frontend/app.js` 的 `loadTimelineData()` 方法中修改文件路径:
```javascript
const response = await fetch('../data/你的目录/你的文件.titles');
```

### 配置后端API

在 `backend/main.go` 中修改 MiniMax API 配置:
```go
const (
    APIKey   = "你的API密钥"
    Model    = "speech-02-hd"
    VoiceID  = "Chinese (Mandarin)_Warm_Bestie"
)
```

## 📝 开发日志

- 2025-11-25: 初始版本发布
  - ✅ 音频播放器功能
  - ✅ Markdown文本渲染
  - ✅ 音频文本同步高亮
  - ✅ 键盘快捷键支持
  - ✅ 前后端代码分离
  - ✅ 项目结构优化

## 📄 许可证

MIT License

## 🤝 贡献

欢迎提交Issue和Pull Request!

---

**享受音频同步阅读的乐趣! 📖🎵**

