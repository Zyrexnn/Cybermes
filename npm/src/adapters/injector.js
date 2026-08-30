/**
 * npm/src/adapters/injector.js
 * Universal client injector, uninstaller, and status evaluator.
 * Supports JSON, JSONC, YAML, and TOML configuration files.
 * 100% Zero external dependencies.
 */

const fs = require('fs');
const path = require('path');

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
    try {
      return JSON.parse(content);
    } catch (_) {
      const clean = content
        .replace(/("(?:\\.|[^"\\])*")|\/\/.*$|\/\*[\s\S]*?\*\//gm, (m, g1) => g1 || '')
        .replace(/,\s*([}\]])/g, '$1');
      return JSON.parse(clean);
    }
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
    if (existingContent.includes('cybermes:') && (existingContent.includes('cybermes-mcp'))) {
      result.status = 'unchanged';
      result.details = 'Already up-to-date in YAML';
      return result;
    }

    const cmdStr = JSON.stringify(client.definition.command);
    const argsStr = JSON.stringify(client.definition.args);
    let block = '';
    if (existingContent.includes('mcp_servers:')) {
      block = `\n  cybermes:\n    command: ${cmdStr}\n    args: ${argsStr}\n`;
    } else {
      block = `\nmcp_servers:\n  cybermes:\n    command: ${cmdStr}\n    args: ${argsStr}\n`;
    }

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
    return { client: client.name, path: filePath, installed: false, configured: false, details: 'Not detected' };
  }

  if (client.type === 'json-mcpServers' || client.type === 'json-cline') {
    const json = safeReadJson(filePath);
    if (json && json._parseError) {
      return { client: client.name, path: filePath, installed: true, configured: false, error: json._parseError, details: 'Corrupted JSON' };
    }
    const hasCybermes = Boolean(json && json.mcpServers && json.mcpServers.cybermes);
    return { client: client.name, path: filePath, installed: true, configured: hasCybermes, details: hasCybermes ? 'Configured' : 'Client detected (Cybermes missing)' };
  }

  if (client.type === 'json-mcp_servers') {
    const json = safeReadJson(filePath);
    if (json && json._parseError) {
      return { client: client.name, path: filePath, installed: true, configured: false, error: json._parseError, details: 'Corrupted JSON' };
    }
    const hasCybermes = Boolean(json && json.mcp_servers && json.mcp_servers.cybermes);
    return { client: client.name, path: filePath, installed: true, configured: hasCybermes, details: hasCybermes ? 'Configured' : 'Client detected (Cybermes missing)' };
  }

  if (client.type === 'json-continue') {
    const json = safeReadJson(filePath);
    if (json && json._parseError) {
      return { client: client.name, path: filePath, installed: true, configured: false, error: json._parseError, details: 'Corrupted JSON' };
    }
    const list = (json && json.experimental && json.experimental.modelContextProtocolServers) || [];
    const hasCybermes = list.some(s => s.name === 'cybermes');
    return { client: client.name, path: filePath, installed: true, configured: hasCybermes, details: hasCybermes ? 'Configured' : 'Client detected (Cybermes missing)' };
  }

  if (client.type === 'json-zed') {
    const json = safeReadJson(filePath);
    if (json && json._parseError) {
      return { client: client.name, path: filePath, installed: true, configured: false, error: json._parseError, details: 'Corrupted JSON' };
    }
    const hasCybermes = Boolean(json && json.context_servers && json.context_servers.cybermes);
    return { client: client.name, path: filePath, installed: true, configured: hasCybermes, details: hasCybermes ? 'Configured' : 'Client detected (Cybermes missing)' };
  }

  if (client.type === 'yaml-hermes' || client.type === 'toml-codex') {
    try {
      const content = fs.readFileSync(filePath, 'utf8');
      const hasCybermes = content.includes('cybermes');
      return { client: client.name, path: filePath, installed: true, configured: hasCybermes, details: hasCybermes ? 'Configured' : 'Client detected (Cybermes missing)' };
    } catch (_) {
      return { client: client.name, path: filePath, installed: true, configured: false, details: 'Read error' };
    }
  }

  return { client: client.name, path: filePath, installed: true, configured: false, details: 'Unknown format' };
}

module.exports = {
  createBackup,
  safeReadJson,
  safeWriteJson,
  injectClientConfig,
  removeClientConfig,
  checkClientStatus
};
