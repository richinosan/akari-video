package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	akariv1 "github.com/richinosan/akari-video/apps/cli/gen/akari/v1"
)

func newBatchCommand() *cobra.Command {
	var planOnly bool
	var approve bool
	var continueOnError bool
	var force bool
	cmd := &cobra.Command{
		Use:   "batch <project-root> [more...]",
		Short: "Plan or render multiple projects (ConnectRPC RenderBatch)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !planOnly && !approve {
				return fmt.Errorf("--approve-plan is required when not using --plan-only")
			}
			client, err := newRPCClient(context.Background())
			if err != nil {
				return err
			}
			var worst akariv1.ExitClass
			err = client.RunBatch(context.Background(), &akariv1.RenderBatchRequest{
				ProjectRoots:      args,
				PlanOnly:          planOnly,
				PlanApproved:      approve,
				ContinueOnError:   continueOnError,
				ForceLintOverride: force,
			}, func(event *akariv1.RenderBatchEvent) error {
				if event.GetExitClass() != akariv1.ExitClass_EXIT_CLASS_OK && event.GetExitClass() != akariv1.ExitClass_EXIT_CLASS_UNSPECIFIED {
					if worst == akariv1.ExitClass_EXIT_CLASS_UNSPECIFIED || event.GetExitClass() == akariv1.ExitClass_EXIT_CLASS_EXECUTION_ERROR {
						worst = event.GetExitClass()
					}
				}
				_, _ = fmt.Fprintf(
					cmd.OutOrStdout(),
					"[%d/%d] %s %s %s\n",
					event.GetIndex(),
					event.GetTotal(),
					event.GetProjectRoot(),
					event.GetPhase().String(),
					event.GetExitClass().String(),
				)
				if event.GetMessage() != "" {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", event.GetMessage())
				}
				if event.GetOutputPath() != "" {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  output: %s\n", event.GetOutputPath())
				}
				return nil
			})
			if err != nil {
				return err
			}
			if worst == akariv1.ExitClass_EXIT_CLASS_UNSPECIFIED {
				return nil
			}
			return exitFromClass(worst)
		},
	}
	cmd.Flags().BoolVar(&planOnly, "plan-only", false, "plan each project without rendering")
	cmd.Flags().BoolVar(&approve, "approve-plan", false, "explicit approval for render phase")
	cmd.Flags().BoolVar(&continueOnError, "continue-on-error", false, "continue batch after a project fails")
	cmd.Flags().BoolVar(&force, "force", false, "override lint refusal")
	return cmd
}
