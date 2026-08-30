/**
 * npm/src/index.js
 * Programmatic public API exports for cybermes-mcp.
 */

const { getClientDefinitions, resolveClientTarget } = require('./adapters/clients');
const { injectClientConfig, removeClientConfig, checkClientStatus } = require('./adapters/injector');
const { loadConfig, saveConfig, setConfigValue, getConfigValue } = require('./utils/config-store');
const { getOrDownloadBinary, findLocalDevBinary } = require('./utils/binary');
const { runInstaller } = require('./commands/install');
const { runDoctor } = require('./commands/doctor');
const { runStatus } = require('./commands/status');
const { runTools } = require('./commands/tools');
const { runSkills } = require('./commands/skills');
const { runConfig } = require('./commands/config');
const { runHelp } = require('./commands/help');
const { runServe } = require('./commands/serve');

module.exports = {
  getClientDefinitions,
  resolveClientTarget,
  injectClientConfig,
  removeClientConfig,
  checkClientStatus,
  loadConfig,
  saveConfig,
  setConfigValue,
  getConfigValue,
  getOrDownloadBinary,
  findLocalDevBinary,
  runInstaller,
  runDoctor,
  runStatus,
  runTools,
  runSkills,
  runConfig,
  runHelp,
  runServe
};
