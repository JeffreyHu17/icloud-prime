# iCloud Prime

iCloud Prime 是一个本地运行的 iCloud Hide My Email 隐私邮箱管理工具。它提供网页管理台和 HTTP API，可用于管理多个 iCloud 账号、创建隐藏邮箱别名、查看别名列表，并读取发往隐藏邮箱别名的邮件。

仓库地址：[https://github.com/forever94yu/icloud-prime](https://github.com/forever94yu/icloud-prime)

## 重要安全说明

1. 本仓库不会包含任何真实账号信息。
2. Windows 10 便携版 Release 包不会包含真实 `accounts.json`。
3. 你的账号 Cookie、App 专用密码、代理地址等只应该保存在本机 `data/accounts.json` 中。
4. 不要把 `data/accounts.json`、`.env`、`logs/`、`build/` 或任何 `.exe` 文件提交到 GitHub。
5. Cookie 和 App 专用密码等同于账号访问凭据，请只在自己的电脑上使用。

## 功能

- 本地网页管理台，默认地址为 `http://127.0.0.1:8081`
- 多 iCloud 账号管理
- 创建 iCloud Hide My Email 隐藏邮箱别名
- 查看账号下的隐藏邮箱别名
- 停用、重新启用、删除隐藏邮箱别名
- 读取邮件：优先使用 IMAP App 专用密码，失败后可回退到 iCloud Web Cookie
- 支持 `icloud.com` 和 `icloud.com.cn`
- 支持 HTTP/SOCKS5 代理配置

## 方式一：Windows 10 便携版

这是最简单的使用方式，不需要安装 Go 或 Node.js。

### 第 1 步：下载

1. 打开 Releases 页面：
   [https://github.com/forever94yu/icloud-prime/releases](https://github.com/forever94yu/icloud-prime/releases)
2. 下载文件：
   `icloud-prime-windows10-portable-v0.1.0.zip`

### 第 2 步：解压

1. 右键 zip 文件。
2. 选择“全部解压缩”。
3. 建议解压到一个固定目录，例如：
   `D:\Tools\icloud-prime`

解压后目录大致如下：

```text
icloud-prime-windows10-portable-v0.1.0/
├── icloud-prime.exe
├── start.bat
├── stop.bat
├── README-Usage.txt
├── data/
│   └── accounts.example.json
└── logs/
```

### 第 3 步：启动

1. 双击 `start.bat`。
2. 等待几秒钟。
3. 打开浏览器访问：
   [http://127.0.0.1:8081](http://127.0.0.1:8081)

如果你想手动启动，也可以在该目录打开命令行运行：

```powershell
.\icloud-prime.exe -addr :8081 -data .\data
```

### 第 4 步：添加账号

打开网页管理台后：

1. 找到账号管理区域。
2. 添加一个账号名称，例如 `主账号`。
3. `host` 一般填写 `icloud.com`。
4. 如果你使用国区 iCloud，可以填写 `icloud.com.cn`。
5. 如果暂时没有 Cookie，可以先添加账号，后续再登录或更新 Cookie。

账号数据会保存到便携目录里的：

```text
data/accounts.json
```

这个文件由程序自动生成，里面可能包含真实 Cookie 和 App 专用密码，请不要上传或分享。

### 第 5 步：配置 Cookie

Cookie 用于创建、管理隐藏邮箱别名，也可作为读取邮件的 Web API 回退方式。

获取 Cookie 的常见方式：

1. 在浏览器打开 [https://www.icloud.com](https://www.icloud.com) 或 [https://www.icloud.com.cn](https://www.icloud.com.cn)。
2. 登录你的 iCloud 账号。
3. 打开浏览器开发者工具。
4. 进入 Application 或 Storage 面板。
5. 找到 Cookies。
6. 复制 iCloud 相关 Cookie。
7. 在管理台中更新账号 Cookie。

程序支持两种 Cookie 输入格式。

格式一：Header 字符串：

```text
X-APPLE-WEBAUTH-TOKEN=你的值; X-APPLE-WEBAUTH-USER=你的值; X-APPLE-DS-WEB-SESSION-TOKEN=你的值
```

格式二：JSON 对象：

```json
{
  "X-APPLE-WEBAUTH-TOKEN": "你的值",
  "X-APPLE-WEBAUTH-USER": "你的值",
  "X-APPLE-DS-WEB-SESSION-TOKEN": "你的值"
}
```

Cookie 会过期。如果创建别名或读取邮件返回 401/403，请重新获取 Cookie。

### 第 6 步：配置 App 专用密码

App 专用密码用于 IMAP 读信。它不是你的 Apple ID 登录密码。

生成步骤：

1. 打开 [https://appleid.apple.com](https://appleid.apple.com)。
2. 登录 Apple ID。
3. 进入“登录与安全”。
4. 找到“App 专用密码”。
5. 创建一个新的 App 专用密码。
6. 在管理台中填写 iCloud 邮箱地址和 App 专用密码。

示例字段：

```json
{
  "icloud_email": "your_email@icloud.com",
  "app_password": "xxxx-xxxx-xxxx-xxxx"
}
```

### 第 7 步：创建隐藏邮箱别名

1. 在网页管理台选择账号。
2. 输入一个标签，例如 `注册某网站`。
3. 点击创建。
4. 成功后会得到一个隐藏邮箱地址，例如 `example@icloud.com`。

也可以使用 API：

```bash
curl -X POST http://127.0.0.1:8081/api/create \
  -H "Content-Type: application/json" \
  -d "{\"account_id\":\"acc_1\",\"label\":\"注册某网站\"}"
```

### 第 8 步：读取邮件

在网页管理台选择账号和别名后读取邮件。

也可以使用 API：

```bash
curl "http://127.0.0.1:8081/api/inbox?account_id=acc_1&alias=your_alias@icloud.com&limit=20&days=7"
```

读取优先级：

1. 已配置 App 专用密码时，优先使用 IMAP。
2. IMAP 不可用时，尝试使用 Cookie 走 Web API。

### 第 9 步：停止

双击：

```text
stop.bat
```

或者在命令行中结束进程：

```powershell
taskkill /F /IM icloud-prime.exe
```

## 方式二：从源码运行

### 第 1 步：安装依赖

需要安装：

- Git
- Go 1.23 或更高版本
- Node.js 20 或更高版本，仅在需要重新构建前端时使用

### 第 2 步：克隆仓库

```bash
git clone https://github.com/forever94yu/icloud-prime.git
cd icloud-prime
```

### 第 3 步：下载 Go 依赖

```bash
go mod download
```

### 第 4 步：可选，重新构建前端

仓库已包含已构建的前端静态文件。如果你修改了 `web/` 目录，需要重新构建：

```bash
cd web
npm install
npm run build
cd ..
```

`web/vite.config.ts` 会把构建结果输出到：

```text
internal/server/static/dist
```

Go 程序会把这个目录嵌入到最终可执行文件里。

### 第 5 步：构建 Windows 可执行文件

```bash
go build -ldflags="-s -w" -o icloud-prime.exe .
```

### 第 6 步：启动

```bash
.\icloud-prime.exe -addr :8081 -data .\data
```

打开：

[http://127.0.0.1:8081](http://127.0.0.1:8081)

## 常用启动参数

```bash
.\icloud-prime.exe
```

默认监听 `:8081`，默认数据目录为 `.\data`。

```bash
.\icloud-prime.exe -addr :9000
```

指定端口为 `9000`。

```bash
.\icloud-prime.exe -data D:\icloud-prime-data
```

指定账号数据目录。

```bash
.\icloud-prime.exe -debug
```

启用调试日志。

## API 简表

### 账号管理

```text
GET    /api/accounts
POST   /api/accounts
DELETE /api/accounts/:id
POST   /api/accounts/:id/password
PUT    /api/accounts/:id/cookies
POST   /api/accounts/:id/login
```

### 隐藏邮箱别名

```text
POST   /api/create
GET    /api/aliases?account_id=acc_1
POST   /api/aliases/:id/deactivate
POST   /api/aliases/:id/reactivate
DELETE /api/aliases/:id
```

### 邮件

```text
GET /api/inbox?account_id=acc_1&alias=alias@icloud.com&limit=20&days=7
GET /api/mailboxes?account_id=acc_1
```

更完整的接口说明见 [API.md](API.md)。

## 数据文件格式示例

程序实际使用的是：

```text
data/accounts.json
```

示例：

```json
{
  "accounts": {
    "acc_1": {
      "id": "acc_1",
      "name": "主账号",
      "host": "icloud.com",
      "cookies": {
        "X-APPLE-WEBAUTH-TOKEN": "你的 Cookie 值",
        "X-APPLE-WEBAUTH-USER": "你的 Cookie 值",
        "X-APPLE-DS-WEB-SESSION-TOKEN": "你的 Cookie 值"
      },
      "icloud_email": "your_email@icloud.com",
      "app_password": "xxxx-xxxx-xxxx-xxxx",
      "status": "active"
    }
  }
}
```

Release 包里只会包含 `accounts.example.json` 示例文件，不会包含上面的真实配置文件。


## 常见问题

### 1. 打不开网页怎么办？

先确认程序是否正在运行，然后访问：

```text
http://127.0.0.1:8081
```

如果端口被占用，可以换端口：

```powershell
.\icloud-prime.exe -addr :9000 -data .\data
```

然后访问：

```text
http://127.0.0.1:9000
```

### 2. 创建别名返回 401 或 403 怎么办？

通常是 Cookie 过期。重新登录 iCloud，重新复制 Cookie，然后在管理台更新账号 Cookie。

### 3. 邮件读取失败怎么办？

优先检查 App 专用密码是否正确。IMAP 需要可访问：

```text
imap.mail.me.com:993
```

如果没有配置 App 专用密码，程序会尝试使用 Cookie 作为 Web API 回退路径。


## 许可证

MIT License
