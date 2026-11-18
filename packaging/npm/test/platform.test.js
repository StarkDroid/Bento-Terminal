'use strict';

const test = require('node:test');
const assert = require('node:assert/strict');

const { assetFor, PLATFORMS } = require('../lib/platform');
const { checksumFor } = require('../lib/install');

test('every release asset has a platform mapping', () => {
  const assets = Object.values(PLATFORMS).flatMap((arches) =>
    Object.values(arches),
  );
  assert.deepEqual(assets.sort(), [
    'bento_darwin_amd64',
    'bento_darwin_arm64',
    'bento_linux_amd64',
    'bento_linux_arm64',
    'bento_windows_amd64.exe',
    'bento_windows_arm64.exe',
  ]);
});

test('assetFor picks the binary name per platform', () => {
  assert.deepEqual(assetFor('darwin', 'arm64'), {
    platform: 'darwin',
    arch: 'arm64',
    asset: 'bento_darwin_arm64',
    binary: 'bento',
  });
  assert.equal(assetFor('win32', 'x64').binary, 'bento.exe');
  assert.equal(assetFor('linux', 'x64').asset, 'bento_linux_amd64');
});

test('assetFor rejects unsupported hosts', () => {
  assert.throws(() => assetFor('freebsd', 'x64'), /unsupported platform/);
  assert.throws(() => assetFor('darwin', 'ia32'), /unsupported architecture/);
});

test('checksumFor reads the SHA256SUMS formats the release writes', () => {
  const sum = 'a'.repeat(64);
  const sums = [
    `${sum}  bento_darwin_arm64`,
    `${'b'.repeat(64)}  bento_linux_arm64`,
    `${'c'.repeat(64)} *bento_windows_amd64.exe`,
  ].join('\n');
  assert.equal(checksumFor(sums, 'bento_darwin_arm64'), sum);
  assert.equal(checksumFor(sums, 'bento_windows_amd64.exe'), 'c'.repeat(64));
  assert.equal(checksumFor(sums, 'missing'), '');
});
