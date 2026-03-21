//go:build !linux && !darwin && !freebsd && !openbsd && !netbsd

package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// readPassword on unsupported platforms reads without hiding input.
func readPassword(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	reader := bufio.NewReader(os.Stdin)
	pw, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("readPassword: %w", err)
	}
	return strings.TrimRight(pw, "\r\n"), nil
}
