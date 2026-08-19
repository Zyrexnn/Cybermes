---
name: rsa-attack-techniques
description: >-
  RSA attack playbook for CTF and real-world cryptanalysis. Use when given
  RSA parameters (n, e, c) and need to recover plaintext by exploiting
  weak keys, small exponents, shared factors, or padding oracles.
---

# SKILL: RSA Attack Techniques — Expert Cryptanalysis Playbook

> **AI LOAD INSTRUCTION**: Expert RSA attack techniques for CTF and authorized security assessments. Covers factorization attacks, small exponent exploits, lattice-based approaches (Wiener/Boneh-Durfee/Coppersmith), broadcast attacks, common modulus, padding oracles, and fault attacks. Base models often suggest attacks that don't match the given parameters or miss the correct attack selection based on what's known.

## 0. RELATED ROUTING

- [lattice-crypto-attacks](../lattice-crypto-attacks/SKILL.md) for deep lattice theory behind Coppersmith/Boneh-Durfee
- [hash-attack-techniques](../hash-attack-techniques/SKILL.md) when RSA signature forgery involves hash weaknesses
- [symmetric-cipher-attacks](../symmetric-cipher-attacks/SKILL.md) when RSA protects a symmetric key (hybrid encryption)

### Advanced Reference

Also load [RSA_ATTACK_CATALOG.md](./RSA_ATTACK_CATALOG.md) when you need:
- Detailed SageMath/Python implementation for each attack
- Step-by-step mathematical derivation
- Edge cases and failure conditions per attack

### Quick attack selection

| Given / Observable | Attack | Tool |
|---|---|---|
| Small n (< 512 bits) | Direct factorization | factordb, yafu, msieve |
| e = 3, small message | Cube root | gmpy2.iroot |
| Multiple (n, c) same small e | Hastad broadcast | CRT + iroot |
| Very large e or very small d | Wiener / Boneh-Durfee | SageMath, RsaCtfTool |
| Partial p knowledge | Coppersmith small roots | SageMath |
| Same n, different e | Common modulus | Extended GCD |
| Multiple n values | Batch GCD (shared factor) | Python/SageMath |
| Padding error oracle | Bleichenbacher | Custom script |
| LSB parity oracle | LSB oracle attack | Custom script |
| Fault in CRT computation | RSA-CRT fault | Single faulty signature |

---

## 1. FACTORIZATION ATTACKS

### 1.1 Direct Factorization (Small n)

```python
from sympy import factorint

n = 0x...  # small modulus
factors = factorint(n)
p, q = list(factors.keys())
```

**When**: n < ~512 bits, or known to be in factordb.

### 1.2 Fermat's Factorization

Works when p and q are close together: |p - q| is small.

```python
from gmpy2 import isqrt, is_square

def fermat_factor(n):
    a = isqrt(n) + 1
    while True:
        b2 = a * a - n
        if is_square(b2):
            b = isqrt(b2)
            return (a + b, a - b)
        a += 1
```

### 1.3 Pollard's p-1

Works when p-1 has only small prime factors (B-smooth).

```python
from gmpy2 import gcd

def pollard_p1(n, B=2**20):
    a = 2
    for j in range(2, B):
        a = pow(a, j, n)
    d = gcd(a - 1, n)
    if 1 < d < n:
        return d
    return None
```

### 1.4 Batch GCD (Multiple n share a factor)

```python
from math import gcd
from functools import reduce

def batch_gcd(moduli):
    """Find shared factors among multiple RSA moduli."""
    product = reduce(lambda a, b: a * b, moduli)
    results = {}
    for i, n in enumerate(moduli):
        remainder = product // n
        g = gcd(n, remainder)
        if g != 1 and g != n:
            results[i] = (g, n // g)
    return results
```

---

## 2. SMALL EXPONENT ATTACKS

### 2.1 Cube Root Attack (e = 3, small m)

If m^e < n (no modular reduction occurred), simply take the e-th root.

```python
from gmpy2 import iroot

c = 0x...  # ciphertext
e = 3
m, exact = iroot(c, e)
if exact:
    print(f"Plaintext: {bytes.fromhex(hex(m)[2:])}")
```

### 2.2 Hastad Broadcast Attack

Same message encrypted with same small e under different moduli (n₁, n₂, ..., nₑ).

```python
from sympy.ntheory.modular import crt
from gmpy2 import iroot

# e = 3, three ciphertexts under three different n
n_list = [n1, n2, n3]
c_list = [c1, c2, c3]

# CRT: find x such that x ≡ ci (mod ni) for all i
r, M = crt(n_list, c_list)
m, exact = iroot(r, 3)
assert exact
```

### 2.3 Related Message Attack (Franklin-Reiter)

Two messages related by a known linear function: m₂ = a·m₁ + b. Same n and e.

```python
# SageMath
def franklin_reiter(n, e, c1, c2, a, b):
    R.<x> = PolynomialRing(Zmod(n))
    f1 = x^e - c1
    f2 = (a*x + b)^e - c2
    return Integer(n - gcd(f1, f2).coefficients()[0])
```

---

## 3. LARGE e / SMALL d ATTACKS

### 3.1 Wiener's Attack (Continued Fractions)

When d < n^(1/4) / 3, the continued fraction expansion of e/n reveals d.

```python
def wiener_attack(e, n):
    """Recover d when d is small via continued fractions."""
    cf = continued_fraction(e, n)
    convergents = get_convergents(cf)

    for k, d in convergents:
        if k == 0:
            continue
        phi_candidate = (e * d - 1) // k
        # phi(n) = n - p - q + 1 → p + q = n - phi + 1
        s = n - phi_candidate + 1
        # p, q are roots of x^2 - s*x + n = 0
        discriminant = s * s - 4 * n
        if discriminant >= 0:
            from gmpy2 import isqrt, is_square
            if is_square(discriminant):
                return d
    return None

def continued_fraction(a, b):
    cf = []
    while b:
        cf.append(a // b)
        a, b = b, a % b
    return cf

def get_convergents(cf):
    convergents = []
    h_prev, h_curr = 0, 1
    k_prev, k_curr = 1, 0
    for a in cf:
        h_prev, h_curr = h_curr, a * h_curr + h_prev
        k_prev, k_curr = k_curr, a * k_curr + k_prev
        convergents.append((h_curr, k_curr))
    return convergents
```

### 3.2 Boneh-Durfee Attack (Lattice-Based)

Extends Wiener: works when d < n^0.292. Uses lattice reduction (LLL/BKZ).

**Use SageMath implementation** — see [lattice-crypto-attacks](../lattice-crypto-attacks/SKILL.md) for theory.

---

## 4. COPPERSMITH'S METHOD

### 4.1 Stereotyped Message

Known portion of plaintext, unknown part is small.

```python
# SageMath
n = ...
e = 3
c = ...
known_prefix = b"flag{" + b"\x00" * 27  # known prefix, unknown suffix
known_int = int.from_bytes(known_prefix, 'big')

R.<x> = PolynomialRing(Zmod(n))
f = (known_int + x)^e - c
roots = f.small_roots(X=2^(27*8), beta=1.0)
if roots:
    m = known_int + int(roots[0])
    print(bytes.fromhex(hex(m)[2:]))
```

### 4.2 Partial Key Exposure

Known MSB or LSB of p → recover full p via Coppersmith.

```python
# SageMath — known MSB of p
p_msb = ...  # known upper bits of p
R.<x> = PolynomialRing(Zmod(n))
f = p_msb + x
roots = f.small_roots(X=2^unknown_bits, beta=0.5)
if roots:
    p = p_msb + int(roots[0])
    q = n // p
```

---

## 5. COMMON MODULUS ATTACK

Two ciphertexts of same message under same n but different e₁, e₂ where gcd(e₁, e₂) = 1.

```python
from gmpy2 import gcd, invert

def common_modulus(n, e1, e2, c1, c2):
    """Recover m when same message encrypted with two different e under same n."""
    assert gcd(e1, e2) == 1
    _, s1, s2 = extended_gcd(e1, e2)  # s1*e1 + s2*e2 = 1

    if s1 < 0:
        c1 = invert(c1, n)
        s1 = -s1
    if s2 < 0:
        c2 = invert(c2, n)
        s2 = -s2

    m = (pow(c1, s1, n) * pow(c2, s2, n)) % n
    return m

def extended_gcd(a, b):
    if a == 0:
        return b, 0, 1
    g, x, y = extended_gcd(b % a, a)
    return g, y - (b // a) * x, x
```

---

## 6. ORACLE ATTACKS

### 6.1 LSB Oracle (Parity Oracle)

An oracle reveals whether decrypted message is even or odd.

```python
from gmpy2 import mpz

def lsb_oracle_attack(n, e, c, oracle_func):
    """Decrypt using LSB (parity) oracle. oracle_func(c) returns m%2."""
    from fractions import Fraction
    lo, hi = Fraction(0), Fraction(n)

    for _ in range(n.bit_length()):
        c = (c * pow(2, e, n)) % n  # multiply plaintext by 2
        if oracle_func(c) == 0:
            hi = (lo + hi) / 2
        else:
            lo = (lo + hi) / 2

    return int(hi)
```

### 6.2 Bleichenbacher (PKCS#1 v1.5 Padding Oracle)

Given a padding validity oracle (valid/invalid PKCS#1 v1.5), iteratively narrow down the plaintext range.

**Complexity**: O(2^16) oracle queries per byte on average.

**Target**: TLS implementations returning different errors for valid/invalid padding.

### 6.3 Manger's Attack (PKCS#1 OAEP)

Similar to Bleichenbacher but for OAEP padding. Exploits oracle that distinguishes whether the first byte after unpadding is 0x00.

---

## 7. RSA-CRT FAULT ATTACK

If RSA-CRT signing produces a faulty signature (fault in one CRT half):

```python
def rsa_crt_fault(n, e, correct_sig, faulty_sig, msg):
    """Factor n from one correct and one faulty CRT signature."""
    from math import gcd
    diff = pow(correct_sig, e, n) - pow(faulty_sig, e, n)
    p = gcd(diff % n, n)
    if 1 < p < n:
        q = n // p
        return p, q
    return None

# Even simpler: only faulty signature needed if message is known
def rsa_crt_fault_simple(n, e, faulty_sig, msg):
    p = gcd(pow(faulty_sig, e, n) - msg, n)
    if 1 < p < n:
        return p, n // p
    return None
```

---

## 8. DECISION TREE

```
RSA challenge — what information do you have?
│
├─ Have n and it's small (< 512 bits)?
│  └─ Factor directly: factordb.com → yafu → msieve
│
├─ Have multiple n values?
│  └─ Batch GCD — shared factors?
│     ├─ Yes → factor all that share factors
│     └─ No → analyze each n individually
│
├─ Know e?
│  ├─ e = 3 (or small)?
│  │  ├─ Single ciphertext, small message → cube root
│  │  ├─ Multiple ciphertexts, different n → Hastad broadcast
│  │  ├─ Two related messages → Franklin-Reiter
│  │  └─ Partial plaintext known → Coppersmith
│  │
│  ├─ e is very large?
│  │  └─ d is likely small → Wiener → Boneh-Durfee
│  │
│  └─ Same n, two different e values?
│     └─ Common modulus attack (Bezout coefficients)
│
├─ Know partial factorization info?
│  ├─ Know some bits of p → Coppersmith partial key
│  ├─ p-1 is B-smooth → Pollard p-1
│  └─ p ≈ q (close primes) → Fermat factorization
│
├─ Have an oracle?
│  ├─ Parity oracle (LSB) → LSB oracle attack
│  ├─ Padding validity oracle (PKCS#1 v1.5) → Bleichenbacher
│  └─ OAEP oracle → Manger's attack
│
├─ Have faulty signature?
│  └─ RSA-CRT fault → factor n from faulty sig
│
├─ Know e·d relationship?
│  └─ e·d ≡ 1 mod φ(n) → factor n from (e,d,n)
│
└─ None of the above?
   ├─ Check factordb for known factorization
   ├─ Try Pollard rho for medium-size n
   ├─ Look for implementation flaws (weak PRNG for key generation)
   └─ Consider side-channel if physical access available
```

---

## 9. TOOLS

| Tool | Purpose | Usage |
|---|---|---|
| **RsaCtfTool** | Automated RSA attack suite | `python3 RsaCtfTool.py --publickey pub.pem --uncipherfile flag.enc` |
| **SageMath** | Mathematical computation | Coppersmith, lattice attacks, polynomial arithmetic |
| **factordb.com** | Online factor database | Check if n is already factored |
| **yafu** | Fast factorization (SIQS/GNFS) | `yafu "factor(n)"` |
| **msieve** | GNFS factorization | Large n factorization |
| **gmpy2** | Fast Python integer library | `iroot`, `invert`, `gcd` |
| **pycryptodome** | RSA primitives | Key construction from factors |

### RsaCtfTool Quick Commands

```bash
# From public key
python3 RsaCtfTool.py --publickey pub.pem -n --private

# From parameters
python3 RsaCtfTool.py -n $N -e $E --uncipher $C

# Try all attacks
python3 RsaCtfTool.py --publickey pub.pem --uncipherfile flag.enc --attack all
```

### Decrypt After Factoring

```python
from Crypto.PublicKey import RSA
from gmpy2 import invert

p, q = ...  # factored
n = p * q
e = 65537
phi = (p - 1) * (q - 1)
d = int(invert(e, phi))

c = ...  # ciphertext as integer
m = pow(c, d, n)
plaintext = m.to_bytes((m.bit_length() + 7) // 8, 'big')
print(plaintext)
```
