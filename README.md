# 微信API代理服务器

一个轻量级的微信API HTTP/HTTPS透明代理服务器，专为解决动态IP环境下的白名单问题而设计。

## 📋 适用场景

本项目适用于以下场景：
1. 本地开发环境使用动态IP，但需要调用微信API（如公众号、小程序等）
2. 微信API要求IP白名单，但本地IP经常变动
3. 需要将微信API请求通过固定IP的服务器转发

> **重要提示**：使用本项目前，请确保：
> - 已有一台具有固定公网IP的云服务器（如阿里云、腾讯云等）
> - 该服务器IP已添加到微信API的IP白名单中
> - 本地环境能够访问该云服务器

## 🎯 功能特性

- ✅ **轻量级设计** - 纯Go标准库实现，无外部依赖
- ✅ **透明HTTPS代理** - 不解析HTTPS内容，保持端到端加密
- ✅ **域名白名单控制** - 仅代理指定的微信相关域名
- ✅ **资源占用极低** - 内存占用约5MB，CPU使用率接近0%
- ✅ **端口灵活** - 可使用任意闲置端口，不影响其他服务

## 🚀 快速开始

### 1. 准备云服务器

1. 准备一台具有固定公网IP的云服务器（如阿里云、腾讯云等）
2. 将服务器IP添加到微信API的IP白名单中：
   - 登录微信公众平台
   - 进入"设置与开发" -> "基本配置"
   - 在"IP白名单"中添加服务器IP

### 2. 部署代理服务

1. 在本地下载所需文件：
   ```bash
   # 下载二进制文件（以 Linux x86_64 为例）
   wget https://github.com/wootile/wechat-proxy-go/releases/latest/download/wechat-proxy-linux-amd64 -O wechat-proxy
   
   # 注意：部署脚本会在同目录查找 wechat-proxy 文件，请确保文件名正确
   
   # 下载部署脚本
   wget https://raw.githubusercontent.com/wootile/wechat-proxy-go/main/deploy-offline.sh
   ```

2. 将文件上传到云服务器：
   ```bash
   scp wechat-proxy deploy-offline.sh user@your-server-ip:~/
   ```

3. 登录云服务器并部署：
   ```bash
   ssh user@your-server-ip
   chmod +x ~/wechat-proxy
   chmod +x ~/deploy-offline.sh
   ```

4. 使用 systemd 服务（推荐）或直接运行：
   ```bash
   # 使用 systemd 服务（推荐）
   sudo ./deploy-offline.sh

   # 或直接运行
   ./wechat-proxy
   ```

## 📝 使用示例

### Python

```python
import requests

# 仅对微信API请求使用代理
proxies = {
    'http': 'http://your-server-ip:8080',
    'https': 'http://your-server-ip:8080'
}

# 获取微信 access_token
response = requests.get(
    'https://api.weixin.qq.com/cgi-bin/token',
    params={
        'grant_type': 'client_credential',
        'appid': 'YOUR_APPID',
        'secret': 'YOUR_SECRET'
    },
    proxies=proxies,  # 仅对微信API请求使用代理
    verify=True
)

# 其他API请求不使用代理
other_response = requests.get('https://other-api.com')
```

### Node.js

```javascript
const axios = require('axios');
const HttpsProxyAgent = require('https-proxy-agent');

// 注意：必须使用 https-proxy-agent 来处理 HTTPS 请求
const agent = new HttpsProxyAgent('http://your-server-ip:8080');

// 获取微信 access_token
const response = await axios.get('https://api.weixin.qq.com/cgi-bin/token', {
    params: {
        grant_type: 'client_credential',
        appid: 'YOUR_APPID',
        secret: 'YOUR_SECRET'
    },
    httpsAgent: agent  // 仅对微信API请求使用代理
});

// 其他API请求不使用代理
const otherResponse = await axios.get('https://other-api.com');
```

> **重要提示**：
> 1. 本项目仅用于代理微信API请求，不建议配置全局代理
> 2. 在 Node.js 中使用时，必须安装并使用 `https-proxy-agent` 包来处理 HTTPS 请求：
>    ```bash
>    npm install https-proxy-agent
>    ```

## ⚙️ 配置说明

### 环境变量

| 环境变量 | 默认值 | 说明 |
|---------|--------|------|
| `PROXY_PORT` | `8080` | 代理服务器监听端口，可使用任意闲置端口 |

### 自定义端口

```bash
# 可以使用任意未被占用的端口
PROXY_PORT=9090 ./wechat-proxy
```

## 🌐 支持的域名

- `api.weixin.qq.com` - 微信公众号API
- `api.wechat.com` - 微信开放平台API
- `mp.weixin.qq.com` - 微信公众号管理后台
- `qyapi.weixin.qq.com` - 企业微信API

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。