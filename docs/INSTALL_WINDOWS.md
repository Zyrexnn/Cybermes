# 🪟 Windows Installation & Setup Guide

This guide provides step-by-step instructions for running **Cybermes** on **Windows 10 / 11**.

---

## ⚡ Quick Comparison: Which Setup Should You Choose?

| Method | Recommended For | Prerequisites | Pros |
| :--- | :--- | :--- | :--- |
| **Method 1: PowerShell Native** | Beginners & daily Windows users | Python 3.11+, Git | Fast setup, no virtualization needed |
| **Method 2: WSL2 (Ubuntu / Kali)** | Power users & Linux tooling | WSL2 enabled | Full Linux binary compatibility |
| **Method 3: Docker Desktop** | Isolated environments | Docker Desktop | Zero local tool dependency conflicts |

---

## 🛠️ Prerequisites (For All Methods)

1. **Git for Windows**: Download and install from [git-scm.com](https://gitforwindows.org/).
2. **Python 3.11 or newer**:
   - Download from [python.org](https://www.python.org/downloads/).
   - ⚠️ **Important**: Check the box **"Add Python to PATH"** during installation.
3. **Node.js (Optional, for Browser MCP & Puppeteer)**:
   - Download LTS from [nodejs.org](https://nodejs.org/).

---

## 🚀 Method 1: PowerShell Native Setup (Recommended)

### Step 1: Open PowerShell as User
Open Windows PowerShell (or Windows Terminal) and navigate to your workspace directory.

### Step 2: Enable Script Execution (If Not Already Enabled)
Run this command once to allow local PowerShell scripts to execute:
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### Step 3: Clone Cybermes Repository
```powershell
git clone https://github.com/Zyrexnn/Cybermes.git
cd Cybermes
```

### Step 4: Run the Automated Windows Installer
```powershell
.\setup_windows.ps1
```
This automated script will:
- Verify Python version.
- Create an isolated Python virtual environment (`venv`).
- Install all required Python packages.
- Initialize folder structures (`reports/`, `recon/`, `output/`, `logs/`, etc.).
- Prepare `.env` and `.hermes/config.yaml`.

### Step 5: Configure API Keys
Open and edit the `.env` file with your preferred editor (Notepad, VS Code, etc.):
```powershell
notepad .env
```
Provide your LLM API credentials (e.g. `ROUTER_API_KEY`, `OPENROUTER_API_KEY`, or local endpoint).

### Step 6: Verify Your System Setup
Run the diagnostic script to ensure everything is operational:
```powershell
python tools\windows_compat_check.py
```

### Step 7: Launch Cybermes
You can launch assessments directly using either wrapper:
```powershell
# Using the root wrapper
.\hermes.bat "Assess http://127.0.0.1:8888"

# Or load the environment in your active PowerShell session
. .\env.ps1
hermes "Assess http://127.0.0.1:8888"
```

---

## 🐧 Method 2: WSL2 (Windows Subsystem for Linux)

WSL2 provides the exact native Linux experience inside Windows.

### Step 1: Enable WSL (Run in PowerShell as Administrator)
```powershell
wsl --install
```
*Restart your computer if prompted.*

### Step 2: Launch WSL (Ubuntu) and Clone Repo
Inside your WSL terminal:
```bash
git clone https://github.com/Zyrexnn/Cybermes.git
cd Cybermes
```

### Step 3: Run the Linux Automated Installer
```bash
chmod +x setup.sh
./setup.sh
```

### Step 4: Configure and Run
```bash
nano .env
source env.sh
./hermes "Assess http://127.0.0.1:8888"
```

---

## 🐳 Method 3: Docker Desktop on Windows

If you prefer a sandboxed container environment without installing Python locally:

### Step 1: Install Docker Desktop
Make sure [Docker Desktop for Windows](https://www.docker.com/products/docker-desktop/) is running.

### Step 2: Configure Environment
```powershell
copy .env.example .env
notepad .env
```

### Step 3: Start Containers via Helper Batch Script
```powershell
.\bin\docker_windows.bat up
```

### Step 4: Execute Assessments Inside Container
```powershell
docker compose exec hermes-cybermes hermes "Assess http://127.0.0.1:8888"
```

To stop containers:
```powershell
.\bin\docker_windows.bat down
```

---

## 🔧 Windows Troubleshooting & FAQs

### 1. `Execution of scripts is disabled on this system` (PSSecurityException)
**Fix:** Run:
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### 2. `'python' is not recognized as an internal or external command`
**Fix:** 
- Re-run the Python installer and ensure **"Add python.exe to PATH"** is checked.
- Alternatively, test if `py` works: `py -3 -m venv venv`.

### 3. Windows Defender / Antivirus Blocks Security Binaries
**Symptoms:** Antivirus flags tools in `tools/bin` or payloads in `knowledge/`.
**Fix:** Add the Cybermes project directory to your Windows Defender exclusions:
- Windows Security -> *Virus & threat protection* -> *Manage settings* -> *Exclusions* -> *Add an exclusion (Folder)* -> Select the `Cybermes` root directory.

### 4. UTF-8 Console Character Glitches
**Fix:** Set your active Windows console codepage to UTF-8:
```powershell
chcp 65001
```

### 5. Windows Long Path Issues (MAX_PATH limit)
**Fix:** In PowerShell as Administrator:
```powershell
New-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Control\FileSystem" -Name "LongPathsEnabled" -Value 1 -PropertyType DWORD -Force
```
