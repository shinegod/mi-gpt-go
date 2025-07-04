# Mi-GPT-Go

基于 Go 语言重构的小米音箱 AI 控制系统，提供高性能的语音对话和设备控制功能。

## 🚀 快速开始

### Docker 部署 (推荐)

1. **克隆项目**
   ```bash
   git clone https://github.com/shinegod/mi-gpt-go.git
   cd mi-gpt-go
   ```

2. **启动服务**
   ```bash
   # 使用 Docker Compose
   docker-compose up -d
   
   # 查看日志
   docker-compose logs -f
   
   # 停止服务
   docker-compose down
   ```

3. **配置系统**
   
   打开浏览器访问 `http://localhost:8080` 进行配置：
   - 在配置管理页面设置小米账号和密码
   - 配置 AI 服务提供商和 API 密钥
   - 测试连接并保存配置

### 直接构建 Docker 镜像

```bash
# 构建镜像
docker build -t mi-gpt-go .

# 运行容器
docker run -d \
  --name mi-gpt-go \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/logs:/app/logs \
  mi-gpt-go
```

## 📋 系统配置

所有配置都通过 Web 管理面板进行设置和管理：

1. **小米设备配置** - 在配置管理页面设置小米账号和密码
2. **AI 服务配置** - 选择 AI 提供商并设置 API 密钥
3. **设备管理** - 获取和选择小米音箱设备
4. **实时保存** - 所有配置都实时保存到数据库中

## 🔧 常用 Docker 命令

```bash
# 查看容器状态
docker-compose ps

# 查看实时日志
docker-compose logs -f mi-gpt-go

# 进入容器
docker-compose exec mi-gpt-go sh

# 重启服务
docker-compose restart mi-gpt-go

# 更新镜像
docker-compose pull
docker-compose up -d
```

## 📁 数据持久化

项目会自动创建以下目录来持久化数据：

- `./data/` - 数据库文件和配置数据
- `./logs/` - 应用日志文件

## 🌐 Web 管理面板

启动后访问 `http://localhost:8080` 使用 Web 管理面板：

- **仪表盘**: 查看系统状态和设备信息
- **配置管理**: 设置 AI 服务和小米设备
- **音箱控制**: TTS 测试和设备控制
- **系统日志**: 实时查看运行日志

## 🆘 故障排除

### 小米登录失败
- 检查账号密码是否正确
- 尝试使用手机号、邮箱或数字 ID
- 查看日志了解具体错误信息

### AI 服务连接失败
- 检查 API 密钥是否有效
- 确认网络连接正常
- 验证 API 基础 URL 是否正确

### 容器启动失败
```bash
# 查看详细错误信息
docker-compose logs mi-gpt-go

# 检查配置文件
docker-compose config
```

## 📖 更多信息

详细文档请查看项目根目录的 [README.md](../README.md) 