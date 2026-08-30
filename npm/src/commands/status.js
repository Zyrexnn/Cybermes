/**
 * npm/src/commands/status.js
 * Visual discovery matrix of client configs across 13+ AI IDEs and CLI agents.
 * 100% Zero external dependencies.
 */

const fs = require('fs');
const os = require('os');
const { ANSI, printBanner, printDivider, badge } = require('../utils/ui');
const { VERSION } = require('../utils/binary');
const { getClientDefinitions } = require('../adapters/clients');
const { checkClientStatus } = require('../adapters/injector');

async function runStatus() {
  printBanner(`Client Discovery & Configuration Matrix (${os.platform()} ${os.arch()})`);

  const clients = getClientDefinitions();

  console.log(`  ${ANSI.bold}${'CLIENT / IDE'.padEnd(26)} ${'STATUS'.padEnd(16)} CONFIG LOCATION${ANSI.reset}`);
  printDivider(74);

  for (const client of clients) {
    const existingPath = client.paths.find(p => fs.existsSync(p));
    const targetPath = existingPath || client.paths[0];
    const status = checkClientStatus(client, targetPath);

    let statusBadge = badge('NOT DETECTED', 'not_detected');
    let nameStr = `${ANSI.darkGray}${client.name.padEnd(26)}${ANSI.reset}`;
    let pathStr = `${ANSI.darkGray}(Not installed)${ANSI.reset}`;

    if (status.configured) {
      statusBadge = badge('CONFIGURED', 'configured');
      nameStr = `${ANSI.bold}${ANSI.white}${client.name.padEnd(26)}${ANSI.reset}`;
      pathStr = `${ANSI.gray}${status.path}${ANSI.reset}`;
    } else if (status.installed) {
      if (status.error) {
        statusBadge = badge('CORRUPT JSON', 'error');
        nameStr = `${ANSI.red}${client.name.padEnd(26)}${ANSI.reset}`;
        pathStr = `${ANSI.red}${status.path} (${status.error})${ANSI.reset}`;
      } else {
        statusBadge = badge('NOT WIRED', 'unwired');
        nameStr = `${ANSI.white}${client.name.padEnd(26)}${ANSI.reset}`;
        pathStr = `${ANSI.yellow}${status.path}${ANSI.reset}`;
      }
    }

    console.log(`  ${nameStr} ${statusBadge} ${pathStr}`);
  }

  printDivider(74);
  console.log(`  ${ANSI.gray}Run ${ANSI.cyan}npx cybermes-mcp install${ANSI.gray} to configure un-wired clients with 1 click.${ANSI.reset}\n`);
}

module.exports = {
  runStatus
};
