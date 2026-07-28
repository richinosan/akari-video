package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	akariv1 "github.com/richinosan/akari-video/apps/cli/gen/akari/v1"
)

func newRenderCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "render",
		Short: "Plan or render a single project",
	}
	cmd.AddCommand(newRenderPlanCommand())
	cmd.AddCommand(newRenderRunCommand())
	return cmd
}

func newRenderPlanCommand() *cobra.Command {
	var outputPath string
	var force bool
	cmd := &cobra.Command{
		Use:   "plan <project-root>",
		Short: "Run render-cut --plan-only via OrchestratorService",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newRPCClient(context.Background())
			if err != nil {
				return err
			}
			resp, err := client.PlanRender(context.Background(), &akariv1.PlanRenderRequest{
				ProjectRoot:       args[0],
				OutputPath:        outputPath,
				ForceLintOverride: force,
			})
			if err != nil {
				return err
			}
			printExitClass(cmd, resp.GetExitClass())
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), resp.GetMessage())
			if resp.GetOutputPath() != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "output: %s\n", resp.GetOutputPath())
			}
			return exitFromClass(resp.GetExitClass())
		},
	}
	cmd.Flags().StringVar(&outputPath, "out", "", "output path override")
	cmd.Flags().BoolVar(&force, "force", false, "override lint refusal (render-cut --force)")
	return cmd
}

func newRenderRunCommand() *cobra.Command {
	var outputPath string
	var force bool
	var approve bool
	cmd := &cobra.Command{
		Use:   "run <project-root>",
		Short: "Run render-cut after explicit plan approval",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !approve {
				return fmt.Errorf("--approve-plan is required for non-interactive render")
			}
			client, err := newRPCClient(context.Background())
			if err != nil {
				return err
			}
			resp, err := client.RenderProject(context.Background(), &akariv1.RenderProjectRequest{
				ProjectRoot:       args[0],
				OutputPath:        outputPath,
				PlanApproved:      true,
				ForceLintOverride: force,
			})
			if err != nil {
				return err
			}
			printExitClass(cmd, resp.GetExitClass())
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), resp.GetMessage())
			if resp.GetOutputPath() != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "output: %s\n", resp.GetOutputPath())
			}
			return exitFromClass(resp.GetExitClass())
		},
	}
	cmd.Flags().StringVar(&outputPath, "out", "", "output path override")
	cmd.Flags().BoolVar(&force, "force", false, "override lint refusal (render-cut --force)")
	cmd.Flags().BoolVar(&approve, "approve-plan", false, "explicit approval gate for render")
	return cmd
}

func printExitClass(cmd *cobra.Command, class akariv1.ExitClass) {
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "exit_class: %s\n", class.String())
}

func exitFromClass(class akariv1.ExitClass) error {
	switch class {
	case akariv1.ExitClass_EXIT_CLASS_OK:
		return nil
	case akariv1.ExitClass_EXIT_CLASS_REFUSAL:
		return &exitError{code: 1}
	case akariv1.ExitClass_EXIT_CLASS_EXECUTION_ERROR:
		return &exitError{code: 2}
	default:
		return &exitError{code: 2}
	}
}

type exitError struct {
	code int
}

func (e *exitError) Error() string {
	return fmt.Sprintf("exit %d", e.code)
}
