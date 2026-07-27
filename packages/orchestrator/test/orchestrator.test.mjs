import assert from "node:assert/strict";
import test from "node:test";

import { Code, ConnectError } from "@connectrpc/connect";

import { ExitClass } from "../gen/akari/v1/orchestrator_pb.js";
import { exitClassFromCode } from "../src/exit-class.mjs";
import { absProjectRoot } from "../src/project-path.mjs";
import { createOrchestratorService } from "../src/service.mjs";
import {
  lintJsonPath,
  renderJsonPath,
  renderReportHtmlPath,
} from "../src/sidecar-json.mjs";

test("exitClassFromCode maps edit-lint / render-cut exit codes", () => {
  assert.equal(exitClassFromCode(0), ExitClass.OK);
  assert.equal(exitClassFromCode(1), ExitClass.REFUSAL);
  assert.equal(exitClassFromCode(2), ExitClass.EXECUTION_ERROR);
  assert.equal(exitClassFromCode(9), ExitClass.EXECUTION_ERROR);
});

test("absProjectRoot rejects empty input", () => {
  assert.throws(() => absProjectRoot(""), /project_root is required/);
});

test("sidecar paths are under .akari", () => {
  const root = "/tmp/project";
  assert.equal(lintJsonPath(root), "/tmp/project/.akari/lint.json");
  assert.equal(renderJsonPath(root), "/tmp/project/.akari/render.json");
  assert.equal(
    renderReportHtmlPath(root),
    "/tmp/project/.akari/reports/render-report.html",
  );
});

test("renderProject rejects missing plan_approved", async () => {
  const service = createOrchestratorService();
  await assert.rejects(
    () => service.renderProject({ projectRoot: "/tmp/project", planApproved: false }),
    (error) => error instanceof ConnectError && error.code === Code.FailedPrecondition,
  );
});

test("renderBatch rejects empty project_roots", async () => {
  const service = createOrchestratorService();
  const iterator = service.renderBatch({ projectRoots: [] });
  await assert.rejects(() => iterator.next(), (error) => error instanceof ConnectError && error.code === Code.InvalidArgument);
});
