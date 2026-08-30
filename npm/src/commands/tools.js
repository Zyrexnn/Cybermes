/**
 * npm/src/commands/tools.js
 * Visual matrix of all registered Cybermes MCP tools, descriptions, and autoApprove status.
 * 100% Zero external dependencies.
 */

const { ANSI, printHeader, printDivider, badge } = require('../utils/ui');
const { VERSION } = require('../utils/binary');
const { DEFAULT_AUTO_APPROVE } = require('../adapters/clients');

const TOOLS_CATALOG = [
  {
    name: 'cybermes_generate_pdf',
    category: 'Reporting & Deliverables',
    description: 'Generates polished executive PDF security report via native chromedp',
    safety: 'read-only',
  },
  {
    name: 'cybermes_aggregate_report',
    category: 'Reporting & Deliverables',
    description: 'Compiles SUMMARY.md, metadata.json, and interactive report.html',
    safety: 'write',
  },
  {
    name: 'cybermes_validate_scope',
    category: 'Guardrails & Safety',
    description: 'Validates target domains/IPs against wildcard & CIDR rules in scope.yaml',
    safety: 'read-only',
  },
  {
    name: 'cybermes_http_probe',
    category: 'Reconnaissance',
    description: 'Web technology detection, TLS certificate analysis & header fingerprinting',
    safety: 'active-probe',
  },
  {
    name: 'cybermes_recon_crawl',
    category: 'Reconnaissance',
    description: 'Smart Pipe token-budgeted SPA crawler and JS bundle miner',
    safety: 'active-probe',
  },
  {
    name: 'cybermes_subdomain_discovery',
    category: 'Reconnaissance',
    description: 'Passive subdomain discovery with stream deduplication',
    safety: 'read-only',
  },
  {
    name: 'cybermes_search_knowledge',
    category: 'Intelligence & Payloads',
    description: 'Sub-50ms query against 50,000+ curated payloads (HackTricks, PayloadsAllTheThings)',
    safety: 'read-only',
  },
  {
    name: 'cybermes_list_skills',
    category: 'Intelligence & SOPs',
    description: 'Lists and filters 200+ offensive security playbooks',
    safety: 'read-only',
  },
  {
    name: 'cybermes_get_skill',
    category: 'Intelligence & SOPs',
    description: 'Loads complete step-by-step offensive methodology SOP into LLM memory',
    safety: 'read-only',
  },
  {
    name: 'cybermes_scan_secrets',
    category: 'Secret Mining',
    description: '48-pattern credential leak detector with automated masking & entropy checks',
    safety: 'read-only',
  },
  {
    name: 'cybermes_fuzz_endpoints',
    category: 'Active Auditing',
    description: 'Directory & parameter mutation fuzzer with Smart Pipe rate-limiting',
    safety: 'active-probe',
  },
  {
    name: 'cybermes_filter_stream',
    category: 'Token Optimization',
    description: 'Stream output parser filtering raw noisy output to save context tokens',
    safety: 'read-only',
  },
  {
    name: 'cybermes_record_finding',
    category: 'Reporting & Deliverables',
    description: 'Saves validated zero-false-positive vulnerability report & standalone PoC',
    safety: 'write',
  },
  {
    name: 'cybermes_list_findings',
    category: 'Reporting & Deliverables',
    description: 'Lists confirmed findings and severity breakdown per target slug',
    safety: 'read-only',
  },
];

async function runTools() {
  printHeader('CYBERMES MCP — CAPABILITY & TOOL MATRIX', `Package: cybermes-mcp v${VERSION} | Total Tools: ${TOOLS_CATALOG.length}`);

  console.log(`  ${ANSI.bold}${'TOOL NAME'.padEnd(30)} ${'CATEGORY'.padEnd(24)} ${'AUTO-APPROVE'}${ANSI.reset}`);
  printDivider(74);

  for (const tool of TOOLS_CATALOG) {
    const isAutoApproved = DEFAULT_AUTO_APPROVE.includes(tool.name);
    const autoBadge = isAutoApproved ? `${ANSI.green}✔ Enabled${ANSI.reset}` : `${ANSI.yellow}○ Ask User${ANSI.reset}`;
    console.log(`  ${ANSI.cyan}${tool.name.padEnd(30)}${ANSI.reset} ${ANSI.gray}${tool.category.padEnd(24)}${ANSI.reset} ${autoBadge}`);
    console.log(`    ${ANSI.dim}${tool.description}${ANSI.reset}`);
  }

  printDivider(74);
  console.log(`  ${ANSI.bold}MCP RESOURCES:${ANSI.reset}`);
  console.log(`    ${ANSI.purple}skills://{skill_name}${ANSI.reset}        - Read-only direct URI access to offensive playbook SOPs`);
  console.log(`    ${ANSI.purple}reports://{target}/summary${ANSI.reset}  - Read-only direct URI access to target summary & findings`);

  console.log(`\n  ${ANSI.bold}MCP PROMPT TEMPLATES:${ANSI.reset}`);
  console.log(`    ${ANSI.teal}cybermes_hunt${ANSI.reset}              - Autonomous offensive security research workflow`);
  console.log(`    ${ANSI.teal}cybermes_triage${ANSI.reset}            - Deterministic vulnerability verification & PoC generation\n`);
}

module.exports = {
  runTools,
  TOOLS_CATALOG
};
