"use strict";

const assert = require("assert");
const crypto = require("crypto");
const fs = require("fs");
const os = require("os");
const path = require("path");
const { spawnSync } = require("child_process");
const {
  createInstallPlan,
  inspectInstalledBinary,
  isInstallDirOnPath,
  joinInstallPath,
  parseArgs,
  parseLauncherOptions,
  pathHint,
  resolveRedirectURL,
  resolveTarget,
  universalMCPConfig,
  verifyChecksum
} = require("../bin/patchxnote-agent.js");

const bin = path.resolve(__dirname, "..", "bin", "patchxnote-agent.js");
const packageVersion = require("../package.json").version;

assert.deepStrictEqual(resolveTarget("linux", "x64"), {
  platform: "linux",
  arch: "amd64",
  ext: ""
});
assert.deepStrictEqual(resolveTarget("win32", "arm64"), {
  platform: "windows",
  arch: "arm64",
  ext: ".exe"
});
assert.throws(() => resolveTarget("sunos", "x64"), /unsupported platform/);
assert.strictEqual(joinInstallPath("/tmp/patchxnote-agent-bin", "patchxnote", "linux"), "/tmp/patchxnote-agent-bin/patchxnote");
assert.strictEqual(joinInstallPath("C:\\PatchXNote", "patchxnote.exe", "windows"), "C:\\PatchXNote\\patchxnote.exe");
assert.strictEqual(resolveRedirectURL(`https://github.com/ZsTs119/patchxnote-agent/releases/download/v${packageVersion}/checksums.txt`, "/download"), "https://github.com/download");
assert.throws(() => resolveRedirectURL("https://github.com/a", "http://example.invalid/b"), /non-https redirect/);
assert.strictEqual(pathHint("/tmp/patchxnote-agent-bin", "linux"), "export PATH=\"/tmp/patchxnote-agent-bin:$PATH\"");
assert.match(pathHint("C:\\PatchXNote", "windows"), /SetEnvironmentVariable/);
assert.strictEqual(isInstallDirOnPath("/tmp/patchxnote-agent-bin", "/usr/bin:/tmp/patchxnote-agent-bin", "linux"), true);
assert.strictEqual(isInstallDirOnPath("/tmp/patchxnote-agent-bin", "/usr/bin:/bin", "linux"), false);
assert.strictEqual(isInstallDirOnPath("C:\\PatchXNote", "C:\\Windows;C:\\PatchXNote", "windows"), true);
assert.deepStrictEqual(parseArgs(["mcp", "config"]), {
  command: "mcp",
  subcommand: "config",
  options: {},
  passthroughArgs: []
});
assert.deepStrictEqual(parseArgs(["mcp", "serve", "--install-dir", "/tmp/px", "--", "--profile", "work"]), {
  command: "mcp",
  subcommand: "serve",
  options: { installDir: "/tmp/px" },
  passthroughArgs: ["--profile", "work"]
});
assert.deepStrictEqual(parseArgs(["login", "--install-dir", "/tmp/px", "--profile", "work"]), {
  command: "login",
  options: { installDir: "/tmp/px" },
  passthroughArgs: ["--profile", "work"]
});
assert.deepStrictEqual(parseLauncherOptions(["--platform", "linux", "--arch", "x64", "--from-local", "/tmp/bin"]), {
  options: { platform: "linux", arch: "x64", fromLocal: "/tmp/bin" },
  passthroughArgs: []
});
assert.throws(() => parseArgs(["mcp"]), /usage: patchxnote-agent mcp/);
assert.throws(() => parseArgs(["mcp", "serve", "--print-config"]), /not valid/);
assert.throws(() => parseArgs(["login", "--print-config"]), /not valid/);
assert.deepStrictEqual(universalMCPConfig(), {
  mcpServers: {
    patchxnote: {
      command: "npx",
      args: ["-y", "patchxnote-agent@latest", "mcp", "serve"]
    }
  }
});

const plan = createInstallPlan("install", {
  platform: "linux",
  arch: "x64",
  installDir: "/tmp/patchxnote-agent-bin"
});
assert.strictEqual(plan.asset_name, `patchxnote_${packageVersion}_linux_amd64`);
assert.strictEqual(plan.install_path, "/tmp/patchxnote-agent-bin/patchxnote");

const binary = Buffer.from("patchxnote-binary-fixture");
const checksum = crypto.createHash("sha256").update(binary).digest("hex");
assert.doesNotThrow(() => verifyChecksum(binary, `patchxnote_${packageVersion}_linux_amd64`, `${checksum}  patchxnote_${packageVersion}_linux_amd64\n`));
assert.throws(() => verifyChecksum(binary, `patchxnote_${packageVersion}_linux_amd64`, `bad  patchxnote_${packageVersion}_linux_amd64\n`), /checksum mismatch/);

const dryRun = spawnSync(process.execPath, [
  bin,
  "install",
  "--dry-run",
  "--print-config",
  "--platform",
  "linux",
  "--arch",
  "x64",
  "--install-dir",
  "/tmp/patchxnote-agent-bin"
], { encoding: "utf8" });

assert.strictEqual(dryRun.status, 0, dryRun.stderr);
assert.match(dryRun.stdout, /PatchXNote Agent install dry run/);
assert.match(dryRun.stdout, new RegExp(`patchxnote_${packageVersion.replaceAll(".", "\\.")}_linux_amd64`));
assert.match(dryRun.stdout, /install_dir_on_path/);
assert.match(dryRun.stdout, /path_hint/);
assert.match(dryRun.stdout, /"args": \[\s+"mcp",\s+"serve"\s+\]/);
assert.doesNotMatch(dryRun.stdout, /access_token|refresh_token|otp|sk_|protocol_mac/i);
assert.doesNotMatch(dryRun.stdout, /PATCHXNOTE_AUTH_INSECURE_FILE_KEYCHAIN/);

for (const [platform, arch, asset] of [
  ["linux", "x64", `patchxnote_${packageVersion}_linux_amd64`],
  ["darwin", "arm64", `patchxnote_${packageVersion}_darwin_arm64`],
  ["win32", "x64", `patchxnote_${packageVersion}_windows_amd64.exe`]
]) {
  const result = spawnSync(process.execPath, [
    bin,
    "install",
    "--dry-run",
    "--platform",
    platform,
    "--arch",
    arch
  ], { encoding: "utf8" });
  assert.strictEqual(result.status, 0, result.stderr);
  assert.match(result.stdout, new RegExp(asset.replace(".", "\\.")));
}

const uninstallDryRun = spawnSync(process.execPath, [
  bin,
  "uninstall",
  "--dry-run",
  "--platform",
  "linux",
  "--arch",
  "x64",
  "--install-dir",
  "/tmp/patchxnote-agent-bin"
], { encoding: "utf8" });
assert.strictEqual(uninstallDryRun.status, 0, uninstallDryRun.stderr);
assert.match(uninstallDryRun.stdout, /"action": "uninstall"/);
assert.match(uninstallDryRun.stdout, /\/tmp\/patchxnote-agent-bin\/patchxnote/);

const bad = spawnSync(process.execPath, [
  bin,
  "install",
  "--dry-run",
  "--platform",
  "plan9",
  "--arch",
  "x64"
], { encoding: "utf8" });
assert.notStrictEqual(bad.status, 0);
assert.match(bad.stderr, /unsupported platform/);

const universalConfigResult = spawnSync(process.execPath, [
  bin,
  "mcp",
  "config"
], { encoding: "utf8" });
assert.strictEqual(universalConfigResult.status, 0, universalConfigResult.stderr);
assert.deepStrictEqual(JSON.parse(universalConfigResult.stdout), universalMCPConfig());
assert.strictEqual(universalConfigResult.stderr, "");
assert.doesNotMatch(universalConfigResult.stdout, /MCP config:|access_token|refresh_token|otp|sk_|protocol_mac/i);

const fixtureRoot = fs.mkdtempSync(path.join(os.tmpdir(), "patchxnote-agent-test-"));
const goodBinary = buildFakePatchXNote(fixtureRoot, packageVersion, "good");
const oldBinary = buildFakePatchXNote(fixtureRoot, "0.0.1", "old");
const badVersionBinary = buildFakePatchXNote(fixtureRoot, packageVersion, "badversion");

const serveDir = path.join(fixtureRoot, "PatchX Note Agent", "serve-bin");
const serveLog = path.join(fixtureRoot, "serve.log");
const serveResult = runWrapper([
  "mcp",
  "serve",
  "--from-local",
  goodBinary,
  "--install-dir",
  serveDir
], { input: initializeLine(), logPath: serveLog });
assert.strictEqual(serveResult.status, 0, serveResult.stderr);
assertSingleJSONRPCLine(serveResult.stdout, 1);
assert.doesNotMatch(serveResult.stdout, /Installed|PatchXNote Agent|binary missing|reinstalling/);
assert.match(serveResult.stderr, /binary missing; installing/);
assert.deepStrictEqual(readJSONLines(serveLog).at(-1), ["mcp", "serve"]);
assert.deepStrictEqual(inspectInstalledBinary(createInstallPlan("install", { installDir: serveDir })), {
  exists: true,
  ok: true,
  version: packageVersion
});

const offlineWarmResult = runWrapper([
  "mcp",
  "serve",
  "--install-dir",
  serveDir
], {
  input: initializeLine(),
  logPath: serveLog,
  env: { PATCHXNOTE_AGENT_RELEASE_BASE_URL: "https://example.invalid/patchxnote-agent" }
});
assert.strictEqual(offlineWarmResult.status, 0, offlineWarmResult.stderr);
assertSingleJSONRPCLine(offlineWarmResult.stdout, 1);

const eofResult = runWrapper([
  "mcp",
  "serve",
  "--install-dir",
  serveDir
], { input: "", logPath: serveLog });
assert.strictEqual(eofResult.status, 0, eofResult.stderr);
assert.strictEqual(eofResult.stdout, "");

const staleDir = path.join(fixtureRoot, "stale-bin");
const staleLog = path.join(fixtureRoot, "stale.log");
const staleInstall = runWrapper([
  "install",
  "--from-local",
  oldBinary,
  "--install-dir",
  staleDir
]);
assert.strictEqual(staleInstall.status, 0, staleInstall.stderr);
const staleRepair = runWrapper([
  "mcp",
  "serve",
  "--from-local",
  goodBinary,
  "--install-dir",
  staleDir
], { input: initializeLine(), logPath: staleLog });
assert.strictEqual(staleRepair.status, 0, staleRepair.stderr);
assert.match(staleRepair.stderr, /does not match package/);
assertSingleJSONRPCLine(staleRepair.stdout, 1);
assert.strictEqual(inspectInstalledBinary(createInstallPlan("install", { installDir: staleDir })).version, packageVersion);

const brokenDir = path.join(fixtureRoot, "broken-bin");
const brokenInstall = runWrapper([
  "install",
  "--from-local",
  badVersionBinary,
  "--install-dir",
  brokenDir
]);
assert.strictEqual(brokenInstall.status, 0, brokenInstall.stderr);
const brokenRepair = runWrapper([
  "mcp",
  "serve",
  "--from-local",
  goodBinary,
  "--install-dir",
  brokenDir
], { input: initializeLine() });
assert.strictEqual(brokenRepair.status, 0, brokenRepair.stderr);
assert.match(brokenRepair.stderr, /preflight failed/);
assertSingleJSONRPCLine(brokenRepair.stdout, 1);

const loginDir = path.join(fixtureRoot, "login-bin");
const loginLog = path.join(fixtureRoot, "login.log");
const loginResult = runWrapper([
  "login",
  "--from-local",
  goodBinary,
  "--install-dir",
  loginDir,
  "--profile",
  "work"
], { logPath: loginLog });
assert.strictEqual(loginResult.status, 0, loginResult.stderr);
assert.match(loginResult.stdout, /fake login/);
assert.deepStrictEqual(readJSONLines(loginLog).at(-1), ["login", "--profile", "work"]);
assert.doesNotMatch(loginResult.stdout + loginResult.stderr, /access_token|refresh_token|otp|sk_|protocol_mac/i);

console.log("installer tests passed");

function runWrapper(args, options = {}) {
  return spawnSync(process.execPath, [bin, ...args], {
    encoding: "utf8",
    input: options.input,
    env: {
      ...process.env,
      ...options.env,
      ...(options.logPath ? { PATCHXNOTE_FAKE_LOG: options.logPath } : {})
    }
  });
}

function initializeLine() {
  return '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}\n';
}

function assertSingleJSONRPCLine(stdout, id) {
  const lines = stdout.trim().split(/\r?\n/).filter(Boolean);
  assert.strictEqual(lines.length, 1, stdout);
  const response = JSON.parse(lines[0]);
  assert.strictEqual(response.jsonrpc, "2.0");
  assert.strictEqual(response.id, id);
  assert.ok(response.result, stdout);
}

function readJSONLines(filePath) {
  return fs.readFileSync(filePath, "utf8")
    .trim()
    .split(/\r?\n/)
    .filter(Boolean)
    .map(line => JSON.parse(line));
}

function buildFakePatchXNote(root, version, mode) {
  const sourceDir = path.join(root, `fake-${mode}`);
  fs.mkdirSync(sourceDir, { recursive: true });
  const sourcePath = path.join(sourceDir, "main.go");
  const outputPath = path.join(sourceDir, process.platform === "win32" ? "patchxnote.exe" : "patchxnote");
  fs.writeFileSync(sourcePath, fakeGoSource(version, mode));
  const build = spawnSync("go", ["build", "-o", outputPath, sourcePath], {
    cwd: sourceDir,
    encoding: "utf8",
    env: { ...process.env, GO111MODULE: "off" }
  });
  assert.strictEqual(build.status, 0, build.stderr || build.stdout);
  return outputPath;
}

function fakeGoSource(version, mode) {
  return `
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

const version = ${JSON.stringify(version)}
const mode = ${JSON.stringify(mode)}

func main() {
	args := os.Args[1:]
	logArgs(args)
	if len(args) >= 2 && args[0] == "version" && args[1] == "--output" {
		if mode == "badversion" {
			fmt.Fprintln(os.Stderr, "broken version")
			os.Exit(2)
		}
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"version": version})
		return
	}
	if len(args) >= 2 && args[0] == "mcp" && args[1] == "serve" {
		serveMCP()
		return
	}
	if len(args) >= 1 && args[0] == "login" {
		fmt.Println("fake login")
		return
	}
	fmt.Fprintf(os.Stderr, "unexpected args: %v\\n", args)
	os.Exit(2)
}

func serveMCP() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var request map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &request); err != nil {
			continue
		}
		id := request["id"]
		switch request["method"] {
		case "initialize":
			_ = json.NewEncoder(os.Stdout).Encode(map[string]any{
				"jsonrpc": "2.0",
				"id": id,
				"result": map[string]any{
					"protocolVersion": "2025-06-18",
					"serverInfo": map[string]any{"name": "fake-patchxnote", "version": version},
					"capabilities": map[string]any{"tools": map[string]any{}},
				},
			})
		case "tools/list":
			_ = json.NewEncoder(os.Stdout).Encode(map[string]any{
				"jsonrpc": "2.0",
				"id": id,
				"result": map[string]any{"tools": []any{}},
			})
		}
	}
}

func logArgs(args []string) {
	path := os.Getenv("PATCHXNOTE_FAKE_LOG")
	if path == "" {
		return
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return
	}
	defer file.Close()
	_ = json.NewEncoder(file).Encode(args)
}
`;
}
