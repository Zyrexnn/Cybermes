---
name: steganography-techniques
description: >-
  Steganography detection and extraction playbook. Use when analyzing images (LSB, PNG chunks, JPEG DCT, EXIF), audio (spectrogram, DTMF), files (polyglots, appended data, ADS), and text (whitespace, zero-width, homoglyphs) for hidden data.
---

# SKILL: Steganography Techniques — Expert Analysis Playbook

> **AI LOAD INSTRUCTION**: Expert steganography detection and extraction techniques. Covers image steganography (LSB, PNG chunk hiding, JPEG DCT, EXIF metadata, dimension tricks, palette manipulation), audio steganography (spectrogram, LSB, DTMF, morse), file steganography (polyglots, binwalk, NTFS ADS, steghide), and text steganography (whitespace, zero-width Unicode, homoglyphs). Base models miss the systematic file-type-based analysis approach and tool-specific extraction workflows.

## 0. RELATED ROUTING

Before going deep, consider loading:

- [traffic-analysis-pcap](../traffic-analysis-pcap/SKILL.md) for extracting files from network captures before stego analysis
- [memory-forensics-volatility](../memory-forensics-volatility/SKILL.md) for extracting files from memory dumps
- [classical-cipher-analysis](../classical-cipher-analysis/SKILL.md) if extracted hidden data is further encrypted/encoded

### Tool Reference

Also load [STEGO_TOOLS_GUIDE.md](./STEGO_TOOLS_GUIDE.md) when you need:
- Tool installation instructions and dependencies
- Detailed command reference for each stego tool
- Workflow patterns for specific file types

---

## 1. IMAGE STEGANOGRAPHY

### LSB (Least Significant Bit)

LSB embeds data in the least significant bits of pixel color channels.

```bash
# zsteg — LSB analysis for PNG/BMP
zsteg image.png                       # auto-detect all LSB patterns
zsteg image.png -a                    # try all known methods
zsteg image.png -b 1                  # extract bit plane 1
zsteg image.png -E "b1,rgb,lsb,xy"   # specific extraction pattern

# StegSolve (Java GUI)
java -jar StegSolve.jar
# Navigate color planes: Red 0, Green 0, Blue 0 → look for hidden image/text
# Data Extractor: specify bit planes + byte order

# stegoveritas — comprehensive automated analysis
stegoveritas image.png
# Runs: exiftool, binwalk, zsteg, foremost, color plane extraction
```

### PNG Specific

```bash
# pngcheck — validate structure, find hidden chunks
pngcheck -v image.png

# Hidden chunks: tEXt, zTXt (compressed text), iTXt (international text)
# Custom/private chunks may contain hidden data

# CRC vs dimensions trick
# If CRC doesn't match declared dimensions → image was cropped
# Fix: brute-force correct width/height → reveals hidden rows/columns
python3 -c "
import struct, zlib
with open('image.png','rb') as f:
    data = f.read()
# Check IHDR CRC at offset 29
ihdr = data[12:29]
for h in range(1,2000):
    for w in range(1,2000):
        new_ihdr = struct.pack('>II',w,h) + ihdr[8:]
        if zlib.crc32(b'IHDR'+new_ihdr) & 0xffffffff == struct.unpack('>I',data[29:33])[0]:
            print(f'Width: {w}, Height: {h}')
"

# APNG (animated PNG) — hidden frames
# Use apngdis to extract all frames: apngdis image.png
```

### JPEG Specific

```bash
# steghide — embed/extract from JPEG (DCT coefficient modification)
steghide extract -sf image.jpg                 # extract (no passphrase)
steghide extract -sf image.jpg -p PASSWORD     # extract with passphrase
steghide info image.jpg                        # check if data is embedded

# stegcracker — brute force steghide passphrase
stegcracker image.jpg wordlist.txt

# jsteg — JPEG LSB steganography
jsteg reveal image.jpg output.txt

# JPEG structure analysis
exiftool -v3 image.jpg       # verbose metadata + structure
jpegdump image.jpg           # raw JPEG marker analysis
```

### EXIF Metadata

```bash
# exiftool — comprehensive metadata extraction
exiftool image.jpg
exiftool -b -ThumbnailImage image.jpg > thumb.jpg   # extract thumbnail
exiftool -all= image.jpg                             # strip all metadata

# Hidden data in EXIF fields (comment, artist, copyright, etc.)
exiftool -Comment image.jpg
exiftool -UserComment image.jpg
strings image.jpg | grep -i "flag\|key\|secret"
```

### Palette-Based (GIF)

```bash
# GIF color table manipulation — data in color palette order
gifsicle -I image.gif                    # info
gifsicle --color-info image.gif          # palette details
# Check for animation frames: convert -coalesce image.gif frame_%d.png
```

---

## 2. AUDIO STEGANOGRAPHY

### Spectrogram Analysis

```bash
# Sonic Visualiser — best for spectrogram viewing
# Layer → Add Spectrogram → look for visual patterns (text/images)

# Audacity
# Analyze → Plot Spectrum
# Select audio → change view to Spectrogram

# sox for command-line spectrogram generation
sox audio.wav -n spectrogram -o spectro.png
```

### Audio LSB

```bash
# DeepSound — hide/extract files in audio (Windows)
# GUI tool: open audio file → extract hidden files

# WavSteg — LSB in WAV files
python3 WavSteg.py -r -i audio.wav -o output.txt -n 1   # extract 1 LSB
python3 WavSteg.py -r -i audio.wav -o output.txt -n 2   # extract 2 LSBs
```

### DTMF / Morse Code

```bash
# DTMF decoder (phone tones)
multimon-ng -t wav -a DTMF audio.wav

# Morse code
# Audacity → visual inspection of on/off pattern
# Online decoder or manual: .- = A, -... = B, etc.

# SSTV (Slow-Scan Television) — image in audio
qsstv                    # GUI decoder
# Or: RX-SSTV (Windows)
```

### WAV Header Manipulation

```bash
# Check for data appended after WAV audio data
# WAV data chunk size vs actual file size
python3 -c "
import wave
w = wave.open('audio.wav','rb')
print(f'Frames: {w.getnframes()}, Channels: {w.getnchannels()}, Width: {w.getsampwidth()}')
expected = w.getnframes() * w.getnchannels() * w.getsampwidth() + 44  # 44 = WAV header
import os
actual = os.path.getsize('audio.wav')
if actual > expected:
    print(f'Extra data: {actual - expected} bytes appended')
"
```

---

## 3. FILE STEGANOGRAPHY

### Polyglot Files

A single file that is valid in two or more formats simultaneously.

```bash
# Detection: check file with multiple tools
file suspicious_file
xxd suspicious_file | head          # check magic bytes
binwalk suspicious_file             # find embedded files

# Common polyglots: PDF+ZIP, JPEG+ZIP, JPEG+RAR, PNG+ZIP
# Try unzip on image files:
unzip image.jpg -d extracted/
7z x image.jpg -oextracted/
```

### Appended / Embedded Data

```bash
# binwalk — scan for embedded files and data
binwalk image.png                   # scan
binwalk -e image.png                # extract embedded files
binwalk --dd='.*' image.png         # extract everything

# foremost — file carving
foremost -i suspicious_file -o output_dir/

# dd — manual extraction
# If binwalk shows embedded ZIP at offset 0x1234:
dd if=suspicious_file bs=1 skip=$((0x1234)) of=extracted.zip
```

### NTFS Alternate Data Streams (ADS)

```cmd
:: List ADS (Windows)
dir /r file.txt
Get-Item file.txt -Stream *

:: Read hidden stream
more < file.txt:hidden_stream
Get-Content file.txt -Stream hidden_stream

:: Create ADS (for testing)
echo "hidden data" > file.txt:secret
```

### Steghide Brute Force

```bash
# stegcracker — wordlist attack on steghide passphrase
stegcracker image.jpg /usr/share/wordlists/rockyou.txt

# stegseek — faster alternative
stegseek image.jpg /usr/share/wordlists/rockyou.txt
# stegseek is ~10000x faster than stegcracker
```

---

## 4. TEXT STEGANOGRAPHY

### Whitespace Encoding

```bash
# Tabs and spaces encode binary (tab=1, space=0 or vice versa)
# stegsnow — whitespace steganography
stegsnow -C message.txt                # extract hidden message
stegsnow -C -p PASSWORD message.txt    # extract with password

# Manual detection:
cat -A file.txt | head     # show tabs (^I) and line endings ($)
xxd file.txt | grep "09 20\|20 09"    # look for tab/space patterns
```

### Zero-Width Characters

```bash
# Unicode invisible characters used for encoding:
# U+200B (Zero-Width Space), U+200C (ZWNJ), U+200D (ZWJ), U+FEFF (BOM)

# Detection:
python3 -c "
text = open('message.txt','r').read()
hidden = [c for c in text if ord(c) in [0x200b, 0x200c, 0x200d, 0xfeff]]
print(f'Found {len(hidden)} zero-width characters')
binary = ''.join('0' if ord(c)==0x200b else '1' for c in hidden)
# Convert binary to ASCII
"

# Online tools: holloway.nz/steg, Unicode Steganography decoders
```

### Homoglyph Substitution

```bash
# Visually identical characters from different Unicode blocks
# e.g., Latin 'a' (U+0061) vs Cyrillic 'а' (U+0430)

# Detection:
python3 -c "
text = open('message.txt','r').read()
for i, c in enumerate(text):
    if ord(c) > 127:
        print(f'Position {i}: char={c} ord={ord(c)} name={__import__(\"unicodedata\").name(c,\"?\")}')
"
```

---

## 5. DECISION TREE

```
Suspect hidden data — what file type?
│
├── Image (PNG/BMP)?
│   ├── Check metadata: exiftool (§1 EXIF)
│   ├── Check structure: pngcheck, binwalk (§1 PNG)
│   ├── LSB analysis: zsteg, StegSolve (§1 LSB)
│   ├── Check dimensions vs CRC: height/width brute force (§1 PNG)
│   ├── Check for appended data: binwalk -e (§3)
│   └── Try as polyglot: unzip/7z (§3)
│
├── Image (JPEG)?
│   ├── Check metadata: exiftool (§1 EXIF)
│   ├── Try steghide: steghide extract (§1 JPEG)
│   │   └── Password protected? → stegseek brute force (§3)
│   ├── Try jsteg: jsteg reveal (§1 JPEG)
│   ├── Check for appended data: binwalk -e (§3)
│   └── Check thumbnail: exiftool -b -ThumbnailImage (§1 EXIF)
│
├── Image (GIF)?
│   ├── Check frames: extract all animation frames (§1 Palette)
│   ├── Check palette: gifsicle --color-info (§1 Palette)
│   └── Check for appended data: binwalk -e (§3)
│
├── Audio (WAV/MP3/FLAC)?
│   ├── Spectrogram: Sonic Visualiser / Audacity (§2)
│   ├── LSB: WavSteg (§2)
│   ├── DTMF tones: multimon-ng (§2)
│   ├── Morse code: manual or decoder (§2)
│   ├── SSTV: qsstv (§2)
│   └── Check file size vs expected: header analysis (§2)
│
├── Text file?
│   ├── Check whitespace: cat -A, stegsnow (§4)
│   ├── Check zero-width chars: Unicode analysis (§4)
│   ├── Check homoglyphs: non-ASCII detection (§4)
│   └── Check encoding: multiple base decodings
│
├── Any file type?
│   ├── strings: strings -n 8 file | grep -i "flag\|key\|pass"
│   ├── binwalk: binwalk -e file (embedded files) (§3)
│   ├── file: file suspicious_file (true type)
│   ├── xxd: check magic bytes, compare headers
│   └── NTFS? → check ADS: dir /r (§3)
│
└── Password/passphrase needed?
    ├── steghide → stegseek / stegcracker (§3)
    ├── Check challenge description for hints
    └── Try common passwords: password, file name, challenge name
```
