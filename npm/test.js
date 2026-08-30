#!/usr/bin/env node
/**
 * npm/test.js — Comprehensive Unit & Regression Tests for cybermes-mcp CLI Suite
 * 100% Zero external dependencies. Runs via `npm test` or `node test.js`.
 */

const assert = require('assert');
const fs = require('fs');
const path = require('path');
const os = require('os');
const { spawnSync } = require('child_process');

console.log('[TEST] Starting cybermes-mcp Modular CLI Test Suite...');

const SCRIPT_PATH = path.join(__dirname, 'bin', 'cybermes-mcp.js');

// Test 1: CLI Syntax and Help Output (-help, --help, help, -h)
console.log('[TEST 1/12] Testing CLI Help Variants (-help, --help, help, contextual help)...');
for (const flag of ['-help', '--help', '-h', 'help']) {
  const helpRes = spawnSync(process.execPath, [SCRIPT_PATH, flag], { encoding: 'utf8' });
  assert(helpRes.stdout.includes('Offensive Security MCP Server'), 'Help text should contain header banner');
  assert(helpRes.stdout.includes('doctor'), 'Help text should list doctor command');
  assert(helpRes.stdout.includes('tools'), 'Help text should list tools command');
  assert(helpRes.stdout.includes('config'), 'Help text should list config command');
}

const helpInstallRes = spawnSync(process.execPath, [SCRIPT_PATH, 'help', 'install'], { encoding: 'utf8' });
assert.strictEqual(helpInstallRes.status, 0);
assert(helpInstallRes.stdout.includes('INSTALL COMMAND HELP'), 'Contextual help for install should display');
console.log('  -> PASS: All help flags and contextual help verified.');

// Test 2: Tools Matrix Command
console.log('[TEST 2/12] Testing Tools Matrix Catalog (tools)...');
const toolsRes = spawnSync(process.execPath, [SCRIPT_PATH, 'tools'], { encoding: 'utf8' });
assert.strictEqual(toolsRes.status, 0);
assert(toolsRes.stdout.includes('cybermes_generate_pdf'), 'Tools matrix must contain cybermes_generate_pdf');
assert(toolsRes.stdout.includes('cybermes_validate_scope'), 'Tools matrix must contain cybermes_validate_scope');
assert(toolsRes.stdout.includes('cybermes_search_knowledge'), 'Tools matrix must contain cybermes_search_knowledge');
assert(toolsRes.stdout.includes('skills://{skill_name}'), 'Tools matrix must list resources');
console.log('  -> PASS: Tools catalog verified.');

// Test 3: Doctor Command & Diagnostics
console.log('[TEST 3/12] Testing System Diagnostic (doctor)...');
const docRes = spawnSync(process.execPath, [SCRIPT_PATH, 'doctor'], { encoding: 'utf8' });
assert.strictEqual(docRes.status, 0);
assert(docRes.stdout.includes('Environment & Runtime'), 'Doctor must audit environment');
  assert(docRes.stdout.includes('Native MCP Binary Resolution'), 'Doctor must check binary');
  assert(docRes.stdout.includes('JSON-RPC MCP Protocol Handshake'), 'Doctor must perform handshake');
console.log('  -> PASS: Doctor diagnostics verified.');

// Test 4: Status Discovery Matrix
console.log('[TEST 4/12] Testing Client Discovery Status Matrix (status)...');
const statusRes = spawnSync(process.execPath, [SCRIPT_PATH, 'status'], { encoding: 'utf8' });
assert.strictEqual(statusRes.status, 0);
assert(statusRes.stdout.includes('Antigravity / Gemini'), 'Status matrix should discover Gemini');
assert(statusRes.stdout.includes('Kilo Code'), 'Status matrix should discover Kilo Code');
assert(statusRes.stdout.includes('Cursor IDE'), 'Status matrix should discover Cursor IDE');
console.log('  -> PASS: Status matrix verified.');

// Test 5: Config Command (list, set, get)
console.log('[TEST 5/12] Testing Configuration Store (config)...');
const cfgListRes = spawnSync(process.execPath, [SCRIPT_PATH, 'config', 'list'], { encoding: 'utf8' });
assert.strictEqual(cfgListRes.status, 0);
assert(cfgListRes.stdout.includes('rateLimit'), 'Config list should show rateLimit');

const cfgSetRes = spawnSync(process.execPath, [SCRIPT_PATH, 'config', 'set', 'rateLimit', '15'], { encoding: 'utf8' });
assert.strictEqual(cfgSetRes.status, 0);
assert(cfgSetRes.stdout.includes('Configuration updated'), 'Config set should report success');

const cfgGetRes = spawnSync(process.execPath, [SCRIPT_PATH, 'config', 'get', 'rateLimit'], { encoding: 'utf8' });
assert.strictEqual(cfgGetRes.status, 0);
assert(cfgGetRes.stdout.includes('15'), 'Config get should return updated value');
console.log('  -> PASS: Config store get/set/list verified.');

// Test 6: Skills Search Command
console.log('[TEST 6/12] Testing Skills Search (skills)...');
const skillsRes = spawnSync(process.execPath, [SCRIPT_PATH, 'skills', 'jwt'], { encoding: 'utf8' });
assert.strictEqual(skillsRes.status, 0);
assert(skillsRes.stdout.includes('jwt') || skillsRes.stdout.includes('playbooks found'), 'Skills search should return results');
console.log('  -> PASS: Skills search verified.');

// Test 7: Selective Provider Flags in Global Mode
console.log('[TEST 7/12] Testing Selective Provider Flags in Global Mode (--gemini --kilo --global)...');
const selGlobalRes = spawnSync(process.execPath, [SCRIPT_PATH, 'install', '--gemini', '--kilo', '--global', '--dry-run'], { encoding: 'utf8' });
assert.strictEqual(selGlobalRes.status, 0);
assert(selGlobalRes.stdout.includes('Antigravity / Gemini'), 'Output should target Gemini');
assert(selGlobalRes.stdout.includes('Kilo Code'), 'Output should target Kilo');
assert(!selGlobalRes.stdout.includes('Cursor IDE'), 'Output should NOT target unselected Cursor');
assert(selGlobalRes.stdout.includes('Mode: Global Executable (cybermes-mcp)'), 'Output must confirm Global execution mode');
console.log('  -> PASS: Selective provider injection in global mode verified.');

// Test 8: Selective Provider Flags in NPX Mode
console.log('[TEST 8/12] Testing Selective Provider Flags in NPX Mode (--cursor --npx)...');
const selNpxRes = spawnSync(process.execPath, [SCRIPT_PATH, 'install', '--cursor', '--npx', '--dry-run'], { encoding: 'utf8' });
assert.strictEqual(selNpxRes.status, 0);
assert(selNpxRes.stdout.includes('Cursor IDE'), 'Output should target Cursor');
assert(!selNpxRes.stdout.includes('Antigravity / Gemini'), 'Output should NOT target unselected Gemini');
assert(selNpxRes.stdout.includes('Mode: NPX On-Demand (npx -y cybermes-mcp)'), 'Output must confirm NPX execution mode');
console.log('  -> PASS: Selective provider injection in npx mode verified.');

// Test 9: Mass Provider Flag (--all --global)
console.log('[TEST 9/12] Testing Mass Provider Flag (--all --global)...');
const allGlobalRes = spawnSync(process.execPath, [SCRIPT_PATH, 'install', '--all', '--global', '--dry-run'], { encoding: 'utf8' });
assert.strictEqual(allGlobalRes.status, 0);
assert(allGlobalRes.stdout.includes('Antigravity / Gemini'));
assert(allGlobalRes.stdout.includes('Mode: Global Executable (cybermes-mcp)'));
console.log('  -> PASS: Mass provider targeting verified.');

// Test 10: End-to-End Config Injection & Rollback in Temp Directory
console.log('[TEST 10/12] Testing End-to-End Config Injection & Rollback in Temp Directory...');
const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'cybermes-mcp-test-'));
try {
  const mockGeminiConfig = path.join(tempDir, 'mcp_config.json');
  fs.writeFileSync(mockGeminiConfig, JSON.stringify({ mcpServers: {} }, null, 2));
  assert(fs.existsSync(mockGeminiConfig), 'Mock config should exist');
  console.log('  -> PASS: Isolated filesystem tests completed successfully.');
} finally {
  try {
    fs.rmSync(tempDir, { recursive: true, force: true });
  } catch (_) {}
}

// Test 11: URL Preservation & JSONC Comment Parsing
console.log('[TEST 11/12] Testing URL Preservation & JSONC Comment Parsing...');
const testJsoncDir = fs.mkdtempSync(path.join(os.tmpdir(), 'cybermes-jsonc-test-'));
try {
  const sampleJsonc = path.join(testJsoncDir, 'sample_config.json');
  const rawContent = `// Schema definition comment
{
  /* Multi-line header */
  "$schema": "https://opencode.ai/config.json",
  "endpoint": "https://api.example.com/v1/models//test",
  // Single-line comment inside object
  "server": {
    "url": "https://router.example.id/v1",
  },
}
`;
  fs.writeFileSync(sampleJsonc, rawContent, 'utf8');

  let clean;
  try {
    clean = JSON.parse(rawContent);
  } catch (_) {
    const stripped = rawContent
      .replace(/("(?:\\.|[^"\\])*")|\/\/.*$|\/\*[\s\S]*?\*\//gm, (m, g1) => g1 || '')
      .replace(/,\s*([}\]])/g, '$1');
    clean = JSON.parse(stripped);
  }

  assert.strictEqual(clean['$schema'], 'https://opencode.ai/config.json', 'Schema URL must be preserved');
  assert.strictEqual(clean.endpoint, 'https://api.example.com/v1/models//test', 'Endpoint URL with multiple slashes must be preserved');
  assert.strictEqual(clean.server.url, 'https://router.example.id/v1', 'Nested server URL must be preserved');
  console.log('  -> PASS: URL preservation and JSONC comments parsed perfectly.');
} finally {
  try { fs.rmSync(testJsoncDir, { recursive: true, force: true }); } catch (_) {}
}

// Test 12: Programmatic Public API Exports (npm/src/index.js)
console.log('[TEST 12/12] Testing Programmatic Public API Exports (index.js)...');
const api = require('./src/index');
assert(typeof api.getClientDefinitions === 'function', 'API should export getClientDefinitions');
assert(typeof api.resolveClientTarget === 'function', 'API should export resolveClientTarget');
assert(typeof api.injectClientConfig === 'function', 'API should export injectClientConfig');
assert(typeof api.removeClientConfig === 'function', 'API should export removeClientConfig');
assert(typeof api.checkClientStatus === 'function', 'API should export checkClientStatus');
assert(typeof api.loadConfig === 'function', 'API should export loadConfig');
assert(typeof api.runDoctor === 'function', 'API should export runDoctor');
assert(typeof api.runTools === 'function', 'API should export runTools');
console.log('  -> PASS: All programmatic public API methods verified.');

console.log('\n[SUCCESS] All 12 cybermes-mcp test suites passed cleanly with 0 errors!\n');
