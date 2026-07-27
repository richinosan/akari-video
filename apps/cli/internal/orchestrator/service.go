package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"connectrpc.com/connect"

	akariv1 "github.com/richinosan/akari-video/apps/cli/gen/akari/v1"
	"github.com/richinosan/akari-video/apps/cli/internal/lintjson"
	"github.com/richinosan/akari-video/apps/cli/internal/nodecli"
	"github.com/richinosan/akari-video/apps/cli/internal/renderjson"
)

type Service struct {
	runner *nodecli.Runner
}

func NewService(repoRoot string) *Service {
	return &Service{
		runner: nodecli.NewRunner(repoRoot, nodecli.NodeFromEnv()),
	}
}

func (s *Service) LintProject(
	ctx context.Context,
	req *connect.Request[akariv1.LintProjectRequest],
) (*connect.Response[akariv1.LintProjectResponse], error) {
	projectRoot, err := absProject(req.Msg.GetProjectRoot())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	code, output, runErr := s.runner.EditLint(ctx, projectRoot, req.Msg.GetIncludeMedia())
	resp := &akariv1.LintProjectResponse{
		ExitClass:    nodecli.ExitClass(code),
		LintJsonPath: lintjson.JSONPath(projectRoot),
		Message:      output,
	}
	if runErr != nil {
		resp.ExitClass = akariv1.ExitClass_EXIT_CLASS_EXECUTION_ERROR
		resp.Message = runErr.Error()
		return connect.NewResponse(resp), nil
	}

	if doc, err := lintjson.Read(projectRoot); err == nil {
		resp.Verdict = doc.Verdict
	}
	return connect.NewResponse(resp), nil
}

func (s *Service) PlanRender(
	ctx context.Context,
	req *connect.Request[akariv1.PlanRenderRequest],
) (*connect.Response[akariv1.PlanRenderResponse], error) {
	projectRoot, err := absProject(req.Msg.GetProjectRoot())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	code, output, runErr := s.runner.RenderCut(ctx, projectRoot, true, req.Msg.GetOutputPath(), req.Msg.GetForceLintOverride())
	resp := &akariv1.PlanRenderResponse{
		ExitClass:       nodecli.ExitClass(code),
		RenderJsonPath:  renderjson.JSONPath(projectRoot),
		ReportHtmlPath:  renderjson.ReportHTMLPath(projectRoot),
		Message:         output,
	}
	if runErr != nil {
		resp.ExitClass = akariv1.ExitClass_EXIT_CLASS_EXECUTION_ERROR
		resp.Message = runErr.Error()
		return connect.NewResponse(resp), nil
	}
	if doc, err := renderjson.Read(projectRoot); err == nil {
		resp.OutputPath = doc.Plan.Output
		resp.PredictedDurationSeconds = doc.Plan.PredictedDurationSeconds
	}
	return connect.NewResponse(resp), nil
}

func (s *Service) RenderProject(
	ctx context.Context,
	req *connect.Request[akariv1.RenderProjectRequest],
) (*connect.Response[akariv1.RenderProjectResponse], error) {
	projectRoot, err := absProject(req.Msg.GetProjectRoot())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if !req.Msg.GetPlanApproved() {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("plan_approved must be true for non-interactive render"))
	}

	code, output, runErr := s.runner.RenderCut(ctx, projectRoot, false, req.Msg.GetOutputPath(), req.Msg.GetForceLintOverride())
	resp := &akariv1.RenderProjectResponse{
		ExitClass:      nodecli.ExitClass(code),
		RenderJsonPath: renderjson.JSONPath(projectRoot),
		Message:        output,
	}
	if runErr != nil {
		resp.ExitClass = akariv1.ExitClass_EXIT_CLASS_EXECUTION_ERROR
		resp.Message = runErr.Error()
		return connect.NewResponse(resp), nil
	}
	if doc, err := renderjson.Read(projectRoot); err == nil {
		resp.OutputPath = doc.Plan.Output
		resp.VerifyVerdict = doc.Verify.Verdict
	}
	return connect.NewResponse(resp), nil
}

type BatchEmitter func(*akariv1.RenderBatchEvent) error

func (s *Service) RenderBatch(
	ctx context.Context,
	req *connect.Request[akariv1.RenderBatchRequest],
	stream *connect.ServerStream[akariv1.RenderBatchEvent],
) error {
	return s.RunBatch(ctx, req.Msg, stream.Send)
}

func (s *Service) RunBatch(ctx context.Context, req *akariv1.RenderBatchRequest, emit BatchEmitter) error {
	roots := req.GetProjectRoots()
	total := int32(len(roots))
	if total == 0 {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("project_roots must not be empty"))
	}
	if !req.GetPlanOnly() && !req.GetPlanApproved() {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("plan_approved must be true when plan_only is false"))
	}

	for i, root := range roots {
		projectRoot, err := absProject(root)
		if err != nil {
			return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("project_roots[%d]: %w", i, err))
		}
		index := int32(i + 1)

		lintResp, err := s.LintProject(ctx, connect.NewRequest(&akariv1.LintProjectRequest{
			ProjectRoot:  projectRoot,
			IncludeMedia: false,
		}))
		if err != nil {
			return err
		}
		if err := emit(&akariv1.RenderBatchEvent{
			ProjectRoot: projectRoot,
			Index:       index,
			Total:       total,
			Phase:       akariv1.RenderBatchEvent_PHASE_LINT,
			ExitClass:   lintResp.Msg.GetExitClass(),
			Message:     lintResp.Msg.GetMessage(),
		}); err != nil {
			return err
		}
		if lintResp.Msg.GetExitClass() != akariv1.ExitClass_EXIT_CLASS_OK && !req.GetForceLintOverride() {
			if !req.GetContinueOnError() {
				return nil
			}
			continue
		}

		if req.GetPlanOnly() {
			planResp, err := s.PlanRender(ctx, connect.NewRequest(&akariv1.PlanRenderRequest{
				ProjectRoot:       projectRoot,
				ForceLintOverride: req.GetForceLintOverride(),
			}))
			if err != nil {
				return err
			}
			if err := emit(&akariv1.RenderBatchEvent{
				ProjectRoot: projectRoot,
				Index:       index,
				Total:       total,
				Phase:       akariv1.RenderBatchEvent_PHASE_PLAN,
				ExitClass:   planResp.Msg.GetExitClass(),
				Message:     planResp.Msg.GetMessage(),
				OutputPath:  planResp.Msg.GetOutputPath(),
			}); err != nil {
				return err
			}
			if planResp.Msg.GetExitClass() != akariv1.ExitClass_EXIT_CLASS_OK && !req.GetContinueOnError() {
				return nil
			}
			continue
		}

		planResp, err := s.PlanRender(ctx, connect.NewRequest(&akariv1.PlanRenderRequest{
			ProjectRoot:       projectRoot,
			ForceLintOverride: req.GetForceLintOverride(),
		}))
		if err != nil {
			return err
		}
		if err := emit(&akariv1.RenderBatchEvent{
			ProjectRoot: projectRoot,
			Index:       index,
			Total:       total,
			Phase:       akariv1.RenderBatchEvent_PHASE_PLAN,
			ExitClass:   planResp.Msg.GetExitClass(),
			Message:     planResp.Msg.GetMessage(),
			OutputPath:  planResp.Msg.GetOutputPath(),
		}); err != nil {
			return err
		}
		if planResp.Msg.GetExitClass() != akariv1.ExitClass_EXIT_CLASS_OK {
			if !req.GetContinueOnError() {
				return nil
			}
			continue
		}

		renderResp, err := s.RenderProject(ctx, connect.NewRequest(&akariv1.RenderProjectRequest{
			ProjectRoot:       projectRoot,
			PlanApproved:      true,
			ForceLintOverride: req.GetForceLintOverride(),
		}))
		if err != nil {
			return err
		}
		if err := emit(&akariv1.RenderBatchEvent{
			ProjectRoot: projectRoot,
			Index:       index,
			Total:       total,
			Phase:       akariv1.RenderBatchEvent_PHASE_RENDER,
			ExitClass:   renderResp.Msg.GetExitClass(),
			Message:     renderResp.Msg.GetMessage(),
			OutputPath:  renderResp.Msg.GetOutputPath(),
		}); err != nil {
			return err
		}
		if err := emit(&akariv1.RenderBatchEvent{
			ProjectRoot: projectRoot,
			Index:       index,
			Total:       total,
			Phase:       akariv1.RenderBatchEvent_PHASE_DONE,
			ExitClass:   renderResp.Msg.GetExitClass(),
			Message:     renderResp.Msg.GetMessage(),
			OutputPath:  renderResp.Msg.GetOutputPath(),
		}); err != nil {
			return err
		}
		if renderResp.Msg.GetExitClass() != akariv1.ExitClass_EXIT_CLASS_OK && !req.GetContinueOnError() {
			return nil
		}
	}
	return nil
}

func absProject(projectRoot string) (string, error) {
	if projectRoot == "" {
		return "", errors.New("project_root is required")
	}
	abs, err := filepath.Abs(projectRoot)
	if err != nil {
		return "", err
	}
	return abs, nil
}
