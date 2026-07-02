import { spawn } from 'node:child_process';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const mode = process.argv[2];

if (!mode) {
  console.error('Usage: node scripts/run-vite-build.mjs <mode>');
  process.exit(1);
}

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(scriptDir, '..');
const viteBin = path.join(rootDir, 'node_modules', 'vite', 'bin', 'vite.js');
const env = { ...process.env };
const nodeMajor = Number(process.versions.node.split('.')[0]);

if (nodeMajor >= 22) {
  const storageDir = path.join(os.tmpdir(), 'pole-control-plane-vite-localstorage');
  fs.mkdirSync(storageDir, { recursive: true });
  const localStorageOption = `--localstorage-file=${path.join(storageDir, `${mode}.localstorage`)}`;
  env.NODE_OPTIONS = [env.NODE_OPTIONS, localStorageOption].filter(Boolean).join(' ');
}

const child = spawn(process.execPath, [viteBin, 'build', '--mode', mode], {
  cwd: rootDir,
  env,
  stdio: 'inherit',
});

child.on('exit', (code, signal) => {
  if (signal) {
    process.kill(process.pid, signal);
    return;
  }
  process.exit(code ?? 1);
});
