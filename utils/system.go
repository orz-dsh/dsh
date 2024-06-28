package utils

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

var systemInfo *SystemInfo

type SystemInfo struct {
	Os         string
	Arch       string
	Hostname   string
	Username   string
	HomeDir    string
	WorkingDir string
}

func GetSystemInfo() (*SystemInfo, error) {
	if systemInfo != nil {
		return systemInfo, nil
	}

	hostname, err := GetSystemHostname()
	if err != nil {
		return nil, err
	}
	username, err := GetSystemUsername()
	if err != nil {
		return nil, err
	}
	homedir, err := GetSystemHomeDir()
	if err != nil {
		return nil, err
	}
	workingDir, err := GetSystemWorkingDir()
	if err != nil {
		return nil, err
	}

	systemInfo = &SystemInfo{
		Os:         GetSystemOs(),
		Arch:       GetSystemArch(),
		Hostname:   hostname,
		Username:   username,
		HomeDir:    homedir,
		WorkingDir: workingDir,
	}
	return systemInfo, nil
}

func GetSystemOs() string {
	return strings.ToLower(runtime.GOOS)
}

func GetSystemArch() string {
	arch := strings.ToLower(runtime.GOARCH)
	if arch == "amd64" {
		arch = "x64"
	} else if arch == "386" {
		arch = "x32"
	}
	return arch
}

func GetSystemHostname() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", errW(err, "get system hostname error")
	}
	return hostname, nil
}

func GetSystemUsername() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", errW(err, "get system username error",
			reason("get current user error"),
		)
	}
	username := currentUser.Username
	if strings.Contains(username, "\\") {
		username = strings.Split(username, "\\")[1]
	}
	return username, nil
}

func GetSystemHomeDir() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", errW(err, "get system home dir error")
	}
	path, err := filepath.Abs(dir)
	if err != nil {
		return "", errW(err, "get system home dir error",
			reason("get abs path error"),
			kv("dir", dir),
		)
	}
	return path, nil
}

func GetSystemWorkingDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", errW(err, "get system working dir error")
	}
	path, err := filepath.Abs(dir)
	if err != nil {
		return "", errW(err, "get system working dir error",
			reason("get abs path error"),
			kv("dir", dir),
		)
	}
	return path, nil
}
