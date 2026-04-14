#!/usr/bin/env node

const https = require('https');
const http = require('http');
const fs = require('fs');
const path = require('path');
const { pipeline } = require('stream');
const { promisify } = require('util');
const streamPipeline = promisify(pipeline);

const REPO = 'largeoliu/redmine-cli';
const BINARY_NAME = 'redmine';

function getPlatform() {
    switch (process.platform) {
        case 'win32': return 'windows';
        case 'darwin': return 'darwin';
        case 'linux': return 'linux';
        default: throw new Error(`Unsupported platform: ${process.platform}`);
    }
}

function getArch() {
    switch (process.arch) {
        case 'x64': return 'amd64';
        case 'arm64': return 'arm64';
        default: throw new Error(`Unsupported architecture: ${process.arch}`);
    }
}

async function fetchJson(url) {
    return new Promise((resolve, reject) => {
        https.get(url, { headers: { 'User-Agent': 'redmine-cli-npm' } }, (res) => {
            if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
                return fetchJson(res.headers.location).then(resolve).catch(reject);
            }
            if (res.statusCode !== 200) {
                return reject(new Error(`HTTP ${res.statusCode}`));
            }
            let data = '';
            res.on('data', chunk => data += chunk);
            res.on('end', () => {
                try {
                    resolve(JSON.parse(data));
                } catch (e) {
                    reject(e);
                }
            });
        }).on('error', reject);
    });
}

async function getLatestVersion() {
    const releases = await fetchJson(`https://api.github.com/repos/${REPO}/releases/latest`);
    return releases.tag_name;
}

async function downloadFile(url, dest) {
    return new Promise((resolve, reject) => {
        const protocol = url.startsWith('https') ? https : http;
        protocol.get(url, (res) => {
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
    const targetDir = path.dirname(destDir);
    
    if (isZip) {
        const unzipper = require('unzipper');
        await new Promise((resolve, reject) => {
            fs.createReadStream(archivePath)
                .pipe(unzipper.Extract({ path: targetDir }))
                .on('close', resolve)
                .on('error', reject);
        });
    } else {
        const tar = require('tar');
        await tar.x({ file: archivePath, cwd: targetDir });
    }
}

async function main() {
    const binDir = path.join(__dirname, '..', 'bin');
    const platform = getPlatform();
    const arch = getArch();
    
    console.log(`[INFO] Installing redmine-cli for ${platform}/${arch}...`);
    
    let version;
    try {
        version = await getLatestVersion();
        console.log(`[INFO] Latest version: ${version}`);
    } catch (e) {
        console.error('[ERROR] Failed to get latest version:', e.message);
        process.exit(1);
    }
    
    const ext = platform === 'windows' ? 'zip' : 'tar.gz';
    const archiveName = `${BINARY_NAME}_${version.replace('v', '')}_${platform}_${arch}.${ext}`;
    const downloadUrl = `https://github.com/${REPO}/releases/download/${version}/${archiveName}`;
    
    const tmpDir = fs.mkdtempSync(path.join(require('os').tmpdir(), 'redmine-cli-'));
    const archivePath = path.join(tmpDir, archiveName);
    
    try {
        console.log(`[INFO] Downloading ${archiveName}...`);
        await downloadFile(downloadUrl, archivePath);
        
        console.log('[INFO] Extracting...');
        await extractArchive(archivePath, binDir, platform === 'windows');
        
        const binaryName = platform === 'windows' ? `${BINARY_NAME}.exe` : BINARY_NAME;
        const extractedPath = path.join(tmpDir, binaryName);
        const destPath = path.join(binDir, binaryName);
        
        if (fs.existsSync(extractedPath)) {
            fs.renameSync(extractedPath, destPath);
        }
        
        if (platform !== 'windows') {
            fs.chmodSync(destPath, 0o755);
        }
        
        console.log('[INFO] Successfully installed redmine-cli');
    } catch (e) {
        console.error('[ERROR] Installation failed:', e.message);
        process.exit(1);
    } finally {
        fs.rmSync(tmpDir, { recursive: true, force: true });
    }
}

main();
