# Business Logic Vulnerability Testing Methodology / 业务逻辑漏洞测试方法论

> Companion to [SKILL.md](./SKILL.md) and [CHECKLIST.md](./CHECKLIST.md). Distilled from real-world payment / captcha / authentication / data-exposure logic flaw cases and live Java code-audit walkthroughs.

业务逻辑漏洞与传统注入/溢出类漏洞最大的不同在于：**它没有固定的特征签名，几乎无法被自动化工具发现，依赖测试者对目标业务的深度理解和经验**。本方法论给出一套可重复、可检索的工程化流程，把"凭感觉挖洞"压缩为"按图索骥"。

---

## 0. Why this methodology / 为什么不是直接套 checklist

业务逻辑漏洞通常具有以下四个特征：

- **黑盒难发现**：规律性弱、条件苛刻（依赖会话 Cookie、特定登录态、特定业务流转），扫描器几乎无效。
- **必须人工**：每一类漏洞都强耦合于具体业务（支付、退款、注册、抽奖），换一个系统几乎要从零理解。
- **代码审计痛但有效**：黑盒覆盖不到所有 API；只看代码又脱离业务；必须双管齐下。
- **看到现象只是开始**：单个 200/302/`success:true` 不能说明问题，要回到业务上去推"它是不是本来就该这样"。

因此本方法论的核心是 **5 个阶段 + 1 张索引表**，每个阶段都给出"该做什么 / 不该做什么 / 看到什么算异常"。

---

## 1. Five-Phase Workflow / 五阶段工作流

```
┌─────────────────────────────────────────────────────────────────┐
│ PHASE 1: Business Modeling                                       │
│   业务建模：把目标产品当成一台状态机来读                         │
│   ├── 角色清单 (admin / user / vip / guest / 内部接口调用方)     │
│   ├── 资源清单 (订单 / 余额 / 优惠券 / 积分 / VIP / 个人资料)    │
│   ├── 状态字段 (paid / unpaid / refunded / shipped / received)   │
│   └── 关键金钱/资产流向 (谁付钱、谁拿钱、谁能撤销)                │
└────────────────────────────┬────────────────────────────────────┘
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│ PHASE 2: State Machine & Data Flow Analysis                      │
│   状态机/数据流分析：每一步状态由谁决定？                       │
│   ├── happy path 走通                                            │
│   ├── 抓全所有跳转请求 + 响应                                    │
│   ├── 标出哪些字段是"前端传的"vs"服务端回写的"                  │
│   └── 找跨步骤复用的 token / order_id / coupon_id / status       │
└────────────────────────────┬────────────────────────────────────┘
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│ PHASE 3: Attack Surface Classification                           │
│   攻击面分类：用 5×N 矩阵代替"凭感觉"                            │
│   ├── 5 类操作：参数篡改 / 流程跳跃 / 重放 / 并发 / 替换身份     │
│   └── N 个业务模块：注册 / 登录 / 找回 / 支付 / 优惠 / 订单 / ...│
└────────────────────────────┬────────────────────────────────────┘
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│ PHASE 4: Checklist-Driven Testing                                │
│   按 CHECKLIST.md 逐条复测，每条带 "为什么会出问题 + 如何复测"  │
└────────────────────────────┬────────────────────────────────────┘
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│ PHASE 5: Human Judgement & Reporting                             │
│   人脑判断：服务端是否真的接受？业务上是不是真的"占便宜"？      │
└─────────────────────────────────────────────────────────────────┘
```

---

## 2. Phase 1: Business Modeling / 业务建模

> 业务逻辑漏洞的本质：**不了解业务功能，就找不到脆弱点。**

### 2.1 Role enumeration / 列清角色

不仅是 UI 上能切换的角色，**审计代码 / 抓包响应 / 错误回显**里出现过的所有角色都要列：

```text
guest               不登录可以走到哪一步
user (普通)         绑定手机后能做什么
vip / paid          付费用户的特权接口列表
admin / super       管理员功能与 admin URL 路径
internal / cron     /api/internal/* 这种"按理不该外网可达"的路径
third-party caller  webhook、回调、签约方调起的接口
```

**实战要点**：很多 Java Web 项目仅依赖 `if (user.role == "admin")` 跳页面，**没有 Filter 也没有 Spring Security 全局拦截**，意味着所有 `/teacher/*` 这类"按理只有特定角色能访问"的 URL，guest 直接 GET 也能拿到数据。这类系统应当默认所有非登录页都未授权。

### 2.2 Resource and asset inventory / 列清资产

凡涉及"钱、积分、库存、VIP 时长、优惠券、邀请码、个人隐私"的字段都列出来：

```text
amount / total / price / discount_amount / shipping_fee
balance / points / coupon_balance / gift_card_balance
quantity / stock / max_per_user
vip_level / vip_expire_at / membership_status
real_name / id_card / phone / email / avatar
```

任何**前端可见但又能被请求体改动**的资产字段，都是攻击优先级 P0。

### 2.3 Money / privilege flow / 画清钱权流向

```
支付场景：              退款场景：
  user → pay → order      order → refund → user
                          (能不能让钱回流的同时又不退货？)

VIP 场景：              邀请场景：
  user → pay → vip(time)  inviter ← reward ← invitee_register
                          (能不能 invitee 自己注册自己刷返佣？)
```

参考自 8.逻辑漏洞.pdf page_00009 的支付攻击面图：

- **订单**：改价 / 负数 / 库存冲突 / 零库存购买 / 改订单金额
- **结算**：优惠券复用 / 拦截改金额、改支付方式 / 伪造刷单
- **支付**：伪造第三方确认 / 窃取付款信息
- **退货**：绕过商家或商品类型限制
- **收货**：绕过客户确认

---

## 3. Phase 2: State Machine and Data Flow / 状态机分析

### 3.1 三步法（标准排查流程）

来自 逻辑漏洞checklist.pdf page_00003：

```
1. 明确流程       让真实账号把 happy path 完整走一遍，全程 Burp 抓包
2. 找可操控环节   每一步都问：这一步的状态是前端传的，还是服务端独立判定的
3. 改参重放对比   只改其中一个变量，与原始包对比响应差异
```

### 3.2 Server-side vs client-side fields / 区分谁说了算

每抓到一个请求都要做一个分类：

| 字段类型 | 例子 | 风险 |
|---|---|---|
| 服务端必须强校验 | `price`, `amount`, `quantity`, `user_id`, `vip_level`, `payment_status` | P0：前端可改即漏洞 |
| 服务端二次计算 | `total = sum(items.price * items.qty)` | 看后端是否信任前端传过来的 total |
| 服务端只用作 hint | `currency_display`, `ui_theme` | 一般不构成漏洞 |
| 客户端伪状态 | `success:true`、`error_code:1`、`is_paid:1` | P0：响应包改了能影响后续业务 |

**真实案例**（8.逻辑漏洞.pdf page_00022/page_00023）：修改手机号场景，服务端返回 `{"error_msg":"旧密码错误","error_code":99999,"success":false}`，攻击者把响应改成 `{"success": true,"error_code": 1}`，前端"看到 success 就放行"导致绕过密码校验直接换绑手机号。**根因是客户端过度信任服务端响应里的状态字段**，必须在服务端独立持久化"该用户是否已通过密码验证"。

### 3.3 跨步骤数据追踪

找出一切**会跨步骤复用**的标识符：

```
order_id     一笔订单从下单 → 支付 → 发货 → 售后用的是不是同一个？是否可被替换？
coupon_id    应用优惠券、提交订单、支付时是否每一步都重新校验？
auth token   找回密码 token / 邮箱验证 token / 邀请 token 是否有效期、是否一次性、是否绑会话
session_id   登录前后 session 是否换发？修改邮箱 / 密码后是否吊销
nonce        防重放参数有没有？服务端是否真的拒绝重复 nonce？
```

逻辑漏洞checklist.pdf page_00011 给出一个经典例子：找回密码 URL `resetpassword.php?id=MD5`，如果 auth 是用 PHP 老版本 Windows 下的 `rand()`（最大值 32768）生成 md5，攻击者可以本地遍历 0~32768 全量计算 md5 字典，**整个找回密码机制相当于失效**。

```php
$a=0; for ($a=0;$a<=32768;$a++){ $b=md5($a); echo "\r\n"; echo $b; }
```

---

## 4. Phase 3: Attack Surface Matrix / 攻击面 5×N 矩阵

把攻击面拆成"操作 × 业务模块"的二维表，每个格子有固定测试套路：

|   | 注册 | 登录 | 找回密码 | 支付/充值 | 优惠/积分 | 订单 | VIP/会员 | 验证码 | 隐私/资料 | 第三方/回调 |
|---|------|------|---------|----------|----------|------|---------|--------|-----------|-------------|
| **参数篡改** | 改 token | 改 user_id | 改邮箱/手机号 | 改金额/数量 | 改面值/复用 | 改总价 | 改等级 | 删验证码字段 | 改 user_id | 改回调 url |
| **流程跳跃** | 跳过验证码 | 跳 2FA | 跳验证步直接到改密 | 跳支付到确认 | 跳活动时间窗口 | 跳收货 | 跳付费直接开通 | 跳人机验证 | 跳所有人审 | 跳签名校验 |
| **重放** | 重放注册请求 | 重放登录 | 重放重置 token | 重放支付回调 | 重放领券 | 重放下单 | 重放充值 | 重放短信触发 | 重放修改 | 重放回调 |
| **并发** | 同邮箱并发注册 | 撞库爆破 | 抢 token 窗口 | 双花 / 并发付款 | 优惠券双开 | 超卖 | 多端并发升级 | 验证码并发触发 | 并发改资料 | 回调并发触发 |
| **替换身份** | 改 cookie 注册他人 | 越权登录 | 改手机号收他人验证码 | 替换收款人 | 替换领券人 | 替换订单归属 | 替换 vip user | 改手机号收码 | 改 user_id 越权 | 替换商户号 |

**用法**：每个项目至少把矩阵全过一遍，能想出测试用例就在 CHECKLIST.md 里把对应行打钩。

### 4.1 真实案例索引到矩阵

| 案例 | 矩阵位置 | 出处 |
|------|---------|------|
| 删除 `prizeIdList` 字段实现 0 元购 | 参数篡改 × 支付 | 8.逻辑漏洞.pdf page_00010 (Keep 跑步活动) |
| `quantity:0.02` 实现半价 | 参数篡改 × 订单 | 8.逻辑漏洞.pdf page_00011 (商城) |
| `quantity=999999999` 整数溢出归零 | 参数篡改 × 支付 | 8.逻辑漏洞.pdf page_00008 |
| Burp Turbo Intruder `concurrentConnections=30` 绕过短信限频 | 并发 × 验证码 | 8.逻辑漏洞.pdf page_00016 |
| 删除 `JSESSIONID` 重置验证码次数 | 参数篡改 × 验证码 | 8.逻辑漏洞.pdf page_00017 |
| 修改响应 `success:false → true` 绕过改手机号 | 参数篡改 × 隐私 | 8.逻辑漏洞.pdf page_00022 |
| 故意填错身份证使审核驳回，反复修改实名 | 流程跳跃 × 隐私 | 8.逻辑漏洞.pdf page_00029 |
| 抓包发现 `previewPullUrl: *.flv` 直链下载 VIP 资源 | 参数篡改 × VIP | 8.逻辑漏洞.pdf page_00032 |
| 多设备并发签约领新人优惠 | 并发 × VIP | 逻辑漏洞checklist.pdf page_00012 |
| Cookie 替换实现水平/垂直越权 | 替换身份 × 隐私 | Java Web 鉴权审计经典案例 |
| `..//` 路径截断绕过 Filter | 流程跳跃 × 登录 | Java Filter 路径规范化缺陷 |
| `;` 分号截断 Servlet 路径绕过 Filter | 流程跳跃 × 登录 | Servlet getRequestURI 不规范化 |
| 数据库 `sys_menu` 重复行致权限残留 | 替换身份 × VIP | RBAC 数据一致性审计 |
| `host_scan.py` Host 碰撞 | 业务建模 × 第三方 | 内部资产发现实战 |

---

## 5. Phase 4: Checklist-Driven Testing / 按 CHECKLIST 复测

实际操作阶段直接打开 [CHECKLIST.md](./CHECKLIST.md)，**按业务模块从上到下走**，每条都问：

1. **为什么会出问题**：这条 checklist 项的根因（前端校验、状态机缺失、并发竞态、签名缺失……）
2. **如何复测**：精确到 Burp 步骤 / payload / 期望响应，能让另一个人按描述复现

CHECKLIST.md 里每条都已经按这两栏给出，本节不再重复。

### 5.1 工具栈最佳实践

| 工作 | 推荐工具 | 备注 |
|------|---------|------|
| 抓包/重放 | Burp Suite Repeater | 必备 |
| 并发竞态 | Burp Turbo Intruder / Repeater Group "Send in parallel" | 课件多次出现 |
| 撞库爆破 | Burp Intruder + 字典 | 4 位纯数字验证码 5 分钟可枚举完毕 |
| Host 碰撞 | `python3 host_scan.py -d target.com -t 100` | github.com/AlphaSec/host_scan |
| OCR 验证码 | tesseract | 弱验证码可识别 |
| Java 代码审计 | IntelliJ IDEA + Find in Path 搜 `Filter`、`@PreAuthorize`、`request.getRequestURI` | Spring/Servlet 鉴权代码审计标配 |
| 数据库一致性核对 | Navicat | 验证"前端删了、库里还在" |
| 自动化签名爆破 | 本地 PHP/Python `for $i = 0..32768: md5($i)` | 弱随机数 token 直接全量 |

---

## 6. Phase 5: Human Judgement / 人脑判断要点

> 看到 200 不等于打到漏洞。

### 6.1 三个关键追问

每个疑似漏洞触发后，**先问自己这三件事再写报告**：

1. **业务真的占了便宜吗？**
   - 例：响应 `success:true` 但实际个人中心没看到订单 → 不算
   - 例：抢到了优惠券但提交订单时被服务端二次校验拒绝 → 不算
   - 必须看到"会员时长真增加了/钱真扣了别人的/数据真返了不该返的"才算实证

2. **是技术成功还是业务成功？**
   - 0.019 元充值实际只扣 0.01 元但钱包真增加 0.02 元 → 业务成功（8.逻辑漏洞.pdf page_00013）
   - 抓到 flv 直链但带 token/Referer 校验导致 VLC 播不开 → 仅技术成功，要继续测试看能否绕过签名

3. **可复现吗？**
   - 并发漏洞：复测 5 次至少 3 次成功才算稳
   - 整数溢出：注意可能引发服务端崩溃，**必须有书面授权**才打（8.逻辑漏洞.pdf page_00008）

### 6.2 别误把功能当成漏洞

业务逻辑漏洞实践中反复强调一个判定标准：

> "判断是否为有效漏洞的标准是：最终获得的权益价值 > 正常单独购买所需成本。"

例：发现可以"先支付 1 元试用，再补 99 元升级 VIP" → 这本来就是产品设计的会员升级路径，不是漏洞。
真的漏洞是："多端并发补差价 → 多次到账 VIP 时长" → 同一笔补差价获得了 N 倍权益。

### 6.3 法律边界

8.逻辑漏洞.pdf page_00001 / page_00002 把法律红线讲得非常具体，**任何团队培训时都应该带着读一遍**：

- 《网络安全法》第 22 条：禁止任何形式的网络入侵、干扰、窃密以及工具/方法提供
- 《网络安全法》第 38 条：严禁非法获取、出售或提供公民个人信息
- 《刑法》第 285 条 + 修正案 7：非法控制或获数情节严重可判七年；提供入侵工具/明知他人犯罪仍提供支持，按共犯论处

实操底线：

- 必须有书面授权（且明确目标系统范围、时间窗口、可触及的数据类型）
- 不接触/不存储/不传输任何真实公民个人信息（实名认证类漏洞复测严禁用真人身份证）
- 测试工具不应为"专为入侵而设计"（同时审查脚本来源、用途说明、是否被列为恶意软件）
- 触发整数溢出/资源耗尽类漏洞前需取得二次确认，避免造成服务停摆

参见 8.逻辑漏洞.pdf page_00001 的合规 checklist：

```
□ 测试活动未涉及未经授权的系统访问或数据提取  
  why: 违反网安法第 22 条及刑法 285 条可能构成犯罪
  verify: 检查授权书范围 / 目标系统清单 / 数据采集日志

□ 所用测试工具非专为入侵或非法控制设计
  why: 刑法 285 条修正案禁止提供专门用于侵入的工具
  verify: 审查工具来源 / 用途说明 / 是否被公开列为恶意软件

□ 不接触、存储或传输任何公民个人信息
  why: 网安法第 38 条禁止非法获取或提供个人信息
  verify: 审计测试数据样本 / 数据库查询记录 / 导出文件内容
```

---

## 7. Quick Reference: A One-Page Decision Tree / 单页决策树

```text
看到一个新接口 / 新业务流，按以下顺序问问题：

Q1  这一步的关键状态字段（金额、用户身份、订单状态、权限）有没有走前端？
    └── 有 → 改一下试试，参数篡改 P0
        └── 没成功 → Q2
        └── 成功了 → 看 6.1 三个追问，写报告

Q2  能不能跳过这一步直接到下一步？
    └── 后端有没有 Filter / 中间件做全局校验？
        └── 没有 → 直接打下一个 URL（Java Web 仅靠 Servlet 跳页面而无 Filter 的经典缺陷）
        └── 有 → 试试 ../ 截断（Filter 路径规范化缺陷）或 ; 分号截断（getRequestURI 不规范化）
        └── 都不行 → Q3

Q3  能不能并发突破限制？
    └── 限频接口 → Burp Turbo Intruder concurrentConnections=30
    └── 限单接口 → 双花、双领券、双签约
        └── 都不行 → Q4

Q4  能不能改身份？
    └── Cookie 替换：低权 cookie 拿不到，admin cookie 拿到 → 越权
    └── user_id 替换：改请求里的 uid 拿到他人数据 → 水平越权
    └── 替换收款方：改 username 让他人扣款 → 用户替换攻击 (checklist.pdf page_00017)
        └── 都不行 → Q5

Q5  能不能重放？
    └── 支付回调签名校验？没有 → 重放成功 → 重复发货
    └── 找回密码 token 一次性？没有 → 重复改密
        └── 都不行 → 转 Q6

Q6  能不能改响应？
    └── 客户端是不是只看 success/error_code 字段？
    └── Burp 改响应 success:false → success:true 看前端会不会放行
        └── 都不行 → 这个接口暂判可信，去打下一个

Q7  实在没思路 → 回 PHASE 1，对着没读懂的业务再读一遍代码
    └── Find in Path: Filter / @PreAuthorize / request.getRequestURI / sessionAttribute
    └── 找数据库一致性问题（如 sys_menu 重复行导致权限取消未生效）
```

---

## 8. References / 参考索引

- [SKILL.md](./SKILL.md) — 业务逻辑漏洞 attack playbook（英文）
- [SCENARIOS.md](./SCENARIOS.md) — 支付精度/竞态/验证码/找回密码/枚举的细化 scenario（英文）
- [CHECKLIST.md](./CHECKLIST.md) — 按业务模块分组的可勾选检查项（双语）
- 蒸馏原始素材（外部，未入库）：
  - 业务逻辑漏洞审计课程录像
  - 业务逻辑漏洞 PDF 教材与 checklist 教材
  - 蒸馏管道脚本：托管在 yaklang-ai-training-materials 仓库的 `scripts/` 目录下，包含 `distill-vuln-videos-by-omni.yak` / `distill-vuln-pdfs-by-omni.yak` / `aggregate-vuln-dumps.yak` / `distill-business-logic-vuln-videos.yak` / `aggregate-distill-topics.yak` / `clean-distill-buckets-by-llm.yak`。
