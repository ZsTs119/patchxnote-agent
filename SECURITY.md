# Security Policy

PatchNote Agent is a local CLI and MCP bridge for read-only PatchNote account access. Please report security issues privately so we can validate and fix them before public disclosure.

## Supported Versions

| Version | Security support |
| --- | --- |
| `0.1.x` | Beta security fixes while Agent V1 is active. |

## Reporting A Vulnerability

Use GitHub's private vulnerability reporting flow when it is available for this repository:

https://github.com/ZsTs119/patchnote-agent/security/advisories/new

If private reporting is not available, contact the maintainer through the GitHub repository without posting exploit details, credentials, OTPs, personal data, raw recordings, transcripts, or provider payloads in a public issue.

Helpful report details:

- affected `patchnote-agent` npm version and `patchnote version` output
- operating system and CPU architecture
- install command and MCP client used
- concise impact description
- minimal reproduction steps with sanitized test data

## In Scope

- npm installer wrapper behavior
- released `patchnote` binaries and checksum verification
- local MCP server behavior
- Agent login, logout, refresh, and session storage
- read-only `/v1/agent/**` API client behavior
- accidental exposure of tokens, OTPs, SK, full MAC values, raw audio, transcripts, prompts, or provider payloads

## Out Of Scope

- denial-of-service or load testing against PatchNote servers without prior written permission
- social engineering, phishing, spam, or physical attacks
- reports requiring access to another user's account or device
- vulnerabilities in third-party MCP clients unless PatchNote Agent is the direct cause
- requests for raw audio, full transcripts, private model responses, payment flows, Admin APIs, or hardware write actions; these are intentionally not exposed by Agent V1

## Beta Security Notes

The first beta release defaults to the PatchNote test API and requires the explicit file credential fallback flag:

```sh
PATCHNOTE_AUTH_INSECURE_FILE_KEYCHAIN=true
```

This fallback is temporary. OS-native keychain adapters are planned before broad production promotion.
