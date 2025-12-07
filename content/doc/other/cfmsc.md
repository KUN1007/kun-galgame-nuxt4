---
title: '关于鲲 Galgame 的 CFMSC 发布机制'
banner: '/content/other/cfmsc/banner.avif'
description: 'CFMSC（Check -> Fetch -> Modify -> Submit -> Collaborate）是鲲 Galgame 论坛开发的一种高效资源发布和编辑流程，旨在优化 Galgame 资源管理。该机制借鉴了 GitHub 的 Pull Request 流程和恋爱游戏网的资源贡献模式，使用户能够自主创建、修改和提交游戏资源，同时避免重复内容，提高准确性和可扩展性。CFMSC 的核心步骤包括：检查（确保资源的正确性与唯一性）、获取（利用 VNDB API 自动填充信息）、修改（用户优化描述与添加预览图）、提交（创建 Galgame 资源基底）以及协作（多人审核与合并更新请求）。这一流程既降低了管理成本，又提升了用户参与度，并支持未来的去中心化发展愿景，推动构建更开放、透明、用户驱动的 Galgame 社区。'
publishedTime: 2024-12-08
modifiedTime: 2024-12-08
category: other
authorUid: 2
authorName: '鲲'
authorAvatar: 'https://image.kungal.com/avatar/user_2/avatar.webp'
authorHomepage: 'https://github.com/KUN1007'
pin: false
---

## 什么是 CFMSC

CFMSC 是鲲 Galgame 发布以及编辑资源流程 (Check -> Fetch -> Modify -> Submit -> Collaborate) 过程的首字母缩写

## CFMSC 的由来

CFMSC 的灵感来自于曾经的 恋爱游戏网, 以及现在 GitHub 的 `Pull Request` 机制

我们从恋爱游戏网 `一个资源下可以由若干用户贡献若干关于该资源的资源链接`, 以及 GitHub 的 `Close Pull Request, Merge Pull Request, Diff Pull Request Content` 中得到了巨大的灵感

CFMSC 首次被实现是在 [鲲 Galgame 论坛](https://www.kungal.com/) 的发布 Galgame 功能上

## 为什么会有 CFMSC

我们一直在寻找一种, 可以由用户参与, 并且降低管理难度, 同时保证安全性, 还要保持最低开支, 以及高可用性和高增长率的资源分享, 资源编辑方式

要实现这六点, 需要利用用户的主动性来维持社区的活力, 我们认为, 一个好的社区永远是用户为核心动力的, 而不是管理层

## CFMSC 实现流程

CFMSC 是一个衔接紧密, 以用户为核心的交互流程, 可以具体分为五大步骤

从下面开始, 资源实体（可以是任何类型的资源），我们在此处仅以 Galgame 作为举例

### Check - 检查

首先, 如果是针对某一个类型的资源实体（例如 Galgame）, 如果要创建一个 Galgame 实体, 对创建这个行为来讲, 以下两个点是需要注意的

1. 用户创建的资源实体是不是 Galgame

什么是 Galgame, 哪里有参考, 有没有规范的定义, 这是我们首先要解决的问题。如果用户发布的连 Galgame 都算不上, 那么这将会对其它用户的体验造成坏的影响, 同时也不利于社区的管理

比如, 没有人会喜欢在 Galgame 网站看见某一款手游的攻略资源

1. 用户创建的 Galgame 实体是否重复

查重这件事在任何行业都会成为一个问题, 没有人希望搜索结果中有着十几个同样的结果, 并且有的 Galgame 有效而有的 Galgame 无效, 这会令人感到烦躁与困惑

如何实现 Galgame 的查重是一个问题, 在这里我们引入了 VNDB 作为辅助

**为什么要使用 VNDB**

首先对于 Galgame, 或者是 Visual Novel（视觉小说） 来讲, VNDB 是最大最全的网站（没有之一）

再者, VNDB 对 [Galgame 是什么](https://www.kungal.com/topic/1720) 这件事, 做出了合理且详实的定义

最后, VNDB 是开放的, 它有简单易懂的 API 文档, 甚至无任何门槛的提供了 API 接口, 甚至还有全部的数据库 dump! 这是模范! 榜样!

**实现 Check 流程**

对于本网站来说, 实现 Check 流程就是用户将 VNDB_ID（可以理解为 Galgame 的唯一编号, 绝对不会重复）输入, 然后我们做以下两步工作

1. 向我们的后端接口发起请求, 判断用户输入的 VNDB_ID 是否已经有人发过对应的游戏了 (Check Duplicate)
2. 若该游戏未重复, 则向 VNDB 的开放 API 接口发起请求, 请求部分我们需要用到的字段 (Fetching Resources)

这样做之后, 有两个好处, 一是保证了用户发布的 Galgame 绝不可能重复, 二是减轻了用户编写介绍的负担, 因为例如游戏别名, 游戏原名这些元信息是不需要手动查询的, 这会造成大量的心智负担

实现 Check 流程, 只需要用户前往 [VNDB 官网](https://vndb.org/) 查询要发布 Galgame 的 VNDB_ID 即可

### Fetch - 获取

其实上面的流程中, `若该游戏未重复, 则向 VNDB 的开放 API 接口发起请求, 请求部分我们需要用到的字段 (Fetching Resources)`, 这一步就是 Fetch 流程就是用户将

Fetch 的作用有以下三点

1. 减轻用户心智负担

比起手动搜索游戏, 查询游戏的别名, 介绍, 原名等等, 简单输入一个 VNDB_ID 会方便得多

1. 保证准确性

VNDB 的数据是可靠且准确的（大部分而言）, 这将会降低很多新手的压力, 在繁杂的互联网环境下, 填出错误的数据也是有可能的（某些网站甚至写出了错误的游戏原名）

1. 优秀的可扩展性

VNDB 是全名是 (the visual novel database), 这意味着, 关于视觉小说（Galgame） 的绝大部分, 能用到的属性, 都可以在 VNDB 检索到, 如果以后我们的系统扩展, 为 Galgame 增加新的介绍将会是一件易如反掌的事

### Modify - 修改

VNDB 获取到的数据, 或者说其它任何网站获取到的一个数据, 都有一个缺陷性

那就是 —— 这些数据不是以用户为主导的

绝大部分的介绍, 都是摘自游戏的官网, 直接翻译得来, 这带来一些问题, 例如

1. 基本上所有的网站提供的介绍的数据, 都是没有预览图的, 然而大部分情况下用户是想要看到几张游戏预览图的
2. 倘若我要为《永不枯萎的世界与终结之花》这个游戏, 添加一个 `枯花` 的别名, 应该如何处理

对于第一点, `鲲 Galgame 补丁` 与 `鲲 Galgame 论坛` 的处理方式是不同的

鲲 Galgame 补丁 采用用户直接上传图片到游戏介绍的方式, 简单而直接, 因为该网站的作用是下载补丁, 它只会做好自己该做的事, 介绍只是顺带的副产品

鲲 Galgame 论坛 对于该方面的处理还未落实, 它的计划是, 单独创建一个 `galgame_preview` 字段甚至是表, 用它来存放单个 Galgame 的预览图, 这带来了下面的好处

1. 极强的可扩展性, 可以将任何图片, 乃知视频, 甚至是 DOM 元素插入 Galgame 的介绍, 实现完美而又细节的介绍
2. 极强的自定义性, 如果有用户不想看游戏预览图, 因为这很占用页面空间, 那么可以直接关闭预览图
3. 极强的可编程性, 如果要实现一个功能, 对该游戏的预览图进行分类, 全年龄的预览图标记为普通, R18 的预览图标记为 NSFW, 这可以很轻易的实现

**实现 Modify 流程**

在用户获取到数据之后, 可以自己修改获取到的数据, 这就是 Modify 的核心

### Submit - 提交

用户在编辑完成自己的 Galgame 之后, 可以提交这个 Galgame, 这会在我们的数据库中新建一个 Galgame

提交的核心在于, Galgame 只是一个基 (primitive), 之后对于 Galgame 这一个基的联系与扩展, 才是我们真正需要注意的地方

一个 Galgame 下面可以有哪些资源呢, 我们可以这么思考, 那么很高兴知道

1. Galgame 资源

站在绝大部分用户的角度, 这是他们最关心的问题, 包括但不限于游戏本体, 补丁, 合集, 音声相关, 图片相关, AI 相关, 动漫相关, 以及其他

1. 相关链接

任何与该游戏有关的链接, 例如网络文章, 视频链接, 媒体链接等等。这可以帮助任何人, 快速理解游戏的影响力, 以及其他地方对于该游戏的看法等等

1. 游戏评论

关于这个 Galgame 以及这个 Galgame 下面的各种扩展属性, 会有人对这些属性有自己的评价或者是疑问, 那么一个评论区会是很有必要的选择

等等 ... 这些属性您都可以在 `鲲 Galgame 补丁` 以及 `鲲 Galgame 论坛` 中看到

**实现 Submit 流程**

Submit 的核心在于创建一个 Galgame 的基, 之后可以利用这个基做出无限的扩展, 它具有无与伦比的可扩展性

### Collaborate - 合作

CFMSC 的核心是用户, 那么用户合作编辑一个 Galgame 将会是一个理所应当的需求, 这也是维护一个良好的社区所必备的技术要求

对于这一点, 我们借鉴了 GitHub 的 `Pull Request` 流程, 但是 `Pull Request` 这个名字经常会给人带来困扰, 例如, 您可以很轻易的搜索到

[Why is a git 'pull request' not called a 'push request'?](https://stackoverflow.com/questions/21657430/why-is-a-git-pull-request-not-called-a-push-request)

[Why are PRs called "Pull Requests" and not "Merge Requests"?](https://www.quora.com/Why-are-PRs-called-Pull-Requests-and-not-Merge-Requests)

于是, 我们统一将 `Pull Request` 更改为了 `更新请求`, 以降低用户的心智负担

Collaborate 其实可以分为三大步骤, 发起（Request）, 审核（Review）, 实行（Execute）

#### 发起 - Request

一个用户观察到某个 Galgame 中有想要更改的地方, 例如, 他想更改这个游戏的介绍, 想为这个游戏添加一个新的标签等

这时, 他可以编辑这个游戏, 但是编辑完成之后, 他的编辑更改并不会立即生效, 而是会被暂存在 `更新请求` 中, 等待其他人的审核

这个更新请求需要经过游戏发布者 (owner), 或者是网站管理员 (moderator) 的同意 (approve) 后, 才可以被合并 (merge) 到这个游戏中

更新请求被合并后, 提出更新请求的这名用户也会被加入到这个游戏的贡献者 (contributor) 中

#### 审核 - Review

任何人都可以看到这名用户提交的更新请求, 但是只有游戏发布者, 或者是管理员, 有同意 (approve) 和拒绝 (decline) 这个更新请求的权限

在查看更新请求时, 更新请求的更改会被 diff 算法标注, 这一点上 `鲲 Galgame 补丁` 与 `鲲 Galgame 论坛` 的实现是不一样的

对于鲲 Galgame 补丁, 新增的字符会被标记为绿色, 删减的字符会被标记为红色, 合并后文本不会变化, 拒绝后更新请求将保持原样, 这个过程使用了现成的 package [jsdiff](https://github.com/kpdecker/jsdiff)

对于鲲 Galgame 论坛, 新增的字符会被标记为蓝色, 删减的字符会被编辑为红色, 合并后文本会变为绿色, 拒绝后更新请求将不可被查看这个, 这个过程使用了鲲 Galgame 论坛 [自行编写的 diff 算法](https://github.com/KUN1007/kun-galgame-nuxt4)

#### 实行 - Execute

若游戏发布者或管理员觉得这个更新请求合理, 则可以同意这个更新请求, 若不合理则可以拒绝, 拒绝时必须提供拒绝原因以便提出更新请求的用户进行整改

实行之后, 改更新请求会被完成 (complete) 或者是废弃 (obsolete)

同时, 提出更新请求的用户会收到自己更新请求的处理结果

上面的流程, Request -> Review -> Execute 均会被记录在 Galgame 下面的贡献历史中, 以便之后的用户查看变更历史

## 之后的发展

目前已经有两个网站实现了 CFMSC 流程

[鲲 Galgame 论坛](https://www.kungal.com/) - Nuxt4 实现

[鲲 Galgame 补丁](https://www.moyu.moe/) - Next.js 15 实现

CFMSC 流程目前尚在尚在发展中, 但是我们觉得, 它是先进的, 科学的, 用户为中心的, 合作式的, 有活力的, 高增长率的

站在管理的角度来讲, 这个过程还缺少了一个重要的因素, 就是 `激励 (incentive)`

我们认为以上流程还满足不了一个活跃的社区对用户应有的激励手段, 我们之后会将这个流程扩展完善, 甚至演变为一个新流程的一部分, 以便建立一个现代, 开放, 用户为驱动的现代网站

## 去中心化

去中心化 (Decentralized) 在管理中指的是将管理决策权下放到下层部门, 甚至底层操作者身上的方法

[Centralized and Decentralized Management Explained](https://www.personalfinancelab.com/finance-knowledge/management/centralized-and-decentralized-management-explained/)

> Centralized management is the organizational structure where a small handful of individuals make most of the decisions in a company. For example, a small family diner owned by a married couple probably uses centralized management. The couple themselves order inventory, decide the marketing direction, and hire new employees. As a company with centralized management grows, they add new levels of mid and lower level managers, each of whom answers to a superior, with very strictly defined roles in the company.

> Decentralized management is the opposite – the upper levels of management transfers some of the decision-making processes onto lower levels, and even to individual employees. The overall authority is still maintained by top level managers, who make policies that influence the major decisions of the company, but most decision-making responsibility is delegated to the lower levels.

我们有做一个去中心化社区 (Decentralized Community) 的想法, 这就像 Mastodon, Misskey 等网站一样, 随处都可以建设独立的社区, 建立一个分布式的社交媒体网络

值得一提的是, 我们的 CFMSC 就是管理权限下放的体现, 我们想要的是用户之间的思考碰撞, 想要看到的是真正有人在思考, 想要看到人与人之间的分工协作, 友好交流

希望鲲 Galgame 可以一直为您带来良好的体验!
