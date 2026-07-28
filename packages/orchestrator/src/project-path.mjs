import { resolve } from "node:path";

export function absProjectRoot(projectRoot) {
  if (!projectRoot) {
    throw new Error("project_root is required");
  }
  return resolve(projectRoot);
}
