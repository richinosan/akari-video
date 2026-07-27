import { runCli as runEditLint } from "../../edit-lint/src/edit-lint.mjs";
import { runCli as runRenderCut } from "../../render-cut/src/render-cut.mjs";

import { captureIo } from "./capture-io.mjs";
import { exitClassFromCode } from "./exit-class.mjs";

export async function runEditLintProject(projectRoot, { includeMedia = false } = {}) {
  const args = [projectRoot];
  if (includeMedia) {
    args.push("--media");
  }
  const { io, output } = captureIo();
  try {
    const code = await runEditLint(args, io);
    return { code, message: output() };
  } catch (error) {
    return {
      code: 2,
      message: error instanceof Error ? error.message : String(error),
    };
  }
}

export async function runRenderCutProject(
  projectRoot,
  { planOnly = false, outputPath = "", force = false } = {},
) {
  const args = [projectRoot];
  if (planOnly) {
    args.push("--plan-only");
  }
  if (outputPath) {
    args.push("--out", outputPath);
  }
  if (force) {
    args.push("--force");
  }
  const { io, output } = captureIo();
  try {
    const code = await runRenderCut(args, io);
    return { code, message: output(), exitClass: exitClassFromCode(code) };
  } catch (error) {
    return {
      code: 2,
      message: error instanceof Error ? error.message : String(error),
      exitClass: exitClassFromCode(2),
    };
  }
}
