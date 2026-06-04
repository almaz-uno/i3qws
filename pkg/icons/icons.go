package icons

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

type (
	// Resolver resolves window class to icon name
	Resolver struct {
		cache map[string]string
		mu    sync.RWMutex
		dirs  []string
	}

	desktopEntry struct {
		Icon           string
		StartupWMClass string
		Name           string
	}
)

// NewResolver creates a new icon resolver
func NewResolver() *Resolver {
	dirs := []string{
		"/usr/share/applications",
		"/usr/local/share/applications",
	}

	// Add user local directory
	if home := os.Getenv("HOME"); home != "" {
		dirs = append(dirs, filepath.Join(home, ".local/share/applications"))
	}

	// Add XDG_DATA_DIRS
	if xdgDataDirs := os.Getenv("XDG_DATA_DIRS"); xdgDataDirs != "" {
		for _, dir := range strings.Split(xdgDataDirs, ":") {
			if dir != "" {
				dirs = append(dirs, filepath.Join(dir, "applications"))
			}
		}
	}

	return &Resolver{
		cache: make(map[string]string),
		dirs:  dirs,
	}
}

// GetIcon returns icon name for given window class
func (r *Resolver) GetIcon(windowClass string) string {
	if windowClass == "" {
		return ""
	}

	// Check cache first
	r.mu.RLock()
	if icon, found := r.cache[windowClass]; found {
		r.mu.RUnlock()
		return icon
	}
	r.mu.RUnlock()

	// Search in desktop files
	icon := r.searchDesktopFiles(windowClass)

	// Cache the result (even if empty)
	r.mu.Lock()
	r.cache[windowClass] = icon
	r.mu.Unlock()

	return icon
}

func (r *Resolver) searchDesktopFiles(windowClass string) string {
	windowClassLower := strings.ToLower(windowClass)

	// Pass 1: Exact matches (StartupWMClass, Name, exact filename)
	for _, dir := range r.dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".desktop") {
				continue
			}

			desktopFile := filepath.Join(dir, entry.Name())
			if icon := r.matchDesktopFile(desktopFile, entry.Name(), windowClass, windowClassLower, true); icon != "" {
				return icon
			}
		}
	}

	// Pass 2: Fuzzy matches (substring in filename)
	for _, dir := range r.dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".desktop") {
				continue
			}

			desktopFile := filepath.Join(dir, entry.Name())
			if icon := r.matchDesktopFile(desktopFile, entry.Name(), windowClass, windowClassLower, false); icon != "" {
				return icon
			}
		}
	}

	// Fallback: use window class as icon name (in lowercase)
	logrus.WithField("windowClass", windowClass).Debug("Icon not found in desktop files, using window class as icon name")
	return windowClassLower
}

func (r *Resolver) matchDesktopFile(desktopFile, entryName, windowClass, windowClassLower string, exactOnly bool) string {
	de := r.parseDesktopFile(desktopFile)
	if de == nil {
		return ""
	}

	// Match by StartupWMClass (exact, case-insensitive)
	if de.StartupWMClass != "" && strings.EqualFold(de.StartupWMClass, windowClass) {
		logrus.WithFields(logrus.Fields{
			"windowClass": windowClass,
			"icon":        de.Icon,
			"desktopFile": desktopFile,
			"matchBy":     "StartupWMClass",
		}).Debug("Icon found")
		return de.Icon
	}

	// Match by Name (exact, case-insensitive)
	if de.Name != "" && strings.EqualFold(de.Name, windowClass) {
		logrus.WithFields(logrus.Fields{
			"windowClass": windowClass,
			"icon":        de.Icon,
			"desktopFile": desktopFile,
			"matchBy":     "Name",
		}).Debug("Icon found")
		return de.Icon
	}

	// Match by desktop filename (without .desktop)
	baseName := strings.TrimSuffix(entryName, ".desktop")
	
	// Exact match always allowed
	if strings.EqualFold(baseName, windowClass) {
		logrus.WithFields(logrus.Fields{
			"windowClass": windowClass,
			"icon":        de.Icon,
			"desktopFile": desktopFile,
			"matchBy":     "filename-exact",
		}).Debug("Icon found")
		return de.Icon
	}
	
	// Substring match only if not exactOnly
	if !exactOnly && strings.Contains(strings.ToLower(baseName), windowClassLower) {
		logrus.WithFields(logrus.Fields{
			"windowClass": windowClass,
			"icon":        de.Icon,
			"desktopFile": desktopFile,
			"matchBy":     "filename-substring",
		}).Debug("Icon found")
		return de.Icon
	}

	return ""
}

func (r *Resolver) parseDesktopFile(path string) *desktopEntry {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	de := &desktopEntry{}
	inDesktopEntry := false
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for [Desktop Entry] section
		if strings.HasPrefix(line, "[") {
			inDesktopEntry = line == "[Desktop Entry]"
			continue
		}

		if !inDesktopEntry {
			continue
		}

		// Parse key=value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "Icon":
			de.Icon = value
		case "StartupWMClass":
			de.StartupWMClass = value
		case "Name":
			de.Name = value
		}

		// Early exit if we have all we need
		if de.Icon != "" && de.StartupWMClass != "" && de.Name != "" {
			break
		}
	}

	if de.Icon != "" {
		return de
	}

	return nil
}
