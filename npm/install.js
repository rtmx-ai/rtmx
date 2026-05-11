#!/usr/bin/env node
// Install script for rtmx npm package.
// Downloads the appropriate Go binary for the current platform and architecture.

"use strict";

const os = require("os");
const fs = require("fs");
const path = require("path");
const https = require("https");
const { execSync } = require("child_process");

const VERSION = require("./package.json").version;
const REPO = "rtmx-ai/rtmx";

const PLATFORM_MAP = {
  darwin: "darwin",
  linux: "linux",
  win32: "windows",
};

const ARCH_MAP = {
  x64: "amd64",
  arm64: "arm64",
};

function getPlatform() {
  const platform = PLATFORM_MAP[os.platform()];
  if (!platform) {
    console.error(`Unsupported platform: ${os.platform()}`);
    process.exit(1);
  }
  return platform;
}

function getArch() {
  const arch = ARCH_MAP[os.arch()];
  if (!arch) {
    console.error(`Unsupported architecture: ${os.arch()}`);
    process.exit(1);
  }
  return arch;
}

function getBinaryName(platform) {
  return platform === "windows" ? "rtmx.exe" : "rtmx";
}

function getDownloadUrl(platform, arch) {
  const ext = platform === "windows" ? "zip" : "tar.gz";
  return `https://github.com/${REPO}/releases/download/v${VERSION}/rtmx_${VERSION}_${platform}_${arch}.${ext}`;
}

async function main() {
  const platform = getPlatform();
  const arch = getArch();
  const binaryName = getBinaryName(platform);
  const binDir = path.join(__dirname, "bin");
  const binaryPath = path.join(binDir, binaryName);

  // Skip if binary already exists
  if (fs.existsSync(binaryPath)) {
    console.log(`rtmx binary already exists at ${binaryPath}`);
    return;
  }

  const url = getDownloadUrl(platform, arch);
  console.log(`Downloading rtmx v${VERSION} for ${platform}/${arch}...`);
  console.log(`  ${url}`);

  try {
    // Use curl or wget for download (more reliable than https module for redirects)
    const tarball = path.join(binDir, `rtmx.tar.gz`);

    if (!fs.existsSync(binDir)) {
      fs.mkdirSync(binDir, { recursive: true });
    }

    execSync(`curl -sL "${url}" -o "${tarball}"`, { stdio: "inherit" });

    if (platform === "windows") {
      execSync(`unzip -o "${tarball}" -d "${binDir}"`, { stdio: "inherit" });
    } else {
      execSync(`tar -xzf "${tarball}" -C "${binDir}" --strip-components=1 rtmx`, { stdio: "inherit" });
    }

    fs.chmodSync(binaryPath, 0o755);
    fs.unlinkSync(tarball);

    console.log(`Installed rtmx v${VERSION} to ${binaryPath}`);
  } catch (err) {
    console.error(`Failed to install rtmx: ${err.message}`);
    console.error("You can manually download from:");
    console.error(`  https://github.com/${REPO}/releases/tag/v${VERSION}`);
    process.exit(1);
  }
}

main();
