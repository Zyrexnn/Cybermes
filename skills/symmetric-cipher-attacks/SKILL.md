---
name: symmetric-cipher-attacks
description: >-
  Symmetric cipher attack playbook. Use when exploiting block cipher mode
  weaknesses (CBC padding oracle, ECB cut-and-paste, bit flipping), stream
  cipher key reuse, or meet-in-the-middle attacks.
---

# SKILL: Symmetric Cipher Attacks — Expert Cryptanalysis Playbook

> **AI LOAD INSTRUCTION**: Expert techniques for attacking symmetric encryption in CTF and authorized testing. Covers CBC padding oracle, CBC bit flipping, ECB detection and exploitation, stream cipher key reuse, LFSR/LCG state recovery, RC4 biases, and meet-in-the-middle attacks. Base models often confuse ECB and CBC attack strategies or fail to set up byte-at-a-time ECB decryption correctly.

## 0. RELATED ROUTING

- [rsa-attack-techniques](../rsa-attack-techniques/SKILL.md) when symmetric key is protected by RSA
- [hash-attack-techniques](../hash-attack-techniques/SKILL.md) when HMAC or hash-based authentication is involved
- [lattice-crypto-attacks](../lattice-crypto-attacks/SKILL.md) for LCG/LFSR state recovery via lattice methods

### Advanced Reference

Also load [BLOCK_CIPHER_ATTACKS.md](./BLOCK_CIPHER_ATTACKS.md) when you need:
- Detailed attack scripts with full Python implementations
- Step-by-step byte-at-a-time ECB walkthrough
- PadBuster usage and custom padding oracle scripts
- LCG/LFSR recovery implementation

### Quick attack selection

| Observable Behavior | Likely Weakness | Attack |
|---|---|---|
| Same plaintext → same ciphertext (block-aligned) | ECB mode | Cut-and-paste / byte-at-a-time |
| Padding error distinguishable | CBC padding oracle | Decrypt without key |
| Can modify ciphertext, affects next block | CBC mode, no integrity check | Bit flipping |
| Key reused with XOR/stream cipher | Two-time pad | XOR ciphertexts together |
| Predictable PRNG output | LCG or LFSR | State recovery |
| Double encryption used | 2DES-like | Meet in the middle |

---

## 1. PADDING ORACLE ATTACK (CBC MODE)

### 1.1 Mechanism

CBC decryption: `P_i = D_K(C_i) ⊕ C_{i-1}`

If the server reveals whether padding is valid (PKCS#7), we can decrypt any block by manipulating the previous ciphertext block.

### 1.2 Attack Steps

```
Target: decrypt block C_i (with unknown plaintext P_i)

For byte position b = 15 down to 0 (last byte first):
  padding_value = 16 - b
  
  For guess = 0x00 to 0xFF:
    Construct modified C'_{i-1}:
      - Bytes 0..b-1: original C_{i-1} bytes
      - Byte b: guess
      - Bytes b+1..15: calculated to produce correct padding
    
    Send (C'_{i-1} || C_i) to oracle
    
    If oracle says "valid padding":
      intermediate_byte[b] = guess ⊕ padding_value
      plaintext_byte[b] = intermediate_byte[b] ⊕ original_C_{i-1}[b]
```

### 1.3 Python Implementation

```python
def padding_oracle_attack(ciphertext, block_size, oracle):
    """
    oracle(ct) returns True if padding is valid, False otherwise.
    ciphertext includes IV as first block.
    """
    blocks = [ciphertext[i:i+block_size] for i in range(0, len(ciphertext), block_size)]
    plaintext = b""

    for block_idx in range(1, len(blocks)):
        prev_block = bytearray(blocks[block_idx - 1])
        curr_block = blocks[block_idx]
        intermediate = [0] * block_size
        decrypted = [0] * block_size

        for byte_pos in range(block_size - 1, -1, -1):
            padding_val = block_size - byte_pos

            for guess in range(256):
                modified = bytearray(block_size)
                modified[byte_pos] = guess

                for j in range(byte_pos + 1, block_size):
                    modified[j] = intermediate[j] ^ padding_val

                test_ct = bytes(modified) + curr_block
                if oracle(test_ct):
                    if byte_pos == block_size - 1:
                        # Verify it's not a false positive (padding 0x02 0x02)
                        check = bytearray(modified)
                        check[byte_pos - 1] ^= 1
                        if not oracle(bytes(check) + curr_block):
                            continue

                    intermediate[byte_pos] = guess ^ padding_val
                    decrypted[byte_pos] = intermediate[byte_pos] ^ prev_block[byte_pos]
                    break

        plaintext += bytes(decrypted)

    return plaintext
```

### 1.4 Tools

```bash
# PadBuster
padbuster http://target/decrypt?ct= CIPHERTEXT_HEX 16 -encoding 0
padbuster http://target/decrypt?ct= CIPHERTEXT_HEX 16 -encoding 0 -plaintext "admin=true"
```

---

## 2. CBC BIT FLIPPING

### 2.1 Concept

Flipping bit at position j in C_{i-1} flips the same bit at position j in P_i (and corrupts all of P_{i-1}).

```
Original:  P_i[j] = D_K(C_i)[j] ⊕ C_{i-1}[j]
Modified:  P'_i[j] = D_K(C_i)[j] ⊕ C'_{i-1}[j]
                    = P_i[j] ⊕ (C_{i-1}[j] ⊕ C'_{i-1}[j])
```

### 2.2 Practical Example

```python
def cbc_bitflip(ciphertext, block_size, target_byte_pos, old_value, new_value):
    """
    Flip byte in plaintext block N+1 by modifying ciphertext block N.
    target_byte_pos: absolute position in plaintext (0-indexed)
    """
    ct = bytearray(ciphertext)
    block_num = target_byte_pos // block_size
    byte_in_block = target_byte_pos % block_size

    # Modify previous block (block_num - 1) to flip target byte
    modify_pos = (block_num - 1) * block_size + byte_in_block

    # XOR to cancel old value and set new value
    ct[modify_pos] ^= old_value ^ new_value
    return bytes(ct)

# Example: flip "admin=0" to "admin=1"
# If "admin=0" is at byte position 22 (block 1, byte 6):
modified_ct = cbc_bitflip(ciphertext, 16, 22, ord('0'), ord('1'))
```

---

## 3. ECB MODE ATTACKS

### 3.1 Detection

```python
def detect_ecb(ciphertext, block_size=16):
    """ECB produces identical blocks for identical plaintext blocks."""
    blocks = [ciphertext[i:i+block_size] for i in range(0, len(ciphertext), block_size)]
    return len(blocks) != len(set(blocks))

# Force detection: send repeated plaintext
test_input = b"A" * 48  # at least 3 blocks of identical data
# If response has repeated blocks → ECB
```

### 3.2 ECB Cut-and-Paste

Reorder ciphertext blocks to create new valid plaintexts.

```
Original blocks:
  Block 0: "email=foo@bar.c"
  Block 1: "om&role=user&uid"
  Block 2: "=10\x0d\x0d\x0d..."

Attack: craft input so "admin" + padding lands in its own block,
then swap it in place of "user" block.

Step 1: Send email that aligns "admin" + PKCS7 to a block:
  email = "foo@bar.coadmin\x0b\x0b\x0b\x0b\x0b\x0b\x0b\x0b\x0b\x0b\x0b"
  → Block 1 encrypts "admin\x0b\x0b..."  (save this block)

Step 2: Send email that puts "role=" at end of block:
  email = "foo@bar.co"
  → Block 2 = "=user&uid=10..."  (but we replace this)

Step 3: Replace last block with saved "admin\x0b..." block
```

### 3.3 Byte-at-a-Time ECB Decryption

Decrypt unknown appended secret one byte at a time.

```python
def ecb_byte_at_a_time(encrypt_oracle, block_size=16):
    """
    encrypt_oracle(input_bytes) = AES_ECB(input || unknown_secret)
    Returns the unknown_secret.
    """
    secret = b""
    secret_len = len(encrypt_oracle(b"")) 

    for i in range(secret_len):
        block_num = i // block_size
        pad_len = block_size - 1 - (i % block_size)
        padding = b"A" * pad_len

        # Build lookup table
        target_ct = encrypt_oracle(padding)
        target_block = target_ct[block_num * block_size:(block_num + 1) * block_size]

        for byte_val in range(256):
            test_input = padding + secret + bytes([byte_val])
            test_ct = encrypt_oracle(test_input)
            test_block = test_ct[block_num * block_size:(block_num + 1) * block_size]

            if test_block == target_block:
                secret += bytes([byte_val])
                break

    return secret
```

---

## 4. STREAM CIPHER ATTACKS

### 4.1 Known Plaintext / Key Reuse (Two-Time Pad)

```python
def two_time_pad(c1, c2, known_crib=None):
    """
    c1 = m1 ⊕ K, c2 = m2 ⊕ K (same key K)
    c1 ⊕ c2 = m1 ⊕ m2 (key cancels)
    """
    xored = bytes(a ^ b for a, b in zip(c1, c2))

    if known_crib:
        results = []
        for offset in range(len(xored) - len(known_crib) + 1):
            candidate = bytes(
                xored[offset + i] ^ known_crib[i] for i in range(len(known_crib))
            )
            if all(0x20 <= b <= 0x7e for b in candidate):
                results.append((offset, candidate))
        return results
    return xored
```

### 4.2 Single-Byte XOR Brute Force

```python
def single_byte_xor_crack(ciphertext):
    """Brute force single-byte XOR key using frequency analysis."""
    english_freq = {
        'e': 12.7, 't': 9.1, 'a': 8.2, 'o': 7.5, 'i': 7.0,
        'n': 6.7, 's': 6.3, 'h': 6.1, 'r': 6.0, 'd': 4.3,
    }
    best_score, best_key, best_plaintext = 0, 0, b""

    for key in range(256):
        plaintext = bytes(b ^ key for b in ciphertext)
        score = sum(
            english_freq.get(chr(b).lower(), 0)
            for b in plaintext if 0x20 <= b <= 0x7e
        )
        if score > best_score:
            best_score = score
            best_key = key
            best_plaintext = plaintext

    return best_key, best_plaintext
```

### 4.3 Repeating-Key XOR (Kasiski-like)

```python
def repeating_xor_crack(ciphertext, max_keylen=40):
    """Crack repeating-key XOR using Hamming distance for key length."""
    def hamming(a, b):
        return sum(bin(x ^ y).count('1') for x, y in zip(a, b))

    # Find key length
    scores = []
    for kl in range(2, max_keylen + 1):
        blocks = [ciphertext[i:i+kl] for i in range(0, len(ciphertext) - kl, kl)]
        if len(blocks) < 4:
            continue
        dist = sum(hamming(blocks[i], blocks[i+1]) for i in range(min(3, len(blocks)-1)))
        normalized = dist / (min(3, len(blocks)-1) * kl)
        scores.append((normalized, kl))

    best_keylen = sorted(scores)[0][1]

    # Crack each position with single-byte XOR
    key = b""
    for i in range(best_keylen):
        column = bytes(ciphertext[j] for j in range(i, len(ciphertext), best_keylen))
        k, _ = single_byte_xor_crack(column)
        key += bytes([k])

    return key
```

### 4.4 LFSR State Recovery (Berlekamp-Massey)

```python
def berlekamp_massey_gf2(output_bits):
    """Recover LFSR feedback polynomial from output sequence over GF(2)."""
    n = len(output_bits)
    C = [0] * (n + 1)
    B = [0] * (n + 1)
    C[0] = B[0] = 1
    L = 0
    m = 1
    b = 1

    for N in range(n):
        d = output_bits[N]
        for i in range(1, L + 1):
            d ^= C[i] & output_bits[N - i]

        if d == 0:
            m += 1
        elif 2 * L <= N:
            T = C[:]
            for i in range(m, n + 1):
                C[i] ^= B[i - m]
            L = N + 1 - L
            B = T
            b = d
            m = 1
        else:
            for i in range(m, n + 1):
                C[i] ^= B[i - m]
            m += 1

    return C[:L + 1], L
```

### 4.5 RC4 Biases

| Bias | Description | Exploitation |
|---|---|---|
| Initial byte bias | P(K[0] = 0) ≈ 2/256 (double normal) | Statistical plaintext recovery for first bytes |
| Fluhrer-Mantin-Shamir | Weak key scheduling with IV | WEP attack (historical) |
| NOMORE attack | Long-term biases in keystream | TLS/RC4 plaintext recovery (2^24-2^26 ciphertexts) |
| Invariance weakness | Key-dependent biases throughout stream | Statistical attack on many encryptions |

---

## 5. MEET-IN-THE-MIDDLE

### 5.1 Double Encryption Attack

```
Double encryption: C = E_K2(E_K1(P))
Brute force: 2^(2n) expected
MITM:        2^(n+1) + storage for 2^n entries

Attack:
1. Encrypt P with all possible K1 → store (E_K1(P), K1) in table
2. Decrypt C with all possible K2 → check if D_K2(C) matches any entry
3. Match found → (K1, K2) recovered
```

```python
from itertools import product

def meet_in_the_middle(encrypt, decrypt, plaintext, ciphertext, keyspace_bits):
    """MITM attack on double encryption."""
    # Phase 1: build encryption table
    enc_table = {}
    for k1 in range(2**keyspace_bits):
        intermediate = encrypt(plaintext, k1)
        enc_table[intermediate] = k1

    # Phase 2: decrypt and look up
    for k2 in range(2**keyspace_bits):
        intermediate = decrypt(ciphertext, k2)
        if intermediate in enc_table:
            k1 = enc_table[intermediate]
            return k1, k2

    return None
```

---

## 6. DECISION TREE

```
Symmetric cipher challenge — what can you observe?
│
├─ Can you detect the mode?
│  ├─ Repeated input → repeated output blocks?
│  │  └─ Yes → ECB mode
│  │     ├─ Can control prefix → byte-at-a-time decryption
│  │     ├─ Can reorder blocks → cut-and-paste
│  │     └─ Can detect block boundaries → block alignment oracle
│  │
│  ├─ Error message differs for bad padding?
│  │  └─ Yes → Padding oracle (CBC)
│  │     └─ PadBuster or custom script
│  │
│  └─ Can modify ciphertext and observe effect?
│     └─ Next-block plaintext changes → CBC bit flipping
│
├─ Stream cipher or XOR?
│  ├─ Key reused on different messages?
│  │  └─ XOR ciphertexts → crib drag
│  │
│  ├─ Known plaintext-ciphertext pair?
│  │  └─ Recover keystream directly
│  │
│  ├─ Single-byte XOR key?
│  │  └─ Brute force 256 keys with frequency analysis
│  │
│  ├─ Repeating-key XOR?
│  │  └─ Hamming distance → key length → per-position crack
│  │
│  └─ LFSR-based?
│     └─ Berlekamp-Massey for state/polynomial recovery
│
├─ PRNG-based cipher?
│  ├─ LCG → truncated output lattice attack
│  ├─ Mersenne Twister → 624 outputs → full state recovery
│  └─ Custom PRNG → analyze period and state size
│
├─ Double / triple encryption?
│  └─ Meet-in-the-middle
│
└─ RC4 specifically?
   ├─ Single encryption → initial byte bias
   ├─ Many encryptions same key → statistical attack
   └─ IV prepended to key → FMS attack (WEP-like)
```

---

## 7. TOOLS

| Tool | Purpose |
|---|---|
| **PadBuster** | Automated padding oracle exploitation |
| **xortool** | Repeating-key XOR analysis (key length detection + cracking) |
| **CyberChef** | Quick XOR, encoding, block cipher operations |
| **SageMath** | LFSR/LCG analysis, lattice-based recovery |
| **pycryptodome** | AES/DES implementation for testing |
| **hashcat** | Brute force symmetric keys (GPU-accelerated) |
| **Custom Python** | All attacks above implementable in pure Python |
