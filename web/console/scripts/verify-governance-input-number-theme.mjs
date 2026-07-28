import fs from 'node:fs';
import path from 'node:path';

const rootDir = path.resolve('src/pages/Governance');

const walk = (dir) => fs.readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
    const fullPath = path.join(dir, entry.name);
    if (entry.isDirectory()) {
        return walk(fullPath);
    }
    return entry.isFile() && fullPath.endsWith('.tsx') ? [fullPath] : [];
});

const lineOf = (source, index) => source.slice(0, index).split('\n').length;
const hasNormalJsxTheme = (chunk) => /theme\s*=\s*["']normal["']/.test(chunk);
const hasNormalObjectTheme = (chunk) => /theme\s*:\s*["']normal["']/.test(chunk);

const failures = [];

for (const file of walk(rootDir)) {
    const source = fs.readFileSync(file, 'utf8');
    const relative = path.relative(process.cwd(), file);

    for (const match of source.matchAll(/<InputNumber\b[\s\S]*?\/>/g)) {
        const chunk = match[0];
        if (!hasNormalJsxTheme(chunk)) {
            failures.push(`${relative}:${lineOf(source, match.index ?? 0)} InputNumber missing theme="normal"`);
        }
    }

    for (const match of source.matchAll(/component:\s*InputNumber/g)) {
        const start = match.index ?? 0;
        const chunk = source.slice(start, start + 500);
        if (!hasNormalObjectTheme(chunk)) {
            failures.push(`${relative}:${lineOf(source, start)} table InputNumber editor missing theme: 'normal'`);
        }
    }
}

if (failures.length > 0) {
    console.error('Governance InputNumber theme verification failed:');
    for (const failure of failures) {
        console.error(`- ${failure}`);
    }
    process.exit(1);
}

console.log('Governance InputNumber theme verification passed.');
