# Block Cipher Attacks — Detailed Scripts & Walkthrough

> **AI LOAD INSTRUCTION**: Load this when you need full attack implementations, step-by-step walkthroughs, and edge-case handling for block cipher exploitation. Assumes the main [SKILL.md](./SKILL.md) is already loaded for attack selection and decision trees.

---

## 1. PADDING ORACLE — FULL WALKTHROUGH

### 1.1 PKCS#7 Padding Review

```
Block size: 16 bytes
Data "HELLO" (5 bytes) → padded to 16 bytes:
  48 45 4C 4C 4F 0B 0B 0B 0B 0B 0B 0B 0B 0B 0B 0B
  
Valid padding examples (last byte determines):
  ...01           → 1 byte of padding
  ...02 02        → 2 bytes of padding
  ...03 03 03     → 3 bytes of padding
  ...10 10 10 ... → 16 bytes (full block of padding)

Invalid:
  ...03 03 04     → last byte says 4, but 3rd-from-end ≠ 04
  ...00           → 0x00 is never valid PKCS#7
```

### 1.2 Attack Internals — Byte-by-Byte Decryption

```
Target: decrypt P[15] (last byte of target block)

CBC decryption internals:
  I[15] = AES_DEC(C_target)[15]    (intermediate value, unknown)
  P[15] = I[15] ⊕ C_prev[15]       (plaintext = intermediate ⊕ prev ciphertext)

Attack for padding 0x01 (valid when last plaintext byte = 0x01):
  We send: C'_prev || C_target
  Where C'_prev[15] = guess

  Server computes: P'[15] = I[15] ⊕ guess
  If P'[15] == 0x01, padding is valid!
  
  Therefore: I[15] = guess ⊕ 0x01
  And: P[15] = I[15] ⊕ original_C_prev[15]

Next: decrypt P[14] (need padding = 0x02 0x02):
  Set C'_prev[15] = I[15] ⊕ 0x02   (forces P'[15] = 0x02)
  Brute force C'_prev[14] until P'[14] = 0x02

Continue until all 16 bytes recovered.
```

### 1.3 Handling False Positives

```python
def is_real_padding(oracle, modified_block, target_block, byte_pos, block_size):
    """
    When cracking the last byte, padding 0x02 0x02 can give false positive.
    Verify by flipping the penultimate byte — if still valid, it was 0x01.
    """
    if byte_pos != block_size - 1:
        return True

    check = bytearray(modified_block)
    check[byte_pos - 1] ^= 1  # flip adjacent byte
    return oracle(bytes(check) + target_block)
```

### 1.4 Encryption via Padding Oracle (CBC-R)

A padding oracle can also encrypt arbitrary plaintext without the key:

```python
def padding_oracle_encrypt(plaintext, block_size, oracle):
    """Encrypt arbitrary plaintext using padding oracle (CBC-R technique)."""
    # Pad plaintext
    pad_len = block_size - (len(plaintext) % block_size)
    padded = plaintext + bytes([pad_len] * pad_len)
    pt_blocks = [padded[i:i+block_size] for i in range(0, len(padded), block_size)]

    # Start with random last ciphertext block
    import os
    ct_blocks = [os.urandom(block_size)]

    # Work backwards
    for pt_block in reversed(pt_blocks):
        # Use padding oracle to find intermediate values for ct_blocks[0]
        intermediate = decrypt_block_intermediate(ct_blocks[0], block_size, oracle)
        # Previous CT block = intermediate ⊕ desired plaintext
        prev_ct = bytes(i ^ p for i, p in zip(intermediate, pt_block))
        ct_blocks.insert(0, prev_ct)

    return b"".join(ct_blocks)  # first block is IV
```

---

## 2. CBC BIT FLIPPING — ADVANCED SCENARIOS

### 2.1 Multi-Byte Flip

```python
def cbc_multibyte_flip(ciphertext, block_size, changes):
    """
    changes: list of (absolute_position, old_byte, new_byte)
    All changes must be in the SAME target block.
    """
    ct = bytearray(ciphertext)

    for pos, old, new in changes:
        target_block = pos // block_size
        byte_in_block = pos % block_size
        prev_block_pos = (target_block - 1) * block_size + byte_in_block
        ct[prev_block_pos] ^= old ^ new

    return bytes(ct)

# Example: change ";admin=false;" to ";admin=true;x"
changes = [
    (32 + 7, ord('f'), ord('t')),   # f → t
    (32 + 8, ord('a'), ord('r')),   # a → r
    (32 + 9, ord('l'), ord('u')),   # l → u
    (32 + 10, ord('s'), ord('e')),  # s → e
    (32 + 11, ord('e'), ord(';')),  # e → ;
    (32 + 12, ord(';'), ord('x')),  # ; → x (pad)
]
```

### 2.2 Dealing with Corrupted Block

The previous block gets corrupted. Strategies:
- If corrupted block is IV: server may not validate IV content
- If corrupted block is a "don't care" field: acceptable corruption
- Two-block technique: use padding oracle to fix the corrupted block

---

## 3. ECB BYTE-AT-A-TIME — COMPLETE WALKTHROUGH

### Step-by-Step Example

```
Server: encrypt(user_input || secret)
Block size: 16
Secret: "FLAG{example}"

Round 1: Find block size
  Send "A", "AA", "AAA", ...
  Watch ciphertext length — when it jumps by 16, block size = 16

Round 2: Confirm ECB
  Send "A" * 32 → check for repeated blocks

Round 3: Decrypt byte 0
  Send "A" * 15 (pad to align secret[0] at end of block 0)
  
  Ciphertext block 0 = E("AAAAAAAAAAAAAAA" + secret[0])
  
  Build table:
    E("AAAAAAAAAAAAAAA" + chr(0))  → block_0_0
    E("AAAAAAAAAAAAAAA" + chr(1))  → block_0_1
    ...
    E("AAAAAAAAAAAAAAA" + "F")     → block_0_70  ← matches!
  
  secret[0] = "F"

Round 4: Decrypt byte 1
  Send "A" * 14
  
  Ciphertext block 0 = E("AAAAAAAAAAAAAA" + "F" + secret[1])
  
  Build table with known prefix "AAAAAAAAAAAAAAAF":
    E("AAAAAAAAAAAAAA" + "F" + chr(0))  → ...
    E("AAAAAAAAAAAAAA" + "F" + "L")     → matches!
  
  secret[1] = "L"

Continue until all bytes recovered.
```

### 3.1 Handling Prefix

If server adds unknown prefix: `encrypt(prefix || user_input || secret)`

```python
def find_prefix_length(encrypt_oracle, block_size):
    """Determine length of unknown prefix."""
    # Find which block the prefix ends in
    base = encrypt_oracle(b"")
    for i in range(1, block_size + 1):
        test = encrypt_oracle(b"A" * i)
        # Find first block that differs from base
        for b in range(len(base) // block_size):
            base_block = base[b*block_size:(b+1)*block_size]
            test_block = test[b*block_size:(b+1)*block_size]
            if base_block != test_block:
                # Prefix ends in block b
                # Now find exact byte offset within block
                for j in range(1, block_size + 1):
                    t1 = encrypt_oracle(b"A" * j)
                    t2 = encrypt_oracle(b"B" * j)
                    if (t1[b*block_size:(b+1)*block_size] !=
                        t2[b*block_size:(b+1)*block_size]):
                        continue
                    return b * block_size + (block_size - j)
    return 0
```

---

## 4. LCG STATE RECOVERY

### 4.1 Known Outputs (Full)

```python
def recover_lcg_full(outputs, modulus):
    """Recover LCG parameters a, b from full consecutive outputs.
    LCG: x_{n+1} = a * x_n + b (mod m)
    """
    # From three consecutive outputs x0, x1, x2:
    x0, x1, x2 = outputs[0], outputs[1], outputs[2]

    # x1 = a*x0 + b (mod m)
    # x2 = a*x1 + b (mod m)
    # x2 - x1 = a*(x1 - x0) (mod m)

    a = ((x2 - x1) * pow(x1 - x0, -1, modulus)) % modulus
    b = (x1 - a * x0) % modulus
    return a, b
```

### 4.2 Truncated Outputs (Lattice-Based)

When only upper bits of each output are known:

```python
# SageMath
def recover_lcg_truncated(known_upper_bits, modulus, a, b, unknown_bits):
    """
    Recover full LCG state from truncated outputs.
    Uses CVP on a lattice.
    """
    n = len(known_upper_bits)
    B = 2^unknown_bits  # bound on unknown part

    # Build lattice
    M = matrix(ZZ, n+1, n+1)
    for i in range(n):
        M[i, i] = modulus
    # Last row encodes the recurrence
    for i in range(n):
        M[n, i] = (a^i) % modulus
    M[n, n] = B

    # LLL reduction
    L = M.LLL()
    # Extract short vector → recover unknown bits
```

---

## 5. MERSENNE TWISTER STATE RECOVERY

```python
def untemper(y):
    """Reverse MT19937 tempering to recover internal state."""
    # Undo: y ^= y >> 18
    y ^= y >> 18

    # Undo: y ^= (y << 15) & 0xEFC60000
    y ^= (y << 15) & 0xEFC60000

    # Undo: y ^= (y << 7) & 0x9D2C5680 (need multiple steps)
    y ^= (y <<  7) & 0x9D2C5680
    y ^= (y << 14) & 0x9D2C5680
    y ^= (y << 21) & 0x9D2C5680
    y ^= (y << 28) & 0x9D2C5680

    # Undo: y ^= y >> 11 (need two steps)
    y ^= (y >> 11)
    y ^= (y >> 22)

    return y & 0xFFFFFFFF

def clone_mt(outputs_624):
    """Clone MT19937 state from 624 consecutive outputs."""
    assert len(outputs_624) == 624
    state = [untemper(o) for o in outputs_624]
    # Now state[] is the internal state array
    # Can predict all future outputs
    return state
```

---

## 6. AES KEY SCHEDULE WEAKNESS EXPLOITATION

### 6.1 Related-Key Attack Setup

```python
from Crypto.Cipher import AES

def related_key_test(key1, key2, plaintext):
    """Demonstrate related-key distinguisher."""
    c1 = AES.new(key1, AES.MODE_ECB).encrypt(plaintext)
    c2 = AES.new(key2, AES.MODE_ECB).encrypt(plaintext)

    # In AES-256, related keys with specific XOR differences
    # can produce predictable ciphertext relationships
    return c1, c2
```

---

## 7. GCM NONCE REUSE

### 7.1 Authentication Key Recovery

When AES-GCM nonce is reused with two different messages:

```
Given: (C1, T1, AAD1) and (C2, T2, AAD2) under same (K, nonce)

The authentication tags are:
  T1 = GHASH_H(AAD1, C1) ⊕ E_K(nonce||1)
  T2 = GHASH_H(AAD2, C2) ⊕ E_K(nonce||1)

Therefore:
  T1 ⊕ T2 = GHASH_H(AAD1, C1) ⊕ GHASH_H(AAD2, C2)

GHASH is polynomial evaluation over GF(2^128):
  GHASH_H(A, C) = A_n * H^(n+1) + ... + C_m * H^2 + len_block * H

T1 ⊕ T2 gives a polynomial equation in H (the auth key).
Factor the polynomial over GF(2^128) to find H.
With H known: forge tags for arbitrary messages.
```

```python
# SageMath
def gcm_nonce_reuse_recover_H(aad1, c1, t1, aad2, c2, t2):
    """Recover GHASH authentication key H from nonce reuse."""
    F.<a> = GF(2^128, modulus=x^128 + x^7 + x^2 + x + 1)

    def bytes_to_gf(b):
        return F(Integer(int.from_bytes(b, 'big')).bits())

    # Build GHASH polynomials and find roots of their difference
    # ... (polynomial construction in GF(2^128))
    # Factor to find H
```

---

## 8. PRACTICAL TIPS

| Scenario | Common Mistake | Correct Approach |
|---|---|---|
| Padding oracle timing | Not accounting for network jitter | Use statistical threshold (multiple requests per guess) |
| ECB byte-at-a-time | Wrong block alignment with prefix | Measure prefix length first, add compensating padding |
| CBC bit flip | Forgetting IV is block -1 | If flipping block 0 content, modify IV bytes |
| XOR key reuse | Trying all cribs simultaneously | Start with high-frequency English words, drag incrementally |
| LFSR recovery | Too few output bits | Need at least 2×LFSR_length bits for Berlekamp-Massey |
| MT clone | Non-consecutive outputs | Must be exactly 624 consecutive outputs, no gaps |
