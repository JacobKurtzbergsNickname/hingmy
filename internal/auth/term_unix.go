//go:build linux || darwin || freebsd || openbsd || netbsd

package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

// readPassword reads a password from the terminal without echoing characters.
// On Unix systems it manipulates terminal flags via ioctl.
//
// SEC: never pass passwords via flags or environment variables — they appear
// in shell history and process listings. Always use interactive prompts.
func readPassword(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)

	fd := int(os.Stdin.Fd())

	var oldState syscall.Termios
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(fd), ioctlReadTermios, uintptr(unsafe.Pointer(&oldState))); errno != 0 {
		// Not a terminal — fall back to plain scan.
		var pw string
		fmt.Scanln(&pw) //nolint:errcheck
		fmt.Fprintln(os.Stderr)
		return pw, nil
	}

	// Disable echo.
	newState := oldState
	newState.Lflag &^= syscall.ECHO
	syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), ioctlWriteTermios, uintptr(unsafe.Pointer(&newState))) //nolint:errcheck

	defer func() {
		// Restore echo even if reading fails.
		syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), ioctlWriteTermios, uintptr(unsafe.Pointer(&oldState))) //nolint:errcheck
		fmt.Fprintln(os.Stderr)
	}()

	reader := bufio.NewReader(os.Stdin)
	password, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("readPassword: %w", err)
	}
	return strings.TrimRight(password, "\r\n"), nil
}
