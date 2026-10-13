package tray

import (
	"os"
	"path/filepath"
	"strings"
)

// LaunchAgentLabel is the user-scope login-item label.
const LaunchAgentLabel = "com.kopio.kopi-o"

// LaunchAgentPath returns ~/Library/LaunchAgents/<label>.plist.
func LaunchAgentPath() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		home = os.Getenv("HOME")
	}
	return filepath.Join(home, "Library", "LaunchAgents", LaunchAgentLabel+".plist")
}

// IsLaunchAtLogin reports whether the login-item plist exists.
func IsLaunchAtLogin() bool {
	_, err := os.Stat(LaunchAgentPath())
	return err == nil
}

// launchPlist renders the plist XML for execPath.
func launchPlist(execPath string) string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>` + LaunchAgentLabel + `</string>
	<key>ProgramArguments</key>
	<array>
		<string>` + execPath + `</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
</dict>
</plist>
`
}

// SetLaunchAtLogin installs (on=true) or removes (on=false) the user-scope
// LaunchAgent plist. Removing a missing file is a no-op.
func SetLaunchAtLogin(on bool, execPath string) error {
	path := LaunchAgentPath()
	if !on {
		err := os.Remove(path)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	if execPath == "" {
		execPath = defaultExecPath()
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(launchPlist(execPath)), 0o644)
}

// defaultExecPath prefers the .app bundle binary, else the current executable.
func defaultExecPath() string {
	if exe, err := os.Executable(); err == nil && exe != "" {
		if i := strings.Index(exe, ".app/Contents/"); i >= 0 {
			return exe[:i] + ".app/Contents/MacOS/kopi-o"
		}
		if strings.HasSuffix(exe, "/kopi-o") {
			return exe
		}
		return exe
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Applications", "Kopi-O.app", "Contents", "MacOS", "kopi-o")
}
