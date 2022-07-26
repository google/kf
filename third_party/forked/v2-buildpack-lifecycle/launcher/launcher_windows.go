package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"code.cloudfoundry.org/buildpackapplifecycle/launcher/profile"

	"golang.org/x/sys/windows"
)

var (
	kernel32       = windows.NewLazySystemDLL("kernel32.dll")
	createProcessW = kernel32.NewProc("CreateProcessW")
)

func runProcess(dir, command string) {
	err := createProcessW.Find()
	handleErr("couldn't find func address", err)

	args, err := syscall.UTF16PtrFromString(command)
	handleErr("casting command failed", err)
	cwd, err := syscall.UTF16PtrFromString(dir)
	handleErr("casting cwd failed", err)

	tmpDir, ok := os.LookupEnv("TMPDIR")
	if !ok {
		handleErr("TMPDIR must be set", errors.New("TMPDIR must be set"))
	}

	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		handleErr("creating TMPDIR", err)
	}

	getenvPath, err := getenvPath()
	handleErr("getting getenv path failed", err)
	envs, err := profile.ProfileEnv(dir, tmpDir, getenvPath, os.Stdout, os.Stderr)
	handleErr("getting environment failed", err)

	for _, v := range envs {
		subs := strings.SplitN(v, "=", 2)
		if strings.ToUpper(subs[0]) == "PATH" {
			if err := os.Setenv("PATH", subs[1]); err != nil {
				handleErr("setting environment failed", err)
			}
		}
	}

	p, _ := syscall.GetCurrentProcess()
	fd := make([]syscall.Handle, 3)
	for i, file := range []*os.File{os.Stdin, os.Stdout, os.Stderr} {
		err := syscall.DuplicateHandle(p, syscall.Handle(file.Fd()), p, &fd[i], 0, true, syscall.DUPLICATE_SAME_ACCESS)
		if err != nil {
			handleErr("DuplicateHandle failed", err)
		}
		defer syscall.CloseHandle(syscall.Handle(fd[i]))
	}
	si := new(syscall.StartupInfo)
	si.Cb = uint32(unsafe.Sizeof(*si))
	si.Flags = syscall.STARTF_USESTDHANDLES
	si.StdInput = fd[0]
	si.StdOutput = fd[1]
	si.StdErr = fd[2]
	pi := new(syscall.ProcessInformation)

	// Change the parent's working directory to the app dir so
	// CreateProcessW will search it when starting the child process
	err = os.Chdir(dir)
	handleErr("couldn't change working directory", err)

	creationFlags := syscall.CREATE_UNICODE_ENVIRONMENT
	// CreateProcessW docs
	// https://msdn.microsoft.com/en-us/library/windows/desktop/ms682425(v=vs.85).aspx
	// Process Creation flags
	// https://msdn.microsoft.com/en-us/library/windows/desktop/ms684863(v=vs.85).aspx
	r, _, e := syscall.Syscall12(createProcessW.Addr(), 10,
		uintptr(uint16(0)),                            // appname
		uintptr(unsafe.Pointer(args)),                 // executable and args
		uintptr(unsafe.Pointer(nil)),                  // process security attributes
		uintptr(unsafe.Pointer(nil)),                  // thread security attributes
		uintptr(uint32(1)),                            // inherit parent's handles
		uintptr(creationFlags),                        // creation flags
		uintptr(unsafe.Pointer(createEnvBlock(envs))), // use generated environment
		uintptr(unsafe.Pointer(cwd)),                  // process working directory
		uintptr(unsafe.Pointer(si)),                   // startup info
		uintptr(unsafe.Pointer(pi)),                   // process info for the created process
		0, 0)

	if r == 0 {
		handleErr(fmt.Sprintf("CreateProcessW failed %s:%s", dir, command), e)
	}
	defer syscall.CloseHandle(syscall.Handle(pi.Thread))
	defer syscall.CloseHandle(syscall.Handle(pi.Process))

	_, err = syscall.WaitForSingleObject(pi.Process, math.MaxUint32)
	handleErr("WaitForSingleObject failed", err)

	var exitCode uint32
	err = syscall.GetExitCodeProcess(pi.Process, &exitCode)
	handleErr("GetExitCodeProcess failed", err)

	os.Exit(int(exitCode))
}

func handleErr(description string, err error) {
	if err != nil {
		fmt.Printf("%s: %s", description, err.Error())
		os.Exit(1)
	}
}

func createEnvBlock(envv []string) *uint16 {
	if len(envv) == 0 {
		return &utf16.Encode([]rune("\x00\x00"))[0]
	}
	length := 0
	for _, s := range envv {
		length += len(s) + 1
	}
	length += 1

	b := make([]byte, length)
	i := 0
	for _, s := range envv {
		l := len(s)
		copy(b[i:i+l], []byte(s))
		copy(b[i+l:i+l+1], []byte{0})
		i = i + l + 1
	}
	copy(b[i:i+1], []byte{0})

	return &utf16.Encode([]rune(string(b)))[0]
}

func getenvPath() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}

	executableDir := filepath.Dir(executable)

	return filepath.Join(executableDir, "getenv.exe"), nil
}
