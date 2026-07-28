#!/usr/bin/env node

import { runServeCli } from "../src/serve-cli.mjs";

const USAGE = `Usage: akari-orchestrator serve [--addr <host:port>]
       akari-orchestrator version`;

const command = process.argv[2];
if (command === "serve") {
  process.exitCode = await runServeCli(process.argv.slice(3));
} else if (command === "version") {
  console.log("akari-orchestrator 0.1.0");
} else {
  console.error(command ? `unknown command: ${command}` : "command required");
  console.error(USAGE);
  process.exitCode = 2;
}
