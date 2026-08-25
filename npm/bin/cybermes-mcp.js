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

function getTargetDefinition(useLocal, localBinPath, workspaceRoot) {
  if (useLocal && localBinPath) {
    return {
      command: localBinPath,
      args: ['-workspace', workspaceRoot || path.resolve(__dirname, '..', '..')],
    };
  }
  return {
    command: 'npx',
    args: ['-y', '@zyrexnn/cybermes-mcp'],
  };
}

function getClientDefinitions(useLocal, localBinPath, workspaceRoot) {
  const home = os.homedir();
  const appdata = getAppdataDir();
  const isWin = os.platform() === 'win32';
  const isMac = os.platform() === 'darwin';

  const defaultDef = getTargetDefinition(useLocal, localBinPath, workspaceRoot);

  return [
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
// 4. CLI Subcommand Handlers: runInstaller, runUninstaller, runStatus
// ============================================================================

async function runInstaller(options) {
  console.log('\n🛡️  Cybermes MCP Server — Universal Auto-Installer v' + VERSION);
  console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  if (options.dryRun) {
    console.log('🔍 DRY RUN MODE ACTIVATED — No configuration files will be modified.\n');
  }

  const clients = getClientDefinitions(options.useLocal, options.localBinPath, options.workspaceRoot);
  const filterList = options.clients ? options.clients.split(',').map(s => s.trim().toLowerCase()) : null;

  let injectedCount = 0;
  let detectedCount = 0;

  for (const client of clients) {
    if (filterList && !filterList.includes(client.id) && !filterList.includes(client.name.toLowerCase())) {
      continue;
    }

    // Pick first matching or existing path, or default to first path if force
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
      console.log(`  ✓ ${client.name.padEnd(26)} -> [${res.status.toUpperCase()}] ${res.path} (${res.details})`);
    } else if (res.status === 'unchanged') {
      console.log(`  = ${client.name.padEnd(26)} -> [UNCHANGED] ${res.path} (${res.details})`);
    } else if (res.status === 'error') {
      console.log(`  ✗ ${client.name.padEnd(26)} -> [ERROR] ${res.path} (${res.details})`);
    }
  }

  console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  if (injectedCount > 0 || detectedCount > 0) {
    console.log(`🎉 Process completed! (Targets evaluated: ${detectedCount}, Updates: ${injectedCount})`);
    console.log('💡 Note: Restart your AI client (Cursor, Claude, Windsurf, etc.) to reload tools.\n');
  } else {
    console.log('ℹ️  No AI client configurations were detected on this system.');
    console.log('   Use --force to automatically generate configuration files for all supported clients.\n');
  }
}

async function runUninstaller(options) {
  console.log('\n🗑️  Cybermes MCP Server — Uninstaller v' + VERSION);
  console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  if (options.dryRun) {
    console.log('🔍 DRY RUN MODE ACTIVATED — No configuration files will be modified.\n');
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
      console.log(`  ✓ ${client.name.padEnd(26)} -> [REMOVED] ${res.path} (${res.details})`);
    } else if (res.status === 'unchanged') {
      console.log(`  = ${client.name.padEnd(26)} -> [NO CHANGE] ${res.path} (${res.details})`);
    }
  }

  console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  console.log(`✨ Cleanup complete! Removed from ${removedCount} client configuration(s).\n`);
}

async function runStatus() {
  console.log('\n📊 Cybermes MCP Server — Client Discovery & Status Matrix v' + VERSION);
  console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');

  const clients = getClientDefinitions();

  for (const client of clients) {
    const existingPath = client.paths.find(p => fs.existsSync(p)) || client.paths[0];
    const status = checkClientStatus(client, existingPath);

    const mark = status.configured ? '✓' : (status.installed ? '!' : '-');
    const stateStr = status.configured ? '[CONFIGURED]' : (status.installed ? '[NOT WIRED]' : '[NOT DETECTED]');
    console.log(`  ${mark} ${client.name.padEnd(26)} ${stateStr.padEnd(16)} ${status.path}`);
  }

  console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  console.log('💡 Run `npx @zyrexnn/cybermes-mcp install` to automatically configure un-wired clients.\n');
}

function printHelp() {
  console.log(`
🛡️  Cybermes Model Context Protocol (MCP) Server — CLI & Auto-Installer v${VERSION}
https://github.com/${REPO}

USAGE:
  npx @zyrexnn/cybermes-mcp [command] [options]

COMMANDS:
  (no command)      Start the Cybermes MCP Server over stdio (JSON-RPC 2.0)
  install           Auto-detect installed AI clients and inject Cybermes configuration
  uninstall         Cleanly remove Cybermes configuration from all AI clients
  status, doctor    Display status and discovery matrix of all supported AI clients
  help, --help, -h  Display this help menu

OPTIONS for 'install':
  --dry-run         Simulate injection without modifying any files
  --force           Generate config files even if AI client is not currently installed
  --clients=<list>  Comma-separated client IDs to target (e.g. cursor,claude-desktop,opencode)
  --local           Wire directly to local compiled binary (tools/bin/cybermes-mcp)

EXAMPLES:
  # 1-Click universal auto-injection for all detected AI clients:
  npx -y @zyrexnn/cybermes-mcp install

  # Test what will change without touching disk:
  npx -y @zyrexnn/cybermes-mcp install --dry-run

  # Check configuration status across your IDEs:
  npx -y @zyrexnn/cybermes-mcp status
`);
}

// ============================================================================
// 5. Main CLI Entrypoint
// ============================================================================

async function main() {
  const args = process.argv.slice(2);
  const firstArg = (args[0] || '').toLowerCase();

  if (firstArg === 'install' || firstArg === '--install' || firstArg === '-i') {
    const dryRun = args.includes('--dry-run');
    const force = args.includes('--force');
    const useLocal = args.includes('--local');
    const clientsArg = args.find(a => a.startsWith('--clients='))?.replace('--clients=', '');
    const localBin = useLocal ? findLocalDevBinary() : null;

    await runInstaller({ dryRun, force, useLocal, localBinPath: localBin, clients: clientsArg });
    return;
  }

  if (firstArg === 'uninstall' || firstArg === '--uninstall' || firstArg === 'remove') {
    const dryRun = args.includes('--dry-run');
    const clientsArg = args.find(a => a.startsWith('--clients='))?.replace('--clients=', '');
    await runUninstaller({ dryRun, clients: clientsArg });
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

