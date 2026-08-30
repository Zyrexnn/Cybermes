/**
 * npm/src/utils/prompt.js
 * Zero-dependency interactive checkbox, radio selection, and confirmation menus using native Node.js readline.
 * Includes dynamic frame tracking, universal line-erasing, smart path truncation, and terminal cursor lifecycle management.
 */

const readline = require('readline');
const { ANSI } = require('./ui');

function getTerminalWidth() {
  return (process.stdout && process.stdout.columns) ? process.stdout.columns : 80;
}

function truncatePath(pathStr, maxLen = 45) {
  if (!pathStr || typeof pathStr !== 'string' || pathStr.length <= maxLen) return pathStr;
  const startLen = Math.max(8, Math.floor((maxLen - 3) * 0.35));
  const endLen = Math.max(12, (maxLen - 3) - startLen);
  return pathStr.substring(0, startLen) + '...' + pathStr.substring(pathStr.length - endLen);
}

function hideCursor() {
  if (process.stdout.isTTY) {
    process.stdout.write('\x1b[?25l');
  }
}

function showCursor() {
  if (process.stdout.isTTY) {
    process.stdout.write('\x1b[?25h');
  }
}

/**
 * Single-select radio prompt.
 * @param {string} title
 * @param {Array<{id: string, name: string, desc?: string}>} options
 * @param {number} defaultIndex
 * @returns {Promise<string>} Selected option ID
 */
function promptRadio(title, options, defaultIndex = 0) {
  return new Promise((resolve) => {
    if (!process.stdin.isTTY) {
      return resolve(options[defaultIndex] ? options[defaultIndex].id : options[0].id);
    }

    let cursor = Math.max(0, Math.min(defaultIndex, options.length - 1));
    let lastRenderedLineCount = 0;

    const rl = readline.createInterface({
      input: process.stdin,
      output: process.stdout,
      terminal: true
    });

    readline.emitKeypressEvents(process.stdin, rl);
    if (process.stdin.isTTY) {
      process.stdin.setRawMode(true);
    }
    hideCursor();

    const cleanTitle = title.replace(/[:\s]+$/, '');

    function render() {
      if (lastRenderedLineCount > 0) {
        readline.cursorTo(process.stdout, 0);
        readline.moveCursor(process.stdout, 0, -lastRenderedLineCount);
      }

      let linesRendered = 0;
      function writeLine(str = '') {
        process.stdout.write(`\x1b[2K\r${str}\n`);
        linesRendered++;
      }

      writeLine(`  ${ANSI.bold}${ANSI.cyan}? ${cleanTitle}:${ANSI.reset}`);
      writeLine(`    ${ANSI.dim}Use [↑/↓] to navigate, [Enter] to select${ANSI.reset}`);

      options.forEach((opt, idx) => {
        const isCurrent = idx === cursor;
        const pointer = isCurrent ? `${ANSI.cyan}❯${ANSI.reset}` : ' ';
        const radio = isCurrent ? `${ANSI.green}(•)${ANSI.reset}` : `${ANSI.darkGray}( )${ANSI.reset}`;
        const nameStr = isCurrent 
          ? `${ANSI.bold}${ANSI.white}${opt.name}${ANSI.reset}` 
          : `${ANSI.gray}${opt.name}${ANSI.reset}`;
        const descStr = opt.desc ? ` ${ANSI.dim}${opt.desc}${ANSI.reset}` : '';

        writeLine(`  ${pointer} ${radio} ${nameStr}${descStr}`);
      });

      lastRenderedLineCount = linesRendered;
    }

    render();

    function cleanup() {
      showCursor();
      process.stdin.removeListener('keypress', onKeyPress);
      if (process.stdin.isTTY) {
        process.stdin.setRawMode(false);
      }
      rl.close();
    }

    function onKeyPress(str, key) {
      if (!key) return;

      if (key.ctrl && key.name === 'c') {
        cleanup();
        console.log(`\n  ${ANSI.yellow}[ABORTED] Operation cancelled by user.${ANSI.reset}\n`);
        process.exit(0);
      }

      if (key.name === 'up') {
        cursor = (cursor - 1 + options.length) % options.length;
        render();
      } else if (key.name === 'down') {
        cursor = (cursor + 1) % options.length;
        render();
      } else if (key.name === 'return' || key.name === 'enter') {
        cleanup();
        console.log('');
        resolve(options[cursor].id);
      }
    }

    process.stdin.on('keypress', onKeyPress);
  });
}

/**
 * Multi-select checkbox prompt.
 * @param {string} title
 * @param {Array<{id: string, name: string, checked: boolean, installed?: boolean, path?: string}>} items
 * @param {string} subtitle
 * @returns {Promise<Array<string>>} Selected item IDs
 */
function promptCheckbox(title, items, subtitle = '') {
  return new Promise((resolve) => {
    if (!process.stdin.isTTY) {
      return resolve(items.filter(i => i.checked).map(i => i.id));
    }

    const state = items.map(item => ({ ...item }));
    let cursor = 0;
    let lastRenderedLineCount = 0;

    const rl = readline.createInterface({
      input: process.stdin,
      output: process.stdout,
      terminal: true
    });

    readline.emitKeypressEvents(process.stdin, rl);
    if (process.stdin.isTTY) {
      process.stdin.setRawMode(true);
    }
    hideCursor();

    const cleanTitle = title.replace(/[:\s]+$/, '');

    function render() {
      if (lastRenderedLineCount > 0) {
        readline.cursorTo(process.stdout, 0);
        readline.moveCursor(process.stdout, 0, -lastRenderedLineCount);
      }

      let linesRendered = 0;
      function writeLine(str = '') {
        process.stdout.write(`\x1b[2K\r${str}\n`);
        linesRendered++;
      }

      const termWidth = getTerminalWidth();
      const maxPathLen = Math.max(20, termWidth - 48);

      writeLine(`  ${ANSI.bold}${ANSI.cyan}? ${cleanTitle}:${ANSI.reset}`);
      if (subtitle) {
        writeLine(`    ${ANSI.teal}${subtitle}${ANSI.reset}`);
      }
      writeLine(`    ${ANSI.dim}Use [↑/↓] to navigate, [Space] to toggle, [a] to select all, [Enter] to confirm${ANSI.reset}`);

      state.forEach((item, idx) => {
        const isCurrent = idx === cursor;
        const pointer = isCurrent ? `${ANSI.cyan}❯${ANSI.reset}` : ' ';
        const box = item.checked ? `${ANSI.green}☒${ANSI.reset}` : `${ANSI.darkGray}☐${ANSI.reset}`;
        const nameStr = isCurrent 
          ? `${ANSI.bold}${ANSI.white}${item.name.padEnd(28)}${ANSI.reset}` 
          : `${ANSI.gray}${item.name.padEnd(28)}${ANSI.reset}`;
        
        let pathBadge = '';
        if (item.installed) {
          const displayPath = truncatePath(item.path, maxPathLen);
          pathBadge = `${ANSI.teal}(Detected)${ANSI.reset} ${ANSI.dim}${displayPath}${ANSI.reset}`;
        } else {
          pathBadge = `${ANSI.darkGray}(Not installed)${ANSI.reset}`;
        }

        writeLine(`  ${pointer} ${box} ${nameStr} ${pathBadge}`);
      });

      lastRenderedLineCount = linesRendered;
    }

    render();

    function cleanup() {
      showCursor();
      process.stdin.removeListener('keypress', onKeyPress);
      if (process.stdin.isTTY) {
        process.stdin.setRawMode(false);
      }
      rl.close();
    }

    function onKeyPress(str, key) {
      if (!key) return;

      if (key.ctrl && key.name === 'c') {
        cleanup();
        console.log(`\n  ${ANSI.yellow}[ABORTED] Operation cancelled by user.${ANSI.reset}\n`);
        process.exit(0);
      }

      if (key.name === 'up') {
        cursor = (cursor - 1 + state.length) % state.length;
        render();
      } else if (key.name === 'down') {
        cursor = (cursor + 1) % state.length;
        render();
      } else if (key.name === 'space') {
        state[cursor].checked = !state[cursor].checked;
        render();
      } else if (key.name === 'a') {
        const allChecked = state.every(i => i.checked);
        state.forEach(i => i.checked = !allChecked);
        render();
      } else if (key.name === 'return' || key.name === 'enter') {
        cleanup();
        console.log('');
        const selectedIds = state.filter(i => i.checked).map(i => i.id);
        resolve(selectedIds);
      }
    }

    process.stdin.on('keypress', onKeyPress);
  });
}

function promptConfirm(question, defaultYes = true) {
  return new Promise((resolve) => {
    if (!process.stdin.isTTY) return resolve(defaultYes);

    const rl = readline.createInterface({
      input: process.stdin,
      output: process.stdout
    });

    const hint = defaultYes ? '[Y/n]' : '[y/N]';
    rl.question(`  ${ANSI.yellow}? ${question} ${ANSI.dim}${hint}${ANSI.reset}: `, (answer) => {
      rl.close();
      const clean = answer.trim().toLowerCase();
      if (!clean) return resolve(defaultYes);
      resolve(clean === 'y' || clean === 'yes');
    });
  });
}

process.on('exit', () => {
  showCursor();
});

module.exports = {
  truncatePath,
  promptRadio,
  promptCheckbox,
  promptConfirm
};
