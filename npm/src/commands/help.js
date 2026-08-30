/**
 * npm/src/commands/help.js
 * Comprehensive CLI documentation, command catalog, flags, and contextual help.
 * Supports: -help, --help, -h, help, help <command>
 * 100% Zero external dependencies.
 */

const { ANSI, printBanner, printHeader, printDivider } = require('../utils/ui');
const { VERSION } = require('../utils/binary');

function printGeneralHelp() {
  printBanner('Autonomous Offensive Security & Diagnostic Model Context Protocol');

  console.log(`  ${ANSI.bold}USAGE:${ANSI.reset}
    ${ANSI.cyan}npx cybermes-mcp${ANSI.reset} [command] [options/flags]
    ${ANSI.cyan}cybermes-mcp${ANSI.reset} [command] [options/flags]

  ${ANSI.bold}CORE COMMANDS:${ANSI.reset}
    ${ANSI.green}(no command)${ANSI.reset}         Start the native Cybermes MCP Server over stdio (JSON-RPC 2.0)
    ${ANSI.green}install${ANSI.reset}              Auto-inject MCP configuration (Interactive Wizard or flags)
    ${ANSI.green}doctor${ANSI.reset}               Deep healthcheck, binary verification & JSON-RPC handshake test
    ${ANSI.green}status${ANSI.reset}               Display configuration & connection matrix for all AI clients
    ${ANSI.green}tools${ANSI.reset}                Display complete catalog of active MCP tools, permissions & SOPs
    ${ANSI.green}skills [query]${ANSI.reset}       Search & inspect 200+ offline security playbooks & SOPs
    ${ANSI.green}config [action]${ANSI.reset}      Manage global settings in ~/.cybermes/config.json (get/set/list)
    ${ANSI.green}uninstall${ANSI.reset}            Cleanly remove Cybermes MCP configuration from AI IDEs
    ${ANSI.green}help, -help, --help${ANSI.reset}  Display this documentation or contextual help for a command

  ${ANSI.bold}TARGET CLIENT FLAGS (Install / Configure specific IDEs only):${ANSI.reset}
    ${ANSI.cyan}--gemini, --antigravity${ANSI.reset}   Google Antigravity / Gemini CLI & IDE
    ${ANSI.cyan}--cursor${ANSI.reset}                  Cursor IDE
    ${ANSI.cyan}--claude, --desktop${ANSI.reset}       Claude Desktop
    ${ANSI.cyan}--kilo, --kilo-code${ANSI.reset}       Kilo Code IDE
    ${ANSI.cyan}--opencode${ANSI.reset}                OpenCode Interpreter
    ${ANSI.cyan}--windsurf${ANSI.reset}                Windsurf IDE (Codeium)
    ${ANSI.cyan}--cline${ANSI.reset}                   Cline (VS Code Extension)
    ${ANSI.cyan}--roo, --roo-code${ANSI.reset}         Roo Code (VS Code Extension)
    ${ANSI.cyan}--claude-code${ANSI.reset}             Claude Code CLI
    ${ANSI.cyan}--continue${ANSI.reset}                Continue.dev
    ${ANSI.cyan}--zed${ANSI.reset}                     Zed Editor
    ${ANSI.cyan}--hermes${ANSI.reset}                  Hermes Agent
    ${ANSI.cyan}--codex${ANSI.reset}                   Codex CLI
    ${ANSI.cyan}--all${ANSI.reset}                     Target all detected clients

  ${ANSI.bold}EXECUTION & INJECTION OPTIONS:${ANSI.reset}
    ${ANSI.yellow}--global, -g${ANSI.reset}            Wire clients to spawn global 'cybermes-mcp' command
    ${ANSI.yellow}--dry-run${ANSI.reset}               Preview configuration changes without writing files to disk
    ${ANSI.yellow}--force${ANSI.reset}                 Force configuration file creation even if client is not detected
    ${ANSI.yellow}--local${ANSI.reset}                 Wire client directly to local binary (tools/bin/cybermes-mcp)

  ${ANSI.bold}POPULAR EXAMPLES:${ANSI.reset}
    # 1. Start interactive setup wizard:
    ${ANSI.cyan}npx -y cybermes-mcp install${ANSI.reset}

    # 2. 1-Click install into Gemini and Cursor with global command:
    ${ANSI.cyan}cybermes-mcp install --gemini --cursor --global${ANSI.reset}

    # 3. Perform deep diagnostic & live MCP handshake test:
    ${ANSI.cyan}npx -y cybermes-mcp doctor${ANSI.reset}

    # 4. View active tool matrix and permissions:
    ${ANSI.cyan}npx -y cybermes-mcp tools${ANSI.reset}

    # 5. Search offensive security skills:
    ${ANSI.cyan}npx -y cybermes-mcp skills jwt${ANSI.reset}

    # 6. Set custom target workspace path:
    ${ANSI.cyan}cybermes-mcp config set workspace "C:\\MyTargets"${ANSI.reset}

    # 7. Get help for a specific command:
    ${ANSI.cyan}npx -y cybermes-mcp help install${ANSI.reset}
`);
}

function printInstallHelp() {
  printHeader('CYBERMES MCP — INSTALL COMMAND HELP', 'Interactive and automated AI Client configuration');
  console.log(`  ${ANSI.bold}USAGE:${ANSI.reset}
    ${ANSI.cyan}cybermes-mcp install${ANSI.reset} [client-flags] [options]

  ${ANSI.bold}BEHAVIOR:${ANSI.reset}
    • Without flags in a terminal: Launches an ${ANSI.green}Interactive Checklist Wizard${ANSI.reset}.
    • With flags (e.g. ${ANSI.cyan}--kilo${ANSI.reset}): Injects directly into specified clients.
    • Automatic backups (${ANSI.yellow}.bak-TIMESTAMP${ANSI.reset}) are created before touching any file.

  ${ANSI.bold}EXAMPLES:${ANSI.reset}
    ${ANSI.cyan}npx cybermes-mcp install${ANSI.reset}
    ${ANSI.cyan}npx cybermes-mcp install --kilo --gemini${ANSI.reset}
    ${ANSI.cyan}npx cybermes-mcp install --all --global${ANSI.reset}
    ${ANSI.cyan}npx cybermes-mcp install --claude --dry-run${ANSI.reset}
`);
}

function printDoctorHelp() {
  printHeader('CYBERMES MCP — DOCTOR COMMAND HELP', 'Diagnostic & JSON-RPC MCP handshake verification');
  console.log(`  ${ANSI.bold}USAGE:${ANSI.reset}
    ${ANSI.cyan}cybermes-mcp doctor${ANSI.reset}

  ${ANSI.bold}CHECKS PERFORMED:${ANSI.reset}
    1. Host OS platform and CPU architecture compatibility.
    2. Node.js runtime version compatibility (>=18.0.0).
    3. Native Go binary resolution, execution test, and version extraction.
    4. Live dry-run MCP handshake (JSON-RPC 2.0 initialize & tools/list).
    5. Client configuration syntax audits (JSON/JSONC/YAML/TOML).
`);
}

function printConfigHelp() {
  printHeader('CYBERMES MCP — CONFIG COMMAND HELP', 'Manage global preferences (~/.cybermes/config.json)');
  console.log(`  ${ANSI.bold}USAGE:${ANSI.reset}
    ${ANSI.cyan}cybermes-mcp config list${ANSI.reset}             View all persistent configuration values
    ${ANSI.cyan}cybermes-mcp config get <key>${ANSI.reset}        Get a specific setting value
    ${ANSI.cyan}cybermes-mcp config set <key> <val>${ANSI.reset}  Set a persistent setting value
    ${ANSI.cyan}cybermes-mcp config reset${ANSI.reset}            Reset configuration to factory defaults

  ${ANSI.bold}CONFIGURABLE KEYS:${ANSI.reset}
    • ${ANSI.bold}workspace${ANSI.reset}   : Default directory for target recon & findings (e.g. C:\\Recon)
    • ${ANSI.bold}binaryMode${ANSI.reset}  : 'auto', 'local', or 'release'
    • ${ANSI.bold}rateLimit${ANSI.reset}   : Default requests/sec safety ceiling (default: 10)
`);
}

function runHelp(args = []) {
  const sub = (args[0] || '').toLowerCase().replace(/^--?/, '');
  if (sub === 'install') {
    printInstallHelp();
  } else if (sub === 'doctor' || sub === 'status') {
    printDoctorHelp();
  } else if (sub === 'config') {
    printConfigHelp();
  } else {
    printGeneralHelp();
  }
}

module.exports = {
  runHelp
};
