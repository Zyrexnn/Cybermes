/**
 * npm/src/utils/updater.js
 * Zero-dependency npm registry version checker and semver comparator.
 * Uses native Node.js https with 24-hour disk cache cooldown.
 */

const https = require('https');
const fs = require('fs');
const path = require('path');
const os = require('os');
const { VERSION } = require('./binary');

const CACHE_FILE = path.join(os.homedir(), '.cybermes', 'update_check.json');
const COOLDOWN_MS = 24 * 60 * 60 * 1000;

function compareSemver(current, latest) {
  if (!current || !latest) return false;
  const parse = (v) => v.replace(/^v/, '').split('.').map((p) => {
    const n = parseInt(p, 10);
    return Number.isFinite(n) ? n : 0;
  });

  const c = parse(current);
  const l = parse(latest);

  for (let i = 0; i < 3; i++) {
    const currPart = c[i] || 0;
    const latestPart = l[i] || 0;
    if (latestPart > currPart) return true;
    if (latestPart < currPart) return false;
  }
  return false;
}

function fetchLatestNpmVersion(packageName = 'cybermes-mcp', timeoutMs = 3500) {
  return new Promise((resolve, reject) => {
    const url = `https://registry.npmjs.org/${packageName}/latest`;
    const req = https.get(
      url,
      { headers: { 'User-Agent': `cybermes-mcp-updater/${VERSION}` }, timeout: timeoutMs },
      (res) => {
        if (res.statusCode !== 200) {
          res.resume();
          return reject(new Error(`NPM registry returned HTTP ${res.statusCode}`));
        }
        let raw = '';
        res.on('data', (chunk) => { raw += chunk; });
        res.on('end', () => {
          try {
            const data = JSON.parse(raw);
            if (data && typeof data.version === 'string') {
              resolve(data.version);
            } else {
              reject(new Error('Invalid response structure from npm registry'));
            }
          } catch (e) {
            reject(e);
          }
        });
      }
    );

    req.on('timeout', () => {
      req.destroy();
      reject(new Error('NPM registry request timed out'));
    });

    req.on('error', reject);
  });
}

async function checkUpdate(force = false) {
  try {
    if (!force && fs.existsSync(CACHE_FILE)) {
      const cached = JSON.parse(fs.readFileSync(CACHE_FILE, 'utf8'));
      if (Date.now() - cached.lastChecked < COOLDOWN_MS) {
        return cached.updateAvailable ? cached : null;
      }
    }

    const latest = await fetchLatestNpmVersion();
    const hasUpdate = compareSemver(VERSION, latest);
    const result = {
      currentVersion: VERSION,
      latestVersion: latest,
      updateAvailable: hasUpdate,
      lastChecked: Date.now(),
    };

    const dir = path.dirname(CACHE_FILE);
    if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });
    fs.writeFileSync(CACHE_FILE, JSON.stringify(result, null, 2));

    return hasUpdate ? result : null;
  } catch (_) {
    // Fail silently when offline or unreachable without halting execution.
    return null;
  }
}

module.exports = {
  compareSemver,
  fetchLatestNpmVersion,
  checkUpdate,
  CACHE_FILE,
  COOLDOWN_MS,
};
