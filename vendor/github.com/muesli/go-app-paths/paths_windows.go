// +build windows

package gap

import (
	"path/filepath"
	"sync"
	"syscall"
	"unsafe"
)

// These are KNOWNFOLDERID constants that are passed to GetKnownFolderPath
var (
	initOnce  sync.Once
	initError error

	getKnownFolderPath uintptr
	coTaskMemFree      uintptr

	rfidLocalAppData = syscall.GUID{
		0xf1b32785,
		0x6fba,
		0x4fcf,
		[8]byte{0x9d, 0x55, 0x7b, 0x8e, 0x7f, 0x15, 0x70, 0x91},
	}
	rfidRoamingAppData = syscall.GUID{
		0x3eb685db,
		0x65f9,
		0x4cf6,
		[8]byte{0xa0, 0x3a, 0xe3, 0xef, 0x65, 0x72, 0x9f, 0x3d},
	}
	rfidProgramData = syscall.GUID{
		0x62ab5d82,
		0xfdc1,
		0x4dc3,
		[8]byte{0xa9, 0xdd, 0x07, 0x0d, 0x1d, 0x49, 0x5d, 0x97},
	}
)

// appendPaths appends the app-name and further variadic parts to a path
func (s *Scope) appendPaths(path string, parts ...string) string {
	paths := []string{path}
	paths = append(paths, parts...)
	return filepath.Join(paths...)
}

// dataDir returns the full path to the data directory.
func (s *Scope) dataDir() (string, error) {
	var rfid syscall.GUID

	switch s.Type {
	case System:
		rfid = rfidProgramData

	case User:
		rfid = rfidLocalAppData

	case CustomHome:
		return s.CustomHome, nil

	default:
		return "", ErrInvalidScope
	}

	path, err := getFolderPath(rfid)
	if err != nil {
		return "", ErrRetrievingPath
	}

	if path, err = filepath.Abs(path); err != nil {
		return "", ErrRetrievingPath
	}

	return filepath.Join(path, s.Vendor, s.App), nil
}

// dataDirs returns a priority-sorted slice of data dirs.
func (s *Scope) dataDirs() ([]string, error) {
	p, err := s.dataDir()
	if err != nil {
		return nil, err
	}

	return []string{p}, nil
}

// configDir returns the full path to the config dir.
func (s *Scope) configDir() (string, error) {
	p, err := s.dataDir()
	if err != nil {
		return p, err
	}

	return filepath.Join(p, "Config"), nil
}

// configDirs returns a priority-sorted slice of config dirs.
func (s *Scope) configDirs() ([]string, error) {
	p, err := s.configDir()
	if err != nil {
		return nil, err
	}

	return []string{p}, nil
}

// cacheDir returns the full path to the cache directory.
func (s *Scope) cacheDir() (string, error) {
	p, err := s.dataDir()
	if err != nil {
		return p, err
	}

	return filepath.Join(p, "Cache"), nil
}

// logDir returns the full path to the log dir.
func (s *Scope) logDir() (string, error) {
	p, err := s.dataDir()
	if err != nil {
		return p, err
	}

	return filepath.Join(p, "Logs"), nil
}

func getFolderPath(rfid syscall.GUID) (string, error) {
	initOnce.Do(initDLL)
	if initError != nil {
		return "", initError
	}

	var res uintptr
	ret, _, callErr := syscall.Syscall6(
		uintptr(getKnownFolderPath),
		4,
		uintptr(unsafe.Pointer(&rfid)),
		0,
		0,
		uintptr(unsafe.Pointer(&res)),
		0,
		0,
	)
	if callErr != 0 && ret != 0 {
		return "", callErr
	}

	defer syscall.Syscall(uintptr(coTaskMemFree), 1, res, 0, 0)
	return ucs2PtrToString(res), nil
}

func ucs2PtrToString(p uintptr) string {
	ptr := (*[4096]uint16)(unsafe.Pointer(p))
	return syscall.UTF16ToString((*ptr)[:])
}

func initDLL() {
	shell32, err := syscall.LoadLibrary("shell32.dll")
	if err != nil {
		initError = err
		return
	}
	getKnownFolderPath, err = syscall.GetProcAddress(shell32, "SHGetKnownFolderPath")
	if err != nil {
		initError = err
		return
	}

	ole32, err := syscall.LoadLibrary("Ole32.dll")
	if err != nil {
		initError = err
		return
	}
	coTaskMemFree, err = syscall.GetProcAddress(ole32, "CoTaskMemFree")
	if err != nil {
		initError = err
		return
	}
}
