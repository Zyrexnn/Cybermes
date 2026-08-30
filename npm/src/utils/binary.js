/**
 * npm/src/utils/binary.js
 * Platform detection, local compilation lookup, and GitHub release binary management.
 * 100% Zero external dependencies.
 */

const fs = require('fs');
const path = require('path');
const os = require('os');
const https = require('https');
const { spawnSync } = require('child_process');
const { ANSI } = require('./ui');

const VERSION = require('../../package.json').version || '3.3.0';
const REPO = 'Zyrexnn/Cybermes';

function getPlatformInfo() {
  const platform = os.platform();
  const arch = os.arch();

  let osName = '';
  let archName = '';
  let ext = '';

  if (platform === 'win32') {
    osName = 'windows';
    ext = '.exe';
  } else if (platform === 'linux') {
    osName = 'linux';
  } else if (platform === 'darwin') {
    osName = 'darwin';
  } else {
    throw new Error(`Unsupported OS platform: ${platform}`);
  }

  if (arch === 'x64') {
    archName = 'amd64';
  } else if (arch === 'arm64') {
    archName = platform === 'win32' ? 'amd64' : 'arm64';
  } else {
    throw new Error(`Unsupported CPU architecture: ${arch}`);
  }

  const binaryFilename = `cybermes-mcp-v${VERSION}-${osName}-${archName}${ext}`;
  return { osName, archName, ext, binaryFilename, platform, arch };
}

function findLocalDevBinary() {
  const isWin = os.platform() === 'win32';
  const binName = `cybermes-mcp${isWin ? '.exe' : ''}`;

  const candidates = [
    path.resolve(__dirname, '..', '..', '..', 'tools', 'bin', binName),
    path.resolve(__dirname, '..', '..', '..', binName),
    path.resolve(process.cwd(), 'tools', 'bin', binName),
    path.resolve(process.cwd(), binName),
  ];

  for (const candidate of candidates) {
    try {
      if (fs.existsSync(candidate) && fs.statSync(candidate).isFile()) {
        return candidate;
      }
    } catch (_) {}
  }
  return null;
}

function downloadBinary(url, destPath) {
  return new Promise((resolve, reject) => {
    process.stderr.write(`[cybermes-mcp] Downloading native binary from GitHub Releases...\n`);
    process.stderr.write(`[cybermes-mcp] URL: ${url}\n`);

    function get(currentUrl, redirectCount = 0) {
      if (redirectCount > 5) {
        return reject(new Error('Too many HTTP redirects'));
      }

      https.get(currentUrl, (res) => {
        if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
          return get(res.headers.location, redirectCount + 1);
        }

        if (res.statusCode !== 200) {
          return reject(new Error(`Failed to download binary: HTTP ${res.statusCode} ${res.statusMessage}`));
        }

        const tempPath = `${destPath}.tmp.${Date.now()}`;
        const fileStream = fs.createWriteStream(tempPath);

        res.pipe(fileStream);

        fileStream.on('finish', () => {
          fileStream.close(() => {
            try {
              if (os.platform() !== 'win32') {
                fs.chmodSync(tempPath, 0o755);
              }
              fs.renameSync(tempPath, destPath);
              process.stderr.write(`[cybermes-mcp] Download complete! Saved to ${destPath}\n`);
              resolve(destPath);
            } catch (err) {
              reject(err);
            }
          });
        });

        fileStream.on('error', (err) => {
          try { fs.unlinkSync(tempPath); } catch (_) {}
          reject(err);
        });
      }).on('error', (err) => {
        reject(err);
      });
    }

    get(url);
  });
}

async function getOrDownloadBinary(preferMode = 'auto') {
  if (preferMode !== 'release') {
    const localDevBin = findLocalDevBinary();
    if (localDevBin) {
      return localDevBin;
    }
  }

  const { binaryFilename } = getPlatformInfo();
  const cacheDir = path.join(os.homedir(), '.cybermes', 'bin');
  const cachedBinPath = path.join(cacheDir, binaryFilename);

  if (fs.existsSync(cachedBinPath)) {
    try {
      if (fs.statSync(cachedBinPath).isFile()) {
        if (os.platform() !== 'win32') {
          fs.chmodSync(cachedBinPath, 0o755);
        }
        return cachedBinPath;
      }
    } catch (_) {}
  }

  if (!fs.existsSync(cacheDir)) {
    fs.mkdirSync(cacheDir, { recursive: true });
  }

  const downloadUrl = `https://github.com/${REPO}/releases/download/v${VERSION}/${binaryFilename}`;
  return await downloadBinary(downloadUrl, cachedBinPath);
}

function testBinaryExecution(binaryPath) {
  if (!binaryPath || !fs.existsSync(binaryPath)) {
    return { ok: false, error: 'Binary path does not exist' };
  }
  try {
    const res = spawnSync(binaryPath, ['-h'], { encoding: 'utf8', timeout: 3000 });
    return { ok: true, output: res.stdout || res.stderr || 'Ready' };
  } catch (err) {
    return { ok: false, error: err.message };
  }
}

module.exports = {
  VERSION,
  REPO,
  getPlatformInfo,
  findLocalDevBinary,
  getOrDownloadBinary,
  testBinaryExecution
};
