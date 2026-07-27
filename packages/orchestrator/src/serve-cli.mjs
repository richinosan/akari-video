import { createServer } from "node:http";

import { createConnectRouter } from "@connectrpc/connect";
import { connectNodeAdapter } from "@connectrpc/connect-node";

import { OrchestratorService } from "../gen/akari/v1/orchestrator_connect.js";
import { createOrchestratorService } from "./service.mjs";

const USAGE = `Usage: akari-orchestrator serve [--addr <host:port>]

Start ConnectRPC OrchestratorService (Node worker). Default: 127.0.0.1:7707`;

export async function runServeCli(argv) {
  const args = [...argv];
  let addr = "127.0.0.1:7707";
  for (let i = 0; i < args.length; i += 1) {
    if (args[i] === "--addr" && args[i + 1]) {
      addr = args[i + 1];
      i += 1;
      continue;
    }
    if (args[i] === "--help" || args[i] === "-h") {
      console.log(USAGE);
      return 0;
    }
    console.error(`unknown flag: ${args[i]}`);
    console.error(USAGE);
    return 2;
  }

  const routes = createConnectRouter().service(
    OrchestratorService,
    createOrchestratorService(),
  );
  const server = createServer(connectNodeAdapter({ routes }));
  await new Promise((resolve, reject) => {
    server.listen(addr, () => {
      console.error(`akari-orchestrator listening on ${addr}`);
      resolve();
    });
    server.on("error", reject);
  });

  const shutdown = () => {
    server.close(() => process.exit(0));
  };
  process.on("SIGINT", shutdown);
  process.on("SIGTERM", shutdown);
  return 0;
}
