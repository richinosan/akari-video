package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	akariv1connect "github.com/richinosan/akari-video/apps/cli/gen/akari/v1/akariv1connect"
	"github.com/richinosan/akari-video/apps/cli/internal/orchestrator"
	"github.com/richinosan/akari-video/apps/cli/internal/reporoot"
)

var (
	repoRootFlag string
)

func Execute() error {
	err := newRootCommand().Execute()
	if err == nil {
		return nil
	}
	var exitErr *exitError
	if errors.As(err, &exitErr) {
		os.Exit(exitErr.code)
	}
	return err
}

func newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "akari",
		Short: "AKARI Video headless CLI",
	}
	root.PersistentFlags().StringVar(&repoRootFlag, "repo-root", "", "akari-video monorepo root (default: auto-detect or AKARI_VIDEO_ROOT)")

	root.AddCommand(newVersionCommand())
	root.AddCommand(newServeCommand())
	root.AddCommand(newRenderCommand())
	root.AddCommand(newBatchCommand())
	return root
}

func resolveRepoRoot() (string, error) {
	if repoRootFlag != "" {
		return reporoot.FromEnvOrResolve(repoRootFlag)
	}
	return reporoot.FromEnvOrResolve("")
}

func newService() (*orchestrator.Service, error) {
	root, err := resolveRepoRoot()
	if err != nil {
		return nil, err
	}
	return orchestrator.NewService(root), nil
}

func newServeCommand() *cobra.Command {
	var addr string
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start ConnectRPC OrchestratorService (IPC)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			svc, err := newService()
			if err != nil {
				return err
			}
			mux := http.NewServeMux()
			path, handler := akariv1connect.NewOrchestratorServiceHandler(svc)
			mux.Handle(path, handler)

			server := &http.Server{
				Addr:    addr,
				Handler: h2c.NewHandler(mux, &http2.Server{}),
			}

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			errCh := make(chan error, 1)
			go func() {
				fmt.Fprintf(cmd.ErrOrStderr(), "akari serve listening on %s (ConnectRPC %s)\n", addr, path)
				errCh <- server.ListenAndServe()
			}()

			select {
			case <-ctx.Done():
				return server.Shutdown(context.Background())
			case err := <-errCh:
				if err == http.ErrServerClosed {
					return nil
				}
				return err
			}
		},
	}
	cmd.Flags().StringVar(&addr, "addr", "127.0.0.1:7707", "listen address")
	return cmd
}
