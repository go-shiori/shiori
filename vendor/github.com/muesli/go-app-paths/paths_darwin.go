// +build darwin

package gap

import (
	"path/filepath"
)

var (
	defaultDataDir   = "/Library/Application Support"
	defaultConfigDir = "/Library/Preferences"
	defaultLogDir    = "/Library/Logs"
	defaultCacheDir  = "/Library/Caches"
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
		return defaultDataDir, nil

	case User:
		return expandUser("~" + defaultDataDir), nil

	case CustomHome:
		return filepath.Join(s.CustomHome, defaultDataDir), nil
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

		fallthrough

	case System:
		sl = append(sl, defaultDataDir)
	}

	return sl, nil
}

// configDir returns the full path to the config dir.
func (s *Scope) configDir() (string, error) {
	switch s.Type {
	case System:
		return defaultConfigDir, nil

	case User:
		return expandUser("~" + defaultConfigDir), nil

	case CustomHome:
		return filepath.Join(s.CustomHome, defaultConfigDir), nil
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

		fallthrough

	case System:
		sl = append(sl, defaultConfigDir)
	}

	return sl, nil
}

// cacheDir returns the full path to the cache directory.
func (s *Scope) cacheDir() (string, error) {
	switch s.Type {
	case System:
		return defaultCacheDir, nil

	case User:
		return expandUser("~" + defaultCacheDir), nil

	case CustomHome:
		return filepath.Join(s.CustomHome, defaultCacheDir), nil
	}

	return "", ErrInvalidScope
}

// logDir returns the full path to the log dir.
func (s *Scope) logDir() (string, error) {
	switch s.Type {
	case System:
		return defaultLogDir, nil

	case User:
		return expandUser("~" + defaultLogDir), nil

	case CustomHome:
		return filepath.Join(s.CustomHome, defaultLogDir), nil
	}

	return "", ErrInvalidScope
}
