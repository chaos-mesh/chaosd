package utils

import (
	"github.com/pingcap/log"
	"go.uber.org/zap"
	"os/exec"
	"time"
)

func ExecWithDeadline(t <-chan time.Time, cmd *exec.Cmd) error {
	done := make(chan error, 1)
	var output []byte
	var err error
	go func() {
		output, err = cmd.CombinedOutput()
		done <- err
	}()

	select {
	case <-t:
		if err := cmd.Process.Kill(); err != nil {
			log.Error("failed to kill process: ", zap.Error(err))
			return err
		}
	case err := <-done:
		if err != nil {
			log.Error(err.Error()+string(output), zap.Error(err))
			return err
		}
		log.Info(string(output))
	}
	return nil
}
