/**
 * npm/src/utils/config-store.js
 * Persistent settings manager for ~/.cybermes/config.json
 * 100% Zero external dependencies.
 */

const fs = require('fs');
const path = require('path');
const os = require('os');

function getConfigDir() {
  return path.join(os.homedir(), '.cybermes');
}

function getConfigPath() {
  return path.join(getConfigDir(), 'config.json');
}

function loadConfig() {
  const configPath = getConfigPath();
  if (!fs.existsSync(configPath)) {
    return {
      version: '1.0',
      workspace: '',
      binaryMode: 'auto', // 'auto', 'local', 'release'
      rateLimit: 10,
      autoApprove: [
        'cybermes_search_knowledge',
        'cybermes_list_skills',
        'cybermes_get_skill',
        'cybermes_scan_secrets',
        'cybermes_validate_scope',
        'cybermes_subdomain_discovery',
        'cybermes_fuzz_endpoints',
        'cybermes_filter_stream',
        'cybermes_generate_pdf'
      ],
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString()
    };
  }
  try {
    const raw = fs.readFileSync(configPath, 'utf8').trim();
    return JSON.parse(raw);
  } catch (err) {
    return {
      error: `Failed to parse config: ${err.message}`,
      rawPath: configPath
    };
  }
}

function saveConfig(config) {
  const dir = getConfigDir();
  if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true });
  }
  config.updatedAt = new Date().toISOString();
  fs.writeFileSync(getConfigPath(), JSON.stringify(config, null, 2) + '\n', 'utf8');
  return true;
}

function setConfigValue(key, value) {
  const config = loadConfig();
  if (config.error) {
    throw new Error(config.error);
  }
  
  if (value === 'true') value = true;
  else if (value === 'false') value = false;
  else if (!isNaN(Number(value)) && value.trim() !== '') value = Number(value);

  config[key] = value;
  saveConfig(config);
  return config;
}

function getConfigValue(key) {
  const config = loadConfig();
  return config[key];
}

function resetConfig() {
  const configPath = getConfigPath();
  if (fs.existsSync(configPath)) {
    fs.unlinkSync(configPath);
  }
  return loadConfig();
}

module.exports = {
  getConfigDir,
  getConfigPath,
  loadConfig,
  saveConfig,
  setConfigValue,
  getConfigValue,
  resetConfig
};
