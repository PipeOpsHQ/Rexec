import { readFileSync } from 'fs';
import { createRequire } from 'module';
import { pathToFileURL } from 'url';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const env = Object.fromEntries(
  readFileSync(path.join(__dirname, '.env'), 'utf8').trim().split('\n').map(l => l.split('='))
);

// Use local built SDK
const mod = await import(pathToFileURL(path.resolve(__dirname, '../../sdk/js/dist/index.js')).href);
const { RexecClient } = mod;

const client = new RexecClient({ baseURL: env.URL, token: env.TOKEN });
console.log('[js] list...');
const before = await client.containers.list();
console.log('[js] list count', before.length);

console.log('[js] create...');
const c = await client.containers.create({ image: 'ubuntu', name: `js-e2e-${Date.now()}` });
console.log('[js] created', c.id, c.status, c.image);

console.log('[js] list after...');
const after = await client.containers.list();
console.log('[js] list count', after.length, 'ids', after.map(x => x.id).slice(0,3));

console.log('[js] get...');
const got = await client.containers.get(c.id);
console.log('[js] get', got.id, got.status);

console.log('[js] delete...');
await client.containers.delete(c.id);
console.log('[js] OK');
