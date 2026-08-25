#!/usr/bin/env node

/**
 * Cybermes MCP Server — Zero-Go NPX Launcher
 * https://github.com/Zyrexnn/Cybermes
 */

const fs = require('fs');
const path = require('path');
const os = require('os');
const https = require('https');
const { spawn } = require('child_process');

const VERSION = require('../package.json').version || '2.1.0';
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
    archName = 'arm64';
  } else {
    throw new Error(`Unsupported CPU architecture: ${arch}`);
  }

  const binaryFilename = `cybermes-mcp-v${VERSION}-${osName}-${archName}${ext}`;
  return { osName, archName, ext, binaryFilename };
}

function findLocalDevBinary() {
  const candidates = [
    path.resolve(__dirname, '..', '..', 'tools', 'bin', `cybermes-mcp${os.platform() === 'win32' ? '.exe' : ''}`),
    path.resolve(__dirname, '..', '..', 'cybermes-mcp' + (os.platform() === 'win32' ? '.exe' : '')),
  ];

  for (const candidate of candidates) {
    if (fs.existsSync(candidate)) {
      return candidate;
    }
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

async function getOrDownloadBinary() {
  // 1. Check local repo build if available
  const localDevBin = findLocalDevBinary();
  if (localDevBin) {
    return localDevBin;
  }

  // 2. Check local user cache ~/.cybermes/bin/
  const { binaryFilename } = getPlatformInfo();
  const cacheDir = path.join(os.homedir(), '.cybermes', 'bin');
  const cachedBinPath = path.join(cacheDir, binaryFilename);

  if (fs.existsSync(cachedBinPath)) {
    return cachedBinPath;
  }

  // 3. Download from GitHub Releases
  if (!fs.existsSync(cacheDir)) {
    fs.mkdirSync(cacheDir, { recursive: true });
  }

  const downloadUrl = `https://github.com/${REPO}/releases/download/v${VERSION}/${binaryFilename}`;
  return await downloadBinary(downloadUrl, cachedBinPath);
}

async function main() {
  try {
    const binPath = await getOrDownloadBinary();

    const child = spawn(binPath, process.argv.slice(2), {
      stdio: 'inherit',
      env: process.env,
    });

    child.on('error', (err) => {
      process.stderr.write(`[cybermes-mcp] Failed to spawn child process: ${err.message}\n`);
      process.exit(1);
    });

    child.on('exit', (code, signal) => {
      if (signal) {
        process.kill(process.pid, signal);
      } else {
        process.exit(code || 0);
      }
    });

    // Forward signals
    process.on('SIGINT', () => child.kill('SIGINT'));
    process.on('SIGTERM', () => child.kill('SIGTERM'));

  } catch (err) {
    process.stderr.write(`[cybermes-mcp] Error: ${err.message}\n`);
    process.exit(1);
  }
}

main();
