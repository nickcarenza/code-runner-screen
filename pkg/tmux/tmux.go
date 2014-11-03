package tmux

import (
	"fmt"
	"os"
	"os/exec"
)

type Tmux struct {
	Session string
	target  string
}

func New(session string, command string) (*Tmux, error) {
	t := new(Tmux)
	t.Session = session
	t.target = fmt.Sprintf("-t %s", t.Session)
	t.Run("new", "-d", fmt.Sprintf("-s %s", t.Session), command)
	return t, nil
}

func (t *Tmux) Run(command string, args ...string) error {
	args = append([]string{command}, args...)
	cmd := exec.Command("tmux", args...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	// cmd.Stderr = os.Stderr
	cmd.Start()
	return cmd.Wait()
}

func (t *Tmux) BindKey(key string, global bool, args ...string) {
	globalFlag := ""
	if global {
		globalFlag = "-n"
	}

	args = append([]string{globalFlag, key}, args...)

	t.Run("bind-key", args...)
}

func (t *Tmux) Split(vertical bool, command string) {
	direction := "-h"
	// fmt.Sprintf("docker run --name '%s' -t -v %s:/app:ro  watch $(eval echo \$$lang)"
	if vertical {
		direction = "-v"
	}
	t.Run("split", t.target, direction, command)
}

func (t *Tmux) Attach() error {
	return t.Run("attach", t.target)
}
