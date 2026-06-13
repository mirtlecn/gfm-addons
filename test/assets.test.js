import test from 'node:test';
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import { existsSync } from 'node:fs';
import { basename, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import {
  assets,
  getAsset,
  getAssetContentType,
  getAssetPath,
  getAssetRemoteUrl,
} from '../assets.js';
import { embeddedAssets, getEmbeddedAsset, getEmbeddedAssetContent } from '../embedded.js';
import { assetDefinitions } from '../scripts/assets.mjs';
import { createMinifiedFilePath, minifyAssetContent } from '../scripts/minify-assets.mjs';

const rootDirectory = fileURLToPath(new URL('..', import.meta.url));
const packageJson = JSON.parse(await readFile(join(rootDirectory, 'package.json'), 'utf8'));
const expectedKeys = [
  'ravel.gfm.css',
  'whitey.gfm.css',
  'newsprint.gfm.css',
  'github.gfm.css',
  'folio.gfm.css',
  'highlight-light.css',
  'highlight-dark.css',
  'highlight-core.js',
  'gfm-addons.css',
  'gfm-addons.js',
];

function expectedRemoteUrl(filePath) {
  return `https://cdn.jsdelivr.net/npm/${packageJson.name}@${packageJson.version}/${createMinifiedFilePath(filePath)}`;
}

test('exports the expected GFM asset keys', () => {
  assert.deepEqual(assets.map((asset) => asset.key), expectedKeys);
});

test('asset keys are generated from file names', () => {
  assert.deepEqual(
    assets.map((asset) => asset.key),
    assets.map((asset) => basename(asset.file)),
  );
  assert.equal(
    assetDefinitions.some((asset) => Object.prototype.hasOwnProperty.call(asset, 'key')),
    false,
  );
});

test('manifest matches the JavaScript asset exports', async () => {
  const manifest = JSON.parse(await readFile(join(rootDirectory, 'manifest.json'), 'utf8'));
  assert.deepEqual(manifest, assets);
});

test('embedded assets match the JavaScript asset exports', () => {
  assert.deepEqual(
    embeddedAssets.map(({ contentBase64, ...asset }) => asset),
    assets,
  );
});

test('each manifest file exists in the package source', () => {
  for (const asset of assets) {
    assert.equal(existsSync(join(rootDirectory, asset.file)), true, `${asset.file} should exist`);
  }
});

test('embedded asset content matches minified packaged files', async () => {
  for (const asset of assets) {
    const fileContent = await readFile(join(rootDirectory, asset.file));
    const minifiedContent = await minifyAssetContent(asset.file, fileContent);
    const embedded = getEmbeddedAsset(asset.key);

    assert.equal(embedded.contentBase64, minifiedContent.toString('base64'));
    assert.equal(
      Buffer.compare(getEmbeddedAssetContent(asset.key, null), minifiedContent),
      0,
      `${asset.key} embedded bytes differ from ${asset.file}`,
    );
  }
});

test('asset helpers return stable file paths, content types, and remote URLs', () => {
  assert.equal(getAssetPath('ravel.gfm.css'), 'assets/ravel.gfm.css');
  assert.equal(getAssetPath('folio.gfm.css'), 'assets/folio.gfm.css');
  assert.equal(getAssetPath('highlight-core.js'), 'assets/highlight-core.js');
  assert.equal(getAssetContentType('ravel.gfm.css'), 'text/css; charset=utf-8');
  assert.equal(getAssetContentType('gfm-addons.js'), 'application/javascript; charset=utf-8');
  assert.equal(getAssetRemoteUrl('ravel.gfm.css'), expectedRemoteUrl('assets/ravel.gfm.css'));
  assert.equal(getAssetRemoteUrl('folio.gfm.css'), expectedRemoteUrl('assets/folio.gfm.css'));
  assert.deepEqual(getAsset('gfm-addons.js'), {
    key: 'gfm-addons.js',
    file: 'assets/gfm-addons.js',
    contentType: 'application/javascript; charset=utf-8',
    remoteUrl: expectedRemoteUrl('assets/gfm-addons.js'),
  });
});

test('unknown asset keys throw', () => {
  assert.throws(() => getAssetPath('missing'), /Unknown GFM asset/);
  assert.throws(() => getAssetRemoteUrl('missing'), /Unknown GFM asset/);
  assert.throws(() => getEmbeddedAsset('missing'), /Unknown GFM asset/);
});
