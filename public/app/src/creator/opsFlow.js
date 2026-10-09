import { runOp } from "../api/ops.js";
import { showOpsToast } from "./opsToast.js";

export function opError(err, fallback = "The operation failed.") {
  return err?.response?.data?.error || err?.message || fallback;
}

export async function performOp(token, action, body) {
  const data = await runOp(token, action, { confirm: true, ...body });
  showOpsToast({
    summary: data.summary || "Done.",
    auditId: data.auditId,
    undoable: !!data.undoable,
    token,
  });
  return data;
}
