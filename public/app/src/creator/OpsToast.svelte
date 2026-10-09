<script>
  import { undoAudit } from "../api/audit.js";
  import { clearOpsToast, opsToast, showOpsToast } from "./opsToast.js";
  import { opError } from "./opsFlow.js";

  let busy = false;
  let error = "";

  $: toastId = $opsToast?.id;
  $: if (toastId) error = "";

  async function undo() {
    const toast = $opsToast;
    if (!toast?.auditId || !toast.token || busy) return;
    busy = true;
    error = "";
    try {
      const data = await undoAudit(toast.token, toast.auditId);
      showOpsToast({
        summary: data.summary || "Undone.",
        auditId: data.auditId,
        undoable: !!data.undoable,
        token: toast.token,
      });
    } catch (err) {
      error = opError(err, "Undo failed.");
    } finally {
      busy = false;
    }
  }
</script>

{#if $opsToast}
  <div class="ops-toast" role="status">
    <p>{$opsToast.summary}</p>
    {#if error}
      <p class="ops-toast-error">{error}</p>
    {/if}
    <div class="ops-toast-actions">
      {#if $opsToast.undoable && $opsToast.auditId}
        <button type="button" on:click={undo} disabled={busy}>{busy ? "Undoing…" : "Undo"}</button>
      {/if}
      <button type="button" class="ops-dismiss" on:click={clearOpsToast}>Dismiss</button>
    </div>
  </div>
{/if}

<style>
  .ops-toast {
    position: fixed;
    right: 16px;
    bottom: 16px;
    z-index: 240;
    width: min(420px, calc(100vw - 32px));
    padding: 14px 16px;
    border-radius: 10px;
    border: 1px solid #92400e;
    background: #1c1410;
    color: #ffedd5;
    box-shadow: 0 16px 40px rgba(0, 0, 0, 0.45);
  }

  p {
    margin: 0;
    font-size: 14px;
    line-height: 1.4;
  }

  .ops-toast-error {
    margin-top: 8px;
    color: #fecaca;
    font-size: 12px;
  }

  .ops-toast-actions {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    margin-top: 12px;
  }

  button {
    border-radius: 6px;
    padding: 6px 12px;
    font-size: 12px;
    font-weight: 700;
    cursor: pointer;
    background: #d97706;
    color: #1c1410;
    border: 0;
  }

  button:disabled {
    opacity: 0.6;
    cursor: default;
  }

  .ops-dismiss {
    background: transparent;
    color: #fdba74;
    border: 1px solid #9a3412;
  }
</style>
