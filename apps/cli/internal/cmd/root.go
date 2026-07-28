package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/richinosan/akari-video/apps/cli/internal/applaunch"
	"github.com/richinosan/akari-video/apps/cli/internal/reporoot"
	"github.com/richinosan/akari-video/apps/cli/internal/rpcclient"
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
		Short: "AKARI Video",
		Long:  "Launch the AKARI Video desktop app. Use `akari cli` for headless commands (render, batch, serve).",
		RunE: func(_ *cobra.Command, args []string) error {
			return applaunch.Run(args)
		},
	}
	root.PersistentFlags().StringVar(&repoRootFlag, "repo-root", "", "akari-video monorepo root (default: auto-detect or AKARI_VIDEO_ROOT)")

	cli := &cobra.Command{
		Use:   "cli",
		Short: "Headless CLI (render, batch, serve)",
	}
	cli.AddCommand(newVersionCommand())
	cli.AddCommand(newServeCommand())
	cli.AddCommand(newRenderCommand())
	cli.AddCommand(newBatchCommand())
	root.AddCommand(cli)
	return root
}

func orchestratorAddr() string {
	if v := os.Getenv("AKARI_ORCHESTRATOR_ADDR"); v != "" {
		return v
	}
	return rpcclient.DefaultAddr
}

func applyRepoRootFlag() error {
	if repoRootFlag == "" {
		return nil
	}
	root, err := reporoot.FromEnvOrResolve(repoRootFlag)
	if err != nil {
		return err
	}
	return os.Setenv("AKARI_VIDEO_ROOT", root)
}

func newRPCClient(ctx context.Context) (*rpcclient.Client, error) {
	if err := applyRepoRootFlag(); err != nil {
		return nil, err
	}
	client := rpcclient.New(orchestratorAddr())
	if err := client.EnsureWorker(ctx); err != nil {
		return nil, fmt.Errorf("start orchestrator worker: %w", err)
	}
	return client, nil
}

func newServeCommand() *cobra.Command {
	var addr string
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start bundled Node orchestrator (ConnectRPC)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := applyRepoRootFlag(); err != nil {
				return err
			}
			if addr == "" {
				addr = orchestratorAddr()
			}
			client := rpcclient.New(addr)
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "akari serve starting node orchestrator on %s\n", addr)
			return client.RunWorkerForeground(ctx)
		},
	}
	cmd.Flags().StringVar(&addr, "addr", "", "listen address (default: AKARI_ORCHESTRATOR_ADDR or 127.0.0.1:7707)")
	return cmd
}
