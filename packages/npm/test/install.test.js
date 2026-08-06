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
assert.doesNotThrow(() => verifyChecksum(binary, "patchnote_0.0.0_linux_amd64", `${checksum}  patchnote_0.0.0_linux_amd64\n`));
assert.throws(() => verifyChecksum(binary, "patchnote_0.0.0_linux_amd64", `bad  patchnote_0.0.0_linux_amd64\n`), /checksum mismatch/);

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
assert.match(dryRun.stdout, /patchnote_0.0.0_linux_amd64/);
assert.match(dryRun.stdout, /"args": \[\s+"mcp",\s+"serve"\s+\]/);
assert.doesNotMatch(dryRun.stdout, /access_token|refresh_token|otp|sk_|protocol_mac/i);

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
