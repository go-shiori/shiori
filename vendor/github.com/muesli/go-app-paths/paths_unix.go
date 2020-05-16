// +build !darwin,!windows

package gap

import (
	"os"
	"path/filepath"
	"strings"
)

var (
	defaultDataDirs   = []string{"/usr/local/share", "/usr/share"}
	defaultConfigDirs = []string{"/etc/xdg", "/etc"}
	defaultLogDir     = "/var/log"
	defaultCacheDir   = "/var/cache"
)

// appendPaths appends the app-name and further variadic parts to a path
func (s *Scope) appendPaths(path string, parts ...string) string {
	paths := []string{path, s.Vendor, s.App}
	paths = append(paths, parts...)
	return filepath.Join(paths...)
}

// dataDir returns the full path to the data directory.
func (s *Scope) dataDir() (string, error) {
	switch s.Type {
	case System:
		return defaultDataDirs[0], nil

	case User:
		path := os.Getenv("XDG_DATA_HOME")
		if path == "" {
			return expandUser("~/.local/share"), nil
		}
		return path, nil

	case CustomHome:
		return filepath.Join(s.CustomHome, ".local/share"), nil
	}

	return "", ErrInvalidScope
}

// dataDirs returns a priority-sorted slice of data dirs.
func (s *Scope) dataDirs() ([]string, error) {
	var sl []string

	switch s.Type {
	case CustomHome:
		path, err := s.dataDir()
		if err != nil {
			return sl, err
		}
		sl = append(sl, path)

	case User:
		path, err := s.dataDir()
		if err != nil {
			return sl, err
		}
		sl = append(sl, path)

		path = os.Getenv("XDG_DATA_DIRS")
		if path != "" {
			paths := strings.Split(path, string(os.PathListSeparator))

			for _, p := range paths {
				sl = append(sl, p)
			}
		}

		fallthrough

	case System:
		sl = append(sl, defaultDataDirs...)
	}

	return sl, nil
}

// configDir returns the full path to the config dir.
func (s *Scope) configDir() (string, error) {
	switch s.Type {
	case System:
		return defaultConfigDirs[0], nil

	case User:
		path := os.Getenv("XDG_CONFIG_HOME")
		if path == "" {
			return expandUser("~/.config"), nil
		}
		return path, nil

	case CustomHome:
		return filepath.Join(s.CustomHome, ".config"), nil
	}

	return "", ErrInvalidScope
}

// configDirs returns a priority-sorted slice of config dirs.
func (s *Scope) configDirs() ([]string, error) {
	var sl []string

	switch s.Type {
	case CustomHome:
		path, err := s.configDir()
		if err != nil {
			return sl, err
		}
		sl = append(sl, path)

	case User:
		path, err := s.configDir()
		if err != nil {
			return sl, err
		}
		sl = append(sl, path)

		path = os.Getenv("XDG_CONFIG_DIRS")
		if path != "" {
			paths := strings.Split(path, string(os.PathListSeparator))

			for _, p := range paths {
				sl = append(sl, p)
			}
		}

		fallthrough

	case System:
		sl = append(sl, defaultConfigDirs...)
	}

	return sl, nil
}

// cacheDir returns the full path to the cache directory.
func (s *Scope) cacheDir() (string, error) {
	switch s.Type {
	case System:
		return defaultCacheDir, nil

	case User:
		path := os.Getenv("XDG_CACHE_HOME")
		if path == "" {
			return expandUser("~/.cache"), nil
		}
		return path, nil

	case CustomHome:
		return filepath.Join(s.CustomHome, ".cache"), nil
	}

	return "", ErrInvalidScope
}

// logDir returns the full path to the log dir.
func (s *Scope) logDir() (string, error) {
	switch s.Type {
	case System:
		return defaultLogDir, nil

	case User:
		fallthrough

	case CustomHome:
		return s.dataDir()
	}

	return "", ErrInvalidScope
}
