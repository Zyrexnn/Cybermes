/**
 * npm/src/adapters/clients.js
 * Client definitions, path discovery, and adapter resolution for 13+ AI IDEs and agents.
 * 100% Zero external dependencies.
 */

const path = require('path');
const os = require('os');
const fs = require('fs');

const DEFAULT_AUTO_APPROVE = [
  'cybermes_search_knowledge',
  'cybermes_list_skills',
  'cybermes_get_skill',
  'cybermes_scan_secrets',
  'cybermes_validate_scope',
  'cybermes_subdomain_discovery',
  'cybermes_fuzz_endpoints',
  'cybermes_filter_stream',
  'cybermes_generate_pdf',
];

function getAppdataDir() {
  return process.env.APPDATA || path.join(os.homedir(), 'AppData', 'Roaming');
}

function getTargetDefinition(useLocal, localBinPath, workspaceRoot, useGlobal) {
  if (useLocal && localBinPath) {
    return {
      command: localBinPath,
      args: ['-workspace', workspaceRoot || path.resolve(__dirname, '..', '..', '..')],
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
        path.join(home, '.config', 'opencode', 'opencode.jsonc'),
        path.join(home, '.config', 'opencode', 'config.json'),
        isWin ? path.join(appdata, 'opencode', 'opencode.json') : null,
        path.join(process.cwd(), 'opencode.json'),
        path.join(process.cwd(), 'opencode.jsonc'),
      ].filter(Boolean),
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
        autoApprove: DEFAULT_AUTO_APPROVE,
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
        autoApprove: DEFAULT_AUTO_APPROVE,
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

module.exports = {
  DEFAULT_AUTO_APPROVE,
  getClientDefinitions,
  resolveClientTarget,
  CLIENT_ALIASES
};
