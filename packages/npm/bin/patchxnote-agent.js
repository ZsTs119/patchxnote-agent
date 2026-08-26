#!/usr/bin/env node
"use strict";

const crypto = require("crypto");
const fs = require("fs");
const https = require("https");
const os = require("os");
const path = require("path");
const { spawn, spawnSync } = require("child_process");

const packageRoot = path.resolve(__dirname, "..");
const packageJSON = require(path.join(packageRoot, "package.json"));
const repo = "ZsTs119/patchxnote-agent";
const lockTimeoutMs = 10000;
const staleLockMs = 10 * 60 * 1000;

async function main(argv) {
  const parsed = parseArgs(argv);

  if (parsed.command === "install" || parsed.command === "update") {
    await runInstall(parsed);
    return;
  }
  if (parsed.command === "uninstall") {
    await runUninstall(parsed);
    return;
  }
  if (parsed.command === "login") {
    await runLogin(parsed);
    return;
  }
  if (parsed.command === "mcp" && parsed.subcommand === "config") {
    printUniversalMCPConfig();
    return;
  }
  if (parsed.command === "mcp" && parsed.subcommand === "serve") {
    await runMCPServe(parsed);
    return;
  }

  throw new Error(usage());
}

async function runInstall(parsed) {
  const plan = createInstallPlan(parsed.command, parsed.options);
  if (parsed.options.dryRun) {
    printPlan(plan);
    if (parsed.options.printConfig) {
      printMCPConfig(plan.install_path);
    }
    return;
  }
  await withInstallLock(plan, async () => installBinary(plan, parsed.options));
  console.log(`Installed PatchXNote Agent ${plan.version} to ${plan.install_path}`);
  printPathGuidance(plan.install_dir, plan.platform);
  if (parsed.options.printConfig) {
    printMCPConfig(plan.install_path);
  } else {
    console.log("Run: patchxnote login");
    console.log("MCP config: patchxnote mcp serve");
  }
}

async function runUninstall(parsed) {
  const plan = createInstallPlan(parsed.command, parsed.options);
  await uninstall(plan, parsed.options);
}

async function runLogin(parsed) {
  const plan = await ensureBinary(parsed.options, { stderr: process.stderr });
  const code = await spawnAndWait(plan.install_path, ["login", ...parsed.passthroughArgs], {
    stdio: "inherit"
  });
  process.exitCode = code;
}

async function runMCPServe(parsed) {
  const plan = await ensureBinary(parsed.options, { stderr: process.stderr });
  const code = await spawnAndWait(plan.install_path, ["mcp", "serve", ...parsed.passthroughArgs], {
    stdio: ["pipe", "inherit", "inherit"],
    pipeStdin: true
  });
  process.exitCode = code;
}

function createInstallPlan(command, options = {}) {
  const target = resolveTarget(options.platform || process.platform, options.arch || process.arch);
  const installDir = options.installDir || defaultInstallDir(process.platform);
  const binaryName = target.platform === "windows" ? "patchxnote.exe" : "patchxnote";
  const installPath = joinInstallPath(installDir, binaryName, target.platform);
  const version = packageJSON.version;
  const assetName = `patchxnote_${version}_${target.platform}_${target.arch}${target.ext}`;
  const releaseBaseURL = process.env.PATCHXNOTE_AGENT_RELEASE_BASE_URL
    || process.env.PATCHNOTE_AGENT_RELEASE_BASE_URL
    || `https://github.com/${repo}/releases/download/v${version}`;
  const assetURL = `${releaseBaseURL}/${assetName}`;
  const checksumsURL = `${releaseBaseURL}/checksums.txt`;

  const plan = {
    version,
    platform: target.platform,
    arch: target.arch,
    install_dir: installDir,
    install_path: installPath,
    install_dir_on_path: isInstallDirOnPath(installDir, process.env.PATH || "", target.platform),
    path_hint: pathHint(installDir, target.platform),
    asset_name: assetName,
    asset_url: assetURL,
    checksums_url: checksumsURL,
    action: command,
    dry_run: Boolean(options.dryRun)
  };
  return plan;
}

async function installBinary(plan, options = {}) {
  let binary;
  if (options.fromLocal) {
    binary = await fs.promises.readFile(options.fromLocal);
  } else {
    binary = await download(plan.asset_url);
    const checksums = await download(plan.checksums_url, "utf8");
    verifyChecksum(binary, plan.asset_name, checksums);
  }

  await fs.promises.mkdir(plan.install_dir, { recursive: true, mode: 0o755 });
  const tempPath = `${plan.install_path}.tmp-${process.pid}-${Date.now()}`;
  try {
    await fs.promises.writeFile(tempPath, binary, { mode: plan.platform === "windows" ? 0o644 : 0o755 });
    if (plan.platform !== "windows") {
      await fs.promises.chmod(tempPath, 0o755);
      await assertExecutable(tempPath);
    }
    await replaceInstalledBinary(tempPath, plan.install_path);
    if (plan.platform !== "windows") {
      await assertExecutable(plan.install_path);
    }
  } finally {
    await fs.promises.rm(tempPath, { force: true }).catch(() => {});
  }
}

async function replaceInstalledBinary(tempPath, installPath) {
  const backupPath = `${installPath}.bak-${process.pid}-${Date.now()}`;
  let hasBackup = false;
  try {
    await fs.promises.rename(installPath, backupPath);
    hasBackup = true;
  } catch (error) {
    if (error.code !== "ENOENT") {
      throw error;
    }
  }

  try {
    await fs.promises.rename(tempPath, installPath);
  } catch (error) {
    if (hasBackup) {
      await fs.promises.rename(backupPath, installPath).catch(() => {});
    }
    throw error;
  }

  if (hasBackup) {
    await fs.promises.rm(backupPath, { force: true }).catch(() => {});
  }
}

async function uninstall(plan, options) {
  if (options.dryRun) {
    printPlan(plan);
    return;
  }
  await fs.promises.rm(plan.install_path, { force: true });
  console.log(`Removed PatchXNote Agent from ${plan.install_path}`);
}

async function assertExecutable(installPath) {
  const stat = await fs.promises.stat(installPath);
  if ((stat.mode & 0o111) === 0) {
    throw new Error(`installed binary is not executable: ${installPath}`);
  }
}

async function ensureBinary(options = {}, io = {}) {
  const plan = createInstallPlan("install", options);
  return withInstallLock(plan, async () => {
    const inspected = inspectInstalledBinary(plan);
    if (inspected.ok && inspected.version === plan.version) {
      return plan;
    }

    if (!inspected.exists) {
      writeDiagnostic(io, `PatchXNote Agent binary missing; installing ${plan.version}.\n`);
    } else if (inspected.ok) {
      writeDiagnostic(io, `PatchXNote Agent binary version ${inspected.version} does not match package ${plan.version}; reinstalling.\n`);
    } else {
      writeDiagnostic(io, `PatchXNote Agent binary preflight failed; reinstalling ${plan.version}.\n`);
    }

    await installBinary(plan, options);
    const afterInstall = inspectInstalledBinary(plan);
    if (!afterInstall.ok) {
      throw new Error(`installed PatchXNote Agent binary failed preflight: ${afterInstall.reason}`);
    }
    if (afterInstall.version !== plan.version) {
      throw new Error(`installed PatchXNote Agent binary version ${afterInstall.version} does not match package ${plan.version}`);
    }
    return plan;
  });
}

function inspectInstalledBinary(plan) {
  if (!fs.existsSync(plan.install_path)) {
    return { exists: false, ok: false, reason: "missing" };
  }
  const result = spawnSync(plan.install_path, ["version", "--output", "json"], {
    encoding: "utf8",
    stdio: ["ignore", "pipe", "pipe"],
    timeout: 5000,
    windowsHide: true,
    maxBuffer: 1024 * 1024
  });
  if (result.error) {
    return { exists: true, ok: false, reason: result.error.message };
  }
  if (result.status !== 0) {
    return { exists: true, ok: false, reason: `exit ${result.status}` };
  }
  let decoded;
  try {
    decoded = JSON.parse(result.stdout.trim());
  } catch (error) {
    return { exists: true, ok: false, reason: "invalid version output" };
  }
  return { exists: true, ok: true, version: String(decoded.version || "") };
}

async function withInstallLock(plan, action) {
  await fs.promises.mkdir(plan.install_dir, { recursive: true, mode: 0o755 });
  const release = await acquireInstallLock(`${plan.install_path}.lock`);
  try {
    return await action();
  } finally {
    await release();
  }
}

async function acquireInstallLock(lockDir) {
  const startedAt = Date.now();
  for (;;) {
    try {
      await fs.promises.mkdir(lockDir);
      return async () => {
        await fs.promises.rmdir(lockDir).catch(() => {});
      };
    } catch (error) {
      if (error.code !== "EEXIST") {
        throw error;
      }
      await removeStaleLock(lockDir).catch(() => {});
      if (Date.now() - startedAt > lockTimeoutMs) {
        throw new Error(`install lock timed out: ${lockDir}`);
      }
      await delay(100);
    }
  }
}

async function removeStaleLock(lockDir) {
  const stat = await fs.promises.stat(lockDir);
  if (Date.now() - stat.mtimeMs > staleLockMs) {
    await fs.promises.rmdir(lockDir);
  }
}

function delay(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

function writeDiagnostic(io, message) {
  if (io.stderr && typeof io.stderr.write === "function") {
    io.stderr.write(message);
  }
}

function spawnAndWait(command, args, options = {}) {
  return new Promise((resolve, reject) => {
    const child = spawn(command, args, {
      stdio: options.stdio || "inherit",
      windowsHide: true
    });
    let settled = false;
    let stdinEndHandler;
    let stdinErrorHandler;
    let exitHandler;

    const cleanup = () => {
      for (const [signal, handler] of signalHandlers) {
        process.off(signal, handler);
      }
      if (stdinEndHandler) {
        process.stdin.off("end", stdinEndHandler);
      }
      if (stdinErrorHandler) {
        process.stdin.off("error", stdinErrorHandler);
      }
      if (exitHandler) {
        process.off("exit", exitHandler);
      }
    };

    const signalHandlers = ["SIGINT", "SIGTERM"].map(signal => {
      const handler = () => {
        if (!child.killed) {
          child.kill(signal);
        }
      };
      process.on(signal, handler);
      return [signal, handler];
    });

    if (options.pipeStdin && child.stdin) {
      child.stdin.on("error", () => {});
      stdinEndHandler = () => child.stdin.end();
      stdinErrorHandler = () => child.stdin.destroy();
      process.stdin.on("end", stdinEndHandler);
      process.stdin.on("error", stdinErrorHandler);
      process.stdin.pipe(child.stdin);
    }

    exitHandler = () => {
      if (!settled && !child.killed) {
        child.kill();
      }
    };
    process.on("exit", exitHandler);

    child.on("error", error => {
      settled = true;
      cleanup();
      reject(error);
    });
    child.on("exit", (code, signal) => {
      settled = true;
      cleanup();
      resolve(code == null ? signalExitCode(signal) : code);
    });
  });
}

function signalExitCode(signal) {
  if (signal === "SIGINT") {
    return 130;
  }
  if (signal === "SIGTERM") {
    return 143;
  }
  return 1;
}

function parseArgs(argv) {
  const [command, ...rest] = argv;
  if (!command) {
    throw new Error(usage());
  }
  if (["install", "update", "uninstall"].includes(command)) {
    return { command, options: parseInstallOptions(command, rest), passthroughArgs: [] };
  }
  if (command === "login") {
    const launcher = parseLauncherOptions(rest, { allowPassthrough: true, rejectPrintConfig: true });
    return { command, options: launcher.options, passthroughArgs: launcher.passthroughArgs };
  }
  if (command === "mcp") {
    const [subcommand, ...subcommandArgs] = rest;
    if (!["serve", "config"].includes(subcommand)) {
      throw new Error(mcpUsage());
    }
    const launcher = parseLauncherOptions(subcommandArgs, {
      allowPassthrough: subcommand === "serve",
      rejectPrintConfig: true
    });
    return { command, subcommand, options: launcher.options, passthroughArgs: launcher.passthroughArgs };
  }
  throw new Error(usage());
}

function parseInstallOptions(command, rest) {
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
        if (command === "uninstall") {
          throw new Error("--from-local is not valid for uninstall");
        }
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
  return options;
}

function parseLauncherOptions(rest, settings = {}) {
  const options = {};
  const passthroughArgs = [];
  for (let index = 0; index < rest.length; index += 1) {
    const arg = rest[index];
    switch (arg) {
      case "--":
        if (settings.allowPassthrough) {
          passthroughArgs.push(...rest.slice(index + 1));
          index = rest.length;
          break;
        }
        throw new Error(`unknown option ${arg}`);
      case "--dry-run":
        throw new Error("--dry-run is only valid for install/update/uninstall");
      case "--print-config":
        if (settings.rejectPrintConfig) {
          throw new Error("--print-config is not valid for this command");
        }
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
        if (!settings.allowPassthrough) {
          throw new Error(`unknown option ${arg}`);
        }
        passthroughArgs.push(arg);
    }
  }
  return { options, passthroughArgs };
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
    return path.join(base, "PatchXNote Agent", "bin");
  }
  return path.join(os.homedir(), ".patchxnote-agent", "bin");
}

function joinInstallPath(installDir, binaryName, targetPlatform) {
  if (targetPlatform === "windows") {
    return path.win32.join(installDir, binaryName);
  }
  return path.posix.join(installDir.replace(/\\/g, "/"), binaryName);
}

function printPlan(plan) {
  console.log(`PatchXNote Agent ${plan.action} dry run:`);
  console.log(JSON.stringify(plan, null, 2));
}

function printMCPConfig(commandPath) {
  console.log("MCP config:");
  console.log(JSON.stringify({
    mcpServers: {
      patchxnote: {
        command: commandPath,
        args: ["mcp", "serve"]
      }
    }
  }, null, 2));
}

function printUniversalMCPConfig() {
  console.log(JSON.stringify(universalMCPConfig(), null, 2));
}

function universalMCPConfig() {
  return {
    mcpServers: {
      patchxnote: {
        command: "npx",
        args: ["-y", "patchxnote-agent@latest", "mcp", "serve"]
      }
    }
  };
}

function isInstallDirOnPath(installDir, pathValue, targetPlatform) {
  if (!pathValue) {
    return false;
  }
  const delimiter = targetPlatform === "windows" ? ";" : ":";
  const normalize = targetPlatform === "windows"
    ? value => path.win32.normalize(value).toLowerCase()
    : value => path.posix.normalize(value.replace(/\\/g, "/"));
  const expected = normalize(installDir);
  return pathValue.split(delimiter).some(entry => entry.trim() && normalize(entry.trim()) === expected);
}

function pathHint(installDir, targetPlatform) {
  if (targetPlatform === "windows") {
    const escaped = installDir.replace(/"/g, '\\"');
    return `[Environment]::SetEnvironmentVariable("Path", [Environment]::GetEnvironmentVariable("Path", "User") + ";${escaped}", "User")`;
  }
  return `export PATH="${installDir.replace(/"/g, '\\"')}:$PATH"`;
}

function printPathGuidance(installDir, targetPlatform) {
  if (isInstallDirOnPath(installDir, process.env.PATH || "", targetPlatform)) {
    return;
  }
  console.log("To use patchxnote directly from your terminal, add the install directory to PATH:");
  console.log(pathHint(installDir, targetPlatform));
  if (targetPlatform === "windows") {
    console.log("Open a new terminal after updating PATH.");
  }
}

function download(url, encoding, redirects = 0) {
  return new Promise((resolve, reject) => {
    https.get(url, response => {
      if (response.statusCode >= 300 && response.statusCode < 400) {
        response.resume();
        if (!response.headers.location) {
          reject(new Error(`download redirect missing location for ${url}`));
          return;
        }
        if (redirects >= 5) {
          reject(new Error(`download redirect limit exceeded for ${url}`));
          return;
        }
        let nextURL;
        try {
          nextURL = resolveRedirectURL(url, response.headers.location);
        } catch (error) {
          reject(error);
          return;
        }
        download(nextURL, encoding, redirects + 1).then(resolve, reject);
        return;
      }
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

function resolveRedirectURL(currentURL, location) {
  const next = new URL(location, currentURL);
  if (next.protocol !== "https:") {
    throw new Error(`refusing non-https redirect to ${next.protocol}`);
  }
  return next.toString();
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

function usage() {
  return "usage: patchxnote-agent <install|update|uninstall|login|mcp> [options]";
}

function mcpUsage() {
  return "usage: patchxnote-agent mcp <serve|config> [options]";
}

if (require.main === module) {
  main(process.argv.slice(2)).catch(error => {
    console.error(error.message);
    process.exitCode = 1;
  });
}

module.exports = {
  createInstallPlan,
  defaultInstallDir,
  inspectInstalledBinary,
  parseLauncherOptions,
  parseArgs,
  resolveTarget,
  joinInstallPath,
  isInstallDirOnPath,
  pathHint,
  resolveRedirectURL,
  universalMCPConfig,
  verifyChecksum
};
