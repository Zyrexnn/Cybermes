# RSA Attack Catalog — Detailed Implementations & Mathematics

> **AI LOAD INSTRUCTION**: Load this when you need full mathematical derivations, complete SageMath/Python implementations, and edge-case handling for each RSA attack. Assumes the main [SKILL.md](./SKILL.md) is already loaded for attack selection and decision trees.

---

## 1. FACTORIZATION METHODS — DETAILED

### 1.1 Trial Division

```python
def trial_division(n, limit=10**6):
    """Factor n by trial division up to limit."""
    factors = []
    d = 2
    while d * d <= n and d <= limit:
        while n % d == 0:
            factors.append(d)
            n //= d
        d += 1
    if n > 1:
        factors.append(n)
    return factors
```

### 1.2 Pollard's Rho

```python
from math import gcd
from random import randint

def pollard_rho(n):
    """Pollard's rho factorization."""
    if n % 2 == 0:
        return 2
    x = randint(2, n - 1)
    y = x
    c = randint(1, n - 1)
    d = 1
    while d == 1:
        x = (x * x + c) % n
        y = (y * y + c) % n
        y = (y * y + c) % n
        d = gcd(abs(x - y), n)
    return d if d != n else None
```

### 1.3 Williams' p+1

Works when p+1 is smooth (complement to Pollard's p-1).

```python
def williams_pp1(n, B=10**6):
    """Williams p+1 factorization."""
    from sympy import nextprime
    
    for A in range(3, 20):
        v = A
        p = 2
        while p < B:
            e = 1
            pe = p
            while pe * p <= B:
                pe *= p
                e += 1
            # Lucas chain multiplication
            v = lucas_chain(v, pe, n)
            p = nextprime(p)
        
        g = gcd(v - 2, n)
        if 1 < g < n:
            return g
    return None
```

### 1.4 Quadratic Sieve (Concept)

For n in range 10^30 to 10^100. Use yafu or msieve for implementation.

```
Algorithm outline:
1. Choose factor base of small primes
2. Sieve for B-smooth values of Q(x) = (x + ⌈√n⌉)² - n
3. Collect enough smooth relations (≥ factor base size + 1)
4. Linear algebra over GF(2) to find subset product = perfect square
5. Compute gcd(x² - y², n) to find factor
```

---

## 2. SMALL EXPONENT ATTACKS — DETAILED

### 2.1 Cube Root with Padding Search

When m^e slightly exceeds n (wraps around a few times):

```python
from gmpy2 import iroot, mpz

def cube_root_with_wrap(c, e, n, max_k=10000):
    """Try c + k*n for small k, check if perfect e-th root."""
    for k in range(max_k):
        m, exact = iroot(mpz(c + k * n), e)
        if exact:
            return int(m)
    return None
```

### 2.2 Hastad with Linear Padding

When messages have linear padding: mᵢ = a·m + bᵢ (different padding per recipient).

```python
# SageMath
def hastad_linear_padding(n_list, c_list, e, a_list, b_list):
    """Hastad attack with linear padding: m_i = a_i * m + b_i."""
    assert len(n_list) == e
    
    # Build polynomial for CRT
    N = product(n_list)
    R.<x> = PolynomialRing(Zmod(N))
    
    g = 0
    for i in range(e):
        Ni = N // n_list[i]
        ti = inverse_mod(Ni, n_list[i])
        fi = (a_list[i] * x + b_list[i])^e - c_list[i]
        g += fi * Ni * ti
    
    g = g.monic()
    roots = g.small_roots()
    if roots:
        return int(roots[0])
    return None
```

### 2.3 Coppersmith's Short Pad Attack

Two encryptions of same message with different short random padding (r₁, r₂):

```python
# SageMath
def short_pad_attack(n, e, c1, c2, pad_bits):
    """Recover message when same message has short random padding."""
    R.<x, y> = PolynomialRing(Zmod(n))
    
    # m1 = m * 2^pad_bits + r1, m2 = m * 2^pad_bits + r2
    # Let y = r1 - r2 (difference of padding)
    g1 = x^e - c1
    g2 = (x + y)^e - c2
    
    # Resultant eliminates x, leaving univariate in y
    res = g1.resultant(g2, x)
    
    Ry.<yy> = PolynomialRing(Zmod(n))
    res_uni = Ry(res.polynomial(y))
    roots = res_uni.small_roots(X=2^pad_bits)
    
    if roots:
        diff = int(roots[0])
        # Now solve for x with known y = diff
        # ...
```

---

## 3. WIENER'S ATTACK — COMPLETE IMPLEMENTATION

```python
def wiener_full(e, n):
    """Full Wiener's attack implementation with validation."""
    from gmpy2 import isqrt, is_square, mpz
    
    cf = []
    a, b = e, n
    while b:
        cf.append(a // b)
        a, b = b, a % b
    
    # Generate convergents
    p_prev, p_curr = 0, 1
    q_prev, q_curr = 1, 0
    
    for a_i in cf:
        p_prev, p_curr = p_curr, a_i * p_curr + p_prev
        q_prev, q_curr = q_curr, a_i * q_curr + q_prev
        
        k, d = p_curr, q_curr
        
        if k == 0:
            continue
        
        # Check if (e*d - 1) / k is integer (= phi(n))
        if (e * d - 1) % k != 0:
            continue
        
        phi = (e * d - 1) // k
        
        # p + q = n - phi + 1
        # p * q = n
        s = n - phi + 1
        discriminant = mpz(s * s - 4 * n)
        
        if discriminant < 0:
            continue
        
        if is_square(discriminant):
            sqrt_disc = isqrt(discriminant)
            p = (s + sqrt_disc) // 2
            q = (s - sqrt_disc) // 2
            if p * q == n:
                return d, int(p), int(q)
    
    return None
```

---

## 4. BONEH-DURFEE — SAGE IMPLEMENTATION

```python
# SageMath
def boneh_durfee(n, e, delta=0.292, m=4):
    """
    Boneh-Durfee attack for small d: d < n^delta.
    Based on: e*d = 1 + k*(n+1-p-q) = 1 + k*(n+1-s) where s = p+q.
    Rewrite: 1 + k*(n+1) ≡ k*s (mod e)
    """
    A = Integer((n + 1) // 2)
    P.<x, y> = PolynomialRing(ZZ)
    f = 1 + x * (A + y)
    
    X = Integer(2 * floor(n^delta))
    Y = Integer(floor(n^0.5))
    
    # Build lattice for Coppersmith-type method
    t = m + 1
    shifts = []
    
    for k in range(m + 1):
        for i in range(m - k + 1):
            g = x^i * f^k * e^(m - k)
            shifts.append(g)
    
    for k in range(1, t + 1):
        for i in range(k):
            g = y^i * f^k * e^(m - k)  
            shifts.append(g)
    
    # Construct lattice matrix and apply LLL
    # ... (full lattice construction omitted for brevity)
    # Use standard Coppersmith multivariate implementation
    
    # After LLL: extract small root (x0, y0)
    # k = x0, s = 2*(A + y0)
    # p = (s + sqrt(s^2 - 4n)) / 2
```

---

## 5. LSB ORACLE ATTACK — COMPLETE

```python
from gmpy2 import mpz, invert
from decimal import Decimal, getcontext

def lsb_oracle_full(n, e, c, oracle):
    """
    Full LSB oracle attack.
    oracle(c) → returns LSB of decrypt(c), i.e., m mod 2.
    """
    getcontext().prec = n.bit_length() + 100
    
    lo = Decimal(0)
    hi = Decimal(n)
    
    two_e = pow(2, e, n)
    multiplier = mpz(1)
    
    for i in range(n.bit_length()):
        multiplier = (multiplier * two_e) % n
        test_c = (c * multiplier) % n
        
        lsb = oracle(int(test_c))
        
        mid = (lo + hi) / 2
        if lsb == 0:
            hi = mid
        else:
            lo = mid
    
    return int(hi)
```

---

## 6. BLEICHENBACHER ATTACK — ALGORITHM OUTLINE

```
Input: oracle O(c) = 1 if PKCS-conformant, 0 otherwise
       n, e (public key), c₀ (target ciphertext)

Step 1: Blinding
  Find s₀ such that O(c₀ · s₀ᵉ mod n) = 1
  Set c ← c₀ · s₀ᵉ mod n

Step 2a: Starting search
  Find smallest s₁ ≥ ⌈n/(3B)⌉ such that O(c · s₁ᵉ mod n) = 1
  where B = 2^(8*(k-2)), k = byte length of n

Step 2b: Searching with one interval
  If M = {[a, b]}, find smallest s ≥ 2(b·sᵢ₋₁ - 2B)/n
  such that O(c · sᵉ mod n) = 1

Step 3: Narrowing intervals
  For each (s, [a,b]) in M × {s}:
    r_min = ⌈(a·s - 3B + 1) / n⌉
    r_max = ⌊(b·s - 2B) / n⌋
    For r in [r_min, r_max]:
      Update interval: [max(a, ⌈(2B + r·n)/s⌉), min(b, ⌊(3B - 1 + r·n)/s⌋)]

Step 4: Check
  If M = {[a, a]}: m = a (done!)
  Else: go to Step 2
```

---

## 7. RSA-CRT FAULT — MATHEMATICAL DERIVATION

```
RSA-CRT computation:
  sp = m^d mod p     (computed correctly)
  sq = m^d mod q     (fault injected here → s̃q ≠ sq)
  
  Correct: s = CRT(sp, sq) = sp + p · (((sq - sp) · p⁻¹) mod q)
  Faulty:  s̃ = CRT(sp, s̃q) = sp + p · (((s̃q - sp) · p⁻¹) mod q)

Verification:
  s^e ≡ m (mod p)   ← correct (sp was correct)
  s̃^e ≡ m (mod p)   ← correct (sp was correct)
  
  s^e ≡ m (mod q)   ← correct
  s̃^e ≢ m (mod q)   ← wrong (s̃q was faulty)

Therefore:
  s̃^e - m ≡ 0 (mod p)   but   s̃^e - m ≢ 0 (mod q)
  
  ⟹ gcd(s̃^e - m, n) = p
```

```python
from math import gcd

def rsa_crt_factor(n, e, m, faulty_sig):
    """Factor n from a single faulty CRT signature."""
    diff = (pow(faulty_sig, e, n) - m) % n
    p = gcd(diff, n)
    if 1 < p < n:
        return p, n // p
    return None
```

---

## 8. MULTI-PRIME RSA

When n = p · q · r (three or more primes):

```python
from sympy import factorint
from functools import reduce

def multi_prime_decrypt(n, e, c):
    factors = factorint(n)
    primes = list(factors.keys())
    
    # Euler's totient for multi-prime
    phi = reduce(lambda a, p: a * (p - 1), primes, 1)
    
    d = pow(e, -1, phi)
    m = pow(c, d, n)
    return m
```

---

## 9. KNOWN e·d → FACTOR n

```python
from random import randint
from math import gcd

def factor_from_ed(n, e, d):
    """Factor n given e and d such that e·d ≡ 1 (mod φ(n))."""
    k = e * d - 1
    
    while True:
        g = randint(2, n - 2)
        t = k
        
        while t % 2 == 0:
            t //= 2
            x = pow(g, t, n)
            
            if x > 1 and gcd(x - 1, n) > 1:
                p = gcd(x - 1, n)
                if p != n:
                    return p, n // p
```

---

## 10. ATTACK CONDITIONS SUMMARY

| Attack | Condition | Complexity | Success Rate |
|---|---|---|---|
| Factordb | n is known composite | O(1) lookup | Depends on database |
| Trial division | n has small factor | O(√n) worst case | Good for n < 2^64 |
| Pollard rho | n has factor < n^(1/4) | O(n^(1/4)) | Probabilistic |
| Pollard p-1 | p-1 is B-smooth | O(B log B) | Depends on smoothness |
| Fermat | \|p-q\| < n^(1/4) | O(n^(1/4) / \|p-q\|) | Good when p ≈ q |
| Cube root | m < n^(1/e) | O(1) | Deterministic |
| Hastad | e copies, e different n | O(e log n) | Deterministic |
| Wiener | d < n^(1/4)/3 | O(log n) | Deterministic |
| Boneh-Durfee | d < n^0.292 | Polynomial | High |
| Coppersmith | Small unknown portion | Polynomial | Depends on bounds |
| Common modulus | Same n, gcd(e₁,e₂)=1 | O(log n) | Deterministic |
| Batch GCD | Shared factor among n's | O(n log² n) | Depends on key gen |
| LSB oracle | Parity oracle access | O(log n) queries | Deterministic |
| Bleichenbacher | PKCS#1 padding oracle | O(2^20) queries avg | High |
| CRT fault | Single faulty signature | O(1) | Deterministic |
