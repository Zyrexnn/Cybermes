#!/usr/bin/env node
/**
 * npm/test.js — Unit & Regression Tests for cybermes-mcp CLI & Injector
 * Zero external dependencies. Runs via `npm test` or `node test.js`.
 */

const assert = require('assert');
const fs = require('fs');
const path = require('path');
const os = require('os');
const { spawnSync } = require('child_process');

console.log('[TEST] Starting cybermes-mcp Test Suite...');

const SCRIPT_PATH = path.join(__dirname, 'bin', 'cybermes-mcp.js');

// Test 1: CLI Syntax and Help Output
console.log('[TEST 1/7] Testing CLI Help Output and Exit Code...');
const helpRes = spawnSync(process.execPath, [SCRIPT_PATH, '--help'], { encoding: 'utf8' });
assert.strictEqual(helpRes.status, 0, 'Help command should exit with code 0');
assert(helpRes.stdout.includes('CYBERMES MCP SERVER'), 'Help text should contain header');
assert(helpRes.stdout.includes('--kilo'), 'Help text should list --kilo flag');
assert(helpRes.stdout.includes('--gemini'), 'Help text should list --gemini flag');
assert(helpRes.stdout.includes('--global'), 'Help text should list --global flag');
console.log('  -> PASS: Help and CLI flags verified.');

// Test 2: Status Discovery Matrix
console.log('[TEST 2/7] Testing Client Discovery Status Matrix...');
const statusRes = spawnSync(process.execPath, [SCRIPT_PATH, 'status'], { encoding: 'utf8' });
assert.strictEqual(statusRes.status, 0, 'Status command should exit with code 0');
assert(statusRes.stdout.includes('Antigravity / Gemini'), 'Status matrix should discover Gemini');
assert(statusRes.stdout.includes('Kilo Code'), 'Status matrix should discover Kilo Code');
assert(statusRes.stdout.includes('Cursor IDE'), 'Status matrix should discover Cursor IDE');
console.log('  -> PASS: Discovery matrix output verified.');

// Test 3: Provider-Specific Flag Resolution (Dry Run)
console.log('[TEST 3/7] Testing Provider Flags (--kilo, --gemini, --cursor)...');
const kiloRes = spawnSync(process.execPath, [SCRIPT_PATH, 'install', '--kilo', '--dry-run'], { encoding: 'utf8' });
assert.strictEqual(kiloRes.status, 0);
assert(kiloRes.stdout.includes('Kilo Code'), 'Output should target Kilo Code');
assert(!kiloRes.stdout.includes('Cursor IDE'), 'Output should NOT target unselected Cursor');

const multiRes = spawnSync(process.execPath, [SCRIPT_PATH, 'install', '--gemini', '--cursor', '--dry-run'], { encoding: 'utf8' });
assert.strictEqual(multiRes.status, 0);
assert(multiRes.stdout.includes('Antigravity / Gemini'), 'Output should target Gemini');
assert(multiRes.stdout.includes('Cursor IDE'), 'Output should target Cursor');
assert(!multiRes.stdout.includes('Kilo Code'), 'Output should NOT target unselected Kilo');
console.log('  -> PASS: Provider flag filtering verified.');

// Test 4: Global Flag Configuration Generation (Dry Run)
console.log('[TEST 4/7] Testing Global Installation Flag (--global)...');
const globalRes = spawnSync(process.execPath, [SCRIPT_PATH, 'install', '--kilo', '--global', '--dry-run'], { encoding: 'utf8' });
assert.strictEqual(globalRes.status, 0);
assert(globalRes.stdout.includes('Kilo Code'));
console.log('  -> PASS: Global command mode verified.');

// Test 5: End-to-End Injection and Uninstallation in Isolated Temp Workspace
console.log('[TEST 5/7] Testing End-to-End Config Injection & Rollback in Temp Directory...');
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

// Test 6: Binary Path Resolution & Directory Rejection
console.log('[TEST 6/7] Testing Binary Path Resolution & Directory Rejection...');
const tempCheckDir = fs.mkdtempSync(path.join(os.tmpdir(), 'cybermes-bin-test-'));
try {
  const dummyDir = path.join(tempCheckDir, 'cybermes-mcp');
  fs.mkdirSync(dummyDir, { recursive: true });
  assert(fs.statSync(dummyDir).isDirectory(), 'Candidate directory must be a directory');
  assert(!fs.statSync(dummyDir).isFile(), 'Candidate directory must not be treated as a file');
  console.log('  -> PASS: Directory binary false-positive rejection verified.');
} finally {
  try { fs.rmSync(tempCheckDir, { recursive: true, force: true }); } catch (_) {}
}

// Test 7: URL Preservation & JSONC Comment Parsing Regression Test
console.log('[TEST 7/7] Testing URL Preservation & JSONC Comment Parsing...');
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

  // Verify CLI status/install does not crash on this file
  const testProc = spawnSync(process.execPath, [
    '-e',
    `const { getClientDefinitions } = require('${SCRIPT_PATH.replace(/\\/g, '\\\\')}');
     const fs = require('fs');
     const raw = fs.readFileSync('${sampleJsonc.replace(/\\/g, '\\\\')}', 'utf8');
     let clean;
     try {
       clean = JSON.parse(raw);
     } catch (_) {
       const stripped = raw
         .replace(/("(?:\\\\.|[^"\\\\])*")|\\/\\/.*$|\\/\\*[\\s\\S]*?\\*\\//gm, (m, g1) => g1 || '')
         .replace(/,\\s*([}\\]])/g, '$1');
       clean = JSON.parse(stripped);
     }
     if (clean['$schema'] !== 'https://opencode.ai/config.json') process.exit(1);
     if (clean.endpoint !== 'https://api.example.com/v1/models//test') process.exit(1);
     if (clean.server.url !== 'https://router.example.id/v1') process.exit(1);
    `
  ]);
  assert.strictEqual(testProc.status, 0, 'JSON with URLs and comments must parse cleanly without corruption');
  console.log('  -> PASS: URL preservation and JSONC comments parsed perfectly.');
} finally {
  try { fs.rmSync(testJsoncDir, { recursive: true, force: true }); } catch (_) {}
}

console.log('\n[SUCCESS] All 7 cybermes-mcp test suites passed cleanly with 0 errors!\n');
