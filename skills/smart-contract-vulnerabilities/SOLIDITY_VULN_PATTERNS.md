# Solidity Vulnerability Patterns — Code Reference

> **Load trigger**: When the agent needs side-by-side vulnerable vs fixed Solidity code patterns, gas-optimization-introduced vulnerabilities, or proxy storage slot collision calculations. Assumes the main [SKILL.md](./SKILL.md) is already loaded for conceptual understanding.

---

## 1. REENTRANCY — VULNERABLE VS FIXED

### 1.1 Classic Reentrancy

**Vulnerable:**
```solidity
function withdraw() public {
    uint bal = balances[msg.sender];
    require(bal > 0);
    (bool sent, ) = msg.sender.call{value: bal}("");
    require(sent, "Failed to send");
    balances[msg.sender] = 0; // state update AFTER external call
}
```

**Fixed (Checks-Effects-Interactions):**
```solidity
function withdraw() public {
    uint bal = balances[msg.sender];
    require(bal > 0);
    balances[msg.sender] = 0; // state update BEFORE external call
    (bool sent, ) = msg.sender.call{value: bal}("");
    require(sent, "Failed to send");
}
```

### 1.2 Cross-Function Reentrancy

**Vulnerable:**
```solidity
function withdraw() public {
    uint bal = balances[msg.sender];
    require(bal > 0);
    (bool sent, ) = msg.sender.call{value: bal}("");
    require(sent);
    balances[msg.sender] = 0;
}

function transfer(address to, uint amount) public {
    require(balances[msg.sender] >= amount);
    balances[msg.sender] -= amount;
    balances[to] += amount;
}
// Attacker: during withdraw callback, call transfer() with stale balance
```

**Fixed:**
```solidity
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";

contract Safe is ReentrancyGuard {
    function withdraw() public nonReentrant { ... }
    function transfer(address to, uint amount) public nonReentrant { ... }
}
```

### 1.3 Read-Only Reentrancy

**Vulnerable third-party contract:**
```solidity
function getPrice() external view returns (uint) {
    return pool.get_virtual_price(); // reads stale state during pool callback
}

function deposit(uint amount) external {
    uint price = getPrice(); // inflated during reentrancy window
    uint shares = amount * 1e18 / price;
    _mint(msg.sender, shares); // mints too few shares (or too many)
}
```

**Attacker contract:**
```solidity
function attack() external {
    pool.remove_liquidity(...); // triggers callback
}

receive() external payable {
    // During callback: pool state is intermediate
    // get_virtual_price() returns inflated value
    vulnerableProtocol.deposit{value: 1 ether}(1 ether);
}
```

---

## 2. INTEGER OVERFLOW / UNDERFLOW

### 2.1 Pre-0.8 Balance Underflow

**Vulnerable (Solidity < 0.8):**
```solidity
pragma solidity ^0.7.0;

function transfer(address to, uint256 amount) public {
    require(balances[msg.sender] - amount >= 0); // always true for uint!
    balances[msg.sender] -= amount; // underflows to ~2^256
    balances[to] += amount;
}
```

**Fixed:**
```solidity
pragma solidity ^0.7.0;
import "@openzeppelin/contracts/math/SafeMath.sol";

function transfer(address to, uint256 amount) public {
    balances[msg.sender] = balances[msg.sender].sub(amount); // reverts on underflow
    balances[to] = balances[to].add(amount);
}
```

### 2.2 Timelock Bypass via Overflow

**Vulnerable:**
```solidity
function increaseLockTime(uint _seconds) public {
    lockTime[msg.sender] += _seconds; // overflow wraps to small value
}

function withdraw() public {
    require(block.timestamp > lockTime[msg.sender]);
    // ...
}
// Attack: increaseLockTime(type(uint256).max - lockTime + 1) → wraps to 0
```

### 2.3 Unsafe Casting (Post-0.8 Risk)

```solidity
function processAmount(uint256 amount) external {
    uint128 truncated = uint128(amount); // 0.8 does NOT revert on downcast!
    // amount = 2^128 + 1 → truncated = 1
    _transfer(msg.sender, truncated);
}
```

**Fixed (Solidity ≥ 0.8.0):**
```solidity
function processAmount(uint256 amount) external {
    require(amount <= type(uint128).max, "overflow");
    uint128 safe = uint128(amount);
    _transfer(msg.sender, safe);
}
```

---

## 3. ACCESS CONTROL

### 3.1 tx.origin Phishing

**Vulnerable:**
```solidity
function transferOwnership(address newOwner) public {
    require(tx.origin == owner); // tx.origin = EOA, not immediate caller
    owner = newOwner;
}
```

**Attack contract:**
```solidity
contract PhishingAttack {
    VulnerableContract target;

    function attack() external {
        // If the owner calls this function (tricked via phishing link),
        // tx.origin == owner → passes the check
        target.transferOwnership(address(this));
    }
}
```

**Fixed:**
```solidity
function transferOwnership(address newOwner) public {
    require(msg.sender == owner); // msg.sender = immediate caller
    owner = newOwner;
}
```

### 3.2 Unprotected Selfdestruct

**Vulnerable:**
```solidity
function destroy() external {
    selfdestruct(payable(msg.sender)); // no access control
}
```

### 3.3 Unprotected Initializer (Proxy Pattern)

**Vulnerable:**
```solidity
function initialize(address _owner) public {
    owner = _owner; // can be called by anyone, multiple times
}
```

**Fixed:**
```solidity
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";

function initialize(address _owner) public initializer {
    owner = _owner;
}
```

---

## 4. DELEGATECALL STORAGE COLLISION

### 4.1 Proxy-Implementation Slot Mismatch

```solidity
// Proxy contract
contract Proxy {
    address public implementation; // slot 0
    address public owner;          // slot 1

    fallback() external payable {
        (bool s, ) = implementation.delegatecall(msg.data);
        require(s);
    }
}

// Implementation contract
contract Implementation {
    uint public someValue;   // slot 0 — COLLIDES with Proxy.implementation!
    address public admin;    // slot 1 — COLLIDES with Proxy.owner!

    function setSomeValue(uint _val) public {
        someValue = _val; // overwrites Proxy.implementation address!
    }
}
```

**Fixed (EIP-1967 storage slots):**
```solidity
contract SafeProxy {
    // Implementation stored at keccak256("eip1967.proxy.implementation") - 1
    bytes32 private constant IMPL_SLOT =
        0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc;

    function _implementation() internal view returns (address impl) {
        assembly { impl := sload(IMPL_SLOT) }
    }
}
```

### 4.2 Function Selector Collision in Transparent Proxy

```
admin()          selector: 0xf851a440
collide_func()   selector: 0xf851a440  ← same 4 bytes by coincidence
```

Tool to check: `cast sig "functionName(argTypes)"` computes selector.

---

## 5. RANDOMNESS MANIPULATION

**Vulnerable:**
```solidity
function roll() external payable {
    uint random = uint(keccak256(abi.encodePacked(
        block.timestamp,
        block.difficulty,
        msg.sender
    ))) % 6;
    if (random == 0) {
        payable(msg.sender).transfer(address(this).balance);
    }
}
```

**Attack contract:**
```solidity
contract AttackRoll {
    function attack(VulnerableRoll target) external payable {
        uint random = uint(keccak256(abi.encodePacked(
            block.timestamp,
            block.difficulty,
            address(this)
        ))) % 6;
        require(random == 0, "not winning, skip");
        target.roll{value: msg.value}();
    }
}
```

---

## 6. SIGNATURE REPLAY

**Vulnerable:**
```solidity
function executeWithSig(address to, uint amount, bytes memory sig) external {
    bytes32 hash = keccak256(abi.encodePacked(to, amount));
    address signer = ECDSA.recover(hash, sig);
    require(signer == owner, "invalid sig");
    // no nonce → same signature can be replayed
    payable(to).transfer(amount);
}
```

**Fixed:**
```solidity
mapping(uint256 => bool) public usedNonces;

function executeWithSig(
    address to, uint amount, uint256 nonce, bytes memory sig
) external {
    require(!usedNonces[nonce], "nonce used");
    bytes32 hash = keccak256(abi.encodePacked(to, amount, nonce, block.chainid, address(this)));
    address signer = ECDSA.recover(hash, sig);
    require(signer == owner, "invalid sig");
    usedNonces[nonce] = true;
    payable(to).transfer(amount);
}
```

---

## 7. SELF-DESTRUCT FORCE-SEND ETH

**Vulnerable (balance-dependent logic):**
```solidity
function isGameComplete() public view returns (bool) {
    return address(this).balance == 10 ether; // exact balance check
}
```

**Attack:**
```solidity
contract ForceEth {
    function attack(address target) external payable {
        selfdestruct(payable(target));
        // forces ETH into target, bypassing receive/fallback
        // target.balance now != 10 ether → game logic broken
    }
}
```

**Fixed:**
```solidity
uint public deposits; // track deposits explicitly, don't rely on balance

function isGameComplete() public view returns (bool) {
    return deposits == 10 ether;
}
```

---

## 8. FLASH LOAN ORACLE MANIPULATION

**Vulnerable price oracle:**
```solidity
function getPrice(address token) public view returns (uint) {
    (uint reserve0, uint reserve1, ) = pair.getReserves();
    return reserve0 * 1e18 / reserve1; // spot price — manipulable in same tx
}
```

**Attack flow:**
```
1. Flash borrow 10,000 ETH
2. Swap ETH → Token on AMM (crashes token spot price)
3. Call lending protocol that uses spot price → borrow token at deflated price
4. Swap token back → ETH (restore price)
5. Repay flash loan + fee
6. Profit = borrowed tokens at deflated price - flash loan fee
```

**Fixed (TWAP oracle):**
```solidity
function getPrice(address token) public view returns (uint) {
    // Use Uniswap V3 TWAP or Chainlink aggregator
    (, int256 price, , uint256 updatedAt, ) = chainlinkFeed.latestRoundData();
    require(block.timestamp - updatedAt < 3600, "stale price");
    return uint256(price);
}
```

---

## 9. UNCHECKED RETURN VALUE

**Vulnerable:**
```solidity
function withdraw(uint amount) external {
    payable(msg.sender).send(amount); // send() returns bool, not checked!
    balances[msg.sender] -= amount;   // balance decremented even if send failed
}
```

**Fixed:**
```solidity
function withdraw(uint amount) external {
    (bool success, ) = payable(msg.sender).call{value: amount}("");
    require(success, "transfer failed");
    balances[msg.sender] -= amount;
}
```

---

## 10. DENIAL OF SERVICE — UNEXPECTED REVERT

**Vulnerable (push pattern):**
```solidity
address[] public recipients;

function distribute() external {
    for (uint i = 0; i < recipients.length; i++) {
        // If one recipient is a contract that reverts, entire distribution fails
        payable(recipients[i]).transfer(1 ether);
    }
}
```

**Fixed (pull pattern):**
```solidity
mapping(address => uint) public pendingWithdrawals;

function distribute() external {
    for (uint i = 0; i < recipients.length; i++) {
        pendingWithdrawals[recipients[i]] += 1 ether;
    }
}

function withdraw() external {
    uint amount = pendingWithdrawals[msg.sender];
    pendingWithdrawals[msg.sender] = 0;
    payable(msg.sender).transfer(amount);
}
```

---

## 11. GAS OPTIMIZATION TRAPS

Optimizations that accidentally introduce vulnerabilities:

| Optimization | Vulnerability Introduced |
|---|---|
| `unchecked{}` loop counter | User-controlled bounds → overflow |
| Assembly `sstore` for gas savings | Bypasses Solidity's overflow checks and visibility |
| Packed storage (`uint128, uint128` in one slot) | Incorrect bit masking → value corruption |
| `immutable` used for mutable config | Cannot update → frozen misconfiguration |
| `selfdestruct` for gas refund | Contract destruction as attack vector (pre-Dencun) |
| Skipping zero-address checks | Gas saved but ownership can be burned |
