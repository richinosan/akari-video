import { createServer } from "node:http";

import { connectNodeAdapter } from "@connectrpc/connect-node";

import { OrchestratorService } from "../gen/akari/v1/orchestrator_pb.js";
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

  const { host, port } = parseListenAddr(addr);

  const connectHandler = connectNodeAdapter({
    routes: (router) => {
      router.service(OrchestratorService, createOrchestratorService());
    },
  });
  const server = createServer((req, res) => {
    if (req.url === "/healthz") {
      res.writeHead(200, { "content-type": "text/plain" });
      res.end("ok");
      return;
    }
    connectHandler(req, res);
  });
  await new Promise((resolve, reject) => {
    server.listen({ host, port }, () => {
      console.error(`akari-orchestrator listening on ${host}:${port}`);
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

function parseListenAddr(addr) {
  if (addr.startsWith("[")) {
    const end = addr.indexOf("]");
    if (end === -1) {
      throw new Error(`invalid listen addr: ${addr}`);
    }
    const host = addr.slice(1, end);
    const port = Number(addr.slice(end + 2));
    return { host, port };
  }
  const sep = addr.lastIndexOf(":");
  if (sep === -1) {
    throw new Error(`invalid listen addr: ${addr}`);
  }
  return {
    host: addr.slice(0, sep),
    port: Number(addr.slice(sep + 1)),
  };
}
