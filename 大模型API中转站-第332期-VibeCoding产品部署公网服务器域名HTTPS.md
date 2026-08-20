---
title: "Vibe Coding 产品部署公网：服务器、域名与 HTTPS"
tags:
  - Vibe Coding
  - 网站部署
  - 域名与 HTTPS
description: "从本地项目出发，完成服务器准备、域名解析、静态文件部署和 HTTPS 配置，并说明动态应用需要额外处理的部分。"
---
# Vibe Coding 产品部署公网：服务器、域名与 HTTPS

用 Vibe Coding 做出一个页面，和让别人稳定访问这个页面，是两件不同的事。开发工具里能打开 `localhost:3000`，只说明程序在自己的电脑上运行；它没有说明公网用户能找到它、服务器能长期运行，或者传输过程已经加密。

本文只解决一个问题：怎样把一个已经在本地跑通的 Vibe Coding 产品发布到公网。示例采用一台 Ubuntu Linux 云服务器、一个域名、Nginx 和 Let's Encrypt。读者最终应能得到一个可以通过 `https://example.com` 访问的站点。若项目还包含登录、数据库或文件上传，文中也会指出需要增加的服务边界。

## 一、先判断你的产品属于哪一类

部署前先打开项目根目录的 `package.json`、`README` 和环境变量文件。不要把“Vite 项目”直接等同于“静态网站”，判断依据是构建后是否仍然需要一个长期运行的进程。

| 项目特征 | 常见部署方式 | 需要的运行组件 |
| --- | --- | --- |
| 构建后只有 `dist`、`build` 等文件 | 上传静态文件 | Nginx 或其他静态文件服务器 |
| 前端需要 SSR、接口或服务端渲染 | 运行 Node 服务，再由 Nginx 转发 | Node、进程管理、Nginx |
| 有登录、文章发布、订单或数据保存 | 前后端和数据服务分别配置 | 应用服务、数据库、对象存储、备份 |

先在本地做一次构建：

```powershell
npm ci
npm run build
```

如果命令成功，查看构建目录：

```powershell
Get-ChildItem .\dist
```

项目使用的目录可能叫 `build` 或其他名字，以构建命令的实际输出为准。若 `npm run build` 失败，先修复本地构建；服务器不会自动解决依赖版本或代码错误。

## 二、准备服务器和域名

服务器可以理解为一台持续联网的远程电脑。个人展示页、作品集和低流量博客通常先从一台小规格实例开始即可，但规格不是部署成功的保证。动态应用还要考虑内存、数据库、存储和备份。

服务器准备完成后，记录三个信息：公网 IPv4 地址、SSH 登录用户名、系统版本。不要把 SSH 密码、私钥或云平台密钥写进代码仓库。

首次登录服务器：

```bash
ssh deploy@example-server-ip
```

确认系统和网络：

```bash
cat /etc/os-release
ip -br address
```

建议使用普通部署用户，不要长期使用 `root` 运行应用。下面的命令假设该用户可以通过 `sudo` 管理服务。

开放必要端口。以 Ubuntu 的 UFW 为例：

```bash
sudo apt update
sudo apt install -y nginx
sudo ufw allow OpenSSH
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
sudo systemctl enable --now nginx
```

云平台控制台通常还有一层安全组。服务器防火墙已经放行，不代表安全组也放行；两层都需要允许 TCP `80` 和 `443`，SSH 端口则只开放给你实际使用的端口。

域名购买平台可以和服务器提供商不同。购买完成后，先在 DNS 管理界面添加一条 `A` 记录：主机记录填 `@`，记录值填服务器公网 IPv4。若还需要 `www.example.com`，再添加 `www` 的 `A` 记录，或添加指向主域名的 `CNAME`。

DNS 记录不是立即在所有网络中同时更新。可以用下面的命令检查解析结果：

```powershell
Resolve-DnsName example.com -Type A
```

输出中应该出现你的服务器 IP。如果仍然是旧地址，先确认记录值、记录类型和 DNS 服务商是否填对，再等待缓存更新。不要在解析尚未生效时反复申请证书，证书验证会因为域名找不到服务器而失败。

如果服务器位于中国大陆，网站对外提供服务通常还要根据实际业务和接入平台要求办理 ICP 备案。备案要求、主体材料和审核时间会变化，应以工业和信息化部备案系统及云服务商当前页面为准：[工业和信息化部 ICP/IP 地址/域名信息备案管理系统](https://beian.miit.gov.cn/)。使用境外节点时，备案要求和网络可达性是另一组问题，不能简单套用大陆服务器的流程。

### 服务器、托管平台和静态托管怎么选

部署方式没有脱离场景的唯一答案，可以先按维护责任来选：

| 方式 | 适合什么项目 | 你需要自己负责什么 |
| --- | --- | --- |
| 静态托管 | 作品集、文档、纯前端博客 | 构建、域名、环境变量和回滚 |
| 云服务器 + Nginx | 想控制系统、端口和运行时的个人项目 | 补丁、防火墙、进程、日志和备份 |
| 托管式应用平台 | 不想维护 Linux，希望平台接管运行时 | 平台配置、构建失败、配额和数据迁移 |

静态托管的运维工作较少，但不适合需要长期运行后端进程的项目；云服务器自由度更高，同时也把安全更新和故障恢复交给了你。第一次上线不必为了“以后可能有很多访问量”提前搭建复杂集群，先选择能完成真实功能验证的最小方案更容易发现问题。

费用也应拆成几类看：域名续费、计算资源、流量、数据库、对象存储、备份和增值服务。证书可能有免费方案，但免费不等于自动配置、自动续期或无限制使用。价格和免费额度会随地区、套餐和时间变化，部署前应以服务商当前账单页面为准，不要把某篇文章里的单次价格当作长期预算。

## 三、部署静态 Vibe Coding 产品

### 1. 在服务器创建站点目录

```bash
sudo mkdir -p /var/www/vibe-site
sudo chown -R "$USER":"$USER" /var/www/vibe-site
```

### 2. 从本地上传构建产物

在本地项目根目录执行。下面假设构建输出是 `dist`，`deploy` 是服务器用户名，服务器地址和域名请替换成自己的值：

```powershell
scp -r .\dist\* deploy@example-server-ip:/var/www/vibe-site/
```

上传后回到服务器检查：

```bash
find /var/www/vibe-site -maxdepth 2 -type f | head
```

### 3. 配置 Nginx

创建站点配置：

```bash
sudo nano /etc/nginx/sites-available/vibe-site
```

填入最小配置：

```nginx
server {
    listen 80;
    listen [::]:80;
    server_name example.com www.example.com;

    root /var/www/vibe-site;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

`try_files` 最后一项很重要。React、Vue、Svelte 等单页应用经常由前端路由处理 `/about` 或 `/settings`。没有回退到 `index.html` 时，首页可能正常，直接刷新子路径却会返回 `404`。

启用配置并验证语法：

```bash
sudo ln -s /etc/nginx/sites-available/vibe-site /etc/nginx/sites-enabled/vibe-site
sudo nginx -t
sudo systemctl reload nginx
```

预期结果是 `syntax is ok` 和 `test is successful`。如果提示端口占用、配置重复或权限错误，先看 `sudo nginx -t` 的文件路径和行号，不要直接重启整台服务器。

先用 HTTP 验证：

```bash
curl -I http://example.com
```

如果返回 `200`，说明 DNS、端口和静态文件至少已经连通。若返回 `403`，检查目录权限和 `index.html` 是否真的上传；若返回默认 Nginx 页面，检查 `server_name` 和启用的配置文件。

## 四、给网站配置 HTTPS

HTTPS 需要域名证书。Let's Encrypt 提供自动化证书签发服务，Certbot 可以协助完成申请和 Nginx 配置。官方文档说明了证书签发、自动续期和不同验证方式，实际命令以当前系统和文档为准：[Certbot 用户指南](https://eff-certbot.readthedocs.io/en/stable/using.html)。

在 Ubuntu 上安装 Certbot 的方式可能随系统仓库变化。安装后执行：

```bash
sudo apt install -y certbot python3-certbot-nginx
sudo certbot --nginx -d example.com -d www.example.com
```

根据提示填写邮箱、同意协议，并选择是否将 HTTP 重定向到 HTTPS。成功后检查：

```bash
curl -I https://example.com
sudo certbot renew --dry-run
```

第一条应返回 HTTPS 响应，第二条用于验证自动续期流程是否可执行。证书不是买完永久有效，续期失败时要看 Certbot 输出和定时任务日志。

Nginx 的完整配置和指令可以参考官方文档：[Nginx Documentation](https://nginx.org/en/docs/)。不要把自签名证书直接用于面向公众的正式站点，否则浏览器会显示不受信任警告。

## 五、动态应用需要多做什么

如果项目包含后端，不能只上传前端的 `dist`。先在服务器安装项目要求的 Node.js 版本，然后在应用目录安装依赖并配置生产环境变量：

```bash
cd /srv/vibe-app
npm ci --omit=dev
npm run start
```

`npm run start` 只是示例，必须以项目 `package.json` 的脚本为准。开发服务器常常会绑定到 `0.0.0.0` 并启用热更新，不应直接作为公网生产服务。更合理的结构是：应用只监听 `127.0.0.1:3000`，Nginx 对外监听 `443`，再将请求转发给应用。

Nginx 的动态接口配置可以是：

```nginx
location /api/ {
    proxy_pass http://127.0.0.1:3000;
    proxy_http_version 1.1;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

为了让应用重启后自动恢复，可以使用 systemd、Docker Compose 或其他进程管理方式。无论选择哪一种，都要先确认四件事：应用监听地址、生产环境变量、日志位置、停止和回滚方式。数据库和对象存储也要独立设置权限，不能把管理员密码放到前端代码里。

### 一个最小的 systemd 配置

如果项目是 Node 服务，可以用 systemd 先跑通最小闭环。下面的服务名、用户、目录和启动命令都是示例，必须改成项目真实值：

```ini
[Unit]
Description=Vibe Coding application
After=network.target

[Service]
Type=simple
User=deploy
WorkingDirectory=/srv/vibe-app
EnvironmentFile=/etc/vibe-app.env
ExecStart=/usr/bin/npm run start
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

保存为 `/etc/systemd/system/vibe-app.service` 后执行：

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now vibe-app
sudo systemctl status vibe-app --no-pager
```

如果状态是 `failed`，先查看日志：

```bash
sudo journalctl -u vibe-app -n 100 --no-pager
```

`ExecStart` 的路径不能想当然。可以先用 `which node` 和 `which npm` 检查当前系统路径；如果 Node 是通过版本管理器安装的，systemd 看到的环境可能和你的 SSH 会话不同。应用还要明确监听地址和端口，推荐只监听 `127.0.0.1`，让 Nginx 负责公网入口。

### 发布、验证和回滚

不要把每次发布都直接覆盖唯一的线上目录。可以把构建产物保存为带版本号的目录，再让 Nginx 指向一个稳定的 `current` 路径：

```bash
sudo mkdir -p /srv/releases/2026-07-29-001
sudo cp -r dist/. /srv/releases/2026-07-29-001/
sudo ln -sfn /srv/releases/2026-07-29-001 /srv/current
sudo nginx -t
sudo systemctl reload nginx
```

如果新版本出现白屏，可以把软链接切回上一个已验证目录，再重新加载 Nginx。数据库迁移不能只靠切回前端文件回滚，因此凡是涉及表结构或数据格式变化的发布，都要先做数据库备份，并确认迁移是否向后兼容。

## 六、上线验收清单

部署完成后，不要只打开一次首页就结束。至少检查：

- `Resolve-DnsName` 返回正确服务器 IP。
- `curl -I https://example.com` 返回预期状态码。
- 首页、一个前端子路径和刷新子路径都能打开。
- 浏览器开发者工具里没有混合内容、资源 `404` 或跨域错误。
- 登录、数据写入、图片上传等真实功能能完成一次闭环。
- `sudo systemctl status nginx` 正常，应用日志没有持续报错。
- `sudo certbot renew --dry-run` 能通过。
- 服务器、数据库和上传文件至少有一种可恢复的备份。

建议把验收结果写入项目的 `DEPLOYMENT.md` 或内部文档，记录发布时间、代码提交、服务器地址、域名、运行命令、环境变量名称、备份位置和回滚命令。这里不要记录密钥值。几个月后再次部署时，这份记录往往比记忆和聊天记录更可靠。

## 结论

部署的核心不是记住某个平台的按钮位置，而是建立一条清晰链路：本地构建产物上传到服务器，域名解析到服务器，Nginx 负责接收请求，再通过证书启用 HTTPS。静态站点可以先完成这条最小链路；动态应用还需要进程管理、环境变量、数据库、文件存储和备份。

这套方法适合个人作品、博客和低流量产品的第一次上线。随着访问量、数据敏感性或可用性要求增加，应进一步验证监控、自动部署、访问控制、备份恢复和扩容方案，而不是仅靠增加服务器规格解决所有问题。

## 参考资料

- [Nginx 官方文档](https://nginx.org/en/docs/)，用于核对站点配置、反向代理和重载命令。
- [Certbot 官方文档](https://eff-certbot.readthedocs.io/en/stable/using.html)，用于核对证书申请和续期流程。
- [Let's Encrypt 文档](https://letsencrypt.org/docs/)，用于理解证书签发和验证机制。
- [工业和信息化部备案系统](https://beian.miit.gov.cn/)，用于核对中国大陆网站备案要求。
