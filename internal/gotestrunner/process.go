package gotestrunner

import (
	"os/exec"
	"syscall"
	"time"
)

// GracefulShutdownDelay is the time a child process has to exit after
// receiving SIGTERM before it is forcibly killed.
const GracefulShutdownDelay = 3 * time.Second

// SetProcessGroup configures cmd to run in its own process group and
// to receive SIGTERM (then SIGKILL after GracefulShutdownDelay) when
// its associated context is cancelled.
func SetProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
	}
	cmd.WaitDelay = GracefulShutdownDelay
}
