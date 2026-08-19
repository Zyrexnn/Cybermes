---
name: stego-tools-guide
description: >-
  Steganography tool installation guide, usage patterns, and command reference for image, audio, file, and text stego analysis.
---

# STEGANOGRAPHY TOOLS GUIDE

> Supplementary reference for [steganography-techniques](./SKILL.md). Installation and detailed usage for each tool.

---

## 1. MULTI-PURPOSE TOOLS

### stegoveritas

Automated steganography analysis — runs multiple tools in one pass.

```bash
# Install
pip3 install stegoveritas
stegoveritas_install_deps

# Usage — full automated analysis
stegoveritas image.png
# Runs: exiftool, binwalk, zsteg, foremost, trailing data check
# Extracts: color planes, LSB data, embedded files
# Output: results/ directory with all findings

stegoveritas image.png -meta         # metadata only
stegoveritas image.png -imageTransform   # color plane extraction only
stegoveritas image.png -extractLSB   # LSB extraction only
```

### binwalk

Firmware/file analysis and extraction.

```bash
# Install
sudo apt install binwalk
# or: pip3 install binwalk

# Scan for embedded files
binwalk file.png

# Extract embedded files
binwalk -e file.png                  # extract known types
binwalk --dd='.*' file.png           # extract everything
binwalk -Me file.png                 # recursive extraction

# Entropy analysis (detect encrypted/compressed regions)
binwalk -E file.png
```

### foremost

File carving tool — recovers files from raw data.

```bash
# Install
sudo apt install foremost

# Carve files
foremost -i suspicious_file -o output_dir/
foremost -t all -i disk_image.raw -o carved/

# Specific file types
foremost -t pdf,jpg,zip -i data.bin -o output/
```

---

## 2. IMAGE TOOLS

### zsteg

PNG/BMP LSB steganography detector.

```bash
# Install
gem install zsteg

# Auto-detect all LSB patterns
zsteg image.png

# Verbose scan with all methods
zsteg image.png -a

# Specific extraction
zsteg image.png -b 1                        # bit plane 1
zsteg image.png -E "b1,rgb,lsb,xy"          # specific pattern
zsteg image.png -E "b1,r,lsb,xy"            # red channel only
zsteg image.png -E "b2,bgr,msb,yx"          # MSB, BGR order

# Extract to file
zsteg image.png -E "b1,rgb,lsb,xy" > extracted.bin
```

### StegSolve

Java GUI for visual bit plane analysis.

```bash
# Install
wget http://www.caesum.com/handbook/Stegsolve.jar

# Run
java -jar Stegsolve.jar

# Key features:
# Analyse → File Format: check headers and structure
# Analyse → Data Extract: specify bit planes to extract
# Arrow keys: cycle through color planes (R0-R7, G0-G7, B0-B7, Alpha)
# XOR/AND/OR with second image for comparison
```

### steghide

JPEG/WAV/BMP/AU steganography with encryption.

```bash
# Install
sudo apt install steghide

# Check if data is embedded
steghide info image.jpg

# Extract (no password)
steghide extract -sf image.jpg

# Extract with password
steghide extract -sf image.jpg -p "password"

# Embed data (for testing)
steghide embed -cf cover.jpg -ef secret.txt -p "password"
```

### stegseek

Fast steghide passphrase cracker (~10000x faster than stegcracker).

```bash
# Install
wget https://github.com/RickdeJager/stegseek/releases/latest/download/stegseek_amd64.deb
sudo dpkg -i stegseek_amd64.deb

# Crack passphrase
stegseek image.jpg /usr/share/wordlists/rockyou.txt

# Seed crack (try without wordlist)
stegseek --seed image.jpg
```

### jsteg

JPEG LSB steganography.

```bash
# Install
go install github.com/lukechampine/jsteg@latest

# Extract
jsteg reveal image.jpg

# Embed (for testing)
jsteg hide cover.jpg secret.txt output.jpg
```

### pngcheck

PNG structure validator.

```bash
# Install
sudo apt install pngcheck

# Validate and show chunk info
pngcheck -v image.png

# Show text chunks
pngcheck -t image.png

# Full verbose with data
pngcheck -vtp7f image.png
```

### exiftool

Metadata extraction and manipulation.

```bash
# Install
sudo apt install libimage-exiftool-perl

# All metadata
exiftool image.jpg

# Specific fields
exiftool -Comment image.jpg
exiftool -UserComment image.jpg
exiftool -GPSLatitude -GPSLongitude image.jpg

# Extract thumbnail
exiftool -b -ThumbnailImage image.jpg > thumbnail.jpg

# Verbose structure
exiftool -v3 image.jpg

# Strip all metadata
exiftool -all= image.jpg
```

---

## 3. AUDIO TOOLS

### Sonic Visualiser

Best spectrogram viewer for stego analysis.

```bash
# Install
sudo apt install sonic-visualiser

# Usage:
# 1. Open audio file
# 2. Layer → Add Spectrogram
# 3. Adjust: Window=4096, Overlap=87.5%, Scale=dBV
# 4. Look for patterns (text, QR codes, images in frequency domain)
```

### multimon-ng

Decoder for DTMF, POCSAG, and other digital modes.

```bash
# Install
sudo apt install multimon-ng

# DTMF decode
multimon-ng -t wav -a DTMF audio.wav

# Multiple decoders
multimon-ng -t wav -a DTMF -a MORSE_CW audio.wav

# From raw audio input
sox audio.wav -t raw -r 22050 -e signed -b 16 -c 1 - | multimon-ng -t raw -
```

### DeepSound

Audio steganography tool (Windows).

```bash
# GUI application — Windows only
# 1. Open carrier audio file
# 2. Click "Extract Secret Files"
# 3. Enter password if prompted

# Alternative for Linux: use WavSteg for WAV LSB analysis
```

---

## 4. TEXT TOOLS

### stegsnow

Whitespace steganography in text files.

```bash
# Install
sudo apt install stegsnow

# Extract hidden message
stegsnow -C message.txt

# Extract with password
stegsnow -C -p "password" message.txt

# Embed (for testing)
stegsnow -C -m "hidden message" -p "password" cover.txt stego.txt
```

---

## 5. RECOMMENDED ANALYSIS WORKFLOW

### Quick Triage (Any File)

```bash
file suspicious_file
exiftool suspicious_file
strings -n 8 suspicious_file | head -50
binwalk suspicious_file
xxd suspicious_file | head -20
```

### Image Deep Analysis

```bash
exiftool -v3 image.*
pngcheck -v image.png         # if PNG
steghide info image.jpg       # if JPEG
zsteg -a image.png            # if PNG/BMP
stegoveritas image.*          # comprehensive automated scan
binwalk -e image.*            # embedded file extraction
```

### Audio Deep Analysis

```bash
exiftool audio.*
file audio.*
sox audio.wav -n spectrogram -o spectro.png
multimon-ng -t wav -a DTMF audio.wav
# Open in Sonic Visualiser for spectrogram inspection
# Check file size vs expected duration
```

### Password Recovery for Steghide

```bash
# Fast: stegseek
stegseek image.jpg /usr/share/wordlists/rockyou.txt

# Slow fallback: stegcracker
stegcracker image.jpg /usr/share/wordlists/rockyou.txt

# Manual: try common passwords
for p in password flag secret admin test ""; do
    steghide extract -sf image.jpg -p "$p" 2>/dev/null && echo "Password: '$p'" && break
done
```
