/**
 * npm/src/utils/ui.js
 * Cyberpunk Truecolor Gradient Engine, Block Typography ASCII Art, and UI Tokens.
 * 100% Zero external dependencies.
 */

const ANSI = {
  reset: '\x1b[0m',
  bold: '\x1b[1m',
  dim: '\x1b[2m',
  italic: '\x1b[3m',
  underline: '\x1b[4m',
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

function gradient(text, fromRgb = [6, 182, 212], toRgb = [168, 85, 247]) {
  let result = '';
  const len = text.length;
  for (let i = 0; i < len; i++) {
    const factor = len > 1 ? i / (len - 1) : 0;
    const r = Math.round(fromRgb[0] + factor * (toRgb[0] - fromRgb[0]));
    const g = Math.round(fromRgb[1] + factor * (toRgb[1] - fromRgb[1]));
    const b = Math.round(fromRgb[2] + factor * (toRgb[2] - fromRgb[2]));
    result += `\x1b[38;2;${r};${g};${b}m${text[i]}`;
  }
  return result + ANSI.reset;
}

const RAW_BANNER = [
  '  ██████╗██╗   ██╗██████╗ ███████╗██████╗ ███╗   ███╗███████╗███████╗',
  ' ██╔════╝╚██╗ ██╔╝██╔══██╗██╔════╝██╔══██╗████╗ ████║██╔════╝██╔════╝',
  ' ██║      ╚████╔╝ ██████╔╝█████╗  ██████╔╝██╔████╔██║█████╗  ███████╗',
  ' ██║       ╚██╔╝  ██╔══██╗██╔══╝  ██╔══██╗██║╚██╔╝██║██╔══╝  ╚════██║',
  ' ╚██████╗   ██║   ██████╔╝███████╗██║  ██║██║ ╚═╝ ██║███████╗███████║',
  '  ╚═════╝   ╚═╝   ╚═════╝ ╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝╚══════╝╚══════╝'
];

function printBanner(subtitle = '') {
  console.log('');
  for (const line of RAW_BANNER) {
    console.log(gradient(line, [6, 182, 212], [168, 85, 247]));
  }
  console.log('  ' + ANSI.teal + '🛡️  Autonomous Offensive Security MCP Server' + ANSI.reset + ' ' + ANSI.dim + 'v3.3.0' + ANSI.reset);
  if (subtitle) {
    console.log('  ' + ANSI.gray + subtitle + ANSI.reset);
  }
  console.log('');
}

function printHeader(title, subtitle = '') {
  const width = 74;
  console.log('\n' + ANSI.cyan + '╭' + '─'.repeat(width - 2) + '╮' + ANSI.reset);
  console.log(ANSI.cyan + '│  ' + ANSI.bold + ANSI.white + title.padEnd(width - 5) + ANSI.reset + ANSI.cyan + '│' + ANSI.reset);
  if (subtitle) {
    const sub = subtitle.length <= width - 5 ? subtitle : subtitle.substring(0, width - 8) + '...';
    console.log(ANSI.cyan + '│  ' + ANSI.gray + sub.padEnd(width - 5) + ANSI.reset + ANSI.cyan + '│' + ANSI.reset);
  }
  console.log(ANSI.cyan + '╰' + '─'.repeat(width - 2) + '╯' + ANSI.reset + '\n');
}

function badge(text, type = 'info') {
  const padded = text.padEnd(12);
  switch (type) {
    case 'success':
    case 'injected':
    case 'pass':
    case 'configured':
      return ANSI.green + '[ ' + padded + ' ]' + ANSI.reset;
    case 'warn':
    case 'ready':
    case 'unwired':
      return ANSI.yellow + '[ ' + padded + ' ]' + ANSI.reset;
    case 'error':
    case 'fail':
      return ANSI.red + '[ ' + padded + ' ]' + ANSI.reset;
    case 'dry-run':
    case 'info':
      return ANSI.cyan + '[ ' + padded + ' ]' + ANSI.reset;
    case 'dim':
    case 'not_detected':
    default:
      return ANSI.darkGray + '[ ' + padded + ' ]' + ANSI.reset;
  }
}

function printDivider(len = 74) {
  console.log('  ' + ANSI.darkGray + '─'.repeat(len - 4) + ANSI.reset);
}

module.exports = {
  ANSI,
  gradient,
  printBanner,
  printHeader,
  badge,
  printDivider,
};
