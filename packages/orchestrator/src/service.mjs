import { create } from "@bufbuild/protobuf";
import { Code, ConnectError } from "@connectrpc/connect";

import {
  ExitClass,
  LintProjectResponseSchema,
  PlanRenderResponseSchema,
  RenderBatchEvent_Phase,
  RenderBatchEventSchema,
  RenderProjectResponseSchema,
} from "../gen/akari/v1/orchestrator_pb.js";

import { exitClassFromCode } from "./exit-class.mjs";
import { absProjectRoot } from "./project-path.mjs";
import {
  lintJsonPath,
  readLintJson,
  readRenderJson,
  renderJsonPath,
  renderReportHtmlPath,
} from "./sidecar-json.mjs";
import { runEditLintProject, runRenderCutProject } from "./worker.mjs";

export function createOrchestratorService() {
  return {
    async lintProject(request) {
      let projectRoot;
      try {
        projectRoot = absProjectRoot(request.projectRoot);
      } catch (error) {
        throw new ConnectError(error.message, Code.InvalidArgument);
      }

      const result = await runEditLintProject(projectRoot, {
        includeMedia: request.includeMedia,
      });
      const response = create(LintProjectResponseSchema, {
        exitClass: exitClassFromCode(result.code),
        lintJsonPath: lintJsonPath(projectRoot),
        message: result.message,
      });
      if (result.code !== 2) {
        try {
          const doc = await readLintJson(projectRoot);
          response.verdict = doc.verdict ?? "";
        } catch {
          // lint.json may be missing when edit-lint refused early.
        }
      }
      return response;
    },

    async planRender(request) {
      let projectRoot;
      try {
        projectRoot = absProjectRoot(request.projectRoot);
      } catch (error) {
        throw new ConnectError(error.message, Code.InvalidArgument);
      }

      const result = await runRenderCutProject(projectRoot, {
        planOnly: true,
        outputPath: request.outputPath,
        force: request.forceLintOverride,
      });
      const response = create(PlanRenderResponseSchema, {
        exitClass: result.exitClass,
        renderJsonPath: renderJsonPath(projectRoot),
        reportHtmlPath: renderReportHtmlPath(projectRoot),
        message: result.message,
      });
      if (result.code !== 2) {
        try {
          const doc = await readRenderJson(projectRoot);
          response.outputPath = doc.plan?.output ?? "";
          response.predictedDurationSeconds = doc.plan?.predicted_duration_seconds ?? 0;
        } catch {
          // render.json may be absent on refusal.
        }
      }
      return response;
    },

    async renderProject(request) {
      let projectRoot;
      try {
        projectRoot = absProjectRoot(request.projectRoot);
      } catch (error) {
        throw new ConnectError(error.message, Code.InvalidArgument);
      }
      if (!request.planApproved) {
        throw new ConnectError(
          "plan_approved must be true for non-interactive render",
          Code.FailedPrecondition,
        );
      }

      const result = await runRenderCutProject(projectRoot, {
        planOnly: false,
        outputPath: request.outputPath,
        force: request.forceLintOverride,
      });
      const response = create(RenderProjectResponseSchema, {
        exitClass: result.exitClass,
        renderJsonPath: renderJsonPath(projectRoot),
        message: result.message,
      });
      if (result.code !== 2) {
        try {
          const doc = await readRenderJson(projectRoot);
          response.outputPath = doc.plan?.output ?? "";
          response.verifyVerdict = doc.verify?.verdict ?? "";
        } catch {
          // render.json may be absent on refusal.
        }
      }
      return response;
    },

    async *renderBatch(request) {
      const roots = request.projectRoots ?? [];
      if (roots.length === 0) {
        throw new ConnectError("project_roots must not be empty", Code.InvalidArgument);
      }
      if (!request.planOnly && !request.planApproved) {
        throw new ConnectError(
          "plan_approved must be true when plan_only is false",
          Code.FailedPrecondition,
        );
      }

      const total = roots.length;
      for (const [index, root] of roots.entries()) {
        let projectRoot;
        try {
          projectRoot = absProjectRoot(root);
        } catch (error) {
          throw new ConnectError(`project_roots[${index}]: ${error.message}`, Code.InvalidArgument);
        }

        const lint = await this.lintProject({
          projectRoot,
          includeMedia: false,
        });
        yield create(RenderBatchEventSchema, {
          projectRoot,
          index: index + 1,
          total,
          phase: RenderBatchEvent_Phase.LINT,
          exitClass: lint.exitClass,
          message: lint.message,
        });
        if (lint.exitClass !== ExitClass.OK && !request.forceLintOverride) {
          if (!request.continueOnError) {
            return;
          }
          continue;
        }

        if (request.planOnly) {
          const plan = await this.planRender({
            projectRoot,
            forceLintOverride: request.forceLintOverride,
          });
          yield create(RenderBatchEventSchema, {
            projectRoot,
            index: index + 1,
            total,
            phase: RenderBatchEvent_Phase.PLAN,
            exitClass: plan.exitClass,
            message: plan.message,
            outputPath: plan.outputPath,
          });
          if (plan.exitClass !== ExitClass.OK && !request.continueOnError) {
            return;
          }
          continue;
        }

        const plan = await this.planRender({
          projectRoot,
          forceLintOverride: request.forceLintOverride,
        });
        yield create(RenderBatchEventSchema, {
          projectRoot,
          index: index + 1,
          total,
          phase: RenderBatchEvent_Phase.PLAN,
          exitClass: plan.exitClass,
          message: plan.message,
          outputPath: plan.outputPath,
        });
        if (plan.exitClass !== ExitClass.OK) {
          if (!request.continueOnError) {
            return;
          }
          continue;
        }

        const render = await this.renderProject({
          projectRoot,
          planApproved: true,
          forceLintOverride: request.forceLintOverride,
        });
        yield create(RenderBatchEventSchema, {
          projectRoot,
          index: index + 1,
          total,
          phase: RenderBatchEvent_Phase.RENDER,
          exitClass: render.exitClass,
          message: render.message,
          outputPath: render.outputPath,
        });
        yield create(RenderBatchEventSchema, {
          projectRoot,
          index: index + 1,
          total,
          phase: RenderBatchEvent_Phase.DONE,
          exitClass: render.exitClass,
          message: render.message,
          outputPath: render.outputPath,
        });
        if (render.exitClass !== ExitClass.OK && !request.continueOnError) {
          return;
        }
      }
    },
  };
}
