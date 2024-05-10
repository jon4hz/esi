package manager

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/jon4hz/esi/shell"
)

func (m *Manager) addEnv(key, value string) {
	m.env = append(m.env, fmt.Sprintf("%s=%s", key, value))
}

func (m *Manager) executeSingleCommandWithEnvs(args []string) error {
	command := args[0]
	argsForCommand := args[1:]

	// TODO: debug log

	cmd := exec.Command(command, argsForCommand...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = m.env

	_, err := m.execCmd(cmd)
	return err
}

// ExecCommand executes the SSH command inside a subshell. If no shell could be determined,
// this function will try to execute ssh directly.
func (m *Manager) execSubshell(args []string, env []string) (int, error) {
	shell := m.subShellCmd()
	var cmd *exec.Cmd

	// make sure the shell exists
	_, err := exec.LookPath(shell[0])
	if err != nil {
		// if we can't find the shell, just execute the command directly
		log.Warn("Shell not found in PATH. Executing \"ssh\" directly.")
		cmd = exec.Command(args[0], args[1:]...) // #nosec G204
	} else {
		subCmd := m.buildExecCmd(args)
		args = []string{
			shell[1],
			subCmd,
		}
		cmd = exec.Command(shell[0], args...) // #nosec G204
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = env

	return m.execCmd(cmd)
}

// buildExecCmd combines the given parts into a single command string.
// If the parts contain quotes or backslashes, they will be escaped.
func (m *Manager) buildExecCmd(parts []string) string {
	if len(parts) == 0 {
		return ""
	}

	var s strings.Builder
	for i, part := range parts {
		if i > 0 {
			s.WriteString(" ")
		}
		s.WriteString(shell.EscapeChars(part))
	}

	return s.String()
}

// subShellCmd returns the shell command which will be used to execute the actuall command.
func (m *Manager) subShellCmd() [2]string {
	// default to sh -c
	shell := [...]string{"sh", "-c"}

	currentShell := os.Getenv("SHELL")
	if currentShell != "" {
		shell[0] = currentShell
		log.Debug("Detected shell based on env:", "shell", shell)
	} else if runtime.GOOS == "windows" {
		// if the SHELL env var is not set and we're on Windows, use cmd.exe
		// The SHELL var should always be checked first, in case the user executes
		// cyberark-ssh-utils from something like Git Bash.
		shell = [...]string{"cmd", "/C"}
		log.Debug("Falling back to CMD on windows:", "shell", shell)
	} else {
		log.Debug("No shell detected. Using \"sh\"")
	}

	return shell
}

// execCmd executes the command and waits for its termination.
// Credit: infisical
func (m *Manager) execCmd(cmd *exec.Cmd) (int, error) {
	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel)
	if err := cmd.Start(); err != nil {
		return 1, err
	}
	go func() {
		for {
			sig := <-sigChannel
			switch sig {
			case os.Interrupt, syscall.SIGINT, syscall.SIGTERM:
				m.cleanup()
			}
			_ = cmd.Process.Signal(sig) // process all sigs
		}
	}()
	if err := cmd.Wait(); err != nil {
		_ = cmd.Process.Signal(os.Kill)

		waitStatus := cmd.ProcessState.Sys().(syscall.WaitStatus)
		return waitStatus.ExitStatus(), fmt.Errorf("failed to wait for command termination: %v", err)
	}
	waitStatus := cmd.ProcessState.Sys().(syscall.WaitStatus)
	return waitStatus.ExitStatus(), nil
}
