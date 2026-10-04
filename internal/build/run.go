package build

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Event is one line of process output or the terminal result.
type Event struct {
	Line string
	Err  error
	Done bool
}

// Start runs argv with its working directory set to dir.
// cancel signals the process group. The events channel closes after a final Done event.
func Start(dir string, argv []string) (cancel func(), events <-chan Event, err error) {
	if len(argv) == 0 {
		return nil, nil, errors.New("empty command")
	}
	bin := argv[0]
	if strings.ContainsAny(bin, `/\`) && !filepath.IsAbs(bin) {
		bin = filepath.Join(dir, bin)
	}
	ctx, cancelCtx := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, bin, argv[1:]...)
	cmd.Dir = dir
	cmd.Env = os.Environ()
	prepareCommand(cmd)

	r, w, err := os.Pipe()
	if err != nil {
		cancelCtx()
		return nil, nil, err
	}
	cmd.Stdout = w
	cmd.Stderr = w
	if err := cmd.Start(); err != nil {
		cancelCtx()
		r.Close()
		w.Close()
		return nil, nil, err
	}
	w.Close()

	ch := make(chan Event, 64)
	go readProcess(ctx, cmd, r, ch)
	cancel = func() {
		cancelCtx()
		signalGroup(cmd)
	}
	return cancel, ch, nil
}

func readProcess(ctx context.Context, cmd *exec.Cmd, r io.ReadCloser, ch chan Event) {
	defer close(ch)
	defer r.Close()
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	stop := false
	for sc.Scan() {
		select {
		case <-ctx.Done():
			stop = true
		case ch <- Event{Line: sc.Text()}:
			continue
		}
		break
	}
	if stop {
		signalGroup(cmd)
	}
	waitErr := cmd.Wait()
	if waitErr == nil {
		waitErr = sc.Err()
	}
	ch <- Event{Done: true, Err: waitErr}
}
