package asyncexec

import (
	"fmt"
	"testing"
	"time"
)

func TestNormalCase(t *testing.T) {
	// this command will run infinitely
	// so use a timeout (with channel) in this case
	cmd := &AsyncExec{
		Name: "ping",
		Args: []string{"0.0.0.0"},
	}

	complete := make(chan struct{})

	cmd.EndAction = func() error {
		fmt.Println("ending command...")
		complete <- struct{}{}
		return nil
	}

	cmd.Start()

	// block here until complete or timeout
	select {
	case <-complete:
		fmt.Println("command completed")
	case <-time.After(10 * time.Second):
		fmt.Println("time out")
	}
}

func TestStartWithTimeout(t *testing.T) {
	cmd := &AsyncExec{
		Name: "ping",
		Args: []string{"0.0.0.0"},
	}

	// blocking version
	cmd.StartWithTimeout(10 * time.Second)
}
