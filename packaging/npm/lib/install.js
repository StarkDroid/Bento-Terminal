'use strict';

const crypto = require('node:crypto');
const fs = require('node:fs');
const https = require('node:https');
const os = require('node:os');
const path = require('node:path');
const { pipeline } = require('node:stream/promises');

const pkg = require('../package.json');
const { current } = require('./platform');

const ROOT = path.join(__dirname, '..');
const VENDOR = path.join(ROOT, 'vendor');
const VERSION_FILE = path.join(VENDOR, 'VERSION');
const MAX_REDIRECTS = 8;

function releaseVersion() {
  return process.env.BENTO_VERSION || pkg.version;
}

function releaseBase() {
  const version = releaseVersion();
  if (!version || version === '0.0.0') {
    throw new Error(
      'this package is a development checkout; set BENTO_VERSION to a published release, for example BENTO_VERSION=0.7.0',
    );
  }
  return (
    process.env.BENTO_BASE_URL ||
    `https://github.com/StarkDroid/Bento-Terminal/releases/download/v${version}`
  ).replace(/\/+$/, '');
}

function binaryPath() {
  return path.join(VENDOR, current().binary);
}

function installedVersion() {
  try {
    return fs.readFileSync(VERSION_FILE, 'utf8').trim();
  } catch {
    return '';
  }
}

function isInstalled() {
  const binary = binaryPath();
  try {
    return (
      fs.statSync(binary).size > 0 && installedVersion() === releaseVersion()
    );
  } catch {
    return false;
  }
}

function get(url, redirects = 0) {
  return new Promise((resolve, reject) => {
    if (redirects > MAX_REDIRECTS) {
      reject(new Error(`too many redirects while fetching ${url}`));
      return;
    }
    const req = https.get(
      url,
      {
        headers: {
          'User-Agent': `bento-terminal/${pkg.version} (${process.platform}/${process.arch})`,
          Accept: 'application/octet-stream',
        },
      },
      (res) => {
        const status = res.statusCode || 0;
        if (status >= 300 && status < 400 && res.headers.location) {
          res.resume();
          const next = new URL(res.headers.location, url).toString();
          get(next, redirects + 1).then(resolve, reject);
          return;
        }
        if (status !== 200) {
          res.resume();
          reject(new Error(`GET ${url} returned HTTP ${status}`));
          return;
        }
        resolve(res);
      },
    );
    req.on('error', reject);
    req.setTimeout(30000, () => {
      req.destroy(new Error(`timed out fetching ${url}`));
    });
  });
}

async function download(url, destination) {
  const res = await get(url);
  await pipeline(res, fs.createWriteStream(destination, { mode: 0o755 }));
}

function sha256(file) {
  return new Promise((resolve, reject) => {
    const hash = crypto.createHash('sha256');
    const stream = fs.createReadStream(file);
    stream.on('data', (chunk) => hash.update(chunk));
    stream.on('error', reject);
    stream.on('end', () => resolve(hash.digest('hex')));
  });
}

function checksumFor(sums, filename) {
  for (const line of sums.split(/\r?\n/)) {
    const match = line.match(/^([a-fA-F0-9]{64})\s+\*?(.+?)\s*$/);
    if (match && match[2] === filename) {
      return match[1].toLowerCase();
    }
  }
  return '';
}

async function ensure(options = {}) {
  const quiet = options.quiet === true;
  if (!process.env.BENTO_FORCE_DOWNLOAD && isInstalled()) {
    return binaryPath();
  }

  const info = current();
  const version = releaseVersion();
  const base = releaseBase();
  const target = path.join(VENDOR, info.binary);
  const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'bento-terminal-'));
  const partial = path.join(tmp, info.asset);
  const sumsPath = path.join(tmp, 'SHA256SUMS');

  if (!quiet) {
    process.stderr.write(
      `Downloading Bento terminal v${version} for ${info.platform}/${info.arch}…\n`,
    );
  }

  try {
    await download(`${base}/${info.asset}`, partial);
    await download(`${base}/SHA256SUMS`, sumsPath);
    const expected = checksumFor(fs.readFileSync(sumsPath, 'utf8'), info.asset);
    if (!expected) {
      throw new Error(`SHA256SUMS does not list ${info.asset}`);
    }
    const actual = await sha256(partial);
    if (actual !== expected) {
      throw new Error(
        `checksum mismatch for ${info.asset}: expected ${expected}, got ${actual}`,
      );
    }
    fs.mkdirSync(VENDOR, { recursive: true });
    fs.rmSync(target, { force: true });
    try {
      fs.renameSync(partial, target);
    } catch {
      // The temporary directory can be on another volume on Windows.
      fs.copyFileSync(partial, target);
    }
    fs.chmodSync(target, 0o755);
    fs.writeFileSync(VERSION_FILE, `${version}\n`);
  } finally {
    fs.rmSync(tmp, { recursive: true, force: true });
  }

  if (!quiet) {
    process.stderr.write(`Installed Bento terminal to ${target}\n`);
  }
  return target;
}

module.exports = {
  ensure,
  binaryPath,
  isInstalled,
  releaseVersion,
  releaseBase,
  checksumFor,
};
