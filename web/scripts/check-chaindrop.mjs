#!/usr/bin/env node
/**
 * Guard against ChainDrop / keyv-cacheable compromised releases (2026-08).
 * Fails if lockfiles or node_modules resolve to known-bad versions.
 */
import fs from 'node:fs';
import path from 'node:path';

const BAD = new Map([
  ['keyv', new Set(['6.0.0'])],
  ['flat-cache', new Set(['6.1.24'])],
  ['file-entry-cache', new Set(['11.1.6'])],
  ['cacheable-request', new Set(['13.0.20'])],
  ['cacheable', new Set(['2.5.1'])],
  ['@cacheable/memory', new Set(['2.2.1'])],
  ['cache-manager', new Set(['7.2.10'])],
  ['@cacheable/node-cache', new Set(['3.1.2'])],
  ['@cacheable/utils', new Set(['2.5.1'])],
  ['@cacheable/net', new Set(['2.1.1'])],
  ['ecto', new Set(['5.0.1'])],
]);

const root = process.cwd();
const hits = [];

function checkText(file, text) {
  for (const [name, versions] of BAD) {
    for (const ver of versions) {
      const token = `${name}@${ver}`;
      if (text.includes(token)) hits.push(`${file}: ${token}`);
    }
  }
}

for (const lock of ['pnpm-lock.yaml', 'package-lock.json', 'yarn.lock']) {
  const p = path.join(root, lock);
  if (fs.existsSync(p)) checkText(lock, fs.readFileSync(p, 'utf8'));
}

function checkPackageJson(file) {
  try {
    const pkg = JSON.parse(fs.readFileSync(file, 'utf8'));
    const name = pkg.name;
    const ver = pkg.version;
    if (name && BAD.has(name) && BAD.get(name).has(ver)) {
      hits.push(`${path.relative(root, file)}: ${name}@${ver}`);
    }
    const pre = pkg.scripts?.preinstall;
    if (name && BAD.has(name) && typeof pre === 'string' && /setup\.mjs/.test(pre)) {
      hits.push(`${path.relative(root, file)}: malicious preinstall -> ${pre}`);
    }
  } catch {}
}

function walk(dir) {
  let entries;
  try {
    entries = fs.readdirSync(dir, { withFileTypes: true });
  } catch {
    return;
  }
  for (const ent of entries) {
    const full = path.join(dir, ent.name);
    if (ent.isDirectory()) {
      // prune heavy unrelated trees
      if (ent.name === '.git' || ent.name === 'dist' || ent.name === 'coverage') continue;
      walk(full);
    } else if (ent.name === 'package.json') {
      // only care about installed deps
      if (full.includes(`${path.sep}node_modules${path.sep}`)) checkPackageJson(full);
    } else if (ent.name === 'setup.mjs') {
      // loader IoC inside known package names
      const parent = path.basename(path.dirname(full));
      const grand = path.basename(path.dirname(path.dirname(full)));
      const candidates = new Set([...BAD.keys()].map((n) => n.split('/').pop()));
      if (candidates.has(parent) || (grand.startsWith('@') && candidates.has(`${grand}/${parent}`))) {
        hits.push(`${path.relative(root, full)}: present`);
      }
    }
  }
}

const nm = path.join(root, 'node_modules');
if (fs.existsSync(nm)) walk(nm);

if (hits.length) {
  console.error('ChainDrop check FAILED:');
  for (const h of [...new Set(hits)]) console.error(' -', h);
  process.exit(1);
}
console.log('ChainDrop check OK: no compromised keyv/cacheable releases found.');
