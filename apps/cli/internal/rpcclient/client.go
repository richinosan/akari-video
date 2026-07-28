package rpcclient

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"

	"connectrpc.com/connect"
	akariv1 "github.com/richinosan/akari-video/apps/cli/gen/akari/v1"
	akariv1connect "github.com/richinosan/akari-video/apps/cli/gen/akari/v1/akariv1connect"
	"github.com/richinosan/akari-video/apps/cli/internal/nodebundle"
)

const (
	DefaultAddr = "127.0.0.1:7707"
	healthPath  = "/healthz"
)

// Client talks to the Node orchestrator worker over ConnectRPC.
type Client struct {
	addr       string
	httpClient *http.Client
	api        akariv1connect.OrchestratorServiceClient
	worker     *workerProcess
}

type workerProcess struct {
	cmd *exec.Cmd
}

// New creates a client for the given orchestrator address.
func New(addr string) *Client {
	if addr == "" {
		addr = DefaultAddr
	}
	transport := &http.Transport{}
	httpClient := &http.Client{Transport: transport}
	return &Client{
		addr:       addr,
		httpClient: httpClient,
		api: akariv1connect.NewOrchestratorServiceClient(
			httpClient,
			"http://"+addr,
		),
	}
}

// Addr returns the orchestrator listen address.
func (c *Client) Addr() string {
	return c.addr
}

// EnsureWorker starts the bundled Node worker when health checks fail.
func (c *Client) EnsureWorker(ctx context.Context) error {
	if err := c.ping(ctx); err == nil {
		return nil
	}
	if c.worker != nil {
		_ = c.worker.cmd.Process.Kill()
		c.worker = nil
	}
	wp, err := spawnWorker(c.addr)
	if err != nil {
		return err
	}
	c.worker = wp
	if err := waitHealthy(ctx, c.addr, 30*time.Second); err != nil {
		_ = wp.cmd.Process.Kill()
		c.worker = nil
		return err
	}
	return nil
}

// RunWorkerForeground starts the Node worker and blocks until ctx is cancelled.
func (c *Client) RunWorkerForeground(ctx context.Context) error {
	if err := c.ping(ctx); err == nil {
		return fmt.Errorf("orchestrator already listening on %s", c.addr)
	}
	wp, err := spawnWorker(c.addr)
	if err != nil {
		return err
	}
	c.worker = wp
	defer func() {
		_ = wp.cmd.Process.Kill()
		c.worker = nil
	}()

	if err := waitHealthy(ctx, c.addr, 30*time.Second); err != nil {
		return err
	}

	done := make(chan error, 1)
	go func() {
		done <- wp.cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		_ = wp.cmd.Process.Signal(os.Interrupt)
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			_ = wp.cmd.Process.Kill()
		}
		return ctx.Err()
	case err := <-done:
		if err != nil {
			return fmt.Errorf("node orchestrator exited: %w", err)
		}
		return nil
	}
}

func spawnWorker(addr string) (*workerProcess, error) {
	runtimeRoot, err := nodebundle.RuntimeRoot()
	if err != nil {
		return nil, err
	}
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("parse orchestrator addr %q: %w", addr, err)
	}
	listenAddr := net.JoinHostPort(host, port)
	cmd := exec.Command(nodebundle.NodeBinary(), nodebundle.OrchestratorScript(runtimeRoot), "serve", "--addr", listenAddr)
	cmd.Dir = runtimeRoot
	cmd.Env = append(os.Environ(), "NODE_ENV=production")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start node orchestrator: %w", err)
	}
	return &workerProcess{cmd: cmd}, nil
}

func (c *Client) ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+c.addr+healthPath, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health status %d", resp.StatusCode)
	}
	return nil
}

func waitHealthy(ctx context.Context, addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 2 * time.Second}
	for time.Now().Before(deadline) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+addr+healthPath, nil)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(200 * time.Millisecond):
		}
	}
	return fmt.Errorf("orchestrator at %s did not become healthy within %s", addr, timeout)
}

// LintProject runs edit-lint via RPC.
func (c *Client) LintProject(ctx context.Context, req *akariv1.LintProjectRequest) (*akariv1.LintProjectResponse, error) {
	resp, err := c.api.LintProject(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return resp.Msg, nil
}

// PlanRender returns a render plan via RPC.
func (c *Client) PlanRender(ctx context.Context, req *akariv1.PlanRenderRequest) (*akariv1.PlanRenderResponse, error) {
	resp, err := c.api.PlanRender(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return resp.Msg, nil
}

// RenderProject executes render-cut via RPC.
func (c *Client) RenderProject(ctx context.Context, req *akariv1.RenderProjectRequest) (*akariv1.RenderProjectResponse, error) {
	resp, err := c.api.RenderProject(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return resp.Msg, nil
}

// RunBatch streams batch events to emit until the stream ends.
func (c *Client) RunBatch(
	ctx context.Context,
	req *akariv1.RenderBatchRequest,
	emit func(*akariv1.RenderBatchEvent) error,
) error {
	stream, err := c.api.RenderBatch(ctx, connect.NewRequest(req))
	if err != nil {
		return err
	}
	for stream.Receive() {
		if err := emit(stream.Msg()); err != nil {
			return err
		}
	}
	return stream.Err()
}
