"use strict";

const assert = require("assert");
const crypto = require("crypto");
const path = require("path");
const { spawnSync } = require("child_process");
const { joinInstallPath, resolveTarget, verifyChecksum } = require("../bin/patchnote-agent.js");

const bin = path.resolve(__dirname, "..", "bin", "patchnote-agent.js");

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
assert.strictEqual(joinInstallPath("/tmp/patchnote-agent-bin", "patchnote", "linux"), "/tmp/patchnote-agent-bin/patchnote");
assert.strictEqual(joinInstallPath("C:\\PatchNote", "patchnote.exe", "windows"), "C:\\PatchNote\\patchnote.exe");

const binary = Buffer.from("patchnote-binary-fixture");
const checksum = crypto.createHash("sha256").update(binary).digest("hex");
assert.doesNotThrow(() => verifyChecksum(binary, "patchnote_0.1.0_linux_amd64", `${checksum}  patchnote_0.1.0_linux_amd64\n`));
assert.throws(() => verifyChecksum(binary, "patchnote_0.1.0_linux_amd64", `bad  patchnote_0.1.0_linux_amd64\n`), /checksum mismatch/);

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
  "/tmp/patchnote-agent-bin"
], { encoding: "utf8" });

assert.strictEqual(dryRun.status, 0, dryRun.stderr);
assert.match(dryRun.stdout, /PatchNote Agent install dry run/);
assert.match(dryRun.stdout, /patchnote_0.1.0_linux_amd64/);
assert.match(dryRun.stdout, /"args": \[\s+"mcp",\s+"serve"\s+\]/);
assert.doesNotMatch(dryRun.stdout, /access_token|refresh_token|otp|sk_|protocol_mac/i);

for (const [platform, arch, asset] of [
  ["linux", "x64", "patchnote_0.1.0_linux_amd64"],
  ["darwin", "arm64", "patchnote_0.1.0_darwin_arm64"],
  ["win32", "x64", "patchnote_0.1.0_windows_amd64.exe"]
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
  "/tmp/patchnote-agent-bin"
], { encoding: "utf8" });
assert.strictEqual(uninstallDryRun.status, 0, uninstallDryRun.stderr);
assert.match(uninstallDryRun.stdout, /"action": "uninstall"/);
assert.match(uninstallDryRun.stdout, /\/tmp\/patchnote-agent-bin\/patchnote/);

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
