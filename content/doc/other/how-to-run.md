---
title: '如何本地运行本论坛的开源项目'
banner: '/content/other/how-to-run/banner.avif'
description: '本文详细介绍了如何在本地搭建并运行鲲 Galgame 论坛的开源项目。文章提供了完整的安装和配置指南，包括 Node.js、MongoDB 和 Redis 的安装步骤，以及如何 Clone 代码仓库并配置 .env 文件。用户需安装依赖后，使用 pnpm dev 启动项目，并确保 Redis 和 MongoDB 连接成功。访问 http://127.0.0.1:1007/ 即可看到运行的论坛。由于默认未配置邮件服务，注册用户需手动查看验证码。该文档适合有基础开发经验的用户，帮助他们快速上手并运行该开源项目。'
publishedTime: 2024-10-17
modifiedTime: 2024-10-17
category: other
authorUid: 2
authorName: '鲲'
authorAvatar: 'https://image.kungal.com/avatar/user_2/avatar.webp'
authorHomepage: 'https://github.com/KUN1007'
pin: false
---

## 视频教程

<video src="https://cdn.jsdelivr.net/gh/kun-moe/kun-image@main/blog/202410171807531.mp4" playsinline="" loop="" controls=""></video>

YouTube: https://youtu.be/IlBDhMiDYp0

## 项目地址

本论坛开源项目地址为

https://github.com/KUN1007/kun-galgame-nuxt4

开发文档地址为

https://soft.moe/kun-visualnovel-docs/v2/introduction/run.html

## 安装 Node.js 环境

建议使用 `nvm` 安装 Node.js

如果 Node.js 安装成功, 可以使用

```shell
node -v
```

查看是否安装成功

## 安装 MongoDB 和 Redis

在您安装成功后，可以使用

```shell
mongod --version
```

查看 MongoDB 是否安装成功

使用

```shell
redis-server --version
```

查看 Redis 是否安装成功

## Clone 我们的项目

找一个文件夹，使用

```shell
git clone https://github.com/KUN1007/kun-galgame-nuxt4
```

来 clone 我们的项目

然后使用 VS Code 或者其它开发工具打开项目

## 配置 .env 文件

在项目的根目录新建一个 `.env` 文件, 将 `.env.example` 的内容复制进入 `.env` 文件中

该文件的详细介绍为

```yaml
# 本地 mongodb 的连接
MONGODB_URL = 'mongodb://127.0.0.1:27017/kungalgame'

# 网站连接，保持即可
KUN_GALGAME_URL = 'https://www.kungal.com'

# API 请求地址，保持即可
KUN_GALGAME_API = 'http://www.kungal.com/api'

# Redis 的 host 和 port 配置
REDIS_HOST = '127.0.0.1'
REDIS_PORT = '6379'

# 发送邮件的配置，不设置也可以运行项目，但是无法发送验证码。这个配置项需要使用 smtp 服务，需要自己寻找配置
KUN_VISUAL_NOVEL_EMAIL_FROM = "KUN VISUAL NOVEL"
KUN_VISUAL_NOVEL_EMAIL_HOST = "KUN VISUAL NOVEL MOE EMAIL HOST"
KUN_VISUAL_NOVEL_EMAIL_PORT = "KUN VISUAL NOVEL MOE EMAIL PORT(generally could be 587/465/25)"
KUN_VISUAL_NOVEL_EMAIL_ACCOUNT = "KUN VISUAL NOVEL MOE EMAIL ACCOUNT"
KUN_VISUAL_NOVEL_EMAIL_PASSWORD = "KUN VISUAL NOVEL MOE EMAIL PASSWORD"

# 图床的配置，不设置也可以运行项目，但是无法上传图片。这个配置需要使用图床的 API，需要自己寻找或开发
KUN_VISUAL_NOVEL_IMAGE_BED_ACCESS_KEY = "kkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkk"
KUN_VISUAL_NOVEL_IMAGE_BED_SECRET_KEY = "kkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkk"
KUN_VISUAL_NOVEL_IMAGE_BED_ENDPOINT = "https://image.kungal.com/kun"
KUN_VISUAL_NOVEL_IMAGE_BED_URL = "https://image.kungal.com"

# JWT 的相关配置，保持这个配置也可以，详细描述可以去 JWT 的官网查看
JWT_ISS = 'KUN VISUAL NOVEL ISS'
JWT_AUD = 'KUN VISUAL NOVEL KUNGALGAMER'
JWT_SECRET = 'KUN VISUAL NOVEL SECRET'
```

## 安装依赖

使用

```shell
pnpm i
```

安装依赖，安装完成后项目目录会多一个 `node_modules` 文件夹

## 启动项目

使用

```shell
pnpm dev
```

启动我们的项目，预期的输出应该为

```shell
> kun-galgame-nuxt4@2.19.2 dev E:\MySite\kun-galgame-nuxt4
> nuxt dev

Nuxt 3.13.2 with Nitro 2.9.7

  ➜ Local:    http://127.0.0.1:1007/
  ➜ Network:  use --host to expose

ℹ Nuxt Icon server bundle mode is set to local
✔ Nuxt Icon discovered local-installed 11 collections
✔ Vite client built in 108ms
✔ Vite server built in 2014ms
✔ Nuxt Nitro server built in 3180 ms
ℹ Vite client warmed up in 1ms
ℹ Vite server warmed up in 2343ms
redis: 127.0.0.1:6379 connection successful!
MongoDB: mongodb://127.0.0.1:27017/kungalgame connection successful!
```

必须要看到 `redis` 和 `MongoDB` 的成功连接才算成功

## 打开网页

浏览器访问 http://127.0.0.1:1007/ 即可看到我们的项目

此时的一个空项目，什么都没有

## 注册用户

我们需要注册用户以使用论坛的功能

我们前往 http://127.0.0.1:1007/register 注册用户

由于我们没有配置邮箱服务，所以需要自主输出邮箱

更改 `/server/utils/sendVerificationCodeEmail.ts` 文件为

```typescript
export const sendVerificationCodeEmail = async (
  event: H3Event,
  email: string,
  type: 'register' | 'forgot' | 'reset'
) => {
  const ip = getRemoteIp(event)

  const limitEmail = await useStorage('redis').getItem(`limit:email:${email}`)
  const limitIP = await useStorage('redis').getItem(`limit:ip:${ip}`)
  if (limitEmail || limitIP) {
    return 10301
  }

  const code = generateRandomCode(7)
  console.log(code)

  await useStorage('redis').setItem(email, code, { ttl: 10 * 60 })
  await useStorage('redis').setItem(`limit:email:${email}`, code, { ttl: 30 })
  await useStorage('redis').setItem(`limit:ip:${ip}`, code, { ttl: 30 })

  const transporter = createTransport(
    SMPTransport({
      pool: {
        pool: true
      },
      host: env.KUN_VISUAL_NOVEL_EMAIL_HOST,
      port: Number(env.KUN_VISUAL_NOVEL_EMAIL_PORT) || 587,
      secure: false,
      requireTLS: true,
      auth: {
        user: env.KUN_VISUAL_NOVEL_EMAIL_ACCOUNT,
        pass: env.KUN_VISUAL_NOVEL_EMAIL_PASSWORD
      }
    })
  )

  const mailOptions = {
    from: `${env.KUN_VISUAL_NOVEL_EMAIL_FROM}<${env.KUN_VISUAL_NOVEL_EMAIL_ACCOUNT}>`,
    sender: env.KUN_VISUAL_NOVEL_EMAIL_ACCOUNT,
    to: email,
    subject: 'KUN Visual Novel Verification Code',
    text: getMailContent(type, code)
  }

  transporter.sendMail(mailOptions)
}
```

我们使用 `console.log(code)` 输出了邮箱验证码，这样本地注册的时候就可以看到验证码了

填写好所有的信息，即可注册成功用户

此时即可使用论坛的大部分功能

不能上传图片，因为图床地址没有配置
