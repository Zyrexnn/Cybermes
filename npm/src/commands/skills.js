/**
 * npm/src/commands/skills.js
 * Terminal search and viewer for Cybermes offensive security playbooks & SOPs.
 * 100% Zero external dependencies.
 */

const fs = require('fs');
const path = require('path');
const { ANSI, printHeader, printDivider } = require('../utils/ui');
const { VERSION } = require('../utils/binary');

function findSkillsDir() {
  const candidates = [
    path.resolve(__dirname, '..', '..', '..', 'skills'),
    path.resolve(process.cwd(), 'skills'),
  ];
  for (const c of candidates) {
    if (fs.existsSync(c) && fs.statSync(c).isDirectory()) {
      return c;
    }
  }
  return null;
}

function listSkills(query = '') {
  const skillsDir = findSkillsDir();
  if (!skillsDir) {
    return [];
  }

  const results = [];
  try {
    const entries = fs.readdirSync(skillsDir, { withFileTypes: true });
    for (const ent of entries) {
      if (ent.isDirectory()) {
        const skillPath = path.join(skillsDir, ent.name);
        const skillFile = path.join(skillPath, 'SKILL.md');
        let title = ent.name;
        let desc = '';

        if (fs.existsSync(skillFile)) {
          const content = fs.readFileSync(skillFile, 'utf8');
          const lines = content.split('\n');
          for (const line of lines) {
            if (line.startsWith('# ')) {
              title = line.replace('# ', '').trim();
            } else if (line.startsWith('description:') || (!desc && line.trim() && !line.startsWith('---') && !line.startsWith('#'))) {
              desc = line.replace('description:', '').trim();
            }
          }
        }

        if (!query || ent.name.toLowerCase().includes(query.toLowerCase()) || title.toLowerCase().includes(query.toLowerCase()) || desc.toLowerCase().includes(query.toLowerCase())) {
          results.push({ name: ent.name, title, desc, path: skillFile });
        }
      }
    }
  } catch (_) {}

  return results;
}

async function runSkills(args = []) {
  const query = (args[0] || '').trim();
  printHeader('CYBERMES MCP — SECURITY PLAYBOOKS & SKILLS', query ? `Search filter: "${query}"` : 'Listing all available offensive SOPs');

  const skills = listSkills(query);

  if (skills.length === 0) {
    console.log(`  ${ANSI.yellow}[INFO]${ANSI.reset} No skills found matching "${query}".\n`);
    return;
  }

  console.log(`  ${ANSI.bold}${'SKILL IDENTIFIER'.padEnd(28)} ${'DESCRIPTION'}${ANSI.reset}`);
  printDivider(74);

  for (const s of skills) {
    console.log(`  ${ANSI.cyan}${s.name.padEnd(28)}${ANSI.reset} ${ANSI.white}${s.title}${ANSI.reset}`);
    if (s.desc && s.desc !== s.title) {
      console.log(`    ${ANSI.dim}${s.desc}${ANSI.reset}`);
    }
  }

  printDivider(74);
  console.log(`  ${ANSI.gray}Total: ${skills.length} playbooks found. Accessible via LLM as ${ANSI.purple}skills://{skill_name}${ANSI.gray} URI.${ANSI.reset}\n`);
}

module.exports = {
  runSkills,
  listSkills
};
