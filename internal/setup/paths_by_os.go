package setup

import "github.com/ZsTs119/patchxnote-agent/internal/config"

func windowsConfigTarget(clientID string, env config.PathEnv) (ConfigTarget, error) {
	appData, err := windowsAppData(env)
	if err != nil {
		return ConfigTarget{}, err
	}
	home, err := ensureHome(env, "windows")
	if err != nil {
		return ConfigTarget{}, err
	}

	switch clientID {
	case "vscode":
		root := joinForOS("windows", appData, "Code", "User")
		return ConfigTarget{Path: joinForOS("windows", root, "mcp.json"), ExpectedRoot: root}, nil
	case "cursor":
		root := joinForOS("windows", home, ".cursor")
		return ConfigTarget{Path: joinForOS("windows", root, "mcp.json"), ExpectedRoot: root}, nil
	case "codex":
		root := joinForOS("windows", home, ".codex")
		return ConfigTarget{Path: joinForOS("windows", root, "config.toml"), ExpectedRoot: root}, nil
	case "claude-desktop":
		root := joinForOS("windows", appData, "Claude")
		return ConfigTarget{Path: joinForOS("windows", root, "claude_desktop_config.json"), ExpectedRoot: root}, nil
	case "windsurf":
		root := joinForOS("windows", home, ".codeium", "windsurf")
		return ConfigTarget{Path: joinForOS("windows", root, "mcp_config.json"), ExpectedRoot: root}, nil
	default:
		return ConfigTarget{}, nil
	}
}

func darwinConfigTarget(clientID string, env config.PathEnv) (ConfigTarget, error) {
	home, err := ensureHome(env, "darwin")
	if err != nil {
		return ConfigTarget{}, err
	}

	switch clientID {
	case "vscode":
		root := joinForOS("darwin", home, "Library", "Application Support", "Code", "User")
		return ConfigTarget{Path: joinForOS("darwin", root, "mcp.json"), ExpectedRoot: root}, nil
	case "cursor":
		root := joinForOS("darwin", home, ".cursor")
		return ConfigTarget{Path: joinForOS("darwin", root, "mcp.json"), ExpectedRoot: root}, nil
	case "codex":
		root := joinForOS("darwin", home, ".codex")
		return ConfigTarget{Path: joinForOS("darwin", root, "config.toml"), ExpectedRoot: root}, nil
	case "claude-desktop":
		root := joinForOS("darwin", home, "Library", "Application Support", "Claude")
		return ConfigTarget{Path: joinForOS("darwin", root, "claude_desktop_config.json"), ExpectedRoot: root}, nil
	case "windsurf":
		root := joinForOS("darwin", home, ".codeium", "windsurf")
		return ConfigTarget{Path: joinForOS("darwin", root, "mcp_config.json"), ExpectedRoot: root}, nil
	default:
		return ConfigTarget{}, nil
	}
}

func linuxConfigTarget(clientID string, env config.PathEnv) (ConfigTarget, error) {
	home, err := ensureHome(env, "linux")
	if err != nil {
		return ConfigTarget{}, err
	}
	configHome, err := linuxConfigHome(env)
	if err != nil {
		return ConfigTarget{}, err
	}

	switch clientID {
	case "vscode":
		root := joinForOS("linux", configHome, "Code", "User")
		return ConfigTarget{Path: joinForOS("linux", root, "mcp.json"), ExpectedRoot: root}, nil
	case "cursor":
		root := joinForOS("linux", home, ".cursor")
		return ConfigTarget{Path: joinForOS("linux", root, "mcp.json"), ExpectedRoot: root}, nil
	case "codex":
		root := joinForOS("linux", home, ".codex")
		return ConfigTarget{Path: joinForOS("linux", root, "config.toml"), ExpectedRoot: root}, nil
	case "claude-desktop":
		root := joinForOS("linux", configHome, "Claude")
		return ConfigTarget{Path: joinForOS("linux", root, "claude_desktop_config.json"), ExpectedRoot: root}, nil
	case "windsurf":
		root := joinForOS("linux", home, ".codeium", "windsurf")
		return ConfigTarget{Path: joinForOS("linux", root, "mcp_config.json"), ExpectedRoot: root}, nil
	default:
		return ConfigTarget{}, nil
	}
}
