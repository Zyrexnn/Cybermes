---
name: mobile-ssl-pinning-bypass
description: >-
  Mobile SSL pinning bypass playbook. Use when intercepting HTTPS traffic from mobile applications that implement certificate pinning, public key pinning, or SPKI hash pinning on Android and iOS, including React Native, Flutter, and Xamarin frameworks.
---

# SKILL: Mobile SSL Pinning Bypass — Expert Attack Playbook

> **AI LOAD INSTRUCTION**: Expert SSL pinning bypass techniques for mobile platforms. Covers Android and iOS bypass methods (Frida, Objection, Xposed, SSL Kill Switch), framework-specific bypasses (Flutter, React Native, Xamarin), and troubleshooting non-standard pinning implementations. Base models miss framework-specific hook points and multi-layer pinning configurations.

## 0. RELATED ROUTING

Before going deep, consider loading:

- [android-pentesting-tricks](../android-pentesting-tricks/SKILL.md) for broader Android testing beyond SSL bypass
- [ios-pentesting-tricks](../ios-pentesting-tricks/SKILL.md) for broader iOS testing beyond SSL bypass
- [api-sec](../api-sec/SKILL.md) once traffic is intercepted for API-level testing

---

## 1. SSL PINNING TYPES

| Pinning Type | What Is Pinned | Resilience | Common In |
|---|---|---|---|
| Certificate pinning | Exact leaf certificate (DER/PEM) | Low (breaks on cert rotation) | Legacy apps |
| Public key pinning | Subject Public Key Info | Medium (survives cert renewal if key unchanged) | Modern apps |
| SPKI hash pinning | SHA-256 of SPKI | Medium (same as public key) | OkHttp, AFNetworking |
| CA pinning | Intermediate or root CA cert | High (any cert from that CA works) | Enterprise apps |
| Multi-pin (backup pins) | Primary + backup pins | High (fallback pins) | HPKP-aware apps |

### How Pinning Works

```
TLS Handshake
│
├── Server presents certificate chain
│
├── Standard validation (system trust store)
│   └── Passes? continue : connection fails
│
└── Pin validation (app-level check)
    ├── Extract server cert/pubkey/SPKI hash
    ├── Compare against embedded pins
    └── Match found? → allow : → reject connection
```

---

## 2. ANDROID BYPASS METHODS

### 2.1 Frida Universal SSL Bypass

```javascript
// Hooks TrustManager, OkHttp, Volley, Retrofit, Conscrypt
Java.perform(function() {

    // ── TrustManagerImpl (Android system) ──
    try {
        var TMI = Java.use('com.android.org.conscrypt.TrustManagerImpl');
        TMI.verifyChain.implementation = function() {
            console.log('[Bypass] TrustManagerImpl.verifyChain');
            return arguments[0]; // return untouched chain
        };
    } catch(e) {}

    // ── X509TrustManager (custom implementations) ──
    var TrustManager = Java.registerClass({
        name: 'com.bypass.TrustManager',
        implements: [Java.use('javax.net.ssl.X509TrustManager')],
        methods: {
            checkClientTrusted: function() {},
            checkServerTrusted: function() {},
            getAcceptedIssuers: function() { return []; }
        }
    });

    var SSLContext = Java.use('javax.net.ssl.SSLContext');
    SSLContext.init.overload('[Ljavax.net.ssl.KeyManager;',
        '[Ljavax.net.ssl.TrustManager;', 'java.security.SecureRandom')
        .implementation = function(km, tm, sr) {
        console.log('[Bypass] SSLContext.init');
        this.init(km, [TrustManager.$new()], sr);
    };

    // ── OkHttp3 CertificatePinner ──
    try {
        var CP = Java.use('okhttp3.CertificatePinner');
        CP.check.overload('java.lang.String', 'java.util.List').implementation = function() {
            console.log('[Bypass] OkHttp3 CertificatePinner.check: ' + arguments[0]);
        };
        // check$okhttp variant (OkHttp 4.x)
        try { CP['check$okhttp'].implementation = function() {}; } catch(e) {}
    } catch(e) {}

    // ── Retrofit / OkHttp interceptor ──
    try {
        var OkHttpClient = Java.use('okhttp3.OkHttpClient$Builder');
        OkHttpClient.certificatePinner.implementation = function(pinner) {
            console.log('[Bypass] OkHttpClient.Builder.certificatePinner');
            return this; // return builder without pinner
        };
    } catch(e) {}

    // ── Volley (HurlStack) ──
    try {
        var HurlStack = Java.use('com.android.volley.toolbox.HurlStack');
        HurlStack.createConnection.implementation = function(url) {
            console.log('[Bypass] Volley HurlStack: ' + url);
            var conn = this.createConnection(url);
            // Remove hostname verifier
            conn.setHostnameVerifier(Java.use(
                'javax.net.ssl.HttpsURLConnection').getDefaultHostnameVerifier());
            return conn;
        };
    } catch(e) {}

    // ── Conscrypt / BoringSSL (modern Android) ──
    try {
        var Conscrypt = Java.use('org.conscrypt.ConscryptFileDescriptorSocket');
        Conscrypt.verifyCertificateChain.implementation = function() {
            console.log('[Bypass] Conscrypt verifyCertificateChain');
        };
    } catch(e) {}

    // ── Apache HttpClient (legacy) ──
    try {
        var AbstractVerifier = Java.use('org.apache.http.conn.ssl.AbstractVerifier');
        AbstractVerifier.verify.overload('java.lang.String', '[Ljava.lang.String;',
            '[Ljava.lang.String;', 'boolean').implementation = function() {
            console.log('[Bypass] Apache AbstractVerifier');
        };
    } catch(e) {}

    // ── HostnameVerifier ──
    try {
        var HV = Java.use('javax.net.ssl.HttpsURLConnection');
        HV.setDefaultHostnameVerifier.implementation = function(v) {
            console.log('[Bypass] Ignoring custom HostnameVerifier');
        };
    } catch(e) {}

    console.log('[+] Android universal SSL bypass loaded');
});
```

### 2.2 Objection (One Command)

```bash
objection -g com.target.app explore --startup-command "android sslpinning disable"
```

### 2.3 Network Security Config (Debug Override)

```xml
<!-- AndroidManifest.xml: android:networkSecurityConfig="@xml/network_security_config" -->

<!-- res/xml/network_security_config.xml -->
<network-security-config>
  <base-config>
    <trust-anchors>
      <certificates src="system" />
      <certificates src="user" />     <!-- Trust user-installed CAs -->
    </trust-anchors>
  </base-config>
</network-security-config>
```

Workflow: decompile APK → add/modify config → repackage → re-sign → install.

```bash
apktool d target.apk -o target_dir
# Edit res/xml/network_security_config.xml
# Add reference in AndroidManifest.xml if missing
apktool b target_dir -o target_patched.apk
zipalign -v 4 target_patched.apk target_aligned.apk
apksigner sign --ks my-key.keystore target_aligned.apk
adb install target_aligned.apk
```

### 2.4 Xposed / LSPosed Modules

| Module | Method | Scope | Root Required |
|---|---|---|---|
| JustTrustMe | Hooks TrustManager + OkHttp | Per-app | Yes (Xposed) |
| SSLUnpinning | Hooks certificate validation | Per-app | Yes (LSPosed) |
| TrustMeAlready | Global TrustManager bypass | System-wide | Yes (LSPosed) |

### 2.5 Magisk + System CA Installation

```bash
# Install proxy CA as system cert (Android 7+ requires this for system-level trust)
# Method 1: MagiskTrustUserCerts module
# Moves user CAs to /system/etc/security/cacerts/ via Magisk overlay

# Method 2: Manual (requires root)
adb push burp_ca.pem /sdcard/
adb shell
su
mount -o remount,rw /system
cp /sdcard/burp_ca.pem /system/etc/security/cacerts/9a5ba575.0  # hash-named
chmod 644 /system/etc/security/cacerts/9a5ba575.0
mount -o remount,ro /system

# Get correct hash filename:
openssl x509 -inform PEM -subject_hash_old -in burp_ca.pem | head -1
# Output: 9a5ba575 → filename is 9a5ba575.0
```

### 2.6 Manual Decompile → Patch → Repackage

```bash
# Step 1: Decompile
jadx -d decompiled/ target.apk

# Step 2: Find pinning code
grep -r "CertificatePinner\|X509TrustManager\|checkServerTrusted\|ssl" decompiled/

# Step 3: Identify pinning implementation and patch
# Use smali editing for precise control:
apktool d target.apk
# Edit smali files to NOP out pinning checks
# Look for invoke-virtual {checkServerTrusted} and replace with return-void

# Step 4: Repackage and sign
apktool b target_dir -o patched.apk
apksigner sign --ks debug.keystore patched.apk
```

---

## 3. iOS BYPASS METHODS

### 3.1 Frida (SecTrust Hooks)

```javascript
// Hook core iOS SSL validation functions
var SecTrustEvaluateWithError = Module.findExportByName('Security', 'SecTrustEvaluateWithError');
Interceptor.attach(SecTrustEvaluateWithError, {
    onLeave: function(retval) {
        retval.replace(ptr(1));
    }
});

var SecTrustEvaluate = Module.findExportByName('Security', 'SecTrustEvaluate');
Interceptor.attach(SecTrustEvaluate, {
    onLeave: function(retval) {
        retval.replace(ptr(0));
    }
});

// Hook SSLHandshake (lower-level)
var SSLHandshake = Module.findExportByName('Security', 'SSLHandshake');
if (SSLHandshake) {
    Interceptor.attach(SSLHandshake, {
        onLeave: function(retval) {
            if (retval.toInt32() === -9807) { // errSSLXCertChainInvalid
                retval.replace(ptr(0));
            }
        }
    });
}

// Hook NSURLSession delegate method
try {
    var cls = ObjC.classes.NSURLSession;
    // Hook URLSession:didReceiveChallenge:completionHandler: on delegates
    ObjC.enumerateLoadedClasses({
        onMatch: function(name) {
            try {
                var methods = ObjC.classes[name].$ownMethods;
                for (var i = 0; i < methods.length; i++) {
                    if (methods[i].indexOf('didReceiveChallenge') !== -1 &&
                        methods[i].indexOf('completionHandler') !== -1) {
                        console.log('[SSL] Found delegate: ' + name + ' ' + methods[i]);
                    }
                }
            } catch(e) {}
        },
        onComplete: function() {}
    });
} catch(e) {}
```

### 3.2 Objection (One Command)

```bash
objection -g com.target.app explore --startup-command "ios sslpinning disable"
```

### 3.3 SSL Kill Switch 2 (Jailbreak Tweak)

```bash
# Install via Cydia/Sileo
# Package: com.nablac0d3.sslkillswitch2
# Disables SSL pinning system-wide or per-app via Settings toggle

# Hooks:
# - SecTrustEvaluate
# - SSLHandshake
# - SSLSetSessionOption
# - tls_helper_create_peer_trust
```

### 3.4 Library-Specific Hooks

| Library | iOS Hook Point | Frida Approach |
|---|---|---|
| AFNetworking | `AFSecurityPolicy.evaluateServerTrust:forDomain:` | Return YES |
| Alamofire | `ServerTrustManager.evaluate(_:forHost:)` | Skip evaluation |
| TrustKit | `TSKPinningValidator verifyPublicKeyPin:` | Return success |
| NSURLSession | `URLSession:didReceiveChallenge:completionHandler:` | Call completionHandler with .useCredential |

### 3.5 Manual Binary Patch

```bash
# Find pinning function in binary
strings decrypted_binary | grep -i "pin\|cert\|trust"
# Disassemble and find the validation function
# Replace comparison/branch instruction with NOP or unconditional pass

# LLDB runtime modification
lldb -n TargetApp
(lldb) breakpoint set -n "SecTrustEvaluateWithError"
(lldb) breakpoint command add 1
> thread return 1
> continue
> DONE
```

---

## 4. FRAMEWORK-SPECIFIC BYPASSES

### 4.1 Flutter

Flutter uses Dart's `dart:io` library with BoringSSL underneath. Standard Frida hooks on Java/ObjC layers don't work.

```javascript
// Flutter SSL bypass — must hook BoringSSL directly
// Find ssl_crypto_x509_session_verify_cert_chain in libflutter.so
var libflutter = Process.findModuleByName('libflutter.so');  // Android
// var libflutter = Process.findModuleByName('Flutter');       // iOS

// Hook ssl_verify_peer_cert (BoringSSL function)
// Signature varies by Flutter version — use pattern scanning
var pattern = 'FF C3 ..';  // Example pattern, varies
var matches = Memory.scan(libflutter.base, libflutter.size, pattern, {
    onMatch: function(address, size) {
        console.log('[Flutter] Potential verify function at: ' + address);
        Interceptor.attach(address, {
            onLeave: function(retval) {
                retval.replace(ptr(0));  // SSL_VERIFY_OK
            }
        });
    },
    onComplete: function() {}
});

// Alternative: use reflutter tool for automated patching
// reflutter target.apk
// This patches BoringSSL in the Flutter engine directly
```

**reflutter tool** (recommended for Flutter apps):

```bash
pip install reflutter
reflutter target.apk
# Outputs patched APK that redirects traffic to your proxy
# Also disables SSL verification in the BoringSSL engine
```

### 4.2 React Native

React Native uses platform networking: OkHttp on Android, NSURLSession on iOS.

| Platform | Networking Stack | Bypass Method |
|---|---|---|
| Android | OkHttp3 | Standard OkHttp CertificatePinner hook |
| iOS | NSURLSession | Standard SecTrust hooks |
| Android (Hermes) | Same OkHttp | Same hooks, but Hermes JIT may need additional handling |

```javascript
// React Native Android — same as OkHttp bypass
Java.perform(function() {
    try {
        var CP = Java.use('okhttp3.CertificatePinner');
        CP.check.overload('java.lang.String', 'java.util.List').implementation = function() {};
    } catch(e) { console.log('OkHttp3 not found, trying okhttp2...'); }

    try {
        var CP2 = Java.use('com.squareup.okhttp.CertificatePinner');
        CP2.check.overload('java.lang.String', 'java.util.List').implementation = function() {};
    } catch(e) {}
});
```

### 4.3 Xamarin

```csharp
// Xamarin pinning typically via:
// ServicePointManager.ServerCertificateValidationCallback
// or custom HttpClientHandler
```

```javascript
// Frida bypass for Xamarin (Mono runtime)
// Hook Mono method: System.Net.ServicePointManager.set_ServerCertificateValidationCallback
var mono_method = Module.findExportByName('libmonosgen-2.0.so',
    'mono_runtime_invoke');
// More practical: hook the managed callback at CIL level
// Use Frida's Mono bridge or objection's built-in Xamarin support

// Objection has built-in Xamarin bypass:
// objection -g com.target.app explore
// > android sslpinning disable   (covers Xamarin on Android)
```

---

## 5. CERTIFICATE TRANSPARENCY & HPKP

| Technology | Status | Impact on Testing |
|---|---|---|
| Certificate Transparency (CT) | Active, enforced by browsers | Mobile apps rarely enforce CT; not a bypass obstacle |
| HPKP (HTTP Public Key Pinning) | Deprecated (2018) | Legacy apps may still check; remove header from proxy response |
| Expect-CT header | Deprecated (2024) | Minimal impact on mobile testing |
| CT in mobile apps | Rare | Only Google apps enforce via custom CT checks |

---

## 6. TROUBLESHOOTING

### 6.1 Common Failures

| Symptom | Cause | Fix |
|---|---|---|
| Bypass script loaded but traffic still fails | Multiple pinning layers | Hook ALL layers: TrustManager + OkHttp + custom checks |
| "Client certificate required" | Mutual TLS (mTLS) | Extract client cert from app bundle/keychain, import into proxy |
| Connection works but no HTTP traffic | Non-HTTP protocol (MQTT, gRPC, WebSocket) | Use Wireshark or protocol-specific proxy |
| App crashes after bypass | Anti-tampering detects hooks | Bypass integrity checks first, then SSL |
| Proxy CA not trusted | Android 7+ user CA restrictions | Install CA as system cert (Magisk module) |
| Flutter app ignores hooks | BoringSSL not hooked at native layer | Use reflutter or native BoringSSL hooks |
| Certificate chain validation timeout | OCSP stapling mismatch | Disable OCSP checks or mock OCSP responder |

### 6.2 Diagnostic Steps

```bash
# Verify proxy CA is installed correctly
# Android:
adb shell "ls /system/etc/security/cacerts/ | grep $(openssl x509 -subject_hash_old -in ca.pem | head -1)"

# iOS: Settings → General → About → Certificate Trust Settings

# Check if target app is actually using SSL (vs. plain HTTP)
# Wireshark filter: tcp.port == 443 and ip.addr == <device_ip>

# Check if Frida is hooking the right process
frida-ps -U | grep target

# Verbose Frida output for debugging hooks
frida -U -f com.target.app -l bypass.js --debug
```

---

## 7. SSL PINNING BYPASS DECISION TREE

```
Need to intercept mobile app HTTPS traffic
│
├── Platform?
│   ├── Android ↓
│   │   ├── Rooted device available?
│   │   │   ├── Yes → Frida universal bypass (§2.1) [FIRST TRY]
│   │   │   │   ├── Works? → done
│   │   │   │   └── Fails? → add Conscrypt + Volley hooks
│   │   │   ├── Still fails? → LSPosed + TrustMeAlready (§2.4)
│   │   │   └── Still fails? → install CA as system cert (§2.5)
│   │   └── No root?
│   │       ├── Debug build? → Network Security Config (§2.3)
│   │       └── Release build? → decompile + patch + repackage (§2.6)
│   │
│   └── iOS ↓
│       ├── Jailbroken device available?
│       │   ├── Yes → Objection ios sslpinning disable (§3.2) [FIRST TRY]
│       │   │   ├── Works? → done
│       │   │   └── Fails? → Frida SecTrust hooks (§3.1)
│       │   ├── Still fails? → SSL Kill Switch 2 (§3.3)
│       │   └── Still fails? → library-specific hooks (§3.4)
│       └── No jailbreak?
│           ├── Re-sign with Frida gadget → run Frida hooks
│           └── Binary patch → sideload (§3.5)
│
├── Framework-specific app?
│   ├── Flutter → reflutter tool or BoringSSL native hooks (§4.1)
│   ├── React Native → standard platform hooks (§4.2)
│   └── Xamarin → Objection or Mono runtime hooks (§4.3)
│
├── Bypass works but issues remain?
│   ├── Client cert required? → extract + import to proxy (§6.1)
│   ├── Non-HTTP protocol? → protocol-specific tooling (§6.1)
│   └── App crashes? → fix anti-tampering first (§6.1)
│
└── All methods fail?
    ├── Analyze traffic at network level (Wireshark/tcpdump)
    ├── Check for custom proprietary protocol
    └── Consider iptables + transparent proxy approach
```

---

## 8. PROXY SETUP QUICK REFERENCE

| Proxy Tool | Best For | SSL Bypass Integration |
|---|---|---|
| Burp Suite | Full HTTP analysis | Import CA to device |
| mitmproxy | Scripted interception | `mitmproxy --set confdir=~/.mitmproxy` |
| Charles Proxy | macOS-native, easy setup | Built-in CA installation |
| Proxyman | macOS/iOS native | Direct iOS device support |
| HTTP Toolkit | Quick Android setup | Automated CA + Frida bypass |

```bash
# Android proxy setup
adb shell settings put global http_proxy <host_ip>:8080

# Remove proxy
adb shell settings put global http_proxy :0

# iOS proxy: Settings → Wi-Fi → Configure Proxy → Manual
```
