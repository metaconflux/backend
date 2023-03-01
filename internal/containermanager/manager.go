package containermanager

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/containers/podman/v4/libpod/define"
	"github.com/containers/podman/v4/pkg/bindings"
	"github.com/containers/podman/v4/pkg/bindings/containers"
	"github.com/containers/podman/v4/pkg/bindings/images"
	"github.com/containers/podman/v4/pkg/specgen"
	"github.com/hashicorp/go-multierror"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
)

type CleanFn func() error

type ContainerManager struct {
	connectionCtx context.Context
	containers    map[string]string
	images        map[string]string
}

type Task struct {
	Timeout time.Duration
	Image   string
	Name    string
	Input   interface{}
}

func NewManager(url string) (*ContainerManager, error) {
	conn, err := bindings.NewConnection(context.Background(), "unix://run/user/1000/podman/podman.sock")
	if err != nil {
		return nil, err
	}

	return &ContainerManager{
		connectionCtx: conn,
		containers:    map[string]string{},
		images:        map[string]string{},
	}, nil
}

func (m *ContainerManager) Kill(nameOrId string) error {
	ins, err := containers.Inspect(m.connectionCtx, nameOrId, nil)
	if err != nil {
		return err
	}

	if ins.State.Running {
		signal := "SIGKILL"
		err := containers.Kill(m.connectionCtx, nameOrId, &containers.KillOptions{Signal: &signal})
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *ContainerManager) Remove(nameOrId string, force bool) error {
	_, err := containers.Remove(m.connectionCtx, nameOrId, &containers.RemoveOptions{
		Force: &force,
	})
	if err != nil {
		return err
	}

	delete(m.containers, nameOrId)

	return nil
}

func (m *ContainerManager) PullIfNotPresent(ctx context.Context, image string) error {
	if ctx == nil {
		ctx = m.connectionCtx
	}
	exists, err := images.Exists(ctx, image, nil)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	_, err = images.Pull(ctx, image, nil)
	if err != nil {
		return err
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	return nil
}

func (m *ContainerManager) Create(ctx context.Context, image string, name string, cpu string, memory string, env map[string]string) (string, CleanFn, error) {
	if ctx == nil {
		ctx = m.connectionCtx
	}

	s := specgen.NewSpecGenerator(image, false)
	s.Name = name
	s.ResourceLimits = &specs.LinuxResources{}

	if cpu != "" {
		s.ResourceLimits.CPU = &specs.LinuxCPU{
			Cpus: "100m",
		}
	}

	if memory != "" {
		memoryLimit := int64(26_214_400)
		s.ResourceLimits.Memory = &specs.LinuxMemory{
			Limit: &memoryLimit,
		}
	}

	s.Env = env

	cont, err := containers.CreateWithSpec(ctx, s, nil)
	if err != nil {
		return "", nil, err
	}

	m.containers[s.Name] = cont.ID
	clean := func() error {
		var errs error
		err := m.Kill(name)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
		err = m.Remove(name, true)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
		return errs
	}

	return s.Name, clean, nil
}

func (m *ContainerManager) Start(ctx context.Context, nameOrId string) error {
	if ctx == nil {
		ctx = m.connectionCtx
	}
	err := containers.Start(ctx, nameOrId, nil)
	if err != nil {
		return err
	}

	return nil
}

func (m *ContainerManager) WaitStop(ctx context.Context, nameOrId string) chan error {
	if ctx == nil {
		ctx = m.connectionCtx
	}

	result := make(chan error)

	go func() {
		ret, err := containers.Wait(ctx, nameOrId, &containers.WaitOptions{
			Condition: []define.ContainerStatus{
				define.ContainerStateStopped,
			},
		})
		if err != nil {
			result <- err
			return
		}

		if ret != 0 {
			result <- fmt.Errorf("Container exited with %d", ret)
			return
		}
		result <- nil
	}()

	return result

}

func (m *ContainerManager) Logs(ctx context.Context, nameOrId string) (chan string, chan error, error) {
	if ctx == nil {
		ctx = m.connectionCtx
	}
	stdout := make(chan string)
	stderr := make(chan error)
	outCh := make(chan string)
	errCh := make(chan string)
	done := make(chan struct{})

	go func() {
		out := ""
		errStr := ""
		for {
			select {
			case o := <-outCh:
				out += o
			case o := <-errCh:
				errStr += o
			case <-ctx.Done():
				errStr = ctx.Err().Error()
				done <- struct{}{}
			case <-done:
				stdout <- out
				var err error
				if errStr != "" {
					err = errors.New(errStr)
				}
				stderr <- err
				return
			}
		}
	}()

	options := &containers.LogOptions{}
	truePtr := true
	options.Stderr = &truePtr
	options.Stdout = &truePtr

	go func() {
		defer func() {
			done <- struct{}{}
		}()
		err := containers.Logs(ctx, nameOrId, options, outCh, errCh)
		if err != nil {
			stderr <- err
		}
	}()

	return stdout, stderr, nil
}

func (m *ContainerManager) WithTimeout(timeout time.Duration) (*ContainerManager, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(m.connectionCtx, timeout)
	return &ContainerManager{
		connectionCtx: ctx,
		containers:    m.containers,
		images:        m.images,
	}, cancel
}

func (m *ContainerManager) Task(task Task) (interface{}, error) {
	return m.TaskT(task.Timeout, task.Image, task.Name, task.Input)
}

func (m *ContainerManager) TaskT(timeout time.Duration, image string, name string, input interface{}) (interface{}, error) {
	ctx, cancel := context.WithTimeout(m.connectionCtx, timeout)
	defer cancel()

	data, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	env := map[string]string{
		"SYNTH_NAME": name,
		"DATA":       encoded,
	}

	err = m.PullIfNotPresent(ctx, image)
	if err != nil {
		return nil, err
	}

	id, clean, err := m.Create(ctx, image, name, "", "", env)
	if err != nil {
		return nil, err
	}
	defer clean()

	err = m.Start(ctx, id)
	if err != nil {
		return nil, err
	}

	waitErr := m.WaitStop(ctx, id)
	var waitError error
	select {
	case err := <-waitErr:
		if err != nil {
			waitError = err
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	outCh, errCh, err := m.Logs(ctx, id)
	if err != nil {
		return nil, err
	}

	var result string
	for i := 0; i < 2; i++ {
		select {
		case r := <-outCh:
			result = r
		case e := <-errCh:
			err = e
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if waitError != nil || err != nil {
		return nil, fmt.Errorf("Container stderr: %s", err)
	}

	//logrus.Infof("Getting result from: %s", result)

	re, err := regexp.Compile(fmt.Sprintf("%s: (.*)\n", id))
	if err != nil {
		return nil, err
	}

	match := re.FindAllStringSubmatch(result, 1)

	var taskResult interface{}
	if len(match) > 0 && len(match[0]) > 1 {
		err = json.Unmarshal([]byte(match[0][1]), &taskResult)
		if err != nil {
			return nil, err
		}
	}

	return taskResult, nil
}

func (m *ContainerManager) Close() {

	for _, id := range m.containers {
		err := m.Kill(id)
		if err != nil {
			logrus.Errorf("Failed to kill container %s: %s", id, err)
		}
		err = m.Remove(id, true)
		if err != nil {
			logrus.Errorf("Failed to remove container %s: %s", id, err)
		}
	}

	conn, _ := bindings.GetClient(m.connectionCtx)

	dialer, _ := conn.GetDialer(m.connectionCtx)

	dialer.Close()
}
