package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	tea "github.com/charmbracelet/bubbletea"
)
func (m *osModel) executeDistroboxAction(action string) tea.Cmd {
	name := m.selectedDistrobox
	if name == "" {
		return func() tea.Msg {
			return commandDoneMsg{output: "No distrobox selected.", err: fmt.Errorf("no distrobox selected"), showOutput: true}
		}
	}

	return func() tea.Msg {
		var output string
		var err error

		switch action {
		case "upgrade":
			output, err = runCommand("distrobox", "upgrade", name)
		case "enter":
			output, err = runCommand("distrobox", "enter", name)
		case "stop":
			output, err = runCommand("distrobox", "stop", name, "--yes")
		case "remove":
			output, err = runCommand("distrobox", "rm", name, "--force")
		}

		if err != nil {
			output = fmt.Sprintf("%s\nError: %v", output, err)
		}

		return commandDoneMsg{output: output, err: err, showOutput: true}
	}
}

// assembleIniPath returns the user's distrobox-assemble manifest
func assembleIniPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "distrobox.ini"
	}
	return filepath.Join(home, ".config", "distrobox", "distrobox.ini")
}

// assembleCreateBox adds a section for name to the user's distrobox.ini (if
// missing) and runs distrobox-assemble to create the box. homeDir may be
// empty for a shared-home box.
func assembleCreateBox(name, homeDir string) (string, error) {
	ini := assembleIniPath()
	if err := os.MkdirAll(filepath.Dir(ini), 0755); err != nil {
		return "", err
	}
	data, _ := os.ReadFile(ini)
	if !strings.Contains(string(data), "["+name+"]") {
		section := fmt.Sprintf("\n[%s]\nimage = ghcr.io/vincentvdk/apparatus-box:latest\npull = true\ninit_hooks = distrobox-host-exec butler sync %s\n", name, name)
		if homeDir != "" {
			section += "home = " + homeDir + "\n"
		}
		f, err := os.OpenFile(ini, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return "", err
		}
		if _, err := f.WriteString(section); err != nil {
			f.Close()
			return "", err
		}
		f.Close()
	}
	return runCommand("distrobox-assemble", "create", "--file", ini)
}

func (m *osModel) executeCreateDistrobox(homeMode string) tea.Cmd {
	name := m.createName
	if name == "" {
		return func() tea.Msg {
			return commandDoneMsg{output: "No name provided.", err: fmt.Errorf("no name provided"), showOutput: true}
		}
	}

	return func() tea.Msg {
		var output string
		var err error

		if homeMode == "custom" {
			customHome := m.createCustomHome
			if customHome == "" {
				customHome = name
			}
			homeDir := fmt.Sprintf("%s/distrobox-homes/%s", os.Getenv("HOME"), customHome)
			output, err = assembleCreateBox(name, homeDir)
		} else {
			output, err = assembleCreateBox(name, "")
		}

		if err != nil {
			output = fmt.Sprintf("%s\nError: %v", output, err)
		} else {
			output = fmt.Sprintf("Distrobox '%s' created successfully!\n\nEnter with: distrobox enter %s", name, name)
		}

		return commandDoneMsg{output: output, err: err, showOutput: true}
	}
}

func (m *osModel) executeUpgradeDistrobox(action string) tea.Cmd {
	name := m.selectedDistrobox
	if name == "" {
		return func() tea.Msg {
			return commandDoneMsg{output: "No distrobox selected.", err: fmt.Errorf("no distrobox selected"), showOutput: true}
		}
	}

	return func() tea.Msg {
		var output string
		var err error

		switch action {
		case "upgrade-packages":
			output, err = runCommand("distrobox", "upgrade", name)
			if err == nil {
				output = fmt.Sprintf("Package upgrade complete for '%s'.", name)
			}
		case "upgrade-image":
			homeDir, _ := getDistroboxHomeDir(name)

			_, _ = runCommand("distrobox", "stop", name, "--yes")
			_, _ = runCommand("distrobox", "rm", name, "--force")

			// Recreate from the assemble manifest (pull = true fetches the
			// latest image; the init hook re-syncs configs)
			output, err = assembleCreateBox(name, homeDir)

			if err == nil {
				output = fmt.Sprintf("Image upgrade complete for '%s'.\n\nEnter with: distrobox enter %s", name, name)
			}
		}

		if err != nil {
			output = fmt.Sprintf("%s\nError: %v", output, err)
		}

		return commandDoneMsg{output: output, err: err, showOutput: true}
	}
}

func (m *osModel) executeMonitorAction() tea.Cmd {
	return func() tea.Msg {
		_, err := exec.LookPath("hyprdynamicmonitors")
		if err != nil {
			return commandDoneMsg{
				output: fmt.Sprintf("hyprdynamicmonitors not found.\n\nInstall it first, then run:\n  hyprdynamicmonitors tui"),
				err:    fmt.Errorf("hyprdynamicmonitors not found in PATH"),
			}
		}

		cmd := exec.Command("hyprdynamicmonitors", "tui")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return commandDoneMsg{
				output: fmt.Sprintf("hyprdynamicmonitors exited with error:\n%v", err),
				err:    err,
			}
		}
		return commandDoneMsg{output: "", err: nil, showOutput: false}
	}
}



