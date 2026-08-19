# Frida Script Templates for Android Testing

> **AI LOAD INSTRUCTION**: Load this when you need ready-to-use Frida scripts for Android application testing. Covers SSL pinning bypass, root detection bypass, method tracing, crypto hooks, and WebView debugging. Assumes the main [SKILL.md](./SKILL.md) is already loaded for general Android testing methodology.

---

## 1. UNIVERSAL SSL PINNING BYPASS

```javascript
// frida_ssl_bypass.js — hooks multiple SSL validation points
Java.perform(function() {

    // Hook TrustManager
    var TrustManager = Java.registerClass({
        name: 'com.frida.TrustManager',
        implements: [Java.use('javax.net.ssl.X509TrustManager')],
        methods: {
            checkClientTrusted: function(chain, authType) {},
            checkServerTrusted: function(chain, authType) {},
            getAcceptedIssuers: function() { return []; }
        }
    });

    var SSLContext = Java.use('javax.net.ssl.SSLContext');
    SSLContext.init.overload(
        '[Ljavax.net.ssl.KeyManager;',
        '[Ljavax.net.ssl.TrustManager;',
        'java.security.SecureRandom'
    ).implementation = function(km, tm, sr) {
        this.init(km, [TrustManager.$new()], sr);
    };

    // Hook OkHttp3 CertificatePinner
    try {
        var CertPinner = Java.use('okhttp3.CertificatePinner');
        CertPinner.check.overload('java.lang.String', 'java.util.List')
            .implementation = function(hostname, peerCerts) {
            return; // skip pinning check
        };
    } catch(e) {}

    // Hook OkHttp3 CertificatePinner.check$okhttp
    try {
        var CertPinner = Java.use('okhttp3.CertificatePinner');
        CertPinner['check$okhttp'].implementation = function(hostname, fn) {
            return;
        };
    } catch(e) {}

    // Hook HttpsURLConnection
    try {
        var HttpsConn = Java.use('javax.net.ssl.HttpsURLConnection');
        HttpsConn.setSSLSocketFactory.implementation = function(factory) {
            return; // prevent custom factory
        };
        HttpsConn.setHostnameVerifier.implementation = function(verifier) {
            return; // prevent custom verifier
        };
    } catch(e) {}

    console.log('[+] SSL Pinning bypassed');
});
```

---

## 2. ROOT DETECTION BYPASS

```javascript
// frida_root_bypass.js — comprehensive root detection bypass
Java.perform(function() {

    // Hook File.exists for su binary checks
    var File = Java.use('java.io.File');
    var rootPaths = ['/system/xbin/su', '/sbin/su', '/system/bin/su',
        '/system/su', '/data/local/xbin/su', '/data/local/bin/su',
        '/su/bin/su', '/magisk/.core/bin/su'];

    File.exists.implementation = function() {
        var path = this.getAbsolutePath();
        for (var i = 0; i < rootPaths.length; i++) {
            if (path === rootPaths[i]) {
                console.log('[Root Bypass] File.exists blocked: ' + path);
                return false;
            }
        }
        return this.exists();
    };

    // Hook Build.TAGS
    var Build = Java.use('android.os.Build');
    Build.TAGS.value = 'release-keys';

    // Hook PackageManager for Magisk/SuperSU detection
    var PM = Java.use('android.app.ApplicationPackageManager');
    var rootPackages = ['com.topjohnwu.magisk', 'eu.chainfire.supersu',
        'com.koushikdutta.superuser', 'com.noshufou.android.su'];

    PM.getPackageInfo.overload('java.lang.String', 'int').implementation = function(name, flags) {
        for (var i = 0; i < rootPackages.length; i++) {
            if (name === rootPackages[i]) {
                console.log('[Root Bypass] PM.getPackageInfo blocked: ' + name);
                throw Java.use('android.content.pm.PackageManager$NameNotFoundException').$new(name);
            }
        }
        return this.getPackageInfo(name, flags);
    };

    // Hook Runtime.exec for "which su" checks
    var Runtime = Java.use('java.lang.Runtime');
    Runtime.exec.overload('java.lang.String').implementation = function(cmd) {
        if (cmd.indexOf('su') !== -1 || cmd.indexOf('which') !== -1) {
            console.log('[Root Bypass] Runtime.exec blocked: ' + cmd);
            throw Java.use('java.io.IOException').$new('Command not found');
        }
        return this.exec(cmd);
    };

    // Hook System.getProperty for ro.debuggable etc.
    var System = Java.use('java.lang.System');
    System.getProperty.overload('java.lang.String').implementation = function(key) {
        if (key === 'ro.debuggable') return '0';
        if (key === 'ro.secure') return '1';
        return this.getProperty(key);
    };

    console.log('[+] Root detection bypassed');
});
```

---

## 3. METHOD TRACER

```javascript
// frida_trace_methods.js — trace all methods of a class
Java.perform(function() {

    function traceClass(className) {
        var clazz = Java.use(className);
        var methods = clazz.class.getDeclaredMethods();

        methods.forEach(function(method) {
            var methodName = method.getName();
            var overloads = clazz[methodName].overloads;

            overloads.forEach(function(overload) {
                overload.implementation = function() {
                    var args = Array.prototype.slice.call(arguments);
                    console.log('[TRACE] ' + className + '.' + methodName +
                        '(' + args.map(function(a) { return String(a); }).join(', ') + ')');

                    var ret = this[methodName].apply(this, arguments);
                    console.log('[TRACE] ' + className + '.' + methodName +
                        ' => ' + String(ret));
                    return ret;
                };
            });
        });

        console.log('[+] Tracing ' + methods.length + ' methods in ' + className);
    }

    // Usage: trace target class
    traceClass('com.target.app.api.AuthService');
});
```

---

## 4. CRYPTO HOOK (KEY & PLAINTEXT EXTRACTION)

```javascript
// frida_crypto_hooks.js — capture encryption keys and plaintext
Java.perform(function() {

    // Hook Cipher.init to capture keys
    var Cipher = Java.use('javax.crypto.Cipher');
    Cipher.init.overload('int', 'java.security.Key').implementation = function(mode, key) {
        var algo = this.getAlgorithm();
        var keyBytes = key.getEncoded();
        console.log('[Crypto] Cipher.init: algo=' + algo +
            ' mode=' + (mode === 1 ? 'ENCRYPT' : 'DECRYPT') +
            ' key=' + bytesToHex(keyBytes));
        return this.init(mode, key);
    };

    // Hook Cipher.doFinal to capture plaintext/ciphertext
    Cipher.doFinal.overload('[B').implementation = function(data) {
        console.log('[Crypto] Cipher.doFinal input (' + data.length + ' bytes): ' +
            bytesToHex(data));
        var result = this.doFinal(data);
        console.log('[Crypto] Cipher.doFinal output (' + result.length + ' bytes): ' +
            bytesToHex(result));
        return result;
    };

    // Hook SecretKeySpec constructor
    var SecretKeySpec = Java.use('javax.crypto.spec.SecretKeySpec');
    SecretKeySpec.$init.overload('[B', 'java.lang.String').implementation = function(key, algo) {
        console.log('[Crypto] SecretKeySpec: algo=' + algo + ' key=' + bytesToHex(key));
        return this.$init(key, algo);
    };

    // Hook MessageDigest (hash operations)
    var MessageDigest = Java.use('java.security.MessageDigest');
    MessageDigest.digest.overload('[B').implementation = function(data) {
        var algo = this.getAlgorithm();
        console.log('[Crypto] ' + algo + '.digest input: ' + bytesToHex(data));
        var result = this.digest(data);
        console.log('[Crypto] ' + algo + '.digest output: ' + bytesToHex(result));
        return result;
    };

    function bytesToHex(bytes) {
        if (!bytes) return 'null';
        var hex = [];
        for (var i = 0; i < bytes.length; i++) {
            hex.push(('0' + (bytes[i] & 0xFF).toString(16)).slice(-2));
        }
        return hex.join('');
    }

    console.log('[+] Crypto hooks installed');
});
```

---

## 5. WEBVIEW DEBUGGING ENABLER

```javascript
// frida_webview_debug.js — enable WebView debugging for all WebViews
Java.perform(function() {

    // Enable WebView debugging globally
    var WebView = Java.use('android.webkit.WebView');
    WebView.setWebContentsDebuggingEnabled(true);

    // Hook loadUrl to log all URLs loaded
    WebView.loadUrl.overload('java.lang.String').implementation = function(url) {
        console.log('[WebView] loadUrl: ' + url);
        return this.loadUrl(url);
    };

    WebView.loadUrl.overload('java.lang.String', 'java.util.Map').implementation = function(url, headers) {
        console.log('[WebView] loadUrl: ' + url + ' headers: ' + headers);
        return this.loadUrl(url, headers);
    };

    // Hook evaluateJavascript
    WebView.evaluateJavascript.implementation = function(script, callback) {
        console.log('[WebView] evaluateJavascript: ' + script);
        return this.evaluateJavascript(script, callback);
    };

    // Hook addJavascriptInterface to log bridge objects
    WebView.addJavascriptInterface.implementation = function(obj, name) {
        console.log('[WebView] addJavascriptInterface: name=' + name +
            ' class=' + obj.getClass().getName());
        return this.addJavascriptInterface(obj, name);
    };

    console.log('[+] WebView debugging enabled');
});
```

---

## 6. SHARED PREFERENCES DUMPER

```javascript
// frida_dump_prefs.js — dump all SharedPreferences
Java.perform(function() {
    var ctx = Java.use('android.app.ActivityThread').currentApplication().getApplicationContext();
    var prefsDir = ctx.getFilesDir().getParent() + '/shared_prefs/';

    var File = Java.use('java.io.File');
    var dir = File.$new(prefsDir);
    var files = dir.listFiles();

    if (files) {
        for (var i = 0; i < files.length; i++) {
            var name = files[i].getName().replace('.xml', '');
            var prefs = ctx.getSharedPreferences(name, 0);
            var all = prefs.getAll();
            var entries = all.entrySet().iterator();

            console.log('\n[SharedPrefs] === ' + name + ' ===');
            while (entries.hasNext()) {
                var entry = entries.next();
                console.log('  ' + entry.getKey() + ' = ' + entry.getValue());
            }
        }
    }
});
```

---

## 7. INTENT SNIFFER

```javascript
// frida_intent_sniffer.js — log all intents sent/received
Java.perform(function() {

    var Activity = Java.use('android.app.Activity');
    Activity.startActivity.overload('android.content.Intent').implementation = function(intent) {
        console.log('[Intent] startActivity: ' + intent.toString());
        if (intent.getExtras()) {
            console.log('[Intent] extras: ' + intent.getExtras().toString());
        }
        return this.startActivity(intent);
    };

    Activity.startActivityForResult.overload('android.content.Intent', 'int')
        .implementation = function(intent, requestCode) {
        console.log('[Intent] startActivityForResult: ' + intent.toString() +
            ' requestCode=' + requestCode);
        return this.startActivityForResult(intent, requestCode);
    };

    var Context = Java.use('android.content.ContextWrapper');
    Context.sendBroadcast.overload('android.content.Intent').implementation = function(intent) {
        console.log('[Intent] sendBroadcast: ' + intent.toString());
        return this.sendBroadcast(intent);
    };

    console.log('[+] Intent sniffer installed');
});
```

---

## 8. BIOMETRIC BYPASS

```javascript
// frida_biometric_bypass.js — bypass fingerprint/face authentication
Java.perform(function() {

    // BiometricPrompt (API 28+)
    try {
        var BiometricCallback = Java.use('android.hardware.biometrics.BiometricPrompt$AuthenticationCallback');
        BiometricCallback.onAuthenticationSucceeded.implementation = function(result) {
            console.log('[Biometric] Forcing authentication success');
            this.onAuthenticationSucceeded(result);
        };
        BiometricCallback.onAuthenticationFailed.implementation = function() {
            console.log('[Biometric] Intercepting failure → forcing success');
            // Don't call original — suppress failure
        };
    } catch(e) {}

    // AndroidX BiometricPrompt
    try {
        var AxCallback = Java.use('androidx.biometric.BiometricPrompt$AuthenticationCallback');
        AxCallback.onAuthenticationFailed.implementation = function() {
            console.log('[Biometric] AndroidX failure suppressed');
        };
    } catch(e) {}

    // FingerprintManager (legacy, pre-API 28)
    try {
        var FPCallback = Java.use('android.hardware.fingerprint.FingerprintManager$AuthenticationCallback');
        FPCallback.onAuthenticationFailed.implementation = function() {
            console.log('[Biometric] Legacy FP failure suppressed');
        };
    } catch(e) {}

    console.log('[+] Biometric bypass hooks installed');
});
```

---

## 9. SCRIPT USAGE QUICK REFERENCE

| Script | Command | Use Case |
|---|---|---|
| SSL bypass | `frida -U -l frida_ssl_bypass.js -f com.target` | Intercept HTTPS traffic |
| Root bypass | `frida -U -l frida_root_bypass.js -f com.target` | Run on rooted device without detection |
| Method tracer | `frida -U -l frida_trace_methods.js -f com.target` | Understand app logic |
| Crypto hooks | `frida -U -l frida_crypto_hooks.js -f com.target` | Extract keys, see plaintext |
| WebView debug | `frida -U -l frida_webview_debug.js -f com.target` | Inspect WebView URLs and bridges |
| Prefs dumper | `frida -U -l frida_dump_prefs.js -f com.target` | Extract stored credentials/tokens |
| Intent sniffer | `frida -U -l frida_intent_sniffer.js -f com.target` | Monitor IPC |
| Biometric bypass | `frida -U -l frida_biometric_bypass.js -f com.target` | Skip fingerprint/face auth |

Common flags:
- `-f`: spawn app (fresh start)
- `-U`: USB-connected device
- `-F`: attach to frontmost app
- `--no-pause`: don't pause on spawn (auto-resume)
- `-H host:port`: remote Frida server
