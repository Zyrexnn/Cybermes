# HACK.SKILLS - Agent 的黑客武装

<p align="right"><a href="./README.md">English</a> | 中文</p>

<p align="center">
    <img src="./assets/readme-hero-banner.jpg" alt="HackSkills Hero Banner" width="100%" />
</p>

<p align="center">
    <strong>总入口 → 分类入口 → 深度专题技能</strong><br/>
    一个总入口，六个稳定分类入口，以及横跨 <strong>14 个安全领域</strong> 的 <strong>101</strong> 个按需下钻的深度专题技能。
</p>

这是一个面向 Agent 的安全技能知识库，覆盖 Web 安全、API 安全、认证与授权、操作系统提权（Linux/Windows/macOS）、Active Directory 攻击、移动安全、二进制漏洞利用（Pwn）、逆向工程、密码学攻击、区块链与智能合约安全、AI/ML 与 LLM 安全、网络协议与横向移动、数字取证——服务于漏洞赏金、渗透测试、CTF 竞赛和授权安全研究。

当前分支已经收敛到标准的目录式结构：每一个 skill 都位于单独目录下，统一采用 `skills/{semantic-identifier}/SKILL.md`。设计目标不是把所有零碎技巧都暴露成入口，而是把 loader 真正需要优先看到的入口压缩为一层总入口、六个分类入口，以及按需下钻的深度专题 skill。

目标很直接：把真正能在实战里派上用场、又方便审查和持续维护的安全知识，整理成一套可安装、可检索、可组合的 HackSkills。

## 在线浏览

本仓库以三种形式发布——按你的工作流挑一种就行，三处内容每次推送 `main` 都会同步刷新。

| 入口 | 你能拿到什么 | 适用场景 |
|---|---|---|
| **网页版** —— <https://skills.hackbenchmark.com> | 模糊搜索、分类侧栏、P0/P1/P2 等级筛选、一键复制安装命令、加密 ZIP 下载 | 日常查阅、分享某个 skill 链接、给同事做演示 |
| **GitHub 源码** —— 当前仓库 | 每个 skill 一个 `SKILL.md`，原生 markdown 渲染，PR review 链路 | diff 审阅、贡献代码、深度阅读 |
| **加密 ZIP 包** —— 见 [离线加密 ZIP](#离线加密-zip) | 一次性下载所有 `*.md`，离线/隔离网络可用 | 目标内网无外网、AV 杀掉明文 markdown |

网页版是 `site/` 目录的纯前端静态构建——无追踪、无后端。源码：[`site/`](./site)，部署 workflow：[`.github/workflows/deploy-pages.yml`](./.github/workflows/deploy-pages.yml)。搜索框基于加权模糊索引，覆盖 name / id / category / description，并支持 `category:auth`、`tier:deep`、`lines:>200` 这样的字段限定语法。

```text
                        ┌─────────────────────────────────────┐
                        │   skills.hackbenchmark.com (静态站) │  ── 搜索 / 筛选 / 复制安装命令
                        └─────────────────────────────────────┘
                                          ▲
   github.com/yaklang/hack-skills ───────►┤  同一仓库，三种视图
                                          ▼
                        ┌─────────────────────────────────────┐
                        │   hack-skills.zip (AES-256，公开    │  ── 离线 / 杀软误报场景
                        │   密码 hack-skills，CDN 分发)       │
                        └─────────────────────────────────────┘
```

## 知识来源与蒸馏边界

这个仓库不是外部资料的镜像仓库，而是一个面向 Agent 的蒸馏层。

主要参考公开知识来源（仅用于教育性蒸馏）：

| 来源 | 提供的内容 | 蒸馏方式 |
|---|---|---|
| `swisskyrepo/PayloadsAllTheThings` | 64 个漏洞类别、payload 家族、绕过技术、利用链方向 | 蒸馏为场景化索引、方法矩阵、按引擎/按数据库的 payload 分节 |
| `PentesterSpecialDict` | OS 特化 payload 字典、Java 中间件路径 fuzz 列表、文件扩展名数据库 | 蒸馏为参数命名模式、端点频率表、中间件指纹矩阵 |
| `Dictionary-Of-Pentesting` | BugBounty 绕过技巧（12 专题）、云元数据端点、XXE payload 集合、一行命令工具链 | 蒸馏为绕过模式矩阵、云元数据端点表、WAF 厂商绕过分节 |
| `Hello-CTF` | CTF Web 安全教程，含 PHP/Python/Java 实战技巧 | 蒸馏为 CTF 特定技术段落（handler 绕过、filter chain 技巧、Flask PIN 计算） |
| `ctf-wiki` | CTF 竞赛知识库，覆盖 Pwn、Crypto、逆向工程、取证和 Misc 方向 | 蒸馏为二进制利用技术（栈/堆/内核）、密码学攻击模式（RSA/格/对称）、逆向方法论、隐写术与流量分析技能 |
| `hacktricks` | 渗透测试百科全书，覆盖 Web 技巧、Linux/Windows/macOS 提权、Active Directory、容器、移动端、AI 安全 | 蒸馏为 OS 级提权手册、AD 攻击链（Kerberos/ACL/ADCS）、移动端测试清单、容器逃逸技术、网络穿透策略 |
| 公开安全研究论文与 CVE 公告 | 方法论框架、漏洞模式分类法、统计分布 | 蒸馏为攻击模式矩阵、系统化测试清单、决策树 |

处理原则：

- 不直接搬运超大字典和长 payload 全集。
- 优先提炼为可路由、可组合、可审查的安全 skill。
- 用小而稳定的样本、分类法和交叉引用来提升 Agent 在真实安全场景中的稳定性。
- 不包含任何特定客户信息，不包含可识别的厂商案例细节，纯教育性方法论。

## 快速开始

首选入口是 `hack`：

```bash
npx skills add yaklang/hack-skills
```

如果你的工具支持直接拉单个 SKILL.md，也可以使用：

- frontmatter name: `hack`
- raw URL: `https://raw.githubusercontent.com/yaklang/hack-skills/main/skills/hack/SKILL.md`

装完以后，推荐顺序很简单：先从总入口开始，再进入分类入口，最后才进入深度专题 skill。

## Loader 优先级

| 层级 | 作用 | 推荐暴露方式 | 代表 skill |
|---|---|---|---|
| 总入口 | 负责全局路由、测试顺序和跨类别切换 | 优先暴露 | [hack](./skills/hack/SKILL.md) |
| 分类入口 | 负责按攻击面分流到稳定的专题族 | 优先暴露 | [recon-for-sec](./skills/recon-for-sec/SKILL.md), [api-sec](./skills/api-sec/SKILL.md), [auth-sec](./skills/auth-sec/SKILL.md) |
| 深度专题 | 提供完整攻击手册和执行细节 | 按需加载 | [xss-cross-site-scripting](./skills/xss-cross-site-scripting/SKILL.md), [sqli-sql-injection](./skills/sqli-sql-injection/SKILL.md) |

## 主要入口

| 类型 | Skill | 用途 | 何时优先使用 |
|---|---|---|---|
| 总入口 | [hack](./skills/hack/SKILL.md) | 全局路由、阶段判断、跨类别切换 | 新目标、未知攻击面 |
| 分类入口 | [recon-for-sec](./skills/recon-for-sec/SKILL.md) | 资产发现、技术识别 | 刚接目标、信息不足 |
| 分类入口 | [api-sec](./skills/api-sec/SKILL.md) | REST、GraphQL、移动端后端路由 | 看到 API 接口 |
| 分类入口 | [auth-sec](./skills/auth-sec/SKILL.md) | 认证、会话、OAuth、JWT、授权 | 看到登录、令牌、对象 ID |
| 分类入口 | [injection-checking](./skills/injection-checking/SKILL.md) | XSS、SQLi、SSRF、XXE、SSTI、CMDi、NoSQL 路由 | 输入进入解释器 |
| 分类入口 | [file-access-vuln](./skills/file-access-vuln/SKILL.md) | 上传、下载、LFI、路径控制 | 文件操作 |
| 分类入口 | [business-logic-vuln](./skills/business-logic-vuln/SKILL.md) | 竞态、价格、流程、状态机 | 业务流程测试 |

## 完整技能索引（101 个技能）

### 侦察与方法论

| 技能 | SKILL.md | SCENARIOS.md | 核心内容 |
|---|---|---|---|
| [hack](./skills/hack/SKILL.md) | 161 行 | - | 总路由器、现象到技能的映射、专家直觉 |
| [recon-for-sec](./skills/recon-for-sec/SKILL.md) | 28 行 | - | 侦察阶段分类路由 |
| [recon-and-methodology](./skills/recon-and-methodology/SKILL.md) | 389 行 | - | 方法论框架、Java 中间件指纹矩阵、泄露检测清单 |

### API 安全

| 技能 | SKILL.md | SCENARIOS.md | 核心内容 |
|---|---|---|---|
| [api-sec](./skills/api-sec/SKILL.md) | 48 行 | - | API 测试分类路由 |
| [api-recon-and-docs](./skills/api-recon-and-docs/SKILL.md) | 60 行 | - | API 发现、OpenAPI/Swagger、隐藏端点 |
| [api-authorization-and-bola](./skills/api-authorization-and-bola/SKILL.md) | 47 行 | - | BOLA/BFLA、批量赋值、对象级授权 |
| [api-auth-and-jwt-abuse](./skills/api-auth-and-jwt-abuse/SKILL.md) | 75 行 | - | JWT 攻击、API 密钥滥用、令牌操控 |
| [graphql-and-hidden-parameters](./skills/graphql-and-hidden-parameters/SKILL.md) | 49 行 | - | GraphQL 内省、批量查询、隐藏参数发现 |

### 认证与授权

| 技能 | SKILL.md | SCENARIOS.md | 核心内容 |
|---|---|---|---|
| [auth-sec](./skills/auth-sec/SKILL.md) | 40 行 | - | 认证测试分类路由 |
| [authbypass-authentication-flaws](./skills/authbypass-authentication-flaws/SKILL.md) | 441 行 | - | 密码重置 22 模式矩阵、验证码绕过 20 方法、不安全随机数（UUID v1/mt_rand/ObjectId） |
| [jwt-oauth-token-attacks](./skills/jwt-oauth-token-attacks/SKILL.md) | 301 行 | - | JWT 算法混淆、密钥混淆、声明篡改、JWKS 滥用 |
| [oauth-oidc-misconfiguration](./skills/oauth-oidc-misconfiguration/SKILL.md) | 45 行 | - | OAuth 流程劫持、OIDC 配置错误 |
| [saml-sso-assertion-attacks](./skills/saml-sso-assertion-attacks/SKILL.md) | 40 行 | - | SAML 断言篡改、SSO 绕过 |
| [idor-broken-object-authorization](./skills/idor-broken-object-authorization/SKILL.md) | 336 行 | - | 8 类系统化 IDOR 测试、ORM 过滤链泄露（Django/Prisma/Ransack） |

### 注入攻击

| 技能 | SKILL.md | SCENARIOS.md | 核心内容 |
|---|---|---|---|
| [injection-checking](./skills/injection-checking/SKILL.md) | 49 行 | - | 注入测试分类路由 |
| [xss-cross-site-scripting](./skills/xss-cross-site-scripting/SKILL.md) | 368 行 | 278 行 | Polyglot payload 库、按厂商 WAF 绕过（Cloudflare/Akamai/Incapsula/WordFence）、CSP 绕过、DOM Clobbering、CSS 注入数据外带 |
| [sqli-sql-injection](./skills/sqli-sql-injection/SKILL.md) | 475 行 | 575 行 | DB2/Cassandra/BigQuery/SQLite 专题、SQLite RCE、WAF 绕过矩阵、CTF 技巧（handler/prepare/innodb） |
| [ssrf-server-side-request-forgery](./skills/ssrf-server-side-request-forgery/SKILL.md) | 314 行 | 226 行 | 云元数据 6 平台矩阵、DNS 重绑定、无头浏览器攻击、Gopher/Redis RCE 链 |
| [ssti-server-side-template-injection](./skills/ssti-server-side-template-injection/SKILL.md) | 340 行 | 319 行 | 15+ 引擎覆盖（Jinja2/Twig/Pug/Handlebars/EJS/Razor/EEx/Smarty）、盲注 SSTI、Flask PIN 计算 |
| [cmdi-command-injection](./skills/cmdi-command-injection/SKILL.md) | 494 行 | - | WAF 绕过（通配符/xor/base64）、PHP disable_functions 6 条绕过路径、组件级 RCE（ImageMagick/FFmpeg/ES） |
| [nosql-injection](./skills/nosql-injection/SKILL.md) | 341 行 | - | 盲注自动化提取脚本、重复键绕过、聚合管道注入、$where JS 执行 |
| [xxe-xml-external-entity](./skills/xxe-xml-external-entity/SKILL.md) | 326 行 | 112 行 | 本地 DTD 注入（Windows/Linux/JAR 17+ 路径）、盲 XXE、Gopher/FTP OOB |
| [deserialization-insecure](./skills/deserialization-insecure/SKILL.md) | 714 行 | - | Java/PHP/Python + Ruby Marshal/YAML 链、.NET BinaryFormatter/ViewState/JSON.NET、Node.js node-serialize/funcster |
| [ghost-bits-cast-attack](./skills/ghost-bits-cast-attack/SKILL.md) | 400+ 行 | PAYLOAD_COOKBOOK.md | Java char 转 byte 截断导致的 WAF 绕过（Black Hat Asia 2026）：Ghost Bits 以每个危险 ASCII 字节 255 个 Unicode 候选重新激活被 WAF 拦截的 SQLi/反序列化/文件上传/路径穿越/CRLF/请求走私，覆盖 Tomcat/Spring/Jetty/Jackson/Fastjson/BCEL/HttpClient/Angus Mail |
| [expression-language-injection](./skills/expression-language-injection/SKILL.md) | 243 行 | - | SpEL、OGNL、Java EL 注入及 RCE 链 |
| [jndi-injection](./skills/jndi-injection/SKILL.md) | 265 行 | - | JNDI/LDAP/RMI 利用、Log4Shell 模式 |
| [crlf-injection](./skills/crlf-injection/SKILL.md) | 175 行 | - | 头注入、HTTP 响应拆分 |
| [request-smuggling](./skills/request-smuggling/SKILL.md) | 298 行 | - | CL.TE/TE.CL/TE.TE 含 8 种混淆变体、HTTP/2 降级走私、客户端去同步 |
| [prototype-pollution](./skills/prototype-pollution/SKILL.md) | 190 行 | - | Express 黑盒探测键、EJS/Kibana Gadget 链、CVE-2019-7609 |
| [type-juggling](./skills/type-juggling/SKILL.md) | 291 行 | - | PHP 松散比较表、Magic Hash（MD5/SHA1/SHA256）、HMAC 0e 暴力、CTF 模式 |
| [http-parameter-pollution](./skills/http-parameter-pollution/SKILL.md) | 208 行 | - | 服务器行为矩阵（9 平台）、HPP+WAF 绕过组合 |
| [xslt-injection](./skills/xslt-injection/SKILL.md) | 281 行 | - | 三条 RCE 链（PHP/Java/.NET）、EXSLT 写文件、Vendor 检测 |
| [csv-formula-injection](./skills/csv-formula-injection/SKILL.md) | 144 行 | - | DDE/rundll32 payload、Google Sheets IMPORT* 外带 |

### 文件与路径攻击

| 技能 | SKILL.md | SCENARIOS.md | 核心内容 |
|---|---|---|---|
| [file-access-vuln](./skills/file-access-vuln/SKILL.md) | 32 行 | - | 文件访问测试分类路由 |
| [path-traversal-lfi](./skills/path-traversal-lfi/SKILL.md) | 603 行 | - | LFI→RCE 7 条路径、PHP Wrapper 矩阵（filter chains/oracle/phar）、pearcmd 四法、参数命名词典 |
| [upload-insecure-files](./skills/upload-insecure-files/SKILL.md) | 287 行 | 158 行 | 成功率公式、编辑器路径矩阵、验证缺陷五维分类、IIS/Apache/Nginx 解析技巧 |

### 业务逻辑与会话

| 技能 | SKILL.md | SCENARIOS.md | 核心内容 |
|---|---|---|---|
| [business-logic-vuln](./skills/business-logic-vuln/SKILL.md) | 32 行 | - | 业务逻辑测试分类路由 |
| [business-logic-vulnerabilities](./skills/business-logic-vulnerabilities/SKILL.md) | 339 行 | 298 行 | 支付篡改矩阵（10 种攻击）、状态机绕过方法论、优惠券/库存竞态 |
| [race-condition](./skills/race-condition/SKILL.md) | 286 行 | - | TOCTOU 模型、HTTP/1.1 末字节同步、HTTP/2 单包攻击、Turbo Intruder 模板、CVE-2022-4037 |
| [csrf-cross-site-request-forgery](./skills/csrf-cross-site-request-forgery/SKILL.md) | 324 行 | - | JSON CSRF 三种技术、Multipart 上传 CSRF、CSPT2CSRF 现代变体 |
| [clickjacking](./skills/clickjacking/SKILL.md) | 163 行 | - | 基于 Frame 的攻击、X-Frame-Options/CSP 绕过 |
| [cors-cross-origin-misconfiguration](./skills/cors-cross-origin-misconfiguration/SKILL.md) | 50 行 | 152 行 | Origin 反射、null origin、子域信任滥用 |
| [open-redirect](./skills/open-redirect/SKILL.md) | 184 行 | - | 重定向链滥用、Tabnabbing（反向标签劫持） |
| [web-cache-deception](./skills/web-cache-deception/SKILL.md) | 211 行 | - | 路径混淆、缓存键操控 |

### 高级 Web 安全

| 技能 | 核心内容 |
|---|---|
| [subdomain-takeover](./skills/subdomain-takeover/SKILL.md) | 悬挂 DNS 记录（CNAME/NS/A）、云服务指纹识别、验证绕过、多供应商接管手册 |
| [waf-bypass-techniques](./skills/waf-bypass-techniques/SKILL.md) | 编码链、分块传输技巧、HTTP 走私绕 WAF、厂商级绕过矩阵（Cloudflare/AWS WAF/Akamai/ModSecurity） |
| [csp-bypass-advanced](./skills/csp-bypass-advanced/SKILL.md) | Script Gadget、base-uri 滥用、JSONP 回调注入、可信 CDN 利用、CSP nonce/hash 泄露、strict-dynamic 绕过 |
| [http-host-header-attacks](./skills/http-host-header-attacks/SKILL.md) | 密码重置投毒、Host 头缓存投毒、路由型 SSRF、绝对 URL 覆盖技巧 |
| [dangling-markup-injection](./skills/dangling-markup-injection/SKILL.md) | 无需 JavaScript 的 HTML 注入数据外带、img/form/base 标签滥用、CSP 安全数据窃取 |
| [dns-rebinding-attacks](./skills/dns-rebinding-attacks/SKILL.md) | DNS 重绑定访问内网、TTL 操控、同源策略绕过、浏览器缓解措施绕过 |
| [email-header-injection](./skills/email-header-injection/SKILL.md) | SMTP 头注入、CC/BCC 操控、邮件中继滥用、注入头钓鱼 |
| [http2-specific-attacks](./skills/http2-specific-attacks/SKILL.md) | HTTP/2 请求走私（H2.CL/H2.TE）、HPACK 头压缩攻击、流多路复用滥用、HTTP/2→HTTP/1.1 降级 |
| [prototype-pollution-advanced](./skills/prototype-pollution-advanced/SKILL.md) | 服务端 Gadget 链发现、框架级 PP→RCE（Express/Fastify/Next.js）、AST 注入、构建工具原型投毒 |
| [401-403-bypass-techniques](./skills/401-403-bypass-techniques/SKILL.md) | 路径规范化技巧、HTTP 动词篡改、头部绕过（X-Original-URL/X-Rewrite-URL）、代理配置错误、IP 访问控制绕过 |

### 基础设施与网络

| 技能 | 核心内容 |
|---|---|
| [unauthorized-access-common-services](./skills/unauthorized-access-common-services/SKILL.md) | 服务暴露清单、反向代理配置错误（Nginx off-by-slash、X-Forwarded-For 信任、Caddy 模板注入） |
| [insecure-source-code-management](./skills/insecure-source-code-management/SKILL.md) | .git/.svn/.hg/.bzr 恢复、403 vs 404 检测、备份文件模式 |
| [dependency-confusion](./skills/dependency-confusion/SKILL.md) | npm/pip/gem 公共注册表劫持、清单识别、scope/namespace 防御 |
| [websocket-security](./skills/websocket-security/SKILL.md) | CSWSH、Origin 验证、wsrepl/ws-harness 工具链 |
| [network-protocol-attacks](./skills/network-protocol-attacks/SKILL.md) | ARP 欺骗、DNS 投毒、LLMNR/NBT-NS 投毒、DHCP 耗尽、IPv6 攻击、协议层中间人 |
| [tunneling-and-pivoting](./skills/tunneling-and-pivoting/SKILL.md) | SSH 隧道（本地/远程/动态）、SOCKS 代理链、chisel/ligolo-ng、端口转发、DNS/ICMP 隧道 |
| [reverse-shell-techniques](./skills/reverse-shell-techniques/SKILL.md) | 多语言 Shell 生成、加密反弹 Shell（OpenSSL/ncat）、分阶段 Payload、防火墙绕过、Web Shell |

### Linux 与容器安全

| 技能 | 核心内容 |
|---|---|
| [linux-privilege-escalation](./skills/linux-privilege-escalation/SKILL.md) | SUID/SGID 滥用、内核漏洞利用、sudo 配置错误、cron 任务、Linux Capabilities、可写服务文件、NFS no_root_squash |
| [container-escape-techniques](./skills/container-escape-techniques/SKILL.md) | Docker socket 滥用、特权容器逃逸、cgroup 越狱、runc 漏洞、挂载敏感路径 |
| [linux-security-bypass](./skills/linux-security-bypass/SKILL.md) | SELinux/AppArmor 绕过、seccomp 过滤器逃逸、namespace 滥用、LD_PRELOAD 技巧 |
| [linux-lateral-movement](./skills/linux-lateral-movement/SKILL.md) | SSH 密钥收割、凭据复用、服务利用、NFS/共享挂载滥用、cron 持久化 |
| [kubernetes-pentesting](./skills/kubernetes-pentesting/SKILL.md) | Pod 安全策略绕过、RBAC 滥用、ServiceAccount 令牌窃取、etcd 访问、容器镜像后门、kubelet API |

### Windows 与 Active Directory

| 技能 | 核心内容 |
|---|---|
| [windows-privilege-escalation](./skills/windows-privilege-escalation/SKILL.md) | 令牌操控、服务配置错误、DLL 劫持、UAC 绕过、AlwaysInstallElevated、未引用服务路径、PrintSpoofer/Potato |
| [active-directory-kerberos-attacks](./skills/active-directory-kerberos-attacks/SKILL.md) | Kerberoasting、AS-REP Roasting、黄金/白银票据、委派滥用（非约束/约束/RBCD）、Diamond Ticket |
| [active-directory-acl-abuse](./skills/active-directory-acl-abuse/SKILL.md) | ACL/DACL 利用、DCSync、对象所有权滥用、WriteDACL/GenericAll/GenericWrite 攻击路径、BloodHound 集成 |
| [active-directory-certificate-services](./skills/active-directory-certificate-services/SKILL.md) | ESC1–ESC8 攻击模式、证书模板滥用、PKINIT 利用、Shadow Credentials、CA 持久化 |
| [ntlm-relay-coercion](./skills/ntlm-relay-coercion/SKILL.md) | PetitPotam、PrinterBug、NTLM 中继链、强制认证技术、WebDAV 中继、NTLM 降级 |
| [windows-lateral-movement](./skills/windows-lateral-movement/SKILL.md) | PsExec、WMI、WinRM、DCOM、Pass-the-Hash/Pass-the-Ticket、RDP 劫持、计划任务、服务部署 |
| [windows-av-evasion](./skills/windows-av-evasion/SKILL.md) | AMSI 绕过、ETW 补丁、API Unhooking、Shellcode 加载器、Living-off-the-Land（LOLBins）、Payload 加密/混淆 |

### macOS 安全

| 技能 | 核心内容 |
|---|---|
| [macos-security-bypass](./skills/macos-security-bypass/SKILL.md) | Gatekeeper 绕过、TCC 滥用、SIP/AMFI 考量、LaunchAgent/LaunchDaemon 持久化、隔离标志绕过 |
| [macos-process-injection](./skills/macos-process-injection/SKILL.md) | Dylib 注入/劫持、task_for_pid、XPC 利用、Electron 应用注入、DYLD_INSERT_LIBRARIES |

### 移动安全

| 技能 | 核心内容 |
|---|---|
| [android-pentesting-tricks](./skills/android-pentesting-tricks/SKILL.md) | APK 分析与逆向、Frida Hook、Intent 利用、Root 检测绕过、Content Provider 泄露、WebView 攻击 |
| [ios-pentesting-tricks](./skills/ios-pentesting-tricks/SKILL.md) | IPA 分析、Objective-C Runtime 操控、越狱检测绕过、Keychain 访问、URL Scheme 滥用、二进制保护 |
| [mobile-ssl-pinning-bypass](./skills/mobile-ssl-pinning-bypass/SKILL.md) | Android/iOS 证书钉扎绕过、Frida/Objection 脚本、动态插桩技术、网络安全配置修改 |

### 二进制漏洞利用（Pwn）

| 技能 | 核心内容 |
|---|---|
| [stack-overflow-and-rop](./skills/stack-overflow-and-rop/SKILL.md) | 缓冲区溢出、ROP 链构建、ret2libc、SROP（Sigreturn-Oriented Programming）、栈迁移、one-gadget |
| [heap-exploitation](./skills/heap-exploitation/SKILL.md) | UAF、Double Free、tcache 投毒、fastbin attack、House of 系列技术、safe-linking 绕过 |
| [format-string-exploitation](./skills/format-string-exploitation/SKILL.md) | 格式化字符串读写原语、GOT 覆写、任意地址写入、FORTIFY_SOURCE 绕过 |
| [kernel-exploitation](./skills/kernel-exploitation/SKILL.md) | 内核 ROP、ret2usr、SMEP/SMAP/KPTI 绕过、内核竞态条件、modprobe_path 覆写、msg_msg 利用 |
| [browser-exploitation-v8](./skills/browser-exploitation-v8/SKILL.md) | V8 引擎漏洞利用、JIT 编译 Bug、类型混淆、OOB 越界读写、沙箱逃逸链、wasm 滥用 |
| [sandbox-escape-techniques](./skills/sandbox-escape-techniques/SKILL.md) | 浏览器沙箱逃逸、seccomp 绕过、IPC 滥用、内核利用提权、策略文件篡改 |
| [binary-protection-bypass](./skills/binary-protection-bypass/SKILL.md) | ASLR/NX/PIE/Canary/Full RELRO 绕过技术、信息泄露利用、部分覆写、GOT 解引用 |
| [arbitrary-write-to-rce](./skills/arbitrary-write-to-rce/SKILL.md) | 写原语到代码执行（GOT/__free_hook/__malloc_hook）、FSOP、_IO_FILE 利用、exit handler 覆写 |

### 逆向工程

| 技能 | 核心内容 |
|---|---|
| [anti-debugging-techniques](./skills/anti-debugging-techniques/SKILL.md) | ptrace 检测、时间检查、自修改代码、反 VM 技术、调试标志检查、异常型反调试 |
| [code-obfuscation-deobfuscation](./skills/code-obfuscation-deobfuscation/SKILL.md) | 控制流平坦化、不透明谓词、字符串加密、混淆工具分析（OLLVM/Themida/VMProtect）、自动去混淆 |
| [symbolic-execution-tools](./skills/symbolic-execution-tools/SKILL.md) | angr、Z3、Triton 自动化漏洞发现、约束求解、路径探索、混合执行 |
| [vm-and-bytecode-reverse](./skills/vm-and-bytecode-reverse/SKILL.md) | 自定义 VM/字节码分析、Python/Java/.NET 反编译、VM Handler 还原、操作码映射 |

### 密码学攻击

| 技能 | 核心内容 |
|---|---|
| [rsa-attack-techniques](./skills/rsa-attack-techniques/SKILL.md) | Wiener 攻击、Boneh-Durfee、Hastad 广播攻击、公共模数攻击、Coppersmith 小根、Franklin-Reiter、PKCS#1 v1.5 填充预言机 |
| [symmetric-cipher-attacks](./skills/symmetric-cipher-attacks/SKILL.md) | CBC 填充预言机、比特翻转攻击、ECB 剪切粘贴、中间相遇攻击、已知明文攻击、IV 复用利用 |
| [lattice-crypto-attacks](./skills/lattice-crypto-attacks/SKILL.md) | LLL/BKZ 格规约、隐藏数问题、NTRU 攻击、CVP/SVP 求解、背包密码系统攻击 |
| [hash-attack-techniques](./skills/hash-attack-techniques/SKILL.md) | 长度扩展攻击、生日攻击、哈希碰撞利用、bcrypt/scrypt/argon2 分析、HMAC 时序攻击 |
| [classical-cipher-analysis](./skills/classical-cipher-analysis/SKILL.md) | 频率分析、Vigenère/Kasiski 破解、Hill 密码、置换密码、转置密码、Enigma 式分析、自动化求解 |

### 区块链与智能合约

| 技能 | SKILL.md | 补充文档 | 核心内容 |
|---|---|---|---|
| [smart-contract-vulnerabilities](./skills/smart-contract-vulnerabilities/SKILL.md) | 314 行 | 460 行 | 重入攻击（4 种变体）、整数溢出、delegatecall 存储碰撞、签名重放、CREATE2 利用、闪电贷模式 |
| [defi-attack-patterns](./skills/defi-attack-patterns/SKILL.md) | 355 行 | - | 闪电贷预言机操控、MEV 三明治/JIT/清算、首存者金库攻击、治理闪电贷、跨链桥利用、转账扣费代币 |

### AI/ML 与 LLM 安全

| 技能 | SKILL.md | 补充文档 | 核心内容 |
|---|---|---|---|
| [llm-prompt-injection](./skills/llm-prompt-injection/SKILL.md) | 357 行 | 306 行 | 直接/间接注入、RAG 投毒、工具/函数滥用、Markdown 外带、MCP 安全风险、编码绕过 |
| [ai-ml-security](./skills/ai-ml-security/SKILL.md) | 425 行 | - | Pickle RCE 模型投毒、对抗样本（FGSM/PGD/C&W）、训练数据投毒、模型提取、成员推断、Agent 安全 |

### 取证与隐写术

| 技能 | 核心内容 |
|---|---|
| [memory-forensics-volatility](./skills/memory-forensics-volatility/SKILL.md) | Volatility 框架、进程/模块分析、网络痕迹提取、恶意软件检测、注册表 Hive 分析、时间线重建 |
| [steganography-techniques](./skills/steganography-techniques/SKILL.md) | LSB 提取、文件格式分析、音频/图像隐写工具（zsteg/stegsolve/steghide）、EXIF 元数据、多层嵌入 |
| [traffic-analysis-pcap](./skills/traffic-analysis-pcap/SKILL.md) | Wireshark/tshark 分析、协议解剖、流量数据提取、加密流量识别、流重建 |

## 技能选择建议

| 现象 | 推荐入口 | 说明 |
|---|---|---|
| 新目标、信息不足 | [recon-for-sec](./skills/recon-for-sec/SKILL.md) | 先做方法论和资产理解 |
| REST API、GraphQL、移动端后端 | [api-sec](./skills/api-sec/SKILL.md) | 先分流到 recon、authz、token 或 GraphQL |
| 登录、重置密码、2FA、JWT、OAuth | [auth-sec](./skills/auth-sec/SKILL.md) | 先分认证、授权、协议配置 |
| HTML/JS 反射、模板表达式 | [injection-checking](./skills/injection-checking/SKILL.md) | 先确定 XSS、SQLi、SSRF、XXE 还是 SSTI |
| 文件路径、下载接口、上传链路 | [file-access-vuln](./skills/file-access-vuln/SKILL.md) | 先分 LFI/Traversal 与 Upload |
| 优惠券、支付、状态机 | [business-logic-vuln](./skills/business-logic-vuln/SKILL.md) | 先按业务规则和竞态建模 |
| HTTP 请求解析异常 | [request-smuggling](./skills/request-smuggling/SKILL.md) | 前后端分帧不一致 |
| Node.js `__proto__` 可控 | [prototype-pollution](./skills/prototype-pollution/SKILL.md) | 客户端 PP→XSS、服务端 PP→RCE |
| PHP 弱比较、0e hash | [type-juggling](./skills/type-juggling/SKILL.md) | 松散比较认证绕过 |
| .git/.svn/.env 路径可访问 | [insecure-source-code-management](./skills/insecure-source-code-management/SKILL.md) | 源码恢复 |
| 内部包名出现在清单文件中 | [dependency-confusion](./skills/dependency-confusion/SKILL.md) | 供应链劫持 |
| WebSocket 协议升级 | [websocket-security](./skills/websocket-security/SKILL.md) | CSWSH 和 WS 注入 |
| CSV/Excel 导出功能 | [csv-formula-injection](./skills/csv-formula-injection/SKILL.md) | 导出中的 DDE 注入 |
| 一次性操作（优惠券、奖励） | [race-condition](./skills/race-condition/SKILL.md) | 并发请求突破限额 |
| 智能合约、Solidity、EVM 审计 | [smart-contract-vulnerabilities](./skills/smart-contract-vulnerabilities/SKILL.md) | 重入、溢出、访问控制、delegatecall |
| DeFi 协议、闪电贷、预言机、MEV | [defi-attack-patterns](./skills/defi-attack-patterns/SKILL.md) | 闪电贷、三明治、治理、跨链桥 |
| LLM、聊天机器人、提示词注入、RAG | [llm-prompt-injection](./skills/llm-prompt-injection/SKILL.md) | 直接/间接注入、工具滥用、MCP |
| ML 模型、对抗样本、模型投毒 | [ai-ml-security](./skills/ai-ml-security/SKILL.md) | 供应链、对抗样本、提取、Agent |
| WAF 拦截 payload | [waf-bypass-techniques](./skills/waf-bypass-techniques/SKILL.md) | 编码链、分块传输、厂商级绕过 |
| 子域名悬挂 CNAME/DNS | [subdomain-takeover](./skills/subdomain-takeover/SKILL.md) | 云服务接管、NS 委派劫持 |
| CSP 阻止 XSS 执行 | [csp-bypass-advanced](./skills/csp-bypass-advanced/SKILL.md) | Script Gadget、JSONP、可信 CDN、strict-dynamic |
| 目标端点返回 401/403 | [401-403-bypass-techniques](./skills/401-403-bypass-techniques/SKILL.md) | 路径规范化、动词篡改、头部技巧 |
| HTTP/2 协议端点 | [http2-specific-attacks](./skills/http2-specific-attacks/SKILL.md) | H2 走私、HPACK 滥用、降级攻击 |
| Linux 主机、存在 SUID/sudo | [linux-privilege-escalation](./skills/linux-privilege-escalation/SKILL.md) | 内核、SUID、cron、capabilities、服务 |
| Docker/Kubernetes 环境 | [container-escape-techniques](./skills/container-escape-techniques/SKILL.md) | Docker socket、特权逃逸、cgroup 越狱 |
| Kubernetes 集群访问 | [kubernetes-pentesting](./skills/kubernetes-pentesting/SKILL.md) | RBAC 滥用、SA 令牌、etcd、Pod 安全绕过 |
| Windows 主机、需要本地管理员 | [windows-privilege-escalation](./skills/windows-privilege-escalation/SKILL.md) | 令牌、服务、DLL 劫持、UAC、Potato 攻击 |
| Active Directory、域环境 | [active-directory-kerberos-attacks](./skills/active-directory-kerberos-attacks/SKILL.md) | Kerberoast、AS-REP、黄金/白银票据 |
| AD CS、证书模板 | [active-directory-certificate-services](./skills/active-directory-certificate-services/SKILL.md) | ESC1–ESC8、模板滥用、Shadow Credentials |
| NTLM 哈希、中继机会 | [ntlm-relay-coercion](./skills/ntlm-relay-coercion/SKILL.md) | PetitPotam、PrinterBug、中继链 |
| Windows AV/EDR 阻止执行 | [windows-av-evasion](./skills/windows-av-evasion/SKILL.md) | AMSI 绕过、Unhooking、LOLBins、Payload 混淆 |
| macOS 端点访问 | [macos-security-bypass](./skills/macos-security-bypass/SKILL.md) | Gatekeeper、TCC、SIP 考量 |
| Android/iOS 应用测试 | [android-pentesting-tricks](./skills/android-pentesting-tricks/SKILL.md) | APK 分析、Frida、Intent、Root 检测绕过 |
| SSL Pinning 阻止代理 | [mobile-ssl-pinning-bypass](./skills/mobile-ssl-pinning-bypass/SKILL.md) | Frida/Objection 脚本、动态插桩 |
| 二进制/ELF/PE 漏洞利用 | [stack-overflow-and-rop](./skills/stack-overflow-and-rop/SKILL.md) | 缓冲区溢出、ROP、ret2libc、SROP |
| 堆破坏、UAF | [heap-exploitation](./skills/heap-exploitation/SKILL.md) | tcache/fastbin 攻击、House of 系列 |
| 内核级漏洞利用 | [kernel-exploitation](./skills/kernel-exploitation/SKILL.md) | 内核 ROP、SMEP/SMAP 绕过、modprobe_path |
| 浏览器 0-day、V8/JSC | [browser-exploitation-v8](./skills/browser-exploitation-v8/SKILL.md) | JIT Bug、类型混淆、沙箱逃逸 |
| 混淆/加壳二进制 | [code-obfuscation-deobfuscation](./skills/code-obfuscation-deobfuscation/SKILL.md) | 控制流、不透明谓词、VM 保护 |
| CTF 密码学题（RSA） | [rsa-attack-techniques](./skills/rsa-attack-techniques/SKILL.md) | Wiener、Coppersmith、公共模数、填充预言机 |
| CTF 密码学题（AES/DES） | [symmetric-cipher-attacks](./skills/symmetric-cipher-attacks/SKILL.md) | 填充预言机、比特翻转、ECB 模式攻击 |
| CTF 密码学题（格密码） | [lattice-crypto-attacks](./skills/lattice-crypto-attacks/SKILL.md) | LLL/BKZ、隐藏数问题、背包 |
| CTF 古典密码 | [classical-cipher-analysis](./skills/classical-cipher-analysis/SKILL.md) | 频率分析、Vigenère、置换密码 |
| 内存转储分析 | [memory-forensics-volatility](./skills/memory-forensics-volatility/SKILL.md) | Volatility、进程/网络分析、恶意软件检测 |
| 图像/音频隐藏数据 | [steganography-techniques](./skills/steganography-techniques/SKILL.md) | LSB、格式分析、隐写工具 |
| PCAP 流量抓包 | [traffic-analysis-pcap](./skills/traffic-analysis-pcap/SKILL.md) | Wireshark、协议解剖、流重建 |
| 需要网络穿透/横向 | [tunneling-and-pivoting](./skills/tunneling-and-pivoting/SKILL.md) | SSH 隧道、SOCKS 代理、chisel/ligolo-ng |
| 需要在目标获取反弹 Shell | [reverse-shell-techniques](./skills/reverse-shell-techniques/SKILL.md) | 多语言 Shell、加密、分阶段 Payload |

## 安装方式

### 通用安装

```bash
npx skills add yaklang/hack-skills
```

### Raw URL 安装

```bash
curl -fsSL https://raw.githubusercontent.com/yaklang/hack-skills/main/skills/hack/SKILL.md
```

### 本地作为知识库使用

```bash
git clone https://github.com/yaklang/hack-skills.git
cd hack-skills
```

### 离线加密 ZIP

针对隔离网络、网速慢、或者会被 AV / EDR / 浏览器内容扫描器把明文攻防 markdown 误判删除的场景：

```bash
curl -fsSLO https://oss-qn.yaklang.com/hack-skills/latest/hack-skills.zip
7z x -phack-skills hack-skills.zip
# 或：unzip -P hack-skills hack-skills.zip
```

| 通道 | 链接 |
|---|---|
| 主 CDN | <https://oss-qn.yaklang.com/hack-skills/latest/hack-skills.zip> |
| 备用 CDN | <https://aliyun-oss.yaklang.com/hack-skills/latest/hack-skills.zip> |
| 当前构建版本号 | <https://oss-qn.yaklang.com/hack-skills/latest/version.txt> |

**关于 ZIP 密码。** ZIP 包采用 **AES-256** 加密，密码是**公开常量** `hack-skills`。**这不是访问控制**——任何人都能下载、任何人都能解压，密码在 README、网页版、GitHub Actions workflow 和 CI 日志里都明示。它存在的唯一目的，是绕过 AV / EDR / 浏览器扫描器对明文攻防 markdown 的内容启发式特征命中——避免文件在传输过程被静默拦截或隔离。打包、加密设置和完整性校验细节在 [`.github/workflows/upload-hack-skills.yml`](./.github/workflows/upload-hack-skills.yml)。

同一份 ZIP 在网页版顶部 `ZIP` 按钮和 `Install → Offline ZIP` Tab 里也提供一键下载。

## 设计原则

- 安全知识优先于花哨包装。
- 内容可审查性优先于数量扩张。
- 优先服务授权测试、合法研究与防御验证场景。
- 目录名尽量让人一眼看出安全语义。
- 不包含任何特定客户信息；所有内容均为通用教育性方法论。

## 无害化 PoC 政策

技能文档的目标是教你如何**证明**漏洞、并在授权场景下取得访问权，而非提供"唯一作用就是
不可逆地销毁数据或让目标下线"的 Payload。边界很简单：

- **移除 / 无害化——不可逆的破坏性操作。** 数据销毁类（`DROP` / `TRUNCATE` / 大范围 `DELETE` /
  Redis `FLUSHALL`）与可用性销毁类（服务 `shutdown` / 重启 / DoS 端点）一律替换为能证明同一
  注入点或同等能力的无害化证明。
- **保留——非破坏性的访问原语。** 真实授权测试所需的标准攻击 PoC 予以保留，包括 RCE，例如
  WebShell 写入（`INTO OUTFILE` / `DUMPFILE`）、反弹 Shell、命令执行——它们能取得访问权而
  不会摧毁环境。

| 被规避的破坏性形态 | 改用 | 为何仍能证明问题 |
|---|---|---|
| `...; DROP TABLE users;--`（堆叠查询） | `...; SELECT SLEEP(5);--` | 时间延迟即可证明任意堆叠语句执行，且不丢数据 |
| `DELETE FROM ... WHERE x='' OR '1'='1'`（删空全表） | 在 DELETE 上下文中用时间盲注证明 | 确认注入但不删除任何行 |
| `redis-cli ... FLUSHALL` / gopher 链中的 `flushall` | 删除，或在 RCE 链中替换为 `PING` | 保留 RCE 链完整，仅去掉清库动作 |
| `/actuator/shutdown`（DoS） | `/actuator/env`、`/actuator/heapdump`、`/actuator/threaddump` | 同等侦察/泄密价值，但不会让应用下线 |

约定：

- 优先使用**时间盲注**（`SLEEP`/`pg_sleep`/`WAITFOR DELAY`）或**只读**证明，避免任何销毁
  数据或可用性的语句。
- **RCE 仍在范围内：** WebShell 写入、反弹 Shell、命令执行属于非破坏性访问原语，是授权测试的
  必要组成，刻意予以保留。
- 标准只读侦察命令（如 `xp_cmdshell 'whoami'`、`LOAD_FILE('/etc/passwd')`）予以保留。
- 防御性加固片段（如 `rename-command FLUSHALL ""`）作为防御措施予以保留。

## 相关项目（姊妹站）

HackSkills 是 Yaklang 生态下面向 Agent 的系列知识库之一：

| 项目 | 在线站点 | 定位 |
|---|---|---|
| **HackSkills**（本库） | [skills.hackbenchmark.com](https://skills.hackbenchmark.com) | 面向 AI Agent 的渗透/安全攻防技能库（Web/API/认证/提权/逆向/密码学等） |
| **Yak Skills** | [skills.yaklang.io](https://skills.yaklang.io) · [yaklang/yak-skills](https://github.com/yaklang/yak-skills) | Yaklang 编程 + Yak 热加载技能库（MITM / Web Fuzzer / 全局热加载、前端加密对抗） |
| **训练素材** | [yaklang/yaklang-ai-training-materials](https://github.com/yaklang/yaklang-ai-training-materials) | 更复杂的 Yak 代码、100+ 标准库实战、详细案例 |
| **Benchmark** | [hackbenchmark.com](https://hackbenchmark.com) | AI Agent 对抗真实 Web 漏洞 |

> 当某个发现需要定制工具——代理热加载、流量解密、Fuzz 逻辑——从 HackSkills 切到 **Yak Skills** 获取 Yaklang 实现范式。

## 贡献方式

欢迎提交 PR，重点方向包括：

- 新漏洞类别与高价值案例
- 更好的漏洞赏金与渗透测试方法论
- OS 级提权路径与 AD 攻击链
- CTF 竞赛技巧（Pwn、Crypto、逆向、取证）
- Agent 容易忽略的边界条件
- 风险提示、术语统一与内容去噪

贡献内容建议满足：可验证、可审查、不鼓励未授权攻击、能帮助 Agent 在真实任务中更稳健地推理与执行。
