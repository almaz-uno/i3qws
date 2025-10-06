package icons

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewResolver(t *testing.T) {
	resolver := NewResolver()
	if resolver == nil {
		t.Fatal("NewResolver returned nil")
	}
	if len(resolver.dirs) == 0 {
		t.Error("NewResolver should have at least one directory")
	}
}

func TestGetIcon(t *testing.T) {
	resolver := NewResolver()

	// Test empty class
	icon := resolver.GetIcon("")
	if icon != "" {
		t.Errorf("Expected empty icon for empty class, got %q", icon)
	}

	// Test caching
	class := "TestClass"
	icon1 := resolver.GetIcon(class)
	icon2 := resolver.GetIcon(class)
	if icon1 != icon2 {
		t.Errorf("Cache not working: got different results %q vs %q", icon1, icon2)
	}
}

func TestParseDesktopFile(t *testing.T) {
	// Create temporary desktop file
	tmpDir := t.TempDir()
	desktopFile := filepath.Join(tmpDir, "test.desktop")

	content := `[Desktop Entry]
Name=Test Application
Icon=test-icon
StartupWMClass=TestApp
Type=Application
Exec=/usr/bin/test
`
	err := os.WriteFile(desktopFile, []byte(content), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	resolver := &Resolver{
		cache: make(map[string]string),
		dirs:  []string{},
	}

	de := resolver.parseDesktopFile(desktopFile)
	if de == nil {
		t.Fatal("parseDesktopFile returned nil")
	}

	if de.Icon != "test-icon" {
		t.Errorf("Expected icon %q, got %q", "test-icon", de.Icon)
	}

	if de.StartupWMClass != "TestApp" {
		t.Errorf("Expected StartupWMClass %q, got %q", "TestApp", de.StartupWMClass)
	}

	if de.Name != "Test Application" {
		t.Errorf("Expected Name %q, got %q", "Test Application", de.Name)
	}
}

func TestParseDesktopFileWithComments(t *testing.T) {
	tmpDir := t.TempDir()
	desktopFile := filepath.Join(tmpDir, "test.desktop")

	content := `# Comment
[Desktop Entry]
# Another comment
Name=Test App
Icon=test-icon
# Comment in the middle
StartupWMClass=TestApp

[Desktop Action Window]
Name=New Window
`
	err := os.WriteFile(desktopFile, []byte(content), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	resolver := &Resolver{
		cache: make(map[string]string),
		dirs:  []string{},
	}

	de := resolver.parseDesktopFile(desktopFile)
	if de == nil {
		t.Fatal("parseDesktopFile returned nil")
	}

	if de.Icon != "test-icon" {
		t.Errorf("Expected icon %q, got %q", "test-icon", de.Icon)
	}
}
