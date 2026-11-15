'use strict';

const os = require('node:os');

// assetFor maps a Node platform and architecture to the release asset name.
// Keep this in step with the six assets scripts/release.sh publishes.
const PLATFORMS = {
  darwin: {
    arm64: 'bento_darwin_arm64',
    x64: 'bento_darwin_amd64',
  },
  linux: {
    arm64: 'bento_linux_arm64',
    x64: 'bento_linux_amd64',
  },
  win32: {
    arm64: 'bento_windows_arm64.exe',
    x64: 'bento_windows_amd64.exe',
  },
};

function assetFor(platform = process.platform, arch = process.arch) {
  const assets = PLATFORMS[platform];
  if (!assets) {
    throw new Error(
      `unsupported platform ${platform}. Bento terminal ships binaries for macOS, Linux, and Windows.`,
    );
  }
  const asset = assets[arch];
  if (!asset) {
    throw new Error(
      `unsupported architecture ${arch} on ${platform}. Bento terminal ships arm64 and x64 binaries.`,
    );
  }
  return {
    platform,
    arch,
    asset,
    binary: platform === 'win32' ? 'bento.exe' : 'bento',
  };
}

function current() {
  return assetFor(process.platform, process.arch);
}

module.exports = {
  PLATFORMS,
  assetFor,
  current,
  hostDescription: () => `${os.platform()}/${os.arch()}`,
};
