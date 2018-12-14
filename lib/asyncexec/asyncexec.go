// Package asyncexec implements a type AsyncExec
// which could gracefully exec command and
// and print output line by line, rather than output
// them in a bunch after command completes
package asyncexec

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/fatih/color"
)

var (
	green  = color.New(color.FgGreen).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
)

// AsyncExec is executable itself
type AsyncExec struct {
	Name      string
	Dir       string
	Args      []string
	Env       map[string]string
	EndAction func() error
	Stdout    []string
	Stderr    []string

	// internal use
	complete chan struct{}
	end      func()
}

// SetEnv sets env
func (ae *AsyncExec) SetEnv(key, value string) {
	if ae.Env == nil {
		ae.Env = make(map[string]string)
	}
	ae.Env[key] = value
}

// Start starts execution
func (ae *AsyncExec) Start() error {
	cmd := exec.Command(ae.Name, ae.Args...)
	if ae.Dir != "" {
		cmd.Dir = ae.Dir
	}

	if ae.Env != nil {
		cmd.Env = append(cmd.Env, os.Environ()...)
		for k, v := range ae.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(stdout)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	scannerErr := bufio.NewScanner(stderr)

	go func() {
		defer ae.endWithFatal()
		defer ifPanic(func(err error) {
			_ = stdout.Close()
			_ = stderr.Close()
		})

		if err := cmd.Start(); err != nil {
			fmt.Println(red("error"), err.Error())
			return
		}

		go func() {
			defer ifPanic(nil)
			for scanner.Scan() {
				l := scanner.Text()
				ae.Stdout = append(ae.Stdout, l)
				fmt.Println(green("stdout"), l)
			}
		}()

		go func() {
			defer ifPanic(nil)
			for scannerErr.Scan() {
				l := scannerErr.Text()
				ae.Stderr = append(ae.Stderr, l)
				fmt.Println(yellow("stderr"), l)
			}
		}()

		if err := cmd.Wait(); err != nil {
			fmt.Println(red("error"), err.Error())
		}
	}()
	return nil
}

// StartWithTimeout is the blocking verison of Start()
func (ae *AsyncExec) StartWithTimeout(timeout time.Duration) error {
	ae.complete = make(chan struct{})
	ae.end = func() {
		ae.complete <- struct{}{}
	}

	if err := ae.Start(); err != nil {
		return err
	}

	select {
	case <-ae.complete:
		fmt.Println(green("complete"))
		return nil
	case <-time.After(timeout):
		log.Println(yellow("timeout"))
		return fmt.Errorf("internal:timeout")
	}
}

func (ae *AsyncExec) endWithFatal() {
	if ae.EndAction != nil {
		if err := ae.EndAction(); err != nil {
			log.Fatal(err)
		}
	}

	if ae.end != nil {
		ae.end()
	}
}

func ifPanic(todo func(error)) {
	if r := recover(); r != nil {
		if todo != nil {
			todo(r.(error))
		}
	}
}
