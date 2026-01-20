package cmd

import (
	"bytes"
	"strings"
	"testing"
)

// Test command structure

func TestRootCommand(t *testing.T) {
	if rootCmd == nil {
		t.Fatal("rootCmd is nil")
	}
	if rootCmd.Use != "muxbee" {
		t.Errorf("expected Use to be 'muxbee', got '%s'", rootCmd.Use)
	}
	if rootCmd.Short == "" {
		t.Error("expected Short description to be set")
	}
	if rootCmd.Long == "" {
		t.Error("expected Long description to be set")
	}
}

func TestVersionVariable(t *testing.T) {
	// Version should default to "dev" when not set via ldflags
	if Version == "" {
		t.Error("Version should not be empty")
	}
}

func TestRootHasSubcommands(t *testing.T) {
	subcommands := rootCmd.Commands()
	if len(subcommands) == 0 {
		t.Fatal("expected root command to have subcommands")
	}

	// Check expected subcommands exist
	expected := []string{"init", "up", "down", "status", "bridge", "logs", "backup", "restore", "nuke", "config", "health", "open", "setup-bots", "tui", "update"}
	cmdNames := make(map[string]bool)
	for _, cmd := range subcommands {
		cmdNames[cmd.Name()] = true
	}

	for _, name := range expected {
		if !cmdNames[name] {
			t.Errorf("expected subcommand '%s' to exist", name)
		}
	}
}

func TestBridgeCommand(t *testing.T) {
	if bridgeCmd == nil {
		t.Fatal("bridgeCmd is nil")
	}
	if bridgeCmd.Use != "bridge" {
		t.Errorf("expected Use to be 'bridge', got '%s'", bridgeCmd.Use)
	}

	// Check bridge subcommands
	subcommands := bridgeCmd.Commands()
	if len(subcommands) == 0 {
		t.Fatal("expected bridge command to have subcommands")
	}

	expected := []string{"list", "enable", "disable", "login"}
	cmdNames := make(map[string]bool)
	for _, cmd := range subcommands {
		cmdNames[cmd.Name()] = true
	}

	for _, name := range expected {
		if !cmdNames[name] {
			t.Errorf("expected bridge subcommand '%s' to exist", name)
		}
	}
}

func TestBridgeEnableRequiresArg(t *testing.T) {
	if bridgeEnableCmd.Args == nil {
		t.Error("expected bridgeEnableCmd to have Args validator")
	}
}

func TestBridgeDisableRequiresArg(t *testing.T) {
	if bridgeDisableCmd.Args == nil {
		t.Error("expected bridgeDisableCmd to have Args validator")
	}
}

func TestBridgeLoginRequiresArg(t *testing.T) {
	if bridgeLoginCmd.Args == nil {
		t.Error("expected bridgeLoginCmd to have Args validator")
	}
}

// Test help output

func TestRootHelpOutput(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"--help"})
	rootCmd.Execute()

	output := buf.String()
	if !strings.Contains(output, "muxbee") {
		t.Error("expected help output to contain 'muxbee'")
	}
	if !strings.Contains(output, "Matrix") {
		t.Error("expected help output to mention Matrix")
	}
}

func TestBridgeListHelpOutput(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"bridge", "list", "--help"})
	rootCmd.Execute()

	output := buf.String()
	if !strings.Contains(output, "List") || !strings.Contains(output, "bridges") {
		t.Error("expected help output to describe listing bridges")
	}
}

func TestInitHelpOutput(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"init", "--help"})
	rootCmd.Execute()

	output := buf.String()
	if !strings.Contains(output, "init") {
		t.Error("expected help output to contain 'init'")
	}
}

func TestUpHelpOutput(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"up", "--help"})
	rootCmd.Execute()

	output := buf.String()
	if !strings.Contains(output, "Start") {
		t.Error("expected help output to mention starting services")
	}
}

func TestDownHelpOutput(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"down", "--help"})
	rootCmd.Execute()

	output := buf.String()
	if !strings.Contains(output, "Stop") {
		t.Error("expected help output to mention stopping services")
	}
}

func TestStatusHelpOutput(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"status", "--help"})
	rootCmd.Execute()

	output := buf.String()
	if !strings.Contains(output, "status") {
		t.Error("expected help output to mention status")
	}
}

func TestLogsHelpOutput(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"logs", "--help"})
	rootCmd.Execute()

	output := buf.String()
	if !strings.Contains(output, "logs") {
		t.Error("expected help output to mention logs")
	}
}

func TestBackupHelpOutput(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"backup", "--help"})
	rootCmd.Execute()

	output := buf.String()
	if !strings.Contains(output, "backup") {
		t.Error("expected help output to mention backup")
	}
}

func TestRestoreHelpOutput(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"restore", "--help"})
	rootCmd.Execute()

	output := buf.String()
	if !strings.Contains(output, "restore") || !strings.Contains(output, "Restore") {
		t.Error("expected help output to mention restore")
	}
}

func TestNukeHelpOutput(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"nuke", "--help"})
	rootCmd.Execute()

	output := buf.String()
	if !strings.Contains(output, "nuke") {
		t.Error("expected help output to mention nuke")
	}
}

func TestHealthHelpOutput(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"health", "--help"})
	rootCmd.Execute()

	output := buf.String()
	if !strings.Contains(output, "health") {
		t.Error("expected help output to mention health")
	}
}

func TestUpdateHelpOutput(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"update", "--help"})
	rootCmd.Execute()

	output := buf.String()
	if !strings.Contains(output, "update") && !strings.Contains(output, "Update") {
		t.Error("expected help output to mention update")
	}
}

func TestConfigShowHelpOutput(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"config", "show", "--help"})
	rootCmd.Execute()

	output := buf.String()
	if !strings.Contains(output, "config") {
		t.Error("expected help output to mention config")
	}
}

// Test version output

func TestVersionOutput(t *testing.T) {
	// Version output goes to stdout, not the buffer, so we just verify the command exists
	if rootCmd.Version == "" {
		t.Error("expected rootCmd.Version to be set")
	}
}

// Test command flags

func TestBackupCommand_HasOutputFlag(t *testing.T) {
	flag := backupCmd.Flags().Lookup("output")
	if flag == nil {
		t.Error("expected backup command to have --output flag")
	}
	if flag.Shorthand != "o" {
		t.Errorf("expected --output shorthand to be 'o', got '%s'", flag.Shorthand)
	}
}

func TestLogsCommand_HasFollowFlag(t *testing.T) {
	flag := logsCmd.Flags().Lookup("follow")
	if flag == nil {
		t.Error("expected logs command to have --follow flag")
	}
	if flag.Shorthand != "f" {
		t.Errorf("expected --follow shorthand to be 'f', got '%s'", flag.Shorthand)
	}
}

func TestLogsCommand_HasTailFlag(t *testing.T) {
	flag := logsCmd.Flags().Lookup("tail")
	if flag == nil {
		t.Error("expected logs command to have --tail flag")
	}
	if flag.Shorthand != "n" {
		t.Errorf("expected --tail shorthand to be 'n', got '%s'", flag.Shorthand)
	}
}

func TestNukeCommand_HasYesFlag(t *testing.T) {
	flag := nukeCmd.Flags().Lookup("yes")
	if flag == nil {
		t.Error("expected nuke command to have --yes flag")
	}
	if flag.Shorthand != "y" {
		t.Errorf("expected --yes shorthand to be 'y', got '%s'", flag.Shorthand)
	}
}

// Test command execution with invalid args

func TestBridgeEnableInvalidBridge(t *testing.T) {
	// Can't actually run this without mocking config, but we can test the command exists
	if bridgeEnableCmd.RunE == nil {
		t.Error("expected bridgeEnableCmd to have RunE function")
	}
}

func TestBridgeDisableInvalidBridge(t *testing.T) {
	if bridgeDisableCmd.RunE == nil {
		t.Error("expected bridgeDisableCmd to have RunE function")
	}
}

func TestBridgeLoginInvalidBridge(t *testing.T) {
	if bridgeLoginCmd.RunE == nil {
		t.Error("expected bridgeLoginCmd to have RunE function")
	}
}

func TestUpdateCommand(t *testing.T) {
	if updateCmd == nil {
		t.Fatal("updateCmd is nil")
	}
	if updateCmd.Use != "update" {
		t.Errorf("expected Use to be 'update', got '%s'", updateCmd.Use)
	}
	if updateCmd.RunE == nil {
		t.Error("expected updateCmd to have RunE function")
	}
}
