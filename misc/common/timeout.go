package common

import (
	"errors"
	"time"
)

var ErrTimeout = errors.New("common: timeout")

func TimeoutRun(duration time.Duration, fn func() (bool, error)) error {
	timeout := time.NewTimer(duration)
	defer timeout.Stop()

	for {
		select {
		case <-timeout.C:
			return ErrTimeout
		default:

		}

		finish, err := fn()
		if err != nil {
			return err
		}
		if finish {
			return nil
		}
	}
}
