---
name: lattice-crypto-attacks
description: >-
  Lattice-based cryptanalysis playbook. Use when attacking RSA via Coppersmith
  small roots, recovering DSA/ECDSA nonces from bias, solving knapsack
  problems, or applying LLL/BKZ reduction to cryptographic constructions.
---

# SKILL: Lattice-Based Cryptanalysis — Expert Attack Playbook

> **AI LOAD INSTRUCTION**: Expert lattice techniques for CTF and cryptanalysis. Covers LLL/BKZ reduction, Coppersmith's method (univariate and multivariate), Hidden Number Problem for DSA/ECDSA nonce recovery, knapsack attacks, and NTRU analysis. Base models often fail to construct the correct attack lattice (wrong dimensions, missing scaling factors) or misapply Coppersmith bounds.

## 0. RELATED ROUTING

- [rsa-attack-techniques](../rsa-attack-techniques/SKILL.md) for RSA-specific attacks that use lattice methods (Coppersmith, Boneh-Durfee)
- [symmetric-cipher-attacks](../symmetric-cipher-attacks/SKILL.md) for LCG state recovery via lattice
- [classical-cipher-analysis](../classical-cipher-analysis/SKILL.md) when lattice methods apply to classical cipher analysis

### Quick application guide

| Problem Type | Lattice Technique | Key Parameter |
|---|---|---|
| RSA small roots | Coppersmith (LLL on polynomial lattice) | Root bound X < N^(1/e) |
| RSA small d | Boneh-Durfee (multivariate Coppersmith) | d < N^0.292 |
| DSA/ECDSA nonce bias | Hidden Number Problem → CVP | Bias bits known |
| Knapsack cipher | Low-density lattice attack | Density < 0.9408 |
| LCG truncated output | CVP on recurrence lattice | Unknown bits per output |
| Subset sum | LLL reduction on knapsack lattice | Element size vs count |
| NTRU key recovery | Lattice reduction on NTRU lattice | Dimension and key size |

---

## 1. LATTICE FUNDAMENTALS

### 1.1 Definitions

A **lattice** L is the set of all integer linear combinations of basis vectors:

```
L = { a₁·b₁ + a₂·b₂ + ... + aₙ·bₙ | aᵢ ∈ ℤ }
```

where b₁, ..., bₙ are linearly independent vectors in ℝᵐ.

**Key problems**:
- **SVP** (Shortest Vector Problem): Find the shortest non-zero vector in L
- **CVP** (Closest Vector Problem): Given target t, find v ∈ L closest to t
- **SVP is NP-hard** in general, but LLL finds an approximately short vector in polynomial time

### 1.2 Lattice Quality Metrics

```
Determinant: det(L) = |det(B)| where B is the basis matrix
Gaussian heuristic: shortest vector ≈ √(n/(2πe)) · det(L)^(1/n)
```

---

## 2. LLL ALGORITHM

### 2.1 What LLL Does

Takes a lattice basis B and produces a **reduced basis** B' where:
- Vectors are nearly orthogonal
- First vector is approximately short (within 2^((n-1)/2) factor of SVP)
- Runs in polynomial time: O(n^5 · d · log³ B) where d = dimension, B = max entry size

### 2.2 SageMath Usage

```python
# SageMath
M = matrix(ZZ, [
    [1, 0, 0, large_value_1],
    [0, 1, 0, large_value_2],
    [0, 0, 1, large_value_3],
    [0, 0, 0, modulus],
])

L = M.LLL()
# Short vectors in L reveal the solution
short_vector = L[0]  # first row is typically shortest
```

### 2.3 Python (fpylll)

```python
from fpylll import IntegerMatrix, LLL

n = 4
A = IntegerMatrix(n, n)
# Fill matrix A...
A[0] = (1, 0, 0, large_value_1)
A[1] = (0, 1, 0, large_value_2)
A[2] = (0, 0, 1, large_value_3)
A[3] = (0, 0, 0, modulus)

LLL.reduction(A)
print(A[0])  # shortest vector
```

---

## 3. BKZ (BLOCK KORKINE-ZOLOTAREV)

### 3.1 Comparison with LLL

| Property | LLL | BKZ-β |
|---|---|---|
| Quality | 2^((n-1)/2) approximation | 2^(n/(β-1)) approximation |
| Speed | Polynomial | Exponential in β |
| Block size | Fixed (2) | Configurable β |
| Best for | Quick reduction | High-quality reduction |

### 3.2 Usage

```python
# SageMath
M = matrix(ZZ, [...])
L = M.BKZ(block_size=20)  # β = 20

# fpylll
from fpylll import BKZ
BKZ.reduction(A, BKZ.Param(block_size=20))
```

Rule of thumb: start with LLL, increase to BKZ if needed. BKZ block size 20-40 is usually sufficient for CTF.

---

## 4. COPPERSMITH'S METHOD

### 4.1 Univariate Case

Given f(x) ≡ 0 (mod N) with small root |x₀| < X, find x₀.

**Bound**: X < N^(1/d) where d = degree of f.

```python
# SageMath — built-in small_roots
N = ...
R.<x> = PolynomialRing(Zmod(N))
f = x^3 + a*x^2 + b*x + c  # known polynomial
roots = f.small_roots(X=2^100, beta=1.0, epsilon=1/30)
```

**Parameters**:
- `X`: upper bound on the root
- `beta`: N = p^beta (beta=1.0 for modular root of N itself; beta=0.5 for root mod unknown factor p ≈ √N)
- `epsilon`: smaller = better results but slower (try 1/30 to 1/100)

### 4.2 Stereotyped Message Attack (RSA)

```python
# SageMath
n, e, c = ...  # RSA parameters
known_msb = ...  # known upper portion of message

R.<x> = PolynomialRing(Zmod(n))
f = (known_msb + x)^e - c

# x represents the unknown lower bits
X = 2^(unknown_bit_count)
roots = f.small_roots(X=X, beta=1.0)
if roots:
    m = known_msb + int(roots[0])
```

### 4.3 Partial Key Exposure (Factor p)

Known MSBs of p: `p = p_known + x` where x is small.

```python
# SageMath
n = ...
p_known = ...  # known upper bits of p

R.<x> = PolynomialRing(Zmod(n))
f = p_known + x
roots = f.small_roots(X=2^unknown_bits, beta=0.5)
# beta=0.5 because p ≈ √n
if roots:
    p = p_known + int(roots[0])
    q = n // p
```

### 4.4 Multivariate Coppersmith (Howgrave-Graham)

For f(x, y) ≡ 0 (mod N):
- No polynomial-time algorithm guaranteed
- Heuristic methods work in practice
- Used in Boneh-Durfee for RSA small d

```python
# SageMath — Boneh-Durfee
# e*d ≡ 1 (mod phi) where phi = (p-1)(q-1)
# Rewrite: e*d = 1 + k*((n+1) - (p+q))
# Let x = k, y = (p+q), both small relative to n

R.<x, y> = PolynomialRing(ZZ)
A = (n + 1) // 2
f = 1 + x * (A + y)  # mod e

# Build shift polynomials and construct lattice
# Apply LLL to find small (x₀, y₀)
```

---

## 5. HIDDEN NUMBER PROBLEM (HNP) — DSA/ECDSA NONCE RECOVERY

### 5.1 Problem Statement

Given: signatures (rᵢ, sᵢ) where nonces kᵢ have known bias (leaked MSBs or LSBs).

DSA equation: `s = k⁻¹(H(m) + xr) mod q`

Rearranged: `k = s⁻¹(H(m) + xr) mod q`

If partial bits of k are known: reduces to CVP on a lattice.

### 5.2 Attack Setup

```python
# SageMath
def ecdsa_nonce_attack(signatures, q, known_bits, bit_position='msb'):
    """
    signatures: list of (r, s, hash, known_nonce_bits)
    q: curve order
    known_bits: number of known bits per nonce
    """
    n = len(signatures)

    # Build lattice
    B = 2^(q.nbits() - known_bits)  # bound on unknown part
    M = matrix(QQ, n + 2, n + 2)

    for i in range(n):
        r_i, s_i, h_i, a_i = signatures[i]
        t_i = Integer(inverse_mod(s_i, q) * r_i % q)
        u_i = Integer(inverse_mod(s_i, q) * h_i % q)

        M[i, i] = q
        M[n, i] = t_i
        M[n+1, i] = u_i - a_i  # a_i = known nonce bits

    M[n, n] = B / q
    M[n+1, n+1] = B

    # LLL reduction
    L = M.LLL()

    # Find row containing the private key x
    for row in L:
        x_candidate = Integer(row[n] * q / B) % q
        # Verify x_candidate against one signature
        if verify_private_key(x_candidate, signatures[0], q):
            return x_candidate

    return None
```

### 5.3 Practical Nonce Bias Sources

| Source | Leaked Bits | Required Signatures |
|---|---|---|
| MSB bias (always 0) | 1 bit | ~100 signatures |
| k generated with wrong length | Variable | ~50 signatures |
| Timing side channel | 1-4 bits | 20-100 signatures |
| Insecure PRNG | Many | Few |
| Reused nonce (k₁ = k₂) | All | 2 signatures |

For **reused nonce** (simplest case):

```python
def ecdsa_reused_nonce(r, s1, s2, h1, h2, q):
    """Recover private key when nonce k is reused."""
    # s1 - s2 = k⁻¹(h1 - h2) mod q  (since r is same)
    k = ((h1 - h2) * inverse_mod(s1 - s2, q)) % q
    x = ((s1 * k - h1) * inverse_mod(r, q)) % q
    return x, k
```

---

## 6. KNAPSACK / SUBSET SUM ATTACKS

### 6.1 Low-Density Attack

Knapsack: given weights a₁,...,aₙ and target S, find x₁,...,xₙ ∈ {0,1} such that Σxᵢaᵢ = S.

**Density** d = n / max(log₂ aᵢ). If d < 0.9408, lattice attack works.

```python
# SageMath
def knapsack_lattice(weights, target):
    """Solve subset sum via LLL lattice attack."""
    n = len(weights)

    # Build lattice (Lagarias-Odlyzko style)
    N = ceil(sqrt(n) / 2)  # scaling factor
    M = matrix(ZZ, n + 1, n + 1)

    for i in range(n):
        M[i, i] = 1
        M[i, n] = N * weights[i]
    M[n, n] = N * target

    # Alternative: CJLOSS embedding
    M2 = matrix(ZZ, n + 1, n + 2)
    for i in range(n):
        M2[i, i] = 1
        M2[i, n + 1] = N * weights[i]
    M2[n, n] = 1
    M2[n, n + 1] = N * (-target)

    L = M2.LLL()

    # Look for short vector with entries in {0, 1, -1}
    for row in L:
        if all(v in (0, 1) for v in row[:n]):
            solution = list(row[:n])
            if sum(solution[i] * weights[i] for i in range(n)) == target:
                return solution

    return None
```

---

## 7. NTRU CRYPTANALYSIS

### 7.1 NTRU Lattice

```python
# SageMath
def ntru_lattice_attack(h, q, N):
    """
    Construct NTRU lattice for key recovery.
    h = public key polynomial (mod q)
    q = modulus
    N = dimension
    """
    # NTRU lattice:
    # | qI  0 |
    # | H   I |
    # where H is the circulant matrix of h

    H = matrix(ZZ, N, N)
    for i in range(N):
        for j in range(N):
            H[i, j] = h[(j - i) % N]

    M = block_matrix([
        [q * identity_matrix(N), zero_matrix(N)],
        [H, identity_matrix(N)]
    ])

    L = M.LLL()

    # Short vector in reduced basis = (f, g) private key
    for row in L:
        f = vector(row[:N])
        g = vector(row[N:])
        if f.norm() < q and g.norm() < q:
            return f, g

    return None
```

---

## 8. CONSTRUCTING ATTACK LATTICES — METHODOLOGY

### 8.1 General Recipe

```
1. Express the cryptographic problem as:
   "Find small x such that f(x) ≡ 0 (mod N)"
   or "Find x close to target t in some lattice L"

2. Choose lattice type:
   ├─ Polynomial lattice → Coppersmith-style
   ├─ Modular lattice → HNP-style CVP
   └─ Knapsack lattice → subset sum / CJLOSS

3. Determine dimensions:
   └─ More dimensions = better approximation but slower

4. Set scaling factors:
   └─ Balance the rows so short vector has roughly equal entries
   └─ Common: multiply by N/X where X is the root bound

5. Apply reduction:
   ├─ LLL first (fast, usually sufficient)
   └─ BKZ if LLL fails (increase block size: 20, 30, 40)

6. Extract solution:
   └─ Check reduced basis rows for valid solutions
```

### 8.2 Embedding Technique (CVP → SVP)

Transform CVP into SVP by embedding the target into the lattice:

```python
# SageMath
def cvp_to_svp(basis_matrix, target, scale=1):
    """Convert CVP to SVP via Kannan's embedding."""
    n = basis_matrix.nrows()
    m = basis_matrix.ncols()

    # Augment matrix
    M = matrix(ZZ, n + 1, m + 1)
    for i in range(n):
        for j in range(m):
            M[i, j] = basis_matrix[i, j]
        M[i, m] = 0

    for j in range(m):
        M[n, j] = target[j]
    M[n, m] = scale  # scaling factor (try 1, then adjust)

    L = M.LLL()

    # Look for row with last entry = ±scale
    for row in L:
        if abs(row[m]) == scale:
            return vector(target) - vector(row[:m]) * (row[m] // abs(row[m]))

    return None
```

### 8.3 Dimension Selection Guide

| Problem | Typical Dimension | Notes |
|---|---|---|
| Coppersmith univariate (degree d) | d × m where m ≈ 1/ε | Larger m = smaller root bound |
| HNP with n signatures | n + 2 | n ≥ known_bits_ratio × q_bits |
| Knapsack with n weights | n + 1 or n + 2 | Depends on density |
| LCG with n outputs | n + 1 | More outputs = easier |
| Boneh-Durfee | (m+1)(m+2)/2 | m = parameter depth |

---

## 9. DECISION TREE

```
Lattice approach needed — which construction?
│
├─ RSA-related?
│  ├─ Small unknown part of message → Coppersmith univariate
│  │  └─ Check: unknown_bits < n_bits / e
│  ├─ Partial factor knowledge → Coppersmith mod p
│  │  └─ Use beta=0.5, X=2^unknown_bits
│  ├─ Small private exponent d → Boneh-Durfee
│  │  └─ Check: d < N^0.292
│  └─ Multiple related equations → multivariate Coppersmith
│
├─ DSA/ECDSA-related?
│  ├─ Reused nonce → direct algebraic recovery (no lattice needed)
│  ├─ Partial nonce leakage → HNP → CVP lattice
│  │  └─ Need enough signatures: n ≥ q_bits / leaked_bits
│  └─ Nonce bias → statistical HNP → larger lattice
│
├─ Knapsack / subset sum?
│  ├─ Low density (d < 0.9408) → CJLOSS lattice attack
│  ├─ High density → lattice attack unlikely to work
│  └─ Super-increasing → greedy algorithm (no lattice needed)
│
├─ LCG / PRNG?
│  ├─ Full outputs known → algebraic recovery (no lattice)
│  ├─ Truncated outputs → CVP on recurrence lattice
│  └─ Unknown modulus → use GCD of output differences
│
├─ NTRU?
│  └─ Build circulant lattice → LLL/BKZ for short key vector
│
└─ Custom problem?
   ├─ Express as "find small root of polynomial mod N" → Coppersmith
   ├─ Express as "find lattice point close to target" → CVP
   ├─ Express as "find short vector in lattice" → SVP / LLL
   └─ If none fit → probably not a lattice problem
```

---

## 10. COMMON PITFALLS

| Pitfall | Symptom | Fix |
|---|---|---|
| Root bound too large | `small_roots()` returns empty | Reduce X, increase epsilon, verify bound satisfies Coppersmith criterion |
| Wrong scaling | LLL finds irrelevant short vector | Scale columns so target vector has balanced entries |
| Insufficient dimension | Solution not in reduced basis | Increase m parameter (more shift polynomials) |
| Wrong beta | Coppersmith doesn't find factor | beta=0.5 for half-size factor, beta=1.0 for full modulus |
| Too few signatures (HNP) | Lattice attack fails | Collect more signatures with nonce bias |
| BKZ block size too small | Solution not short enough | Increase block size (try 25, 30, 40) |
| Integer overflow | SageMath crashes | Use ZZ ring explicitly, avoid mixing QQ and ZZ |
