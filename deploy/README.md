# 🚀 VoicePaper 部署脚本使用说明

## 📋 部署脚本列表

| 脚本文件 | 用途 | 执行时间 |
|---------|------|---------|
| `deploy.sh` | 完整部署（前端+后端） | ~2-3分钟 |
| `deploy-backend.sh` | 仅部署后端 | ~30秒 |
| `deploy-frontend.sh` | 仅部署前端 | ~1-2分钟 |

---

## 🎯 快速开始

### 1️⃣ 仅部署后端（推荐用于后端代码更新）

```bash
# 在项目根目录执行
./deploy/deploy-backend.sh
```

**适用场景：**
- ✅ 修改了后端代码（Go代码）
- ✅ 修改了API接口
- ✅ 修改了数据库逻辑
- ✅ 修改了配置文件
- ❌ 不涉及前端页面修改

**部署步骤：**
1. 构建Linux AMD64版本的后端程序
2. 上传后端可执行文件和配置
3. 重启后端服务
4. 自动健康检查

---

### 2️⃣ 仅部署前端（推荐用于前端代码更新）

```bash
# 在项目根目录执行
./deploy/deploy-frontend.sh
```

**适用场景：**
- ✅ 修改了前端页面（React/Vue代码）
- ✅ 修改了前端样式
- ✅ 修改了前端路由
- ✅ 修改了前端配置
- ❌ 不涉及后端API修改

**部署步骤：**
1. 安装前端依赖（如果需要）
2. 构建生产版本
3. 清空服务器旧文件
4. 上传新的构建文件
5. 重载Caddy配置

---

### 3️⃣ 完整部署（前端+后端）

```bash
# 在项目根目录执行
./deploy/deploy.sh
```

**适用场景：**
- ✅ 首次部署
- ✅ 前后端都有修改
- ✅ 大版本更新
- ✅ 配置文件重大变更

---

## 🔧 常见问题

### Q1: 部署失败怎么办？

**检查步骤：**
1. 确认你在**项目根目录**运行脚本
2. 检查服务器SSH连接是否正常
   ```bash
   ssh root@<YOUR_SERVER_IP>
   ```
3. 查看详细错误信息

### Q2: 后端部署后接口不通？

**排查方法：**
```bash
# 1. SSH连接到服务器
ssh root@<YOUR_SERVER_IP>

# 2. 检查后端服务状态
systemctl status voicepaper-backend

# 3. 查看后端日志
journalctl -u voicepaper-backend -f

# 4. 手动测试接口
curl http://localhost:8080/api/v1/ping

# 5. 检查端口占用
netstat -tlnp | grep 8080
```

### Q3: 前端部署后看到的还是旧版本？

**解决方案：**
1. **清除浏览器缓存**
   - Chrome/Edge: `Ctrl + Shift + R` (Windows) 或 `Cmd + Shift + R` (Mac)
   - Firefox: `Ctrl + F5`

2. **验证文件是否更新**
   ```bash
   ssh root@<YOUR_SERVER_IP> "ls -lh /var/www/voicepaper/frontend/dist"
   ```

3. **清除CDN缓存**（如果使用了CDN）

### Q4: 如何回滚到上一个版本？

**后端回滚：**
```bash
# 1. SSH到服务器
ssh root@<YOUR_SERVER_IP>

# 2. 恢复备份文件（如果有）
cp /opt/voicepaper/backend/server.backup /opt/voicepaper/backend/server

# 3. 重启服务
systemctl restart voicepaper-backend
```

**前端回滚：**
```bash
# 使用Git回滚到上一个版本，然后重新部署
git checkout HEAD~1 frontend/
./deploy/deploy-frontend.sh
```

### Q5: 如何查看部署日志？

**后端日志：**
```bash
ssh root@<YOUR_SERVER_IP> 'journalctl -u voicepaper-backend -f'
```

**Caddy日志：**
```bash
ssh root@<YOUR_SERVER_IP> 'journalctl -u caddy -f'
```

---

## 📊 部署后验证

### 验证后端

```bash
# 1. Ping接口
curl https://voicepaper.online/api/v1/ping

# 预期响应: {"message":"pong"}

# 2. 书籍列表接口
curl https://voicepaper.online/api/v1/books

# 预期响应: {"code":0,"message":"success","data":[...]}
```

### 验证前端

```bash
# 1. 检查首页
curl -I https://voicepaper.online

# 预期响应: HTTP/2 200

# 2. 浏览器访问
open https://voicepaper.online  # Mac
# 或在浏览器打开 https://voicepaper.online
```

---

## 🛡️ 安全建议

1. **生产环境密钥管理**
   - 不要在脚本中硬编码密码
   - 使用SSH密钥认证
   - 定期轮换密钥

2. **权限控制**
   ```bash
   # 确保部署脚本只有所有者可执行
   chmod 700 deploy/*.sh
   ```

3. **备份策略**
   - 每次部署前自动备份
   - 保留最近3个版本
   - 定期备份数据库

---

## 🔄 持续集成建议

如果使用GitHub Actions或其他CI/CD工具，可以参考：

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [ main ]

jobs:
  deploy-backend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Deploy Backend
        run: ./deploy/deploy-backend.sh

  deploy-frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Deploy Frontend
        run: ./deploy/deploy-frontend.sh
```

---

## 📞 技术支持

遇到问题？
1. 查看日志文件
2. 检查服务器配置
3. 参考 [完整API文档](../backend/docs/BOOK_API.md)

---

## 📝 更新日志

- **2024-02-13**: 新增独立的前端/后端部署脚本
- **2024-02-13**: 添加自动健康检查
- **2024-02-13**: 优化部署流程，提升速度
