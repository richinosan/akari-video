import { ExitClass } from "../gen/akari/v1/orchestrator_pb.js";

export function exitClassFromCode(code) {
  if (code === 0) {
    return ExitClass.OK;
  }
  if (code === 1) {
    return ExitClass.REFUSAL;
  }
  return ExitClass.EXECUTION_ERROR;
}
