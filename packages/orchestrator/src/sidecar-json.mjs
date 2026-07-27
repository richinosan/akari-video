import { readFile } from "node:fs/promises";
import { join } from "node:path";

export function lintJsonPath(projectRoot) {
  return join(projectRoot, ".akari", "lint.json");
}

export function renderJsonPath(projectRoot) {
  return join(projectRoot, ".akari", "render.json");
}

export function renderReportHtmlPath(projectRoot) {
  return join(projectRoot, ".akari", "reports", "render-report.html");
}

export async function readLintJson(projectRoot) {
  const text = await readFile(lintJsonPath(projectRoot), "utf8");
  return JSON.parse(text);
}

export async function readRenderJson(projectRoot) {
  const text = await readFile(renderJsonPath(projectRoot), "utf8");
  return JSON.parse(text);
}
