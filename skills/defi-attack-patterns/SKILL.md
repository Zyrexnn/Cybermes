---
name: defi-attack-patterns
description: >-
  DeFi attack pattern playbook. Use when analyzing flash loan attacks, price oracle manipulation, MEV sandwich attacks, governance exploits, bridge vulnerabilities, and token standard edge cases in decentralized finance protocols.
---

# SKILL: DeFi Attack Patterns — Expert Attack Playbook

> **AI LOAD INSTRUCTION**: Expert DeFi exploitation techniques. Covers flash loan mechanics, oracle manipulation (spot vs TWAP), MEV extraction (sandwich, JIT, liquidation), precision loss attacks, governance exploits, bridge vulnerabilities, and token standard pitfalls. Base models often miss the single-transaction atomicity constraint of flash loans and the distinction between spot price and TWAP manipulation.

## 0. RELATED ROUTING

- [smart-contract-vulnerabilities](../smart-contract-vulnerabilities/SKILL.md) for underlying Solidity vulnerability patterns (reentrancy, integer overflow, delegatecall)
- [deserialization-insecure](../deserialization-insecure/SKILL.md) when targeting off-chain bridge relayer or indexer infrastructure

---

## 1. FLASH LOAN ATTACKS

### 1.1 Mechanism

Flash loans provide uncollateralized borrowing within a single transaction. The entire borrow → use → repay cycle must complete atomically; if repayment fails, the transaction reverts as if nothing happened.

| Provider | Max Amount | Fee |
|---|---|---|
| Aave V3 | Pool liquidity per asset | 0.05% (can be 0 for approved borrowers) |
| dYdX | Pool liquidity | 0 (uses internal balance manipulation) |
| Uniswap V3 | Pool liquidity per pair | 0.3% (swap fee tier) |
| Balancer | Pool liquidity | Protocol-configurable |

### 1.2 Price Oracle Manipulation

```
1. Flash borrow 100,000 WETH
2. Swap 100,000 WETH → TOKEN on AMM_A
   → TOKEN spot price on AMM_A skyrockets
3. On Lending_Protocol (reads AMM_A spot price as oracle):
   → Deposit small TOKEN collateral (valued at inflated price)
   → Borrow large amount of WETH against it
4. Swap TOKEN back → WETH on AMM_A (restore price)
5. Repay flash loan (100,000 WETH + fee)
6. Keep borrowed WETH from Lending_Protocol minus collateral cost
```

**Key insight**: protocols using AMM spot reserves (`getReserves()`) as price oracles are vulnerable. Must use TWAP or external oracle (Chainlink).

### 1.3 Liquidity Pool Drain via Reentrancy

Flash borrow → deposit into pool → trigger reentrancy during callback → withdraw more than deposited → repay loan.

Exploits the combination of flash loan capital with reentrancy in pool accounting logic.

### 1.4 Governance Flash Borrow

```
1. Flash borrow governance tokens
2. Create/vote on malicious proposal (if no snapshot or timelock)
3. Proposal passes instantly
4. Execute proposal (drain treasury, change admin, etc.)
5. Return governance tokens
```

Defense: snapshot-based voting (Compound Governor Bravo), timelocks, minimum proposal period.

---

## 2. PRICE ORACLE MANIPULATION

### 2.1 Spot Price vs TWAP

| Oracle Type | Manipulation Cost | Time Window |
|---|---|---|
| Spot price (`getReserves()`) | Single large swap (flash loanable) | Same transaction |
| TWAP (Time-Weighted Average) | Sustained multi-block manipulation | Multiple blocks (expensive) |
| Chainlink aggregator | Compromise ≥ majority of oracle nodes | Practically infeasible |

### 2.2 AMM Manipulation Flow

```
Normal state: Pool has 1000 ETH + 1,000,000 USDC → price = 1000 USDC/ETH

Attack:
├── Swap 9000 ETH into pool
│   Pool now: 10000 ETH + 100,000 USDC (constant product)
│   Spot price: 10 USDC/ETH (crashed 100x)
├── Dependent contract reads this price
│   → Liquidates positions at wrong price
│   → Or allows cheap borrowing against ETH collateral
├── Swap back: buy ETH with USDC
│   Price restores to ~1000 USDC/ETH
└── Net profit = value extracted from dependent contract - swap slippage - fees
```

### 2.3 Chainlink Oracle Staleness

```solidity
(, int price, , uint updatedAt, ) = priceFeed.latestRoundData();
// Missing checks:
// 1. price > 0
// 2. updatedAt != 0
// 3. block.timestamp - updatedAt < HEARTBEAT
// 4. answeredInRound >= roundId
```

If oracle is stale (network congestion, L2 sequencer down), price can be hours old → arbitrage against stale price.

**L2 Sequencer Risk**: If Arbitrum/Optimism sequencer is down, Chainlink prices freeze. When it comes back, prices jump → mass liquidations at wrong prices.

---

## 3. MEV (MAXIMAL EXTRACTABLE VALUE)

### 3.1 Sandwich Attack

```
Mempool observation: victim submits swap TOKEN_A → TOKEN_B with slippage 1%

Front-run:  Buy TOKEN_B (increase price)
Victim tx:  Swap executes at worse price (within slippage tolerance)
Back-run:   Sell TOKEN_B (profit from price impact)

Profit = victim's price impact - gas costs × 2
```

### 3.2 JIT (Just-In-Time) Liquidity

```
1. Observe large pending swap in mempool
2. Provide concentrated liquidity in the exact price range (Uniswap V3 tick)
3. Victim's swap executes → JIT LP earns majority of fees
4. Remove liquidity immediately after swap
5. Profit = fee earned - gas - impermanent loss (minimal for single block)
```

### 3.3 Liquidation MEV

```
1. Monitor lending protocols for positions approaching liquidation threshold
2. When price oracle updates → position becomes liquidatable
3. Front-run other liquidators → execute liquidation
4. Receive liquidation bonus (typically 5-15% of collateral)
5. Sell collateral for profit
```

### 3.4 MEV Protection Mechanisms

| Mechanism | How It Works |
|---|---|
| Flashbots Protect | Sends tx to private mempool; only block builder sees it |
| MEV Blocker | RPC endpoint that routes through MEV-aware relayers |
| Cow Protocol (batch auction) | Batch matching eliminates ordering advantage |
| Encrypted mempools | Threshold encryption; decrypt only at block build time |
| MEV-Share | User captures portion of MEV extracted from their tx |

---

## 4. PRECISION LOSS EXPLOITATION

### 4.1 Rounding Errors in Token Calculations

Solidity has no floating point. Integer division truncates:

```
shares = depositAmount * totalShares / totalAssets
```

If `totalAssets` is very large relative to `depositAmount * totalShares`, result rounds to 0 → depositor gets no shares but pool keeps the deposit.

### 4.2 First Depositor / Vault Inflation Attack

```
1. Attacker deposits 1 wei → receives 1 share
2. Attacker donates 1,000,000 tokens directly to vault (not via deposit)
3. Vault state: 1,000,001 tokens, 1 share
4. Victim deposits 999,999 tokens:
   shares = 999,999 * 1 / 1,000,001 = 0 (integer truncation)
5. Victim gets 0 shares; attacker owns 100% of vault (now 2,000,000 tokens)
6. Attacker withdraws all
```

**Defenses:**
- Mint dead shares on first deposit (OpenZeppelin ERC4626 offset)
- Require minimum initial deposit
- Internal accounting with virtual offset

### 4.3 Dust Attack via Precision Truncation

Repeated small operations where each truncation loses 1 wei. Accumulate across thousands of operations → material loss.

---

## 5. GOVERNANCE ATTACKS

### 5.1 Flash Loan Governance

Borrow governance tokens → vote → return. Only works if protocol doesn't snapshot balances before voting.

### 5.2 Timelock Bypass

| Vector | Method |
|---|---|
| Timelock set to 0 | Admin can execute proposals instantly |
| `emergencyExecute` function | Bypasses timelock for "emergencies" |
| Guardian/multisig override | Single point of failure |
| Proposal cancellation by attacker | Front-run with cancel if threshold met |

### 5.3 Quorum Manipulation

```
Protocol requires 10% quorum (10M tokens out of 100M supply)
├── Flash borrow 10M governance tokens
├── Create proposal: set admin = attacker
├── Vote with borrowed tokens → meets quorum
├── If no timelock: execute immediately
└── Return tokens
```

---

## 6. BRIDGE EXPLOITS

### 6.1 Common Bridge Attack Vectors

| Vector | Example |
|---|---|
| Signature verification bypass | Ronin Bridge ($624M) — compromised 5/9 validators |
| Message replay | Replay deposit proof on multiple chains |
| Fake deposit proof | Submit proof for non-existent L1 deposit |
| Validator collusion | Compromised majority of bridge validators |
| Smart contract bug | Wormhole ($320M) — uninitialized guardian set |
| Upgradeable proxy exploit | Attacker gains upgrade authority → swap implementation |

### 6.2 Cross-Chain Message Verification

```
Secure pattern:
├── Source chain: emit event with (destination, amount, nonce, chainId)
├── Relayer: submit proof (Merkle proof of event inclusion)
├── Destination chain: verify proof against known source block header
│   ├── Check nonce not replayed
│   ├── Check chainId matches
│   ├── Verify Merkle proof against trusted root
│   └── Mint/release tokens

Vulnerable pattern:
├── Relayer: submit (destination, amount) signed by N-of-M validators
└── If M is small or keys are compromised → forge signatures
```

---

## 7. TOKEN STANDARD EDGE CASES

### 7.1 ERC-20 Approval Front-Running

```
1. Alice approves Bob for 100 tokens
2. Alice wants to change approval to 50 tokens
3. Bob sees the approval change tx in mempool
4. Bob front-runs: transferFrom(Alice, Bob, 100) — uses old approval
5. Alice's approval change executes: approval = 50
6. Bob calls transferFrom(Alice, Bob, 50) — uses new approval
7. Bob extracted 150 tokens instead of 50
```

Defense: `approve(0)` first, then `approve(newAmount)`. Or use `increaseAllowance/decreaseAllowance`.

### 7.2 ERC-777 Reentrancy via Hooks

ERC-777 tokens call `tokensReceived()` hook on the recipient before completing the transfer → classic reentrancy vector.

```
transfer(attacker, amount)
├── _beforeTokenTransfer hook
├── Balance update
├── tokensReceived() callback to recipient  ← reentrancy window
│   └── attacker re-enters: transfer, swap, deposit, etc.
└── _afterTokenTransfer hook
```

### 7.3 Fee-on-Transfer Tokens

Tokens that deduct a fee on each transfer. Protocol receives less than `amount`:

```solidity
// Vulnerable: assumes received == amount
token.transferFrom(msg.sender, address(this), amount);
deposits[msg.sender] += amount; // overcredits by fee amount

// Fixed: measure actual balance change
uint before = token.balanceOf(address(this));
token.transferFrom(msg.sender, address(this), amount);
uint received = token.balanceOf(address(this)) - before;
deposits[msg.sender] += received;
```

### 7.4 Rebasing Tokens

Tokens that automatically adjust balances (e.g., Aave aTokens, stETH). Protocols holding rebasing tokens may have accounting mismatches if they cache balances.

---

## 8. NOTABLE DEFI EXPLOITS REFERENCE

| Exploit | Date | Loss | Primary Vector |
|---|---|---|---|
| Ronin Bridge | Mar 2022 | $624M | Compromised validator keys |
| Wormhole | Feb 2022 | $320M | Signature verification bug |
| Beanstalk | Apr 2022 | $182M | Flash loan governance |
| Mango Markets | Oct 2022 | $114M | Oracle manipulation |
| Euler Finance | Mar 2023 | $197M | Donation attack + liquidation logic |
| Curve (reentrancy) | Jul 2023 | $73M | Vyper compiler reentrancy bug |

---

## 9. DECISION TREE

```
Analyzing a DeFi protocol?
├── Does it use price oracles?
│   ├── Spot price (AMM reserves)? → Flash loan manipulation (Section 1.2)
│   │   └── Can oracle be manipulated in single tx? → HIGH RISK
│   ├── TWAP? → Multi-block manipulation needed → MEDIUM RISK
│   ├── Chainlink? → Check staleness handling (Section 2.3)
│   │   ├── Heartbeat check present? → OK
│   │   └── L2? → Check sequencer uptime oracle
│   └── Multiple oracles with fallback? → Evaluate each
├── Does it accept external tokens?
│   ├── Yes → Check fee-on-transfer handling (Section 7.3)
│   ├── ERC-777 tokens accepted? → Reentrancy via hooks (Section 7.2)
│   └── Rebasing tokens? → Accounting mismatch (Section 7.4)
├── Does it have governance?
│   ├── Yes → Flash loan governance possible? (Section 5.1)
│   │   ├── Snapshot-based voting? → Safer
│   │   └── Live balance voting? → Flash borrow attack
│   ├── Timelock present? → Check for bypass (Section 5.2)
│   └── Quorum threshold vs flash-loanable supply? (Section 5.3)
├── Is it a vault / yield aggregator?
│   ├── Yes → First depositor attack (Section 4.2)
│   │   └── Virtual offset or dead shares? → Mitigated
│   └── Precision loss in share calculation? (Section 4.1)
├── Is it a bridge?
│   ├── Yes → Load bridge vectors (Section 6)
│   │   ├── Validator set size and key management?
│   │   ├── Replay protection (nonce + chainId)?
│   │   └── Upgradeable? → Who holds upgrade key?
│   └── No → Continue
├── User-facing swap functionality?
│   ├── Yes → MEV exposure (Section 3)
│   │   ├── Slippage protection enforced?
│   │   └── Private mempool integration?
│   └── No → Continue
└── Load [smart-contract-vulnerabilities](../smart-contract-vulnerabilities/SKILL.md)
    for underlying Solidity-level bugs
```
