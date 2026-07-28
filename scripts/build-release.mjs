#!/usr/bin/env node
/**
 * Build a self-contained akari CLI distribution:
 *   dist/akari-<platform>/
 *     akari[.exe]
 *     akari-bundle/
 *
 * Usage:
 *   node scripts/build-release.mjs [--platform host|linux-x64|win-x64|osx-arm64|osx-x64] [--with-shell]
 */
import { spawnSync } from "node:child_process";
import { chmod, cp, mkdir, rm } from "node:fs/promises";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const repoRoot = join(dirname(fileURLToPath(import.meta.url)), "..");

/** @type {Record<string, { goos: string, goarch: string, ext: string, electron: string[] }>} */
const PLATFORMS = {
  "linux-x64": {
    goos: "linux",
    goarch: "amd64",
    ext: "",
    electron: ["--linux", "--x64"],
  },
  "win-x64": {
    goos: "windows",
    goarch: "amd64",
    ext: ".exe",
    electron: ["--win", "--x64"],
  },
  "osx-arm64": {
    goos: "darwin",
    goarch: "arm64",
    ext: "",
    electron: ["--mac", "--arm64"],
  },
  "osx-x64": {
    goos: "darwin",
    goarch: "amd64",
    ext: "",
    electron: ["--mac", "--x64"],
  },
};

function usage() {
  console.error(`Usage: node scripts/build-release.mjs [--platform <id>] [--with-shell]

Platforms: ${Object.keys(PLATFORMS).join(", ")}, host (default)
  --with-shell  Also package apps/shell with electron-builder (slow; mac requires macOS host)`);
}

function parseArgs(argv) {
  const args = [...argv];
  let platform = "host";
  let withShell = false;
  for (let i = 0; i < args.length; i += 1) {
    if (args[i] === "--platform" && args[i + 1]) {
      platform = args[i + 1];
      i += 1;
      continue;
    }
    if (args[i] === "--with-shell") {
      withShell = true;
      continue;
    }
    if (args[i] === "--help" || args[i] === "-h") {
      usage();
      process.exit(0);
    }
    console.error(`unknown argument: ${args[i]}`);
    usage();
    process.exit(2);
  }
  return { platform, withShell };
}

function detectHostPlatform() {
  if (process.platform === "linux") {
    return "linux-x64";
  }
  if (process.platform === "win32") {
    return "win-x64";
  }
  if (process.platform === "darwin") {
    return process.arch === "arm64" ? "osx-arm64" : "osx-x64";
  }
  throw new Error(`unsupported host platform: ${process.platform}`);
}

function run(command, args, options = {}) {
  const result = spawnSync(command, args, {
    cwd: options.cwd ?? repoRoot,
    stdio: "inherit",
    env: { ...process.env, ...options.env },
  });
  if (result.status !== 0) {
    process.exit(result.status ?? 1);
  }
}

async function bundleOrchestrator(bundleRoot) {
  run(process.execPath, ["scripts/bundle-orchestrator.mjs", "--out", bundleRoot]);
}

async function buildCliBinary(platformId, outPath) {
  const spec = PLATFORMS[platformId];
  run(
    "go",
    ["build", "-o", outPath, "./cmd/akari"],
    {
      cwd: join(repoRoot, "apps/cli"),
      env: {
        GOOS: spec.goos,
        GOARCH: spec.goarch,
        CGO_ENABLED: "0",
      },
    },
  );
  if (spec.goos !== "windows") {
    await chmod(outPath, 0o755);
  }
}

async function stageCliDistribution(platformId) {
  const spec = PLATFORMS[platformId];
  const distRoot = join(repoRoot, "dist", `akari-${platformId}`);
  const bundleRoot = join(distRoot, "akari-bundle");
  const cliPath = join(distRoot, `akari${spec.ext}`);

  await rm(distRoot, { recursive: true, force: true });
  await mkdir(distRoot, { recursive: true });

  await bundleOrchestrator(bundleRoot);
  await buildCliBinary(platformId, cliPath);

  console.log(`akari CLI distribution ready: ${distRoot}`);
  return distRoot;
}

async function buildShellPackage(platformId, cliDistRoot) {
  const spec = PLATFORMS[platformId];
  const shellDir = join(repoRoot, "apps/shell");
  const bundledCliRoot = join(shellDir, "resources", "bundled-cli");
  const shellOut = join(repoRoot, "dist", `akari-video-shell-${platformId}`);

  if (process.platform !== "darwin" && platformId.startsWith("osx")) {
    throw new Error(`electron-builder cannot package ${platformId} without a macOS host`);
  }
  if (process.platform === "linux" && platformId === "win-x64") {
    console.error("note: packaging win-x64 Electron from Linux may require wine (electron-builder)");
  }

  await rm(bundledCliRoot, { recursive: true, force: true });
  await mkdir(join(bundledCliRoot, "bin"), { recursive: true });
  await cp(join(cliDistRoot, `akari${spec.ext}`), join(bundledCliRoot, "bin", `akari${spec.ext}`));
  await cp(join(cliDistRoot, "akari-bundle"), join(bundledCliRoot, "bin", "akari-bundle"), {
    recursive: true,
  });

  const installEnv = process.platform === "darwin" ? { PYTHON: "/usr/bin/python3" } : {};
  run("npm", ["install", "--no-workspaces"], { cwd: shellDir, env: installEnv });
  run("npm", ["run", "build"], { cwd: shellDir });

  await rm(shellOut, { recursive: true, force: true });
  run(
    "npx",
    ["electron-builder", "--dir", ...spec.electron, `--config.directories.output=${shellOut}`],
    { cwd: shellDir },
  );

  console.log(`AKARI Video shell package ready: ${shellOut}`);
}

const { platform, withShell } = parseArgs(process.argv.slice(2));
const platformId = platform === "host" ? detectHostPlatform() : platform;
if (!PLATFORMS[platformId]) {
  console.error(`unknown platform: ${platformId}`);
  usage();
  process.exit(2);
}

const cliDistRoot = await stageCliDistribution(platformId);
if (withShell) {
  await buildShellPackage(platformId, cliDistRoot);
}
