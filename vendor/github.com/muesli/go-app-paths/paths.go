package gap

import (
	"errors"
	"os"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
)

var (
	// ErrInvalidScope gets returned when an invalid scope type has been set.
	ErrInvalidScope = errors.New("Invalid scope type")
	// ErrRetrievingPath gets returned when the path could not be resolved.
	ErrRetrievingPath = errors.New("Could not retrieve path")
)

// ScopeType specifies whether returned paths are user-specific or system-wide.
type ScopeType int

const (
	// System is the system-wide scope.
	System ScopeType = iota
	// User is the user-specific scope.
	User
	// CustomHome uses a custom user home as scope.
	CustomHome
)

// Scope holds scope & app-specific information.
type Scope struct {
	Type       ScopeType
	CustomHome string
	Vendor     string
	App        string
}

// NewScope returns a new Scope that lets you query app- & platform-specific
// paths.
func NewScope(t ScopeType, app string) *Scope {
	return &Scope{
		Type: t,
		App:  app,
	}
}

// NewVendorScope returns a new Scope with vendor information that lets you
// query app- & platform-specific paths.
func NewVendorScope(t ScopeType, vendor, app string) *Scope {
	return &Scope{
		Type:   t,
		Vendor: vendor,
		App:    app,
	}
}

// NewCustomHomeScope returns a new Scope that lets you operate on a custom path
// prefix.
func NewCustomHomeScope(path, vendor, app string) *Scope {
	return &Scope{
		Type:       CustomHome,
		CustomHome: path,
		Vendor:     vendor,
		App:        app,
	}
}

// DataDirs returns a priority-sorted slice of all the application's data dirs.
func (s *Scope) DataDirs() ([]string, error) {
	ps, err := s.dataDirs()
	if err != nil {
		return nil, err
	}

	var sl []string
	for _, v := range ps {
		sl = append(sl, s.appendPaths(v))
	}
	return sl, nil
}

// ConfigDirs returns a priority-sorted slice of all of the application's config
// dirs.
func (s *Scope) ConfigDirs() ([]string, error) {
	ps, err := s.configDirs()
	if err != nil {
		return nil, err
	}

	var sl []string
	for _, v := range ps {
		sl = append(sl, s.appendPaths(v))
	}
	return sl, nil
}

// CacheDir returns the full path to the application's default cache dir.
func (s *Scope) CacheDir() (string, error) {
	p, err := s.cacheDir()
	if err != nil {
		return p, err
	}

	return s.appendPaths(p), nil
}

// LogPath returns the full path to the application's default log file.
func (s *Scope) LogPath(filename string) (string, error) {
	p, err := s.logDir()
	if err != nil {
		return p, err
	}

	return s.appendPaths(p, filename), nil
}

// DataPath returns the full path to a file in the application's default data
// directory.
func (s *Scope) DataPath(filename string) (string, error) {
	p, err := s.dataDir()
	if err != nil {
		return p, err
	}

	return s.appendPaths(p, filename), nil
}

// ConfigPath returns the full path to a file in the application's default
// config directory.
func (s *Scope) ConfigPath(filename string) (string, error) {
	p, err := s.configDir()
	if err != nil {
		return p, err
	}

	return s.appendPaths(p, filename), nil
}

// LookupConfig returns all existing configs with this filename.
func (s *Scope) LookupConfig(filename string) ([]string, error) {
	paths, err := s.configDirs()
	if err != nil {
		return nil, err
	}

	return s.findExisting(paths, filename), nil
}

// LookupDataFile returns all existing data files with this filename.
func (s *Scope) LookupDataFile(filename string) ([]string, error) {
	paths, err := s.dataDirs()
	if err != nil {
		return nil, err
	}

	return s.findExisting(paths, filename), nil
}

// expandUser is a helper function that expands the first '~' it finds in the
// passed path with the home directory of the current user.
func expandUser(path string) string {
	if u, err := homedir.Dir(); err == nil {
		return strings.Replace(path, "~", u, -1)
	}
	return path
}

// findExisting tries to find filename in all paths and returns a list of
// existing paths.
func (s *Scope) findExisting(paths []string, filename string) []string {
	var sl []string

	for _, p := range paths {
		f := s.appendPaths(p, filename)
		_, err := os.Stat(f)
		if err == nil || os.IsExist(err) {
			sl = append(sl, f)
		}
	}

	return sl
}
