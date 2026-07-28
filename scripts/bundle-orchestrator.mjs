#!/usr/bin/env node
/**
 * Copy orchestrator runtime (orchestrator + edit-lint + render-cut) into bin/akari-bundle/
 * for distribution alongside the akari Go binary.
 */
import { cp, mkdir, rm, writeFile } from "node:fs/promises";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { spawnSync } from "node:child_process";

const repoRoot = join(dirname(fileURLToPath(import.meta.url)), "..");
const bundleRoot = join(repoRoot, "bin", "akari-bundle");
const packages = ["orchestrator", "edit-lint", "render-cut"];

function shouldSkip(path) {
  return (
    path.includes(`${join("packages", "orchestrator", "test")}`) ||
    path.includes(`${join("packages", "edit-lint", "test")}`) ||
    path.includes(`${join("packages", "render-cut", "test")}`) ||
    path.includes(`${join("packages", "render-cut", "evidence")}`) ||
    path.includes("node_modules")
  );
}

await rm(bundleRoot, { recursive: true, force: true });
await mkdir(bundleRoot, { recursive: true });

for (const name of packages) {
  const source = join(repoRoot, "packages", name);
  const target = join(bundleRoot, "packages", name);
  await cp(source, target, {
    recursive: true,
    filter: (src) => !shouldSkip(src),
  });
}

await writeFile(
  join(bundleRoot, "package.json"),
  `${JSON.stringify(
    {
      name: "akari-bundle",
      private: true,
      workspaces: ["packages/*"],
    },
    null,
    2,
  )}\n`,
);

const install = spawnSync(
  "npm",
  ["install", "--omit=dev", "--ignore-scripts", "--no-audit", "--no-fund", "-w", "@akari-video/orchestrator"],
  { cwd: bundleRoot, stdio: "inherit" },
);
if (install.status !== 0) {
  process.exit(install.status ?? 1);
}

console.log(`akari-bundle ready: ${bundleRoot}`);
