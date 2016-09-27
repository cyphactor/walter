package task

import (
	"bufio"
	"os/exec"
	"sync"

	log "github.com/Sirupsen/logrus"
)

type Task struct {
	Name           string
	Command        string
	Parallel       []Task
	Serial         []Task
	Stdout         []string
	Stderr         []string
	CombinedOutput []string
}

func (t *Task) Run() error {
	if len(t.Parallel) > 0 {
		runParallel(t.Parallel)
	}

	if len(t.Serial) > 0 {
		runSerial(t.Serial)
	}

	if t.Command == "" {
		return nil
	}

	cmd := exec.Command("sh", "-c", t.Command)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			text := scanner.Text()
			t.Stdout = append(t.Stdout, text)
			t.CombinedOutput = append(t.CombinedOutput, text)
			log.Info(text)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			text := scanner.Text()
			t.Stderr = append(t.Stderr, text)
			t.CombinedOutput = append(t.CombinedOutput, text)
			log.Info(text)
		}
	}()

	wg.Wait()
	cmd.Wait()

	return nil
}

func runParallel(tasks []Task) {
	var wg sync.WaitGroup
	for _, t := range tasks {
		wg.Add(1)
		go func(t Task) {
			defer wg.Done()
			t.Run()
		}(t)
	}
	wg.Wait()
}

func runSerial(tasks []Task) {
	for _, t := range tasks {
		t.Run()
	}
}
