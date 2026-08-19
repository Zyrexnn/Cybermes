---
name: classical-cipher-analysis
description: >-
  Classical cipher analysis playbook. Use when encountering substitution
  ciphers, Vigenere, transposition, XOR, or encoded text in CTF challenges
  that requires frequency analysis, Kasiski examination, or known-plaintext
  cryptanalysis.
---

# SKILL: Classical Cipher Analysis — Expert Cryptanalysis Playbook

> **AI LOAD INSTRUCTION**: Expert classical cipher identification and breaking techniques for CTF. Covers cipher identification methodology (frequency analysis, IC, Kasiski), monoalphabetic substitution, Caesar/ROT, Vigenere, Enigma, affine, Hill, transposition ciphers, Bacon/Polybius/Playfair, and XOR ciphers. Base models often skip the identification step and jump to the wrong cipher type, or fail to recognize encoded (base64/hex) ciphertext that needs decoding before analysis.

## 0. RELATED ROUTING

- [symmetric-cipher-attacks](../symmetric-cipher-attacks/SKILL.md) when dealing with modern symmetric ciphers (AES/DES) rather than classical
- [hash-attack-techniques](../hash-attack-techniques/SKILL.md) when the challenge involves hash-based constructions
- [lattice-crypto-attacks](../lattice-crypto-attacks/SKILL.md) when knapsack-based ciphers are encountered

### Quick identification guide

| Observation | Likely Cipher | First Action |
|---|---|---|
| All uppercase letters, uneven frequency | Monoalphabetic substitution | Frequency analysis |
| All uppercase, flat frequency distribution | Polyalphabetic (Vigenere) | IC + Kasiski |
| Only A-Z shifted uniformly | Caesar/ROT | Brute force 25 shifts |
| Base64 alphabet (A-Za-z0-9+/=) | Base64 encoded (decode first) | Base64 decode |
| Hex string (0-9a-f) | Hex encoded (decode first) | Hex decode |
| Binary (0s and 1s) | Binary encoded | Convert to ASCII |
| Dots and dashes | Morse code | Morse decode |
| Raised/normal text pattern | Bacon cipher | Map to A/B, decode |
| 2-digit number pairs (11-55) | Polybius square | Grid lookup |
| Text appears scrambled (right letters, wrong order) | Transposition | Anagram analysis |
| Non-printable bytes XOR-like | XOR cipher | Single/repeating key XOR analysis |

---

## 1. CIPHER IDENTIFICATION METHODOLOGY

### 1.1 Step 1: Character Set Analysis

```python
def analyze_charset(ciphertext):
    """Identify encoding/cipher by character set."""
    chars = set(ciphertext.strip())

    if chars <= set('01 \n'):
        return "Binary encoding"
    if chars <= set('.-/ \n'):
        return "Morse code"
    if chars <= set('0123456789abcdef \n'):
        return "Hex encoding"
    if chars <= set('ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=\n'):
        if '=' in ciphertext or len(ciphertext) % 4 == 0:
            return "Base64 encoding"
    if chars <= set('ABCDEFGHIJKLMNOPQRSTUVWXYZ \n'):
        return "Uppercase only — classical cipher"
    if all(c in '12345' for c in ciphertext.replace(' ', '').replace('\n', '')):
        return "Polybius square (digits 1-5)"

    return "Mixed charset — needs further analysis"
```

### 1.2 Step 2: Frequency Analysis

```python
from collections import Counter

def frequency_analysis(text):
    """Compute letter frequency distribution."""
    text = text.upper()
    letters = [c for c in text if c.isalpha()]
    total = len(letters)
    freq = Counter(letters)

    print("Letter frequencies:")
    for letter, count in freq.most_common():
        pct = count / total * 100
        bar = '#' * int(pct)
        print(f"  {letter}: {pct:5.1f}% {bar}")

    return freq

# English letter frequency (for comparison):
# E T A O I N S H R D L C U M W F G Y P B V K J X Q Z
# 12.7 9.1 8.2 7.5 7.0 6.7 6.3 6.1 6.0 4.3 4.0 2.8 ...
```

### 1.3 Step 3: Index of Coincidence (IC)

```python
def index_of_coincidence(text):
    """
    IC ≈ 0.065 → English / monoalphabetic substitution
    IC ≈ 0.038 → random / polyalphabetic cipher
    """
    text = [c for c in text.upper() if c.isalpha()]
    N = len(text)
    freq = Counter(text)

    ic = sum(f * (f - 1) for f in freq.values()) / (N * (N - 1))
    return ic

# Interpretation:
# IC > 0.060 → monoalphabetic (Caesar, simple substitution, Playfair)
# IC ≈ 0.045-0.055 → polyalphabetic with short key (Vigenere key < 10)
# IC ≈ 0.038-0.042 → polyalphabetic with long key or random
```

### 1.4 Step 4: Kasiski Examination (for Polyalphabetic)

```python
from math import gcd
from functools import reduce

def kasiski(ciphertext, min_len=3):
    """Find repeated sequences and their distances → key length."""
    text = ''.join(c for c in ciphertext.upper() if c.isalpha())
    distances = []

    for length in range(min_len, min(20, len(text) // 3)):
        for i in range(len(text) - length):
            seq = text[i:i+length]
            j = text.find(seq, i + 1)
            while j != -1:
                distances.append(j - i)
                j = text.find(seq, j + 1)

    if not distances:
        return None

    # Key length is likely GCD of common distances
    common_gcds = Counter()
    for d in distances:
        for factor in range(2, min(d + 1, 30)):
            if d % factor == 0:
                common_gcds[factor] += 1

    print("Likely key lengths (by frequency):")
    for length, count in common_gcds.most_common(5):
        print(f"  Key length {length}: {count} occurrences")

    return common_gcds.most_common(1)[0][0]
```

---

## 2. MONOALPHABETIC SUBSTITUTION

### 2.1 Frequency Analysis Attack

```python
def solve_substitution(ciphertext, interactive=False):
    """Solve monoalphabetic substitution via frequency analysis."""
    freq = frequency_analysis(ciphertext)

    # English frequency order
    eng_order = "ETAOINSRHLDCUMWFGYPBVKJXQZ"
    cipher_order = ''.join(c for c, _ in freq.most_common())

    # Initial mapping (frequency-based guess)
    mapping = {}
    for i, c in enumerate(cipher_order):
        if i < len(eng_order):
            mapping[c] = eng_order[i]

    # Apply mapping
    result = ""
    for c in ciphertext.upper():
        result += mapping.get(c, c)

    return result, mapping

# Better approach: use automated solvers
# quipqiup.com — online substitution solver
# dcode.fr/monoalphabetic-substitution — with word pattern matching
```

### 2.2 Known Plaintext (Crib Dragging)

If part of the plaintext is known (e.g., "flag{" prefix):

```python
def crib_drag_substitution(ciphertext, known_plain, known_cipher):
    """Build partial mapping from known plaintext-ciphertext pair."""
    mapping = {}
    for p, c in zip(known_plain.upper(), known_cipher.upper()):
        mapping[c] = p

    # Apply partial mapping
    result = ""
    for c in ciphertext.upper():
        result += mapping.get(c, '?')

    return result, mapping
```

---

## 3. CAESAR / ROT CIPHERS

### 3.1 Brute Force

```python
def caesar_bruteforce(ciphertext):
    """Try all 25 shifts, score by English frequency."""
    results = []
    for shift in range(26):
        decrypted = ""
        for c in ciphertext:
            if c.isalpha():
                base = ord('A') if c.isupper() else ord('a')
                decrypted += chr((ord(c) - base - shift) % 26 + base)
            else:
                decrypted += c

        # Chi-squared scoring against English frequency
        score = chi_squared_score(decrypted)
        results.append((shift, score, decrypted))

    results.sort(key=lambda x: x[1])
    return results[0]  # best match

def chi_squared_score(text):
    """Lower score = closer to English."""
    expected = {
        'E': 12.7, 'T': 9.1, 'A': 8.2, 'O': 7.5, 'I': 7.0,
        'N': 6.7, 'S': 6.3, 'H': 6.1, 'R': 6.0, 'D': 4.3,
        'L': 4.0, 'C': 2.8, 'U': 2.8, 'M': 2.4, 'W': 2.4,
        'F': 2.2, 'G': 2.0, 'Y': 2.0, 'P': 1.9, 'B': 1.5,
        'V': 1.0, 'K': 0.8, 'J': 0.2, 'X': 0.2, 'Q': 0.1, 'Z': 0.1,
    }
    text = text.upper()
    letters = [c for c in text if c.isalpha()]
    total = len(letters)
    if total == 0:
        return float('inf')

    freq = Counter(letters)
    score = sum(
        (freq.get(c, 0) / total * 100 - expected.get(c, 0)) ** 2 / max(expected.get(c, 0.1), 0.1)
        for c in 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'
    )
    return score
```

### 3.2 ROT13 and ROT47

```python
import codecs

# ROT13 (letters only)
rot13 = codecs.decode(ciphertext, 'rot_13')

# ROT47 (ASCII 33-126)
def rot47(text):
    return ''.join(
        chr(33 + (ord(c) - 33 + 47) % 94) if 33 <= ord(c) <= 126 else c
        for c in text
    )
```

---

## 4. VIGENERE CIPHER

### 4.1 Full Attack Workflow

```
Step 1: Confirm polyalphabetic (IC ≈ 0.04-0.05)
Step 2: Find key length (Kasiski + IC per period)
Step 3: For each key position, solve as single Caesar cipher
Step 4: Assemble key → decrypt
```

### 4.2 IC-Based Key Length Detection

```python
def find_vigenere_key_length(ciphertext, max_key=20):
    """Use IC to find Vigenere key length."""
    text = [c for c in ciphertext.upper() if c.isalpha()]
    results = []

    for kl in range(1, max_key + 1):
        # Split text into kl columns
        columns = [[] for _ in range(kl)]
        for i, c in enumerate(text):
            columns[i % kl].append(c)

        # Average IC across columns
        avg_ic = sum(
            index_of_coincidence(''.join(col)) for col in columns
        ) / kl

        results.append((kl, avg_ic))
        print(f"  Key length {kl:2d}: IC = {avg_ic:.4f}")

    # Key length with IC closest to 0.065
    best = max(results, key=lambda x: x[1])
    return best[0]
```

### 4.3 Per-Position Frequency Attack

```python
def crack_vigenere(ciphertext, key_length):
    """Crack Vigenere given known key length."""
    text = [c for c in ciphertext.upper() if c.isalpha()]
    key = ""

    for pos in range(key_length):
        column = ''.join(text[i] for i in range(pos, len(text), key_length))
        # Solve as Caesar cipher
        shift, score, _ = caesar_bruteforce(column)
        key += chr(shift + ord('A'))

    # Decrypt
    plaintext = ""
    ki = 0
    for c in ciphertext:
        if c.isalpha():
            shift = ord(key[ki % key_length]) - ord('A')
            base = ord('A') if c.isupper() else ord('a')
            plaintext += chr((ord(c) - base - shift) % 26 + base)
            ki += 1
        else:
            plaintext += c

    return key, plaintext
```

---

## 5. AFFINE CIPHER

### 5.1 Definition

`E(x) = (a·x + b) mod 26` where gcd(a, 26) = 1.

Valid a values: 1, 3, 5, 7, 9, 11, 15, 17, 19, 21, 23, 25 (12 values).

### 5.2 Brute Force (312 combinations)

```python
def crack_affine(ciphertext):
    """Brute force affine cipher: 12 × 26 = 312 combinations."""
    valid_a = [a for a in range(1, 26) if gcd(a, 26) == 1]

    for a in valid_a:
        a_inv = pow(a, -1, 26)
        for b in range(26):
            plaintext = ""
            for c in ciphertext.upper():
                if c.isalpha():
                    y = ord(c) - ord('A')
                    x = (a_inv * (y - b)) % 26
                    plaintext += chr(x + ord('A'))
                else:
                    plaintext += c

            score = chi_squared_score(plaintext)
            if score < 50:  # reasonable English
                print(f"a={a}, b={b}: {plaintext[:50]}...")
```

### 5.3 Known Plaintext

```python
def affine_from_known(plain1, cipher1, plain2, cipher2):
    """Recover (a, b) from two known plaintext-ciphertext pairs."""
    p1, c1 = ord(plain1) - ord('A'), ord(cipher1) - ord('A')
    p2, c2 = ord(plain2) - ord('A'), ord(cipher2) - ord('A')

    # c1 = a*p1 + b, c2 = a*p2 + b
    # c1 - c2 = a*(p1 - p2) mod 26
    diff_p = (p1 - p2) % 26
    diff_c = (c1 - c2) % 26

    if gcd(diff_p, 26) != 1:
        return None

    a = (diff_c * pow(diff_p, -1, 26)) % 26
    b = (c1 - a * p1) % 26
    return a, b
```

---

## 6. HILL CIPHER

Matrix-based cipher: `C = K · P mod 26` where K is an n×n key matrix.

### 6.1 Known-Plaintext Attack

```python
import numpy as np

def crack_hill(known_plain, known_cipher, n=2):
    """Recover Hill cipher key from known plaintext-ciphertext (mod 26)."""
    # Convert to numbers
    P = [ord(c) - ord('A') for c in known_plain.upper()]
    C = [ord(c) - ord('A') for c in known_cipher.upper()]

    # Build matrices (need at least n pairs of n-grams)
    P_matrix = np.array(P[:n*n]).reshape(n, n).T
    C_matrix = np.array(C[:n*n]).reshape(n, n).T

    # K = C · P⁻¹ mod 26
    # Need modular matrix inverse
    from sympy import Matrix
    P_mat = Matrix(P_matrix.tolist())
    C_mat = Matrix(C_matrix.tolist())

    P_inv = P_mat.inv_mod(26)
    K = (C_mat * P_inv) % 26

    return K
```

---

## 7. TRANSPOSITION CIPHERS

### 7.1 Rail Fence

```python
def rail_fence_decrypt(ciphertext, rails):
    """Decrypt rail fence cipher."""
    n = len(ciphertext)
    # Build the zigzag pattern
    pattern = []
    for i in range(n):
        row = 0
        cycle = 2 * (rails - 1)
        pos = i % cycle
        row = pos if pos < rails else cycle - pos
        pattern.append((row, i))

    pattern.sort()

    # Fill in characters
    result = [''] * n
    ci = 0
    for _, orig_pos in pattern:
        result[orig_pos] = ciphertext[ci]
        ci += 1

    return ''.join(result)

# Brute force all rail counts
for rails in range(2, 20):
    print(f"Rails {rails}: {rail_fence_decrypt(ct, rails)[:50]}")
```

### 7.2 Columnar Transposition

```python
def columnar_decrypt(ciphertext, key):
    """Decrypt columnar transposition given key word."""
    n_cols = len(key)
    n_rows = -(-len(ciphertext) // n_cols)  # ceiling division

    # Determine column order from key
    order = sorted(range(n_cols), key=lambda i: key[i])

    # Calculate column lengths (some may be shorter)
    full_cols = len(ciphertext) % n_cols
    if full_cols == 0:
        full_cols = n_cols

    # Split ciphertext into columns (in key order)
    columns = [''] * n_cols
    pos = 0
    for col_idx in order:
        col_len = n_rows if col_idx < full_cols else n_rows - 1
        columns[col_idx] = ciphertext[pos:pos + col_len]
        pos += col_len

    # Read off row by row
    plaintext = ''
    for row in range(n_rows):
        for col in range(n_cols):
            if row < len(columns[col]):
                plaintext += columns[col][row]

    return plaintext
```

---

## 8. XOR CIPHER

### 8.1 Single-Byte XOR

See [symmetric-cipher-attacks](../symmetric-cipher-attacks/SKILL.md) Section 4.2 for full implementation.

### 8.2 Multi-Byte XOR (xortool)

```bash
# Automatic key length detection and cracking
xortool ciphertext.bin -l 5        # try key length 5
xortool ciphertext.bin -b          # brute force key length
xortool ciphertext.bin -c 20       # assume most common char is space (0x20)
```

### 8.3 Known Plaintext XOR

```python
def xor_known_plaintext(ciphertext, known_plain, offset=0):
    """Recover XOR key from known plaintext at given offset."""
    key_fragment = bytes(
        c ^ p for c, p in zip(ciphertext[offset:], known_plain)
    )
    print(f"Key fragment: {key_fragment}")

    # If repeating key, infer full key from fragment
    return key_fragment
```

---

## 9. SPECIAL CIPHERS

### 9.1 Bacon Cipher

Binary encoding using two typefaces (A=normal, B=bold/italic).

```python
BACON = {
    'AAAAA': 'A', 'AAAAB': 'B', 'AAABA': 'C', 'AAABB': 'D',
    'AABAA': 'E', 'AABAB': 'F', 'AABBA': 'G', 'AABBB': 'H',
    'ABAAA': 'I', 'ABAAB': 'J', 'ABABA': 'K', 'ABABB': 'L',
    'ABBAA': 'M', 'ABBAB': 'N', 'ABBBA': 'O', 'ABBBB': 'P',
    'BAAAA': 'Q', 'BAAAB': 'R', 'BAABA': 'S', 'BAABB': 'T',
    'BABAA': 'U', 'BABAB': 'V', 'BABBA': 'W', 'BABBB': 'X',
    'BAAAA': 'Y', 'BAAAB': 'Z',
}

def decode_bacon(text):
    """Decode Bacon cipher: uppercase=B, lowercase=A (or similar mapping)."""
    binary = ''.join('B' if c.isupper() else 'A' for c in text if c.isalpha())
    result = ''
    for i in range(0, len(binary) - 4, 5):
        chunk = binary[i:i+5]
        result += BACON.get(chunk, '?')
    return result
```

### 9.2 Polybius Square

```
    1 2 3 4 5
  ┌──────────
1 │ A B C D E
2 │ F G H I/J K
3 │ L M N O P
4 │ Q R S T U
5 │ V W X Y Z

"HELLO" = "23 15 31 31 34"
```

### 9.3 Playfair

5×5 grid cipher encrypting digraphs.

```
Key: "MONARCHY" → grid:
  M O N A R
  C H Y B D
  E F G I/J K
  L P Q S T
  U V W X Z

Rules:
  Same row → shift right: HE → FE → "GF"
  Same col → shift down
  Rectangle → swap columns
```

---

## 10. DECISION TREE

```
Unknown ciphertext — how to identify and break?
│
├─ Step 1: Check encoding
│  ├─ Base64 alphabet with padding? → Decode first, then re-analyze
│  ├─ Hex string? → Convert to bytes, re-analyze
│  ├─ Binary (01)? → Convert to ASCII
│  ├─ Morse (.-/)? → Decode Morse
│  └─ Printable text? → Continue to Step 2
│
├─ Step 2: Character set
│  ├─ Only letters (A-Z)?
│  │  ├─ Compute IC
│  │  │  ├─ IC ≈ 0.065 → Monoalphabetic
│  │  │  │  ├─ Uniform shift in freq? → Caesar → brute force 25
│  │  │  │  ├─ Random-looking mapping? → Simple substitution → frequency analysis
│  │  │  │  └─ Digraph patterns? → Playfair → digraph analysis
│  │  │  │
│  │  │  ├─ IC ≈ 0.04-0.05 → Polyalphabetic
│  │  │  │  ├─ Kasiski → find key length
│  │  │  │  └─ Per-position frequency → crack Vigenere
│  │  │  │
│  │  │  └─ IC ≈ 0.038 → Very long key or one-time pad
│  │  │     └─ Look for key reuse or weak key generation
│  │  │
│  │  └─ Letters appear scrambled (right freq, wrong order)?
│  │     └─ Transposition
│  │        ├─ Rail fence → brute force rail count
│  │        └─ Columnar → try common key lengths
│  │
│  ├─ Numbers (digit pairs)?
│  │  ├─ Pairs in range 11-55 → Polybius square
│  │  └─ Numbers mod 26 → numeric substitution
│  │
│  ├─ Mixed case with pattern?
│  │  └─ Upper/lower encodes binary → Bacon cipher
│  │
│  └─ Non-printable bytes?
│     └─ XOR cipher
│        ├─ Single-byte key → brute force 256
│        ├─ Repeating key → xortool / Hamming distance
│        └─ Known plaintext → direct key recovery
│
└─ Step 3: Apply specific attack
   ├─ Substitution → quipqiup.com / frequency analysis
   ├─ Caesar → dcode.fr / brute force
   ├─ Vigenere → Kasiski + per-column Caesar
   ├─ Affine → brute force 312 combinations
   ├─ Hill → known-plaintext matrix attack
   ├─ Transposition → pattern analysis + brute force
   └─ XOR → xortool / crib dragging
```

---

## 11. TOOLS

| Tool | Purpose | URL/Usage |
|---|---|---|
| **CyberChef** | Universal encoding/cipher Swiss army knife | gchq.github.io/CyberChef |
| **dcode.fr** | 200+ cipher solvers online | dcode.fr |
| **quipqiup** | Automated substitution cipher solver | quipqiup.com |
| **xortool** | XOR cipher analysis and cracking | `pip install xortool` |
| **RsaCtfTool** | RSA + some classical cipher support | GitHub |
| **Ciphey** | Automated cipher detection and decryption | `pip install ciphey` |
| **hashID** | Identify hash types | `pip install hashid` |
| **Python** | Custom frequency analysis and scripting | All attacks above |

### CyberChef Recipes (Common)

```
ROT13:               ROT13
Caesar brute force:   ROT13 (with offset slider)
Base64 decode:        From Base64
Hex decode:           From Hex
XOR:                  XOR (key as hex/utf8)
Vigenere:             Vigenère Decode
Morse:                From Morse Code
```
