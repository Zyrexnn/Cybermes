#!/usr/bin/env node

/**
 * Cybermes MCP Server — Zero-Go NPX Launcher & Universal Auto-Injector
 * https://github.com/Zyrexnn/Cybermes
 */

const fs = require('fs');
const path = require('path');
const os = require('os');
const https = require('https');
const { spawn } = require('child_process');

const VERSION = require('../package.json').version || '3.0.0';
const REPO = 'Zyrexnn/Cybermes';

// ============================================================================
// 1. Platform & Binary Helpers
// ============================================================================

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
    path.resolve(process.cwd(), 'tools', 'bin', `cybermes-mcp${os.platform() === 'win32' ? '.exe' : ''}`),
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
  const localDevBin = findLocalDevBinary();
  if (localDevBin) {
    return localDevBin;
  }

  const { binaryFilename } = getPlatformInfo();
  const cacheDir = path.join(os.homedir(), '.cybermes', 'bin');
  const cachedBinPath = path.join(cacheDir, binaryFilename);

  if (fs.existsSync(cachedBinPath)) {
    return cachedBinPath;
  }

  if (!fs.existsSync(cacheDir)) {
    fs.mkdirSync(cacheDir, { recursive: true });
  }

  const downloadUrl = `https://github.com/${REPO}/releases/download/v${VERSION}/${binaryFilename}`;
  return await downloadBinary(downloadUrl, cachedBinPath);
}

// ============================================================================
// 2. Client Auto-Injector & Configuration Definitions
// ============================================================================

function createBackup(filePath) {
  if (fs.existsSync(filePath)) {
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    const backupPath = `${filePath}.bak-${timestamp}`;
    fs.copyFileSync(filePath, backupPath);
    return backupPath;
  }
  return null;
}

function safeReadJson(filePath) {
  if (!fs.existsSync(filePath)) return null;
  try {
    const content = fs.readFileSync(filePath, 'utf8').trim();
    if (!content) return {};
    return JSON.parse(content);
  } catch (err) {
    return { _parseError: err.message };
  }
}

function safeWriteJson(filePath, data, isDryRun) {
  if (isDryRun) return true;
  const dir = path.dirname(filePath);
  if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true });
  }
  fs.writeFileSync(filePath, JSON.stringify(data, null, 2) + '\n', 'utf8');
  return true;
}

function getAppdataDir() {
  return process.env.APPDATA || path.join(os.homedir(), 'AppData', 'Roaming');
}

function getTargetDefinition(useLocal, localBinPath, workspaceRoot, useGlobal) {
  if (useLocal && localBinPath) {
    return {
      command: localBinPath,
      args: ['-workspace', workspaceRoot || path.resolve(__dirname, '..', '..')],
    };
  }
  if (useGlobal) {
    return {
      command: 'cybermes-mcp',
      args: [],
    };
  }
  return {
    command: 'npx',
    args: ['-y', 'cybermes-mcp'],
  };
}

function getClientDefinitions(useLocal, localBinPath, workspaceRoot, useGlobal) {
  const home = os.homedir();
  const appdata = getAppdataDir();
  const isWin = os.platform() === 'win32';
  const isMac = os.platform() === 'darwin';

  const defaultDef = getTargetDefinition(useLocal, localBinPath, workspaceRoot, useGlobal);

  return [
    {
      id: 'gemini',
      name: 'Antigravity / Gemini',
      paths: [
        path.join(home, '.gemini', 'config', 'mcp_config.json'),
      ],
      type: 'json-mcpServers',
      definition: defaultDef,
    },
    {
      id: 'claude-desktop',
      name: 'Claude Desktop',
      paths: [
        isWin
          ? path.join(appdata, 'Claude', 'claude_desktop_config.json')
          : isMac
          ? path.join(home, 'Library', 'Application Support', 'Claude', 'claude_desktop_config.json')
          : path.join(home, '.config', 'Claude', 'claude_desktop_config.json'),
      ],
      type: 'json-mcpServers',
      definition: defaultDef,
    },
    {
      id: 'cursor',
      name: 'Cursor IDE',
      paths: [
        path.join(home, '.cursor', 'mcp.json'),
        path.join(process.cwd(), '.cursor', 'mcp.json'),
      ],
      type: 'json-mcpServers',
      definition: defaultDef,
    },
    {
      id: 'opencode',
      name: 'OpenCode Interpreter',
      paths: [
        path.join(home, '.config', 'opencode', 'opencode.json'),
        path.join(home, '.config', 'opencode', 'config.json'),
      ],
      type: 'json-mcp_servers',
      definition: defaultDef,
    },
    {
      id: 'windsurf',
      name: 'Windsurf IDE (Codeium)',
      paths: [
        path.join(home, '.codeium', 'windsurf', 'mcp_config.json'),
      ],
      type: 'json-mcpServers',
      definition: defaultDef,
    },
    {
      id: 'cline',
      name: 'Cline (VS Code Extension)',
      paths: [
        isWin
          ? path.join(appdata, 'Code', 'User', 'globalStorage', 'saoudrizwan.claude-dev', 'settings', 'cline_mcp_settings.json')
          : isMac
          ? path.join(home, 'Library', 'Application Support', 'Code', 'User', 'globalStorage', 'saoudrizwan.claude-dev', 'settings', 'cline_mcp_settings.json')
          : path.join(home, '.config', 'Code', 'User', 'globalStorage', 'saoudrizwan.claude-dev', 'settings', 'cline_mcp_settings.json'),
      ],
      type: 'json-cline',
      definition: {
        ...defaultDef,
        disabled: false,
        autoApprove: [
          'cybermes_search_knowledge',
          'cybermes_list_skills',
          'cybermes_get_skill',
          'cybermes_scan_secrets',
          'cybermes_validate_scope',
        ],
      },
    },
    {
      id: 'roo-code',
      name: 'Roo Code (VS Code Extension)',
      paths: [
        isWin
          ? path.join(appdata, 'Code', 'User', 'globalStorage', 'rooveterinaryinc.roo-cline', 'settings', 'cline_mcp_settings.json')
          : isMac
          ? path.join(home, 'Library', 'Application Support', 'Code', 'User', 'globalStorage', 'rooveterinaryinc.roo-cline', 'settings', 'cline_mcp_settings.json')
          : path.join(home, '.config', 'Code', 'User', 'globalStorage', 'rooveterinaryinc.roo-cline', 'settings', 'cline_mcp_settings.json'),
      ],
      type: 'json-cline',
      definition: {
        ...defaultDef,
        disabled: false,
        autoApprove: [
          'cybermes_search_knowledge',
          'cybermes_list_skills',
          'cybermes_get_skill',
          'cybermes_scan_secrets',
          'cybermes_validate_scope',
        ],
      },
    },
    {
      id: 'claude-code',
      name: 'Claude Code CLI',
      paths: [
        path.join(home, '.claude.json'),
      ],
      type: 'json-mcpServers',
      definition: defaultDef,
    },
    {
      id: 'continue',
      name: 'Continue.dev',
      paths: [
        path.join(home, '.continue', 'config.json'),
      ],
      type: 'json-continue',
      definition: {
        name: 'cybermes',
        transport: {
          type: 'stdio',
          command: defaultDef.command,
          args: defaultDef.args,
        },
      },
    },
    {
      id: 'zed',
      name: 'Zed Editor',
      paths: [
        path.join(home, '.config', 'zed', 'settings.json'),
        isWin ? path.join(appdata, 'Zed', 'settings.json') : null,
      ].filter(Boolean),
      type: 'json-zed',
      definition: defaultDef,
    },
    {
      id: 'kilo',
      name: 'Kilo Code',
      paths: [
        path.join(home, '.kilo', 'mcp.json'),
      ],
      type: 'json-mcpServers',
      definition: defaultDef,
    },
    {
      id: 'hermes',
      name: 'Hermes Agent',
      paths: [
        path.join(home, '.hermes', 'config.yaml'),
        path.join(process.cwd(), '.hermes', 'config.yaml'),
      ],
      type: 'yaml-hermes',
      definition: defaultDef,
    },
    {
      id: 'codex',
      name: 'Codex CLI',
      paths: [
        path.join(home, '.codex', 'config.toml'),
      ],
      type: 'toml-codex',
      definition: defaultDef,
    },
  ];
}

// ============================================================================
// 3. Injector Engine: Install, Uninstall, Status
// ============================================================================

function injectClientConfig(client, filePath, isDryRun) {
  const result = { client: client.name, path: filePath, status: 'unknown', details: '' };

  if (client.type === 'json-mcpServers' || client.type === 'json-cline') {
    let json = safeReadJson(filePath);
    if (json && json._parseError) {
      result.status = 'error';
      result.details = `Invalid JSON syntax: ${json._parseError}`;
      return result;
    }
    if (!json) json = {};

    json.mcpServers = json.mcpServers || {};
    const existing = json.mcpServers.cybermes;
    if (existing && JSON.stringify(existing) === JSON.stringify(client.definition)) {
      result.status = 'unchanged';
      result.details = 'Already up-to-date';
      return result;
    }

    json.mcpServers.cybermes = client.definition;

    if (!isDryRun) {
      const bak = createBackup(filePath);
      safeWriteJson(filePath, json, false);
      result.status = 'injected';
      result.details = bak ? `Updated (backup: ${path.basename(bak)})` : 'Created new config';
    } else {
      result.status = 'dry-run';
      result.details = 'Would inject mcpServers.cybermes';
    }
    return result;
  }

  if (client.type === 'json-mcp_servers' || client.type === 'json-opencode') {
    let json = safeReadJson(filePath);
    if (json && json._parseError) {
      result.status = 'error';
      result.details = `Invalid JSON syntax: ${json._parseError}`;
      return result;
    }
    if (!json) json = {};

    const opencodeEntry = {
      type: 'local',
      command: [client.definition.command, ...client.definition.args],
      enabled: true,
    };

    if (json.mcp || (!json.mcp_servers && !json.mcpServers)) {
      json.mcp = json.mcp || {};
      const existing = json.mcp.cybermes;
      if (existing && JSON.stringify(existing) === JSON.stringify(opencodeEntry)) {
        result.status = 'unchanged';
        result.details = 'Already up-to-date';
        return result;
      }
      json.mcp.cybermes = opencodeEntry;
    } else {
      json.mcp_servers = json.mcp_servers || {};
      const existing = json.mcp_servers.cybermes;
      if (existing && JSON.stringify(existing) === JSON.stringify(client.definition)) {
        result.status = 'unchanged';
        result.details = 'Already up-to-date';
        return result;
      }
      json.mcp_servers.cybermes = client.definition;
    }

    if (!isDryRun) {
      const bak = createBackup(filePath);
      safeWriteJson(filePath, json, false);
      result.status = 'injected';
      result.details = bak ? `Updated (backup: ${path.basename(bak)})` : 'Created new config';
    } else {
      result.status = 'dry-run';
      result.details = 'Would inject mcp.cybermes';
    }
    return result;
  }

  if (client.type === 'json-continue') {
    let json = safeReadJson(filePath);
    if (json && json._parseError) {
      result.status = 'error';
      result.details = `Invalid JSON syntax: ${json._parseError}`;
      return result;
    }
    if (!json) json = {};

    json.experimental = json.experimental || {};
    json.experimental.modelContextProtocolServers = json.experimental.modelContextProtocolServers || [];

    const idx = json.experimental.modelContextProtocolServers.findIndex(s => s.name === 'cybermes');
    if (idx >= 0 && JSON.stringify(json.experimental.modelContextProtocolServers[idx]) === JSON.stringify(client.definition)) {
      result.status = 'unchanged';
      result.details = 'Already up-to-date';
      return result;
    }

    if (idx >= 0) {
      json.experimental.modelContextProtocolServers[idx] = client.definition;
    } else {
      json.experimental.modelContextProtocolServers.push(client.definition);
    }

    if (!isDryRun) {
      const bak = createBackup(filePath);
      safeWriteJson(filePath, json, false);
      result.status = 'injected';
      result.details = bak ? `Updated (backup: ${path.basename(bak)})` : 'Created new config';
    } else {
      result.status = 'dry-run';
      result.details = 'Would inject experimental.modelContextProtocolServers entry';
    }
    return result;
  }

  if (client.type === 'json-zed') {
    let json = safeReadJson(filePath);
    if (json && json._parseError) {
      result.status = 'error';
      result.details = `Invalid JSON syntax: ${json._parseError}`;
      return result;
    }
    if (!json) json = {};

    json.context_servers = json.context_servers || {};
    const existing = json.context_servers.cybermes;
    if (existing && JSON.stringify(existing) === JSON.stringify(client.definition)) {
      result.status = 'unchanged';
      result.details = 'Already up-to-date';
      return result;
    }

    json.context_servers.cybermes = client.definition;

    if (!isDryRun) {
      const bak = createBackup(filePath);
      safeWriteJson(filePath, json, false);
      result.status = 'injected';
      result.details = bak ? `Updated (backup: ${path.basename(bak)})` : 'Created new config';
    } else {
      result.status = 'dry-run';
      result.details = 'Would inject context_servers.cybermes';
    }
    return result;
  }

  if (client.type === 'yaml-hermes') {
    let existingContent = '';
    if (fs.existsSync(filePath)) {
      existingContent = fs.readFileSync(filePath, 'utf8');
    }
    if (existingContent.includes('cybermes:') && (existingContent.includes('@zyrexnn/cybermes-mcp') || existingContent.includes('cybermes-mcp'))) {
      result.status = 'unchanged';
      result.details = 'Already up-to-date in YAML';
      return result;
    }

    const cmdStr = JSON.stringify(client.definition.command);
    const argsStr = JSON.stringify(client.definition.args);
    const block = `\nmcp_servers:\n  cybermes:\n    command: ${cmdStr}\n    args: ${argsStr}\n`;

    if (!isDryRun) {
      const bak = createBackup(filePath);
      const dir = path.dirname(filePath);
      if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });
      fs.appendFileSync(filePath, block, 'utf8');
      result.status = 'injected';
      result.details = bak ? `Appended (backup: ${path.basename(bak)})` : 'Created config with mcp_servers.cybermes';
    } else {
      result.status = 'dry-run';
      result.details = 'Would append YAML block';
    }
    return result;
  }

  if (client.type === 'toml-codex') {
    let existingContent = '';
    if (fs.existsSync(filePath)) {
      existingContent = fs.readFileSync(filePath, 'utf8');
    }
    if (existingContent.includes('[mcp_servers.cybermes]')) {
      result.status = 'unchanged';
      result.details = 'Already up-to-date in TOML';
      return result;
    }

    const cmdStr = JSON.stringify(client.definition.command);
    const argsArr = client.definition.args.map(a => JSON.stringify(a)).join(', ');
    const block = `\n[mcp_servers.cybermes]\ncommand = ${cmdStr}\nargs = [${argsArr}]\n`;

    if (!isDryRun) {
      const bak = createBackup(filePath);
      const dir = path.dirname(filePath);
      if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });
      fs.appendFileSync(filePath, block, 'utf8');
      result.status = 'injected';
      result.details = bak ? `Appended (backup: ${path.basename(bak)})` : 'Created config with [mcp_servers.cybermes]';
    } else {
      result.status = 'dry-run';
      result.details = 'Would append TOML block';
    }
    return result;
  }

  return result;
}

function removeClientConfig(client, filePath, isDryRun) {
  const result = { client: client.name, path: filePath, status: 'unknown', details: '' };
  if (!fs.existsSync(filePath)) {
    result.status = 'not_found';
    result.details = 'File does not exist';
    return result;
  }

  if (client.type === 'json-mcpServers' || client.type === 'json-cline') {
    const json = safeReadJson(filePath);
    if (!json || json._parseError || !json.mcpServers || !json.mcpServers.cybermes) {
      result.status = 'unchanged';
      result.details = 'Cybermes not present';
      return result;
    }
    delete json.mcpServers.cybermes;
    if (!isDryRun) {
      const bak = createBackup(filePath);
      safeWriteJson(filePath, json, false);
      result.status = 'removed';
      result.details = bak ? `Cleaned (backup: ${path.basename(bak)})` : 'Removed';
    } else {
      result.status = 'dry-run';
      result.details = 'Would remove mcpServers.cybermes';
    }
    return result;
  }

  if (client.type === 'json-mcp_servers') {
    const json = safeReadJson(filePath);
    if (!json || json._parseError || !json.mcp_servers || !json.mcp_servers.cybermes) {
      result.status = 'unchanged';
      result.details = 'Cybermes not present';
      return result;
    }
    delete json.mcp_servers.cybermes;
    if (!isDryRun) {
      const bak = createBackup(filePath);
      safeWriteJson(filePath, json, false);
      result.status = 'removed';
      result.details = bak ? `Cleaned (backup: ${path.basename(bak)})` : 'Removed';
    } else {
      result.status = 'dry-run';
      result.details = 'Would remove mcp_servers.cybermes';
    }
    return result;
  }

  if (client.type === 'json-continue') {
    const json = safeReadJson(filePath);
    if (!json || json._parseError || !json.experimental || !Array.isArray(json.experimental.modelContextProtocolServers)) {
      result.status = 'unchanged';
      result.details = 'Cybermes not present';
      return result;
    }
    const initialLen = json.experimental.modelContextProtocolServers.length;
    json.experimental.modelContextProtocolServers = json.experimental.modelContextProtocolServers.filter(s => s.name !== 'cybermes');
    if (json.experimental.modelContextProtocolServers.length === initialLen) {
      result.status = 'unchanged';
      result.details = 'Cybermes not present';
      return result;
    }
    if (!isDryRun) {
      const bak = createBackup(filePath);
      safeWriteJson(filePath, json, false);
      result.status = 'removed';
      result.details = bak ? `Cleaned (backup: ${path.basename(bak)})` : 'Removed';
    } else {
      result.status = 'dry-run';
      result.details = 'Would remove from experimental.modelContextProtocolServers';
    }
    return result;
  }

  if (client.type === 'json-zed') {
    const json = safeReadJson(filePath);
    if (!json || json._parseError || !json.context_servers || !json.context_servers.cybermes) {
      result.status = 'unchanged';
      result.details = 'Cybermes not present';
      return result;
    }
    delete json.context_servers.cybermes;
    if (!isDryRun) {
      const bak = createBackup(filePath);
      safeWriteJson(filePath, json, false);
      result.status = 'removed';
      result.details = bak ? `Cleaned (backup: ${path.basename(bak)})` : 'Removed';
    } else {
      result.status = 'dry-run';
      result.details = 'Would remove context_servers.cybermes';
    }
    return result;
  }

  result.status = 'skipped';
  result.details = 'Manual cleanup recommended for YAML/TOML';
  return result;
}

function checkClientStatus(client, filePath) {
  if (!fs.existsSync(filePath)) {
    return { client: client.name, path: filePath, installed: false, configured: false, details: 'Not installed' };
  }

  if (client.type === 'json-mcpServers' || client.type === 'json-cline') {
    const json = safeReadJson(filePath);
    const hasCybermes = Boolean(json && json.mcpServers && json.mcpServers.cybermes);
    return { client: client.name, path: filePath, installed: true, configured: hasCybermes, details: hasCybermes ? 'Configured' : 'Client detected (Cybermes missing)' };
  }

  if (client.type === 'json-mcp_servers') {
    const json = safeReadJson(filePath);
    const hasCybermes = Boolean(json && json.mcp_servers && json.mcp_servers.cybermes);
    return { client: client.name, path: filePath, installed: true, configured: hasCybermes, details: hasCybermes ? 'Configured' : 'Client detected (Cybermes missing)' };
  }

  if (client.type === 'json-continue') {
    const json = safeReadJson(filePath);
    const list = (json && json.experimental && json.experimental.modelContextProtocolServers) || [];
    const hasCybermes = list.some(s => s.name === 'cybermes');
    return { client: client.name, path: filePath, installed: true, configured: hasCybermes, details: hasCybermes ? 'Configured' : 'Client detected (Cybermes missing)' };
  }

  if (client.type === 'json-zed') {
    const json = safeReadJson(filePath);
    const hasCybermes = Boolean(json && json.context_servers && json.context_servers.cybermes);
    return { client: client.name, path: filePath, installed: true, configured: hasCybermes, details: hasCybermes ? 'Configured' : 'Client detected (Cybermes missing)' };
  }

  if (client.type === 'yaml-hermes' || client.type === 'toml-codex') {
    const content = fs.readFileSync(filePath, 'utf8');
    const hasCybermes = content.includes('cybermes');
    return { client: client.name, path: filePath, installed: true, configured: hasCybermes, details: hasCybermes ? 'Configured' : 'Client detected (Cybermes missing)' };
  }

  return { client: client.name, path: filePath, installed: true, configured: false, details: 'Unknown format' };
}

// ============================================================================
// 4. Modern Terminal UI & Handlers
// ============================================================================

const ANSI = {
  reset: '\x1b[0m',
  bold: '\x1b[1m',
  dim: '\x1b[2m',
  cyan: '\x1b[38;2;6;182;212m',
  teal: '\x1b[38;2;20;184;166m',
  green: '\x1b[38;2;34;197;94m',
  yellow: '\x1b[38;2;234;179;8m',
  red: '\x1b[38;2;239;68;68m',
  purple: '\x1b[38;2;168;85;247m',
  gray: '\x1b[38;2;148;163;184m',
  darkGray: '\x1b[38;2;71;85;105m',
  white: '\x1b[38;2;248;250;252m',
};

function printHeader(title, subtitle = '') {
  const width = 68;
  console.log(`\n${ANSI.cyan}╭${'─'.repeat(width - 2)}╮${ANSI.reset}`);
  console.log(`${ANSI.cyan}│  ${ANSI.bold}${ANSI.white}${title.padEnd(width - 5)}${ANSI.reset}${ANSI.cyan}│${ANSI.reset}`);
  if (subtitle) {
    const sub = subtitle.length <= width - 5 ? subtitle : subtitle.substring(0, width - 8) + '...';
    console.log(`${ANSI.cyan}│  ${ANSI.gray}${sub.padEnd(width - 5)}${ANSI.reset}${ANSI.cyan}│${ANSI.reset}`);
  }
  console.log(`${ANSI.cyan}╰${'─'.repeat(width - 2)}╯${ANSI.reset}\n`);
}

async function runInstaller(options) {
  printHeader('CYBERMES MCP — AUTO-INJECTOR', `Version: v${VERSION} | Target Clients: 13+ AI IDEs`);
  
  if (options.dryRun) {
    console.log(`  ${ANSI.yellow}[ DRY RUN ]${ANSI.reset} ${ANSI.gray}Simulation mode active — No files will be modified.${ANSI.reset}\n`);
  }

  const clients = getClientDefinitions(options.useLocal, options.localBinPath, options.workspaceRoot, options.useGlobal);
  const filterList = options.clients ? options.clients.split(',').map(s => s.trim().toLowerCase()) : null;

  let injectedCount = 0;
  let detectedCount = 0;

  for (const client of clients) {
    if (filterList && !filterList.includes(client.id) && !filterList.includes(client.name.toLowerCase())) {
      continue;
    }

    let targetPath = client.paths.find(p => fs.existsSync(p));
    if (!targetPath) {
      if (options.force || options.createAll) {
        targetPath = client.paths[0];
      } else {
        continue;
      }
    }

    detectedCount++;
    const res = injectClientConfig(client, targetPath, options.dryRun);

    if (res.status === 'injected' || res.status === 'dry-run') {
      injectedCount++;
      const badge = `${ANSI.green}[ INJECTED ]${ANSI.reset}`;
      console.log(`  ${badge} ${ANSI.white}${client.name.padEnd(24)}${ANSI.reset} ${ANSI.dim}-> ${res.path}${ANSI.reset}`);
    } else if (res.status === 'unchanged') {
      const badge = `${ANSI.cyan}[ READY ]${ANSI.reset}`;
      console.log(`  ${badge} ${ANSI.white}${client.name.padEnd(24)}${ANSI.reset} ${ANSI.dim}-> Up-to-date${ANSI.reset}`);
    } else if (res.status === 'error') {
      const badge = `${ANSI.red}[ ERROR ]${ANSI.reset}`;
      console.log(`  ${badge} ${ANSI.white}${client.name.padEnd(24)}${ANSI.reset} ${ANSI.red}${res.details}${ANSI.reset}`);
    }
  }

  console.log(`\n${ANSI.darkGray}────────────────────────────────────────────────────────────────────${ANSI.reset}`);
  if (injectedCount > 0 || detectedCount > 0) {
    console.log(`  ${ANSI.green}[SUCCESS]${ANSI.reset} Evaluated: ${detectedCount} client(s), Updated: ${injectedCount}`);
    console.log(`  ${ANSI.gray}Note: Restart your AI client (Cursor, Gemini, Claude, etc.) to reload MCP tools.${ANSI.reset}\n`);
  } else {
    console.log(`  ${ANSI.yellow}[INFO]${ANSI.reset} No target client configuration detected. Use --force to generate automatically.\n`);
  }
}

async function runUninstaller(options) {
  printHeader('CYBERMES MCP — UNINSTALLER', `Version: v${VERSION}`);
  
  if (options.dryRun) {
    console.log(`  ${ANSI.yellow}[ DRY RUN ]${ANSI.reset} ${ANSI.gray}Simulation mode active — No files will be modified.${ANSI.reset}\n`);
  }

  const clients = getClientDefinitions();
  const filterList = options.clients ? options.clients.split(',').map(s => s.trim().toLowerCase()) : null;
  let removedCount = 0;

  for (const client of clients) {
    if (filterList && !filterList.includes(client.id) && !filterList.includes(client.name.toLowerCase())) {
      continue;
    }

    const existingPath = client.paths.find(p => fs.existsSync(p));
    if (!existingPath) continue;

    const res = removeClientConfig(client, existingPath, options.dryRun);
    if (res.status === 'removed' || res.status === 'dry-run') {
      removedCount++;
      const badge = `${ANSI.green}[ REMOVED ]${ANSI.reset}`;
      console.log(`  ${badge} ${client.name.padEnd(24)} -> ${res.path}`);
    } else if (res.status === 'unchanged') {
      const badge = `${ANSI.darkGray}[ NO CHANGE ]${ANSI.reset}`;
      console.log(`  ${badge} ${client.name.padEnd(24)} -> Not present`);
    }
  }

  console.log(`\n${ANSI.darkGray}────────────────────────────────────────────────────────────────────${ANSI.reset}`);
  console.log(`  ${ANSI.green}[SUCCESS]${ANSI.reset} Cybermes configuration removed from ${removedCount} client(s).\n`);
}

async function runStatus() {
  printHeader('CYBERMES MCP — CLIENT DISCOVERY MATRIX', `Version: v${VERSION} | Host: ${os.platform()} (${os.arch()})`);

  const clients = getClientDefinitions();

  console.log(`  ${ANSI.bold}${'CLIENT'.padEnd(26)} ${'STATUS'.padEnd(16)} CONFIG PATH${ANSI.reset}`);
  console.log(`  ${ANSI.darkGray}${'─'.repeat(68)}${ANSI.reset}`);

  for (const client of clients) {
    const existingPath = client.paths.find(p => fs.existsSync(p)) || client.paths[0];
    const status = checkClientStatus(client, existingPath);

    let badge = `${ANSI.darkGray}[ NOT DETECTED ]${ANSI.reset}`;
    let nameStr = `${ANSI.darkGray}${client.name.padEnd(26)}${ANSI.reset}`;
    let pathStr = `${ANSI.darkGray}${status.path}${ANSI.reset}`;

    if (status.configured) {
      badge = `${ANSI.green}[ CONFIGURED ]${ANSI.reset}`;
      nameStr = `${ANSI.white}${client.name.padEnd(26)}${ANSI.reset}`;
      pathStr = `${ANSI.gray}${status.path}${ANSI.reset}`;
    } else if (status.installed) {
      badge = `${ANSI.yellow}[ NOT WIRED  ]${ANSI.reset}`;
      nameStr = `${ANSI.white}${client.name.padEnd(26)}${ANSI.reset}`;
      pathStr = `${ANSI.yellow}${status.path}${ANSI.reset}`;
    }

    console.log(`  ${nameStr} ${badge} ${pathStr}`);
  }

  console.log(`  ${ANSI.darkGray}${'─'.repeat(68)}${ANSI.reset}`);
  console.log(`  ${ANSI.gray}Run ${ANSI.cyan}npx cybermes-mcp install${ANSI.gray} to automatically configure un-wired clients.${ANSI.reset}\n`);
}

const CLIENT_ALIASES = {
  'gemini': ['gemini', 'antigravity', 'google'],
  'claude-desktop': ['claude-desktop', 'claude', 'desktop'],
  'cursor': ['cursor', 'cursor-ide', 'cursoride'],
  'opencode': ['opencode', 'open-code', 'interpreter'],
  'windsurf': ['windsurf', 'codeium'],
  'cline': ['cline', 'claude-dev'],
  'roo-code': ['roo-code', 'roo', 'roocode', 'roo-cline'],
  'claude-code': ['claude-code', 'claudecode', 'claude-cli'],
  'continue': ['continue', 'continuedev', 'continue-dev'],
  'zed': ['zed', 'zed-editor'],
  'kilo': ['kilo', 'kilo-code', 'kilocode'],
  'hermes': ['hermes', 'hermes-agent'],
  'codex': ['codex', 'codex-cli'],
};

function resolveClientTarget(token) {
  const clean = token.toLowerCase().replace(/^--?/, '').trim();
  for (const [id, aliases] of Object.entries(CLIENT_ALIASES)) {
    if (id === clean || aliases.includes(clean)) {
      return id;
    }
  }
  return null;
}

function parseTargetClients(args) {
  const targeted = new Set();
  const knownGeneralFlags = new Set(['install', 'uninstall', 'remove', 'status', 'doctor', 'help', '-h', '--help', '--dry-run', '--force', '--local', '--all']);

  for (const arg of args) {
    if (arg.startsWith('--clients=')) {
      const parts = arg.replace('--clients=', '').split(',');
      for (const p of parts) {
        const id = resolveClientTarget(p);
        if (id) targeted.add(id);
      }
      continue;
    }

    if (knownGeneralFlags.has(arg.toLowerCase())) continue;

    const matchedId = resolveClientTarget(arg);
    if (matchedId) {
      targeted.add(matchedId);
    }
  }

  return targeted.size > 0 ? Array.from(targeted) : null;
}

function printHelp() {
  printHeader('CYBERMES MCP SERVER — CLI HELP', `Package: cybermes-mcp v${VERSION}`);
  console.log(`  ${ANSI.bold}USAGE:${ANSI.reset}
    ${ANSI.cyan}npx cybermes-mcp${ANSI.reset} [command] [options/flags]

  ${ANSI.bold}COMMANDS:${ANSI.reset}
    ${ANSI.green}(no command)${ANSI.reset}      Start the Cybermes MCP Server over stdio (JSON-RPC 2.0)
    ${ANSI.green}install${ANSI.reset}           Auto-detect AI clients and inject MCP configuration
    ${ANSI.green}status, doctor${ANSI.reset}    Display status matrix across all supported AI clients
    ${ANSI.green}uninstall${ANSI.reset}         Cleanly remove Cybermes configuration from AI clients
    ${ANSI.green}help, -h${ANSI.reset}          Display this help menu

  ${ANSI.bold}TARGET PROVIDER FLAGS (Install to specific AI clients only):${ANSI.reset}
    ${ANSI.cyan}--gemini${ANSI.reset}, ${ANSI.cyan}--antigravity${ANSI.reset}   Google Antigravity / Gemini CLI & IDE
    ${ANSI.cyan}--kilo${ANSI.reset}, ${ANSI.cyan}--kilo-code${ANSI.reset}       Kilo Code IDE
    ${ANSI.cyan}--cursor${ANSI.reset}                  Cursor IDE
    ${ANSI.cyan}--claude${ANSI.reset}                  Claude Desktop
    ${ANSI.cyan}--claude-code${ANSI.reset}             Claude Code CLI
    ${ANSI.cyan}--windsurf${ANSI.reset}                Windsurf IDE (Codeium)
    ${ANSI.cyan}--cline${ANSI.reset}, ${ANSI.cyan}--roo${ANSI.reset}           Cline / Roo Code (VS Code Extension)
    ${ANSI.cyan}--opencode${ANSI.reset}                OpenCode Interpreter
    ${ANSI.cyan}--zed${ANSI.reset}, ${ANSI.cyan}--continue${ANSI.reset}       Zed Editor / Continue.dev
    ${ANSI.cyan}--hermes${ANSI.reset}, ${ANSI.cyan}--codex${ANSI.reset}         Hermes Agent / Codex CLI

  ${ANSI.bold}OPTIONS:${ANSI.reset}
    ${ANSI.yellow}--global, -g${ANSI.reset}      Wire clients to use global 'cybermes-mcp' command (no npx)
    ${ANSI.yellow}--dry-run${ANSI.reset}         Simulate operations without writing files to disk
    ${ANSI.yellow}--force${ANSI.reset}           Generate configuration files even if client is not detected
    ${ANSI.yellow}--local${ANSI.reset}           Wire directly to local compiled binary (tools/bin/cybermes-mcp)

  ${ANSI.bold}EXAMPLES:${ANSI.reset}
    # 1-Click install globally for all AI clients:
    ${ANSI.cyan}cybermes-mcp install --global${ANSI.reset}

    # Install ONLY into Kilo Code:
    ${ANSI.cyan}npx -y cybermes-mcp install --kilo${ANSI.reset}

    # Install into Gemini and Cursor with global command:
    ${ANSI.cyan}cybermes-mcp install --gemini --cursor --global${ANSI.reset}

    # Check status for specific client:
    ${ANSI.cyan}npx -y cybermes-mcp status --kilo${ANSI.reset}

    # Remove from Claude Desktop only:
    ${ANSI.cyan}npx -y cybermes-mcp uninstall --claude${ANSI.reset}
`);
}

// ============================================================================
// 5. Main CLI Entrypoint
// ============================================================================

async function main() {
  const args = process.argv.slice(2);
  const firstArg = (args[0] || '').toLowerCase();
  const targetClients = parseTargetClients(args);
  const useGlobal = args.includes('--global') || args.includes('-g');

  if (firstArg === 'install' || firstArg === '--install' || firstArg === '-i') {
    const dryRun = args.includes('--dry-run');
    const force = args.includes('--force') || Boolean(targetClients); // Auto-force if user explicitly targeted a client
    const useLocal = args.includes('--local');
    const localBin = useLocal ? findLocalDevBinary() : null;

    await runInstaller({
      dryRun,
      force,
      useLocal,
      useGlobal,
      localBinPath: localBin,
      clients: targetClients ? targetClients.join(',') : null,
    });
    return;
  }

  if (firstArg === 'uninstall' || firstArg === '--uninstall' || firstArg === 'remove') {
    const dryRun = args.includes('--dry-run');
    await runUninstaller({
      dryRun,
      clients: targetClients ? targetClients.join(',') : null,
    });
    return;
  }

  if (firstArg === 'status' || firstArg === '--status' || firstArg === 'doctor') {
    await runStatus();
    return;
  }

  if (firstArg === 'help' || firstArg === '--help' || firstArg === '-h') {
    printHelp();
    return;
  }

  // Check if first arg is directly an install flag like `npx cybermes-mcp --kilo`
  if (targetClients && (firstArg.startsWith('--') || firstArg.startsWith('-'))) {
    const dryRun = args.includes('--dry-run');
    const useLocal = args.includes('--local');
    const localBin = useLocal ? findLocalDevBinary() : null;
    await runInstaller({
      dryRun,
      force: true,
      useLocal,
      useGlobal,
      localBinPath: localBin,
      clients: targetClients.join(','),
    });
    return;
  }

  // Default: Execute native MCP Server over stdio
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

    process.on('SIGINT', () => child.kill('SIGINT'));
    process.on('SIGTERM', () => child.kill('SIGTERM'));

  } catch (err) {
    process.stderr.write(`[cybermes-mcp] Error: ${err.message}\n`);
    process.exit(1);
  }
}

main();

