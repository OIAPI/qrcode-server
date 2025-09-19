# QRCode-Server：二维码生成与识别服务

QRCode-Server 是一个基于豆包编程，Go语言实现的二维码生成与识别的轻量级、高性能的二维码服务，支持通过 API 生成自定义二维码（指定内容、尺寸、格式）和识别二维码（上传文件/远程 URL），内置日志切割、优雅关闭、多环境配置等生产级特性。


## 目录
- [QRCode-Server：二维码生成与识别服务](#qrcode-server二维码生成与识别服务)
  - [目录](#目录)
  - [功能特性](#功能特性)
  - [环境要求](#环境要求)
  - [快速开始](#快速开始)
    - [1. 下载与编译](#1-下载与编译)
    - [2. 配置文件](#2-配置文件)
    - [3. 启动服务](#3-启动服务)
  - [API 使用指南](#api-使用指南)
    - [1. 二维码生成（GET）](#1-二维码生成get)
      - [请求信息](#请求信息)
      - [参数说明](#参数说明)
      - [使用示例](#使用示例)
      - [响应说明](#响应说明)
    - [2. 二维码识别（POST）](#2-二维码识别post)
      - [请求信息](#请求信息-1)
      - [参数说明](#参数说明-1)
      - [使用示例](#使用示例-1)
        - [示例1：上传本地文件识别](#示例1上传本地文件识别)
        - [示例2：通过远程 URL 识别](#示例2通过远程-url-识别)
      - [响应说明](#响应说明-1)
  - [高级配置](#高级配置)
    - [多环境配置切换](#多环境配置切换)
      - [1. 创建多环境配置文件](#1-创建多环境配置文件)
      - [2. 启动时指定配置文件](#2-启动时指定配置文件)
    - [日志配置](#日志配置)
    - [服务配置](#服务配置)
  - [服务管理](#服务管理)
    - [启动参数](#启动参数)
    - [优雅关闭](#优雅关闭)
  - [常见问题](#常见问题)
    - [1. 服务启动失败，提示“无法加载配置”？](#1-服务启动失败提示无法加载配置)
    - [2. 访问 API 提示“404 Not Found”？](#2-访问-api-提示404-not-found)
    - [3. 二维码识别失败，提示“找不到二维码”？](#3-二维码识别失败提示找不到二维码)
    - [4. 日志文件未生成？](#4-日志文件未生成)
  - [联系方式](#联系方式)


## 功能特性
- **双核心能力**：支持二维码生成（PNG/JPEG）和识别（本地文件/远程 URL）；
- **灵活配置**：支持多环境配置（开发/测试/生产），可自定义服务端口、日志规则、二维码参数；
- **日志管理**：日志自动按大小切割、过期清理，同时输出到控制台和文件；
- **优雅关闭**：支持 Ctrl+C 或 `kill` 命令触发优雅关闭，确保正在处理的请求完成；
- **错误处理**：统一的 JSON 错误响应格式，404 路径兜底，便于前端解析。


## 环境要求
- Go 1.20+（推荐 1.21+）
- 依赖管理：Go Modules（默认启用）


## 快速开始

### 1. 下载与编译
```bash
# 1. 克隆代码（或直接下载源码）
git clone https://github.com/your-username/qrcode-server.git
cd qrcode-server

# 2. 安装依赖
go mod tidy

# 3. 编译（根据系统选择命令）
# Linux 64位
GOOS=linux GOARCH=amd64 go build -o qrcode-server-linux main.go
# Windows 64位
GOOS=windows GOARCH=amd64 go build -o qrcode-server-windows.exe main.go
# MacOS 64位
GOOS=darwin GOARCH=amd64 go build -o qrcode-server-darwin main.go
```


### 2. 配置文件
项目根目录需创建 `config.yaml`（默认配置文件），支持自定义服务、日志、二维码参数：
```yaml
# 服务配置
server:
  host: "0.0.0.0"       # 绑定地址：0.0.0.0（外网可访问）、127.0.0.1（仅本地）
  port: "8080"          # 服务端口
  read_timeout: 5       # 读取超时（秒）
  write_timeout: 10     # 写入超时（秒）

# 日志配置
log:
  level: "info"         # 日志级别：debug/info/warn/error
  path: "logs/qrcode.log" # 日志文件路径（自动创建 logs 目录）
  max_size: 10          # 单日志文件最大大小（MB），超过自动切割
  max_age: 7            # 日志保留天数（过期自动删除）
  max_backup: 10        # 最大备份文件数（避免日志文件过多）

# 二维码配置
qrcode:
  default_size: 300     # 生成二维码默认尺寸（像素）
  default_level: "M"    # 默认纠错级别：L(7%)/M(15%)/Q(25%)/H(30%)
  support_types:        # 支持的图片格式
    - "png"
    - "jpeg"
```


### 3. 启动服务
```bash
# 使用默认配置文件（./config.yaml）
./qrcode-server-linux  # Linux
# 或
./qrcode-server-windows.exe  # Windows

# 启动成功后，日志输出如下：
# time=2024-09-20T15:30:00Z level=INFO msg="Logger成功初始化"
# time=2024-09-20T15:30:00Z level=INFO msg="路由成功初始化"
# time=2024-09-20T15:30:00Z level=INFO msg="服务器启动" addr=0.0.0.0:8080
```


## API 使用指南
默认服务地址：`http://localhost:8080`，所有 API 均返回 JSON 格式响应（图片生成接口直接返回图片流）。


### 1. 二维码生成（GET）
通过 URL 参数指定二维码内容、尺寸等，直接返回图片。

#### 请求信息
- 方法：`GET`
- 路径：`/api/qrcode/generate`
- 无需请求体，参数通过 URL 查询串传递。

#### 参数说明
| 参数名   | 类型   | 必传 | 默认值 | 说明                                                                 |
|----------|--------|------|--------|----------------------------------------------------------------------|
| `content`| 字符串 | 是   | -      | 二维码存储的内容（文本、URL、手机号等，支持 UTF-8 中文）             |
| `size`   | 整数   | 否   | 300    | 二维码尺寸（像素），范围 100-2000                                    |
| `type`   | 字符串 | 否   | png    | 图片格式，支持 `png`/`jpeg`                                          |
| `level`  | 字符串 | 否   | M      | 纠错级别：`L`(7%)、`M`(15%)、`Q`(25%)、`H`(30%)                       |

#### 使用示例
```bash
# 示例1：生成默认参数的二维码（300px PNG，M级纠错）
curl "http://localhost:8080/api/qrcode/generate?content=https://example.com" -o example.png

# 示例2：生成自定义参数的二维码（500px JPEG，H级纠错）
curl "http://localhost:8080/api/qrcode/generate?content=测试二维码&size=500&type=jpeg&level=H" -o custom-qrcode.jpg
```

#### 响应说明
- 成功：状态码 `200`，返回图片二进制流（浏览器/工具自动解析为图片）；
- 失败：状态码 `400`（参数错误）/ `500`（服务异常），返回 JSON 错误信息：
  ```json
  {
    "code": -1,                       // 多个错误码，但都为负数
    "message": "content is required"  // 缺少 content 参数
  }
  ```


### 2. 二维码识别（POST）
支持两种识别方式：上传本地文件 或 指定远程图片 URL。

#### 请求信息
- 方法：`POST`
- 路径：`/api/qrcode/decode`
- 内容类型：
  - 上传文件：`multipart/form-data`；
  - 远程 URL：无需指定（参数通过 URL 查询串传递）。

#### 参数说明
| 参数类型   | 参数名   | 类型   | 必传 | 说明                                                                 |
|------------|----------|--------|------|----------------------------------------------------------------------|
| 上传文件   | `file`   | 文件   | 二选一 | 本地二维码图片（支持 PNG/JPEG，建议大小 ≤10MB）                      |
| 远程 URL   | `url`    | 字符串 | 二选一 | 远程图片链接（需公开可访问，支持 PNG/JPEG）                           |

#### 使用示例
##### 示例1：上传本地文件识别
```bash
curl -X POST \
  -F "file=@./test-qrcode.png"  # @后接本地图片路径
  "http://localhost:8080/api/qrcode/decode"
```

##### 示例2：通过远程 URL 识别
```bash
curl -X POST \
  "http://localhost:8080/api/qrcode/decode?url=https://example.com/test-qrcode.jpg"
```

#### 响应说明
- 成功：状态码 `200`，返回识别到的内容：
  ```json
  {
    "code": 1,
    "message": "content", // 二维码的内容
    "data": {
      "content": "https://example.com"  // 二维码中的内容
    }
  }
  ```
- 失败：状态码 `400`（文件无效/URL 错误）/ `500`（识别失败），返回 JSON 错误信息：
  ```json
  {
    "code": -1,                                // 多个错误码，但都为负数
    "message": "图片中未识别到二维码"  // 图片中未识别到二维码
  }
  ```


## 高级配置

### 多环境配置切换
通过命令行参数 `-config` 指定不同环境的配置文件，适配开发、测试、生产场景。

#### 1. 创建多环境配置文件
```
qrcode-server/
├── conf/
│   ├── dev.yaml    # 开发环境（端口 8081，日志级别 debug）
│   ├── test.yaml   # 测试环境（端口 8082，日志级别 info）
│   └── prod.yaml   # 生产环境（端口 80，日志级别 warn）
└── config.yaml     # 默认配置（开发环境）
```

#### 2. 启动时指定配置文件
`qrcode-server-linux` 是编译后的文件，替换为你的文件名。
```bash
# 开发环境
./qrcode-server-linux -config ./conf/dev.yaml

# 生产环境（推荐使用绝对路径）
./qrcode-server-linux -config /etc/qrcode-server/prod.yaml
```


### 日志配置
日志默认同时输出到 **控制台** 和 **文件**，关键配置项说明：
- `log.level`：日志级别，生产环境建议用 `warn`（减少日志量），开发环境用 `debug`（便于调试）；
- `log.path`：日志文件路径，生产环境建议配置为 `/var/log/qrcode-server/qrcode.log`（需确保服务有写入权限并且文件夹必须存在）；
- `log.max_size`：单文件最大大小，建议设置 10-100MB（避免单个日志文件过大）；
- `log.max_age`：日志保留天数，建议 7-30 天（平衡日志追溯和磁盘占用）。


### 服务配置
- `server.host`：绑定地址，生产环境建议配置 `0.0.0.0`（允许外网访问），内网服务可配置为内网 IP（如 `192.168.1.100`）；
- `server.port`：服务端口，生产环境若需用 80/443 端口，需确保服务有 root 权限（Linux）或管理员权限（Windows）；
- `server.read_timeout`/`write_timeout`：超时时间，建议 5-30 秒（避免长连接占用资源）。


## 服务管理

### 启动参数
| 参数名   | 类型   | 默认值         | 说明                                  |
|----------|--------|----------------|---------------------------------------|
| `-config`| 字符串 | `./config.yaml`| 指定配置文件路径（支持相对/绝对路径） |

示例：
```bash
# 使用绝对路径的配置文件
./qrcode-server-linux -config /opt/qrcode-server/conf/prod.yaml
```


### 优雅关闭
服务支持两种优雅关闭方式，关闭前会等待正在处理的请求完成（超时 5 秒）：
1. **Ctrl+C**：直接在终端按下 `Ctrl+C`，服务会输出关闭日志并退出；
2. **kill 命令**：通过进程 ID 关闭服务（适用于后台运行场景）：
   ```bash
   # 1. 查找服务进程 ID
   ps aux | grep qrcode-server-linux
   # 2. 发送 SIGTERM 信号（优雅关闭）
   kill -15 12345  # 12345 替换为实际进程 ID
   ```

关闭成功后，日志输出：
```
time=2024-09-20T16:00:00Z level=INFO msg="服务器关闭..."
time=2024-09-20T16:00:00Z level=INFO msg="服务器关闭成功"
```


## 常见问题

### 1. 服务启动失败，提示“无法加载配置”？
- 检查配置文件路径是否正确（通过 `-config` 参数指定的路径是否存在）；
- 检查配置文件格式是否正确（YAML 格式严格，避免缩进错误）。

### 2. 访问 API 提示“404 Not Found”？
- 检查请求路径是否正确（如生成接口路径是 `/api/qrcode/generate`，而非 `/generate`）；
- 检查服务是否启动成功（查看日志中是否有“服务器启动”的信息）。

### 3. 二维码识别失败，提示“找不到二维码”？
- 检查图片是否清晰（无遮挡、无模糊）；
- 确认图片格式是否支持（仅 PNG/JPEG）；
- 若使用 URL 识别，检查 URL 是否公开可访问（不支持需登录的链接）。

### 4. 日志文件未生成？
- 检查配置的 `log.path` 目录是否存在（服务会自动创建，但需确保父目录有写入权限）；
- 检查服务运行用户是否有日志目录的写入权限（如 Linux 下 `logs` 目录权限是否为 0755）。


## 联系方式
- 作者：OIAPI
- 仓库：https://github.com/OIAPI/qrcode-server
- 问题反馈：欢迎在 GitHub Issues 提交 bug 或需求

> 内容由 豆包 生成
