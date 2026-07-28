#!/usr/bin/env node
/**
 * Build unified akari distribution:
 *   dist/akari-<platform>/
 *     akari[.exe]           # entry point: app launcher + `akari cli` headless commands
 *     lib/
 *       akari-bundle/       # Node orchestrator runtime
 *       app/                # Electron app (linux-unpacked / win-unpacked / .app contents)
 *
 * Usage:
 *   node scripts/build-release.mjs [--platform host|linux-x64|win-x64|osx-arm64|osx-x64]
 */
import { spawnSync } from "node:child_process";
import { chmod, cp, mkdir, readdir, rm } from "node:fs/promises";
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
  console.error(`Usage: node scripts/build-release.mjs [--platform <id>]

Platforms: ${Object.keys(PLATFORMS).join(", ")}, host (default)`);
}

function parseArgs(argv) {
  const args = [...argv];
  let platform = "host";
  for (let i = 0; i < args.length; i += 1) {
    if (args[i] === "--platform" && args[i + 1]) {
      platform = args[i + 1];
      i += 1;
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
  return { platform };
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

function canBuildShellOnHost(platformId) {
  if (platformId.startsWith("osx")) {
    return process.platform === "darwin";
  }
  return true;
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

async function findUnpackedDir(stagingRoot, platformId) {
  const entries = await readdir(stagingRoot, { withFileTypes: true });
  if (platformId.startsWith("osx")) {
    for (const entry of entries) {
      const entryPath = join(stagingRoot, entry.name);
      if (entry.isDirectory() && entry.name.endsWith(".app")) {
        return entryPath;
      }
      if (!entry.isDirectory()) {
        continue;
      }
      const children = await readdir(entryPath, { withFileTypes: true }).catch(() => []);
      for (const child of children) {
        if (child.isDirectory() && child.name.endsWith(".app")) {
          return join(entryPath, child.name);
        }
      }
    }
    throw new Error(`macOS .app not found under ${stagingRoot}`);
  }

  const prefix = platformId.startsWith("win") ? "win" : "linux";
  for (const entry of entries) {
    if (entry.isDirectory() && entry.name.startsWith(`${prefix}-`) && entry.name.endsWith("-unpacked")) {
      return join(stagingRoot, entry.name);
    }
  }
  throw new Error(`electron unpacked dir not found under ${stagingRoot}`);
}

async function buildShellInto(platformId, appRoot) {
  const spec = PLATFORMS[platformId];
  const shellDir = join(repoRoot, "apps/shell");
  const shellStaging = join(repoRoot, "dist", `.electron-staging-${platformId}`);

  if (process.platform === "linux" && platformId === "win-x64") {
    console.error("note: packaging win-x64 Electron from Linux may require wine (electron-builder)");
  }

  const installEnv = process.platform === "darwin" ? { PYTHON: "/usr/bin/python3" } : {};
  run("npm", ["install", "--no-workspaces"], { cwd: shellDir, env: installEnv });
  run("npm", ["run", "build"], { cwd: shellDir });

  await rm(shellStaging, { recursive: true, force: true });
  run(
    "npx",
    ["electron-builder", "--dir", ...spec.electron, `--config.directories.output=${shellStaging}`],
    { cwd: shellDir },
  );

  const unpacked = await findUnpackedDir(shellStaging, platformId);
  await rm(appRoot, { recursive: true, force: true });
  await cp(unpacked, appRoot, { recursive: true });
  await rm(shellStaging, { recursive: true, force: true });
}

async function stageDistribution(platformId) {
  const spec = PLATFORMS[platformId];
  const distRoot = join(repoRoot, "dist", `akari-${platformId}`);
  const libRoot = join(distRoot, "lib");
  const bundleRoot = join(libRoot, "akari-bundle");
  const appRoot = join(libRoot, "app");

  await rm(distRoot, { recursive: true, force: true });
  await mkdir(libRoot, { recursive: true });

  await bundleOrchestrator(bundleRoot);

  if (canBuildShellOnHost(platformId)) {
    await buildShellInto(platformId, appRoot);
  } else {
    console.error(`note: skipping Electron app for ${platformId} on this host (CLI-only dist)`);
  }

  await buildCliBinary(platformId, join(distRoot, `akari${spec.ext}`));

  console.log(`akari distribution ready: ${distRoot}`);
  if (canBuildShellOnHost(platformId)) {
    console.log("  akari                 # launch desktop app");
    console.log("  akari cli version     # headless commands");
  } else {
    console.log("  akari cli ...         # headless commands only (app not packaged on this host)");
  }
}

const { platform } = parseArgs(process.argv.slice(2));
const platformId = platform === "host" ? detectHostPlatform() : platform;
if (!PLATFORMS[platformId]) {
  console.error(`unknown platform: ${platformId}`);
  usage();
  process.exit(2);
}

await stageDistribution(platformId);
