#!/usr/bin/env node

const https = require('https');
const http = require('http');
const fs = require('fs');
const os = require('os');
const path = require('path');
const { pipeline } = require('stream');
const { promisify } = require('util');
const tar = require('tar');
const unzipper = require('unzipper');

const streamPipeline = promisify(pipeline);

const REPO = 'largeoliu/redmine-cli';
const RELEASE_ASSET_PREFIX = 'redmine-cli';
const BINARY_NAME = 'redmine';
const PACKAGE_VERSION = require('../package.json').version;

function getPlatform() {
  switch (process.platform) {
    case 'win32':
      return 'windows';
    case 'darwin':
      return 'darwin';
    case 'linux':
      return 'linux';
    default:
      throw new Error(`Unsupported platform: ${process.platform}`);
  }
}

function getArch() {
  switch (process.arch) {
    case 'x64':
      return 'amd64';
    case 'arm64':
      return 'arm64';
    default:
      throw new Error(`Unsupported architecture: ${process.arch}`);
  }
}

function getRequestedVersion() {
  const version = process.env.npm_package_version || PACKAGE_VERSION;
  if (!version) {
    throw new Error('Unable to determine package version for npm install');
  }
  return `v${version.replace(/^v/, '')}`;
}

function getBinaryName(platform) {
  if (platform === 'windows') {
    return `${BINARY_NAME}.exe`;
  }
  return BINARY_NAME;
}

async function downloadFile(url, dest) {
  return new Promise((resolve, reject) => {
    const protocol = url.startsWith('https') ? https : http;
    protocol.get(url, { headers: { 'User-Agent': 'redmine-cli-npm' } }, (res) => {
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        return downloadFile(res.headers.location, dest).then(resolve).catch(reject);
      }
      if (res.statusCode !== 200) {
        return reject(new Error(`HTTP ${res.statusCode}`));
      }
      const file = fs.createWriteStream(dest);
      streamPipeline(res, file).then(resolve).catch(reject);
    }).on('error', reject);
  });
}

async function extractArchive(archivePath, destDir, isZip) {
  if (isZip) {
    await new Promise((resolve, reject) => {
      fs.createReadStream(archivePath)
        .pipe(unzipper.Extract({ path: destDir }))
        .on('close', resolve)
        .on('error', reject);
    });
    return;
  }

  await tar.x({ file: archivePath, cwd: destDir });
}

function findFile(rootDir, targetName) {
  const entries = fs.readdirSync(rootDir, { withFileTypes: true });
  for (const entry of entries) {
    const entryPath = path.join(rootDir, entry.name);
    if (entry.isDirectory()) {
      const nested = findFile(entryPath, targetName);
      if (nested) {
        return nested;
      }
      continue;
    }

    if (entry.name === targetName) {
      return entryPath;
    }
  }

  return null;
}

async function main() {
  const binDir = path.join(__dirname, '..', 'bin');
  const platform = getPlatform();
  const arch = getArch();
  const version = getRequestedVersion();

  console.log(`[INFO] Installing redmine-cli ${version} for ${platform}/${arch}...`);

  const ext = platform === 'windows' ? 'zip' : 'tar.gz';
  const archiveName = `${RELEASE_ASSET_PREFIX}_${version.replace(/^v/, '')}_${platform}_${arch}.${ext}`;
  const binaryName = getBinaryName(platform);
  const downloadURL = `https://github.com/${REPO}/releases/download/${version}/${archiveName}`;
  const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'redmine-cli-'));
  const archivePath = path.join(tmpDir, archiveName);

  try {
    fs.mkdirSync(binDir, { recursive: true });

    console.log(`[INFO] Downloading ${archiveName}...`);
    await downloadFile(downloadURL, archivePath);

    console.log('[INFO] Extracting release archive...');
    await extractArchive(archivePath, tmpDir, platform === 'windows');

    const extractedBinary = findFile(tmpDir, binaryName);
    if (!extractedBinary) {
      throw new Error(`Binary ${binaryName} not found in ${archiveName}`);
    }

    const destination = path.join(binDir, binaryName);
    fs.copyFileSync(extractedBinary, destination);
    if (platform !== 'windows') {
      fs.chmodSync(destination, 0o755);
    }

    console.log(`[INFO] Installed ${binaryName} to ${destination}`);
  } catch (error) {
    console.error('[ERROR] Installation failed:', error.message);
    process.exit(1);
  } finally {
    fs.rmSync(tmpDir, { recursive: true, force: true });
  }
}

main();
