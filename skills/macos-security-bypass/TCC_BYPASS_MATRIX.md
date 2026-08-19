# TCC Bypass Matrix — Per-Version & Per-Protection Techniques

> **AI LOAD INSTRUCTION**: Load this when you need version-specific TCC bypass details, protection-type-specific techniques, or MDM/PPPC abuse patterns. Assumes the main [SKILL.md](./SKILL.md) is already loaded for TCC fundamentals.

---

## 1. TCC BYPASS BY macOS VERSION

### 1.1 Mojave (10.14)

| CVE / Technique | Bypassed Protection | Method |
|---|---|---|
| SSH TCC bypass | FDA | SSH into localhost inherits FDA of Terminal |
| Finder scripting | File access | AppleScript controls Finder without TCC prompt |
| Direct TCC.db edit | All | User TCC.db modifiable without SIP restriction |
| `tccutil` manipulation | All | Programmatic TCC reset and grant |

### 1.2 Catalina (10.15)

| CVE / Technique | Bypassed Protection | Method |
|---|---|---|
| TCC.db moved under SIP | — | Direct edit now blocked (defense improvement) |
| Mount trick | FDA | Mount disk image with crafted TCC.db over user's |
| Environment variable injection | Varies | Inject into entitled processes via env vars |
| CVE-2020-9934 | All | TCC.db race condition during profile migration |
| Inetloc file handler | Automation | `.inetloc` file triggers app without quarantine check |

### 1.3 Big Sur (11.x)

| CVE / Technique | Bypassed Protection | Method |
|---|---|---|
| CVE-2021-30713 | All | XCSSET malware: forge synthetic clicks via `screencapture` |
| CVE-2021-30970 | FDA | `com.apple.systempreferences` injection via scripts dir |
| Accessibility API abuse | Accessibility | Entitled process → inject into Accessibility-granted app |
| Signed System Volume (SSV) | — | Volume seal limits modification (defense improvement) |

### 1.4 Monterey (12.x)

| CVE / Technique | Bypassed Protection | Method |
|---|---|---|
| CVE-2022-22583 | FDA | `PackageKit` framework exploitation |
| CVE-2021-30892 (Shrootless) | SIP + TCC | `system_installd` post-install script in signed pkg |
| CVE-2022-32900 | Automation | Shortcut app can be abused for automation without consent |
| Python/Perl TCC inheritance | FDA | Script interpreters inherit calling app's TCC |
| Background Task Management Agent | Persistence | New `BTM` visibility (defense improvement) |

### 1.5 Ventura (13.x)

| CVE / Technique | Bypassed Protection | Method |
|---|---|---|
| CVE-2023-32369 (Migraine) | SIP + TCC | `systemmigrationd` abuse during Migration Assistant |
| CVE-2023-32364 | FDA | SQL injection in TCC subsystem via crafted bundle ID |
| CVE-2023-38571 | FDA | Archive Utility bypass — files extracted without quarantine |
| CVE-2022-46689 (MacDirtyCow) | All | COW race condition to overwrite SIP-protected TCC.db |
| Login Items notification | — | Users notified of new Login Items (defense improvement) |

### 1.6 Sonoma (14.x)

| CVE / Technique | Bypassed Protection | Method |
|---|---|---|
| CVE-2024-44133 (HM Surf) | Camera, Mic, Location | Safari TCC bypass via `dscl` user home dir manipulation |
| CVE-2024-44131 | FDA | `FileProvider` symlink race condition |
| Third-party app FDA exploitation | FDA | Piggyback on third-party app's pre-existing FDA grant |
| App Management protection | — | New protection for app modification (defense improvement) |

### 1.7 Sequoia (15.x)

| CVE / Technique | Bypassed Protection | Method |
|---|---|---|
| CVE-2024-44243 | SIP + TCC | StorageKit daemon exploitation for SIP bypass |
| CVE-2025-24086 | FDA | CoreMedia framework exploitation |
| Consent prompt for screen recording | — | Weekly re-consent for screen recording (defense improvement) |
| Network extension prompt changes | — | Tighter notification for network extensions |

---

## 2. TCC BYPASS BY PROTECTION TYPE

### 2.1 Camera & Microphone

| Technique | Requirements | Notes |
|---|---|---|
| Exploit Camera-granted app | Code execution in granted app | Zoom, Teams, FaceTime all have Camera TCC |
| Safari HM Surf (CVE-2024-44133) | Safari-specific bypass | Manipulate user home directory config |
| XPC to VDCAssistant/coreaudiod | Privileged IPC | Camera/audio daemons |
| Virtual camera/mic injection | Kernel extension or DriverKit | Inject virtual device at driver level |

### 2.2 Full Disk Access (FDA)

| Technique | Requirements | Notes |
|---|---|---|
| Terminal/iTerm inheritance | Terminal app has FDA | All child processes inherit FDA |
| Backup software exploitation | Backup apps (CCC, Time Machine) granted FDA | Common in enterprise |
| MDM agent exploitation | MDM agent with FDA | Jamf, Mosyle, Kandji, etc. |
| SSH localhost | FDA-granted Terminal + SSH enabled | SSH session inherits FDA |
| Script interpreter inheritance | Python/Ruby/Perl run from FDA context | Interpreter inherits grants |

### 2.3 Automation (Apple Events)

| Technique | Requirements | Notes |
|---|---|---|
| osascript to Finder | Automation→Finder permission (lower bar) | Move, copy, read files via Finder |
| System Events scripting | Automation→System Events | UI scripting, keystroke injection |
| Terminal `do script` | Automation→Terminal | Run commands in Terminal context |
| Calendar alert exploit | Calendar event with alert action | Execute script on trigger |

### 2.4 Location Services

| Technique | Requirements | Notes |
|---|---|---|
| CoreLocation via entitled app | Exploit app with Location permission | Weather, Maps |
| Wi-Fi SSID for location inference | Network access (no TCC required) | Correlate SSID with location databases |

### 2.5 Screen Recording / Screen Capture

| Technique | Requirements | Notes |
|---|---|---|
| `screencapture` via entitled process | Exploit process with Screen Recording TCC | Zoom, Teams desktop agents |
| CGWindowListCreateImage without TCC | Pre-Catalina only | API deprecated for privacy |
| Window server exploitation | Privileged access | `WindowServer` process has inherent screen access |

---

## 3. MDM / PPPC PROFILE ABUSE

### 3.1 PPPC (Privacy Preferences Policy Control) Profiles

MDM-managed devices can push PPPC profiles that pre-approve TCC permissions without user consent.

```xml
<!-- Example PPPC payload granting FDA to a bundle -->
<dict>
  <key>Authorization</key>
  <string>AllowStandardUserToSetSystemService</string>
  <key>CodeRequirement</key>
  <string>identifier "com.attacker.tool" and anchor apple generic</string>
  <key>IdentifierType</key>
  <string>bundleID</string>
  <key>Identifier</key>
  <string>com.attacker.tool</string>
  <key>Services</key>
  <dict>
    <key>SystemPolicyAllFiles</key>
    <dict>
      <key>Authorization</key>
      <string>Allow</string>
    </dict>
  </dict>
</dict>
```

### 3.2 Attack Vectors

| Vector | Description |
|---|---|
| Rogue MDM enrollment | Social engineering user to enroll in attacker-controlled MDM |
| Compromised MDM server | Exploit Jamf Pro, Mosyle, Kandji, etc. to push malicious profiles |
| Profile injection via local admin | `profiles install -path /path/to/profile.mobileconfig` (requires admin or MDM) |
| Inherited enrollment | Device sold/transferred with MDM enrollment still active |

---

## 4. TCC INSPECTION & RECONNAISSANCE

```bash
# List all TCC databases accessible
find / -name "TCC.db" 2>/dev/null

# Dump user TCC (requires FDA or SIP off)
sqlite3 ~/Library/Application\ Support/com.apple.TCC/TCC.db \
  "SELECT service, client, client_type, auth_value, auth_reason FROM access;"

# auth_value meanings
# 0 = denied, 1 = unknown, 2 = allowed, 3 = limited

# List apps with specific TCC grants
sqlite3 ~/Library/Application\ Support/com.apple.TCC/TCC.db \
  "SELECT client FROM access WHERE service='kTCCServiceSystemPolicyAllFiles' AND auth_value=2;"

# Check effective TCC for a running process
# (no direct API — must infer from TCC.db + entitlements + code signature)

# Enumerate entitlements of all running processes
ps -eo pid,comm | while read pid comm; do
  codesign -d --entitlements :- "$comm" 2>/dev/null
done
```

---

## 5. DEFENSE & DETECTION

| Indicator | Detection Method |
|---|---|
| TCC.db access from unexpected process | Endpoint telemetry (ES_EVENT_TYPE_NOTIFY_OPEN on TCC.db) |
| `xattr -d` on quarantine | File integrity monitoring |
| AppleScript to sensitive apps | Audit Apple Events via Endpoint Security |
| New LaunchAgents/LaunchDaemons | Monitor `~/Library/LaunchAgents/` and `/Library/LaunchDaemons/` |
| MDM profile installation | Alert on unexpected profile additions |
| `csrutil disable` | Monitor SIP status changes at boot |
