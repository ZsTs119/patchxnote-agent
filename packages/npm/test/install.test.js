"use strict";

const assert = require("assert");
const crypto = require("crypto");
const path = require("path");
const { spawnSync } = require("child_process");
const { isInstallDirOnPath, joinInstallPath, pathHint, resolveRedirectURL, resolveTarget, verifyChecksum } = require("../bin/patchxnote-agent.js");

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

console.log("installer tests passed");
