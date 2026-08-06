#!/usr/bin/env node
"use strict";

const crypto = require("crypto");
const fs = require("fs");
const https = require("https");
const os = require("os");
const path = require("path");

const packageRoot = path.resolve(__dirname, "..");
const packageJSON = require(path.join(packageRoot, "package.json"));
const repo = "ZsTs119/patchnote-agent";

async function main(argv) {
  const parsed = parseArgs(argv);
  if (parsed.command !== "install") {
    throw new Error("usage: patchnote-agent install [--dry-run] [--print-config] [--install-dir <path>]");
  }

  const target = resolveTarget(parsed.options.platform || process.platform, parsed.options.arch || process.arch);
  const installDir = parsed.options.installDir || defaultInstallDir(process.platform);
  const binaryName = target.platform === "windows" ? "patchnote.exe" : "patchnote";
  const installPath = joinInstallPath(installDir, binaryName, target.platform);
  const version = packageJSON.version;
  const assetName = `patchnote_${version}_${target.platform}_${target.arch}${target.ext}`;
  const releaseBaseURL = process.env.PATCHNOTE_AGENT_RELEASE_BASE_URL || `https://github.com/${repo}/releases/download/v${version}`;
  const assetURL = `${releaseBaseURL}/${assetName}`;
  const checksumsURL = `${releaseBaseURL}/checksums.txt`;

  const plan = {
    version,
    platform: target.platform,
    arch: target.arch,
    install_dir: installDir,
    install_path: installPath,
    asset_url: assetURL,
    checksums_url: checksumsURL,
    dry_run: Boolean(parsed.options.dryRun)
  };

  if (parsed.options.dryRun) {
    printPlan(plan);
    if (parsed.options.printConfig) {
      printMCPConfig(installPath);
    }
    return;
  }

  let binary;
  if (parsed.options.fromLocal) {
    binary = await fs.promises.readFile(parsed.options.fromLocal);
  } else {
    binary = await download(assetURL);
    const checksums = await download(checksumsURL, "utf8");
    verifyChecksum(binary, assetName, checksums);
  }

  await fs.promises.mkdir(installDir, { recursive: true, mode: 0o755 });
  await fs.promises.writeFile(installPath, binary, { mode: target.platform === "windows" ? 0o644 : 0o755 });
  if (target.platform !== "windows") {
    await fs.promises.chmod(installPath, 0o755);
  }

  console.log(`Installed PatchNote Agent ${version} to ${installPath}`);
  if (parsed.options.printConfig) {
    printMCPConfig(installPath);
  } else {
    console.log("Run: patchnote login");
    console.log("MCP config: patchnote mcp serve");
  }
}

function parseArgs(argv) {
  const [command, ...rest] = argv;
  const options = {};
  for (let index = 0; index < rest.length; index += 1) {
    const arg = rest[index];
    switch (arg) {
      case "--dry-run":
        options.dryRun = true;
        break;
      case "--print-config":
        options.printConfig = true;
        break;
      case "--install-dir":
        options.installDir = requireValue(rest, ++index, arg);
        break;
      case "--from-local":
        options.fromLocal = requireValue(rest, ++index, arg);
        break;
      case "--platform":
        options.platform = requireValue(rest, ++index, arg);
        break;
      case "--arch":
        options.arch = requireValue(rest, ++index, arg);
        break;
      default:
        throw new Error(`unknown option ${arg}`);
    }
  }
  return { command, options };
}

function requireValue(values, index, option) {
  if (index >= values.length || values[index].startsWith("--")) {
    throw new Error(`${option} requires a value`);
  }
  return values[index];
}

function resolveTarget(platform, arch) {
  const platformMap = {
    darwin: "darwin",
    linux: "linux",
    win32: "windows",
    windows: "windows"
  };
  const archMap = {
    x64: "amd64",
    amd64: "amd64",
    arm64: "arm64"
  };
  const resolvedPlatform = platformMap[platform];
  const resolvedArch = archMap[arch];
  if (!resolvedPlatform || !resolvedArch) {
    throw new Error(`unsupported platform or architecture: ${platform}/${arch}`);
  }
  return {
    platform: resolvedPlatform,
    arch: resolvedArch,
    ext: resolvedPlatform === "windows" ? ".exe" : ""
  };
}

function defaultInstallDir(platform) {
  if (platform === "win32") {
    const base = process.env.LOCALAPPDATA || path.join(os.homedir(), "AppData", "Local");
    return path.join(base, "PatchNote Agent", "bin");
  }
  return path.join(os.homedir(), ".patchnote-agent", "bin");
}

function joinInstallPath(installDir, binaryName, targetPlatform) {
  if (targetPlatform === "windows") {
    return path.win32.join(installDir, binaryName);
  }
  return path.posix.join(installDir.replace(/\\/g, "/"), binaryName);
}

function printPlan(plan) {
  console.log("PatchNote Agent install dry run:");
  console.log(JSON.stringify(plan, null, 2));
}

function printMCPConfig(commandPath) {
  console.log("MCP config:");
  console.log(JSON.stringify({
    mcpServers: {
      patchnote: {
        command: commandPath,
        args: ["mcp", "serve"]
      }
    }
  }, null, 2));
}

function download(url, encoding) {
  return new Promise((resolve, reject) => {
    https.get(url, response => {
      if (response.statusCode < 200 || response.statusCode >= 300) {
        response.resume();
        reject(new Error(`download failed with status ${response.statusCode}`));
        return;
      }
      const chunks = [];
      response.on("data", chunk => chunks.push(chunk));
      response.on("end", () => {
        const body = Buffer.concat(chunks);
        resolve(encoding ? body.toString(encoding) : body);
      });
    }).on("error", reject);
  });
}

function verifyChecksum(binary, assetName, checksumsText) {
  const expectedLine = checksumsText.split(/\r?\n/).find(line => line.trim().endsWith(`  ${assetName}`));
  if (!expectedLine) {
    throw new Error(`checksum not found for ${assetName}`);
  }
  const expected = expectedLine.trim().split(/\s+/)[0];
  const actual = crypto.createHash("sha256").update(binary).digest("hex");
  if (actual !== expected) {
    throw new Error(`checksum mismatch for ${assetName}`);
  }
}

if (require.main === module) {
  main(process.argv.slice(2)).catch(error => {
    console.error(error.message);
    process.exitCode = 1;
  });
}

module.exports = {
  defaultInstallDir,
  parseArgs,
  resolveTarget,
  joinInstallPath,
  verifyChecksum
};
