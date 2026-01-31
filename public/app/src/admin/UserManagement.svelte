<script>
  import { getAuth } from "../auth.js";
  import {
    getAllUsers,
    updateUserRole,
    banUser,
    unbanUser,
    deleteUser,
  } from "../api/admin.js";

  const { isAuthenticated, authToken } = getAuth();

  let users = [];
  let loading = true;
  let error = null;

  // Ban modal state
  let banModalOpen = false;
  let banConfirmStep = 1; // 1 = first confirm, 2 = double confirm
  let targetUser = null;

  function loadUsers() {
    loading = true;
    error = null;
    getAllUsers(
      $authToken,
      (data) => {
        users = data;
        loading = false;
      },
      (err) => {
        console.error("Failed to load users:", err);
        error = "Failed to load users";
        loading = false;
      }
    );
  }

  $: if ($isAuthenticated && $authToken) {
    loadUsers();
  }

  function getRoleLabel(user) {
    const role = user.role || "player";
    switch (role) {
      case "admin":
        return "MUD Admin";
      case "creator":
        return "MUD Creator";
      default:
        return "Player";
    }
  }

  function getRoleBadgeClass(user) {
    const role = user.role || "player";
    switch (role) {
      case "admin":
        return "bg-amber-500/20 text-amber-400 border-amber-500/30";
      case "creator":
        return "bg-emerald-500/20 text-emerald-400 border-emerald-500/30";
      default:
        return "bg-slate-500/20 text-slate-400 border-slate-500/30";
    }
  }

  function handlePromote(user) {
    updateUserRole(
      $authToken,
      user.id,
      "creator",
      () => loadUsers(),
      (err) => console.error("Failed to promote user:", err)
    );
  }

  function handleDemote(user) {
    updateUserRole(
      $authToken,
      user.id,
      "player",
      () => loadUsers(),
      (err) => console.error("Failed to demote user:", err)
    );
  }

  function openBanModal(user) {
    targetUser = user;
    banConfirmStep = 1;
    banModalOpen = true;
  }

  function advanceBanConfirm() {
    banConfirmStep = 2;
  }

  function executeBan() {
    banUser(
      $authToken,
      targetUser.id,
      () => {
        closeBanModal();
        loadUsers();
      },
      (err) => console.error("Failed to ban user:", err)
    );
  }

  function closeBanModal() {
    banModalOpen = false;
    banConfirmStep = 1;
    targetUser = null;
  }

  function handleUnban(user) {
    unbanUser(
      $authToken,
      user.id,
      () => loadUsers(),
      (err) => console.error("Failed to unban user:", err)
    );
  }

  function handleDelete(user) {
    deleteUser(
      $authToken,
      user.id,
      () => loadUsers(),
      (err) => console.error("Failed to delete user:", err)
    );
  }

  function isOrphanUser(user) {
    return !user.name && !user.email && !user.nickname;
  }
</script>

<div class="px-6 py-8">
  <div class="max-w-6xl mx-auto space-y-6">
    <div>
      <h1 class="text-3xl font-bold tracking-tight">User Management</h1>
      <p class="text-sm text-slate-500 dark:text-slate-400">
        Manage player accounts, roles, and access levels.
      </p>
    </div>

    {#if loading}
      <div class="card p-12 text-center">
        <p class="text-slate-400">Loading users...</p>
      </div>
    {:else if error}
      <div class="card p-12 text-center">
        <p class="text-red-400">{error}</p>
        <button class="btn btn-outline mt-4" on:click={loadUsers}>Retry</button>
      </div>
    {:else}
      <div class="card overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-slate-200 dark:border-slate-800">
                <th
                  class="px-4 py-3 text-left text-[10px] font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500"
                  >ID</th
                >
                <th
                  class="px-4 py-3 text-left text-[10px] font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500"
                  >Name</th
                >
                <th
                  class="px-4 py-3 text-left text-[10px] font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500"
                  >Nickname</th
                >
                <th
                  class="px-4 py-3 text-left text-[10px] font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500"
                  >Email</th
                >
                <th
                  class="px-4 py-3 text-left text-[10px] font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500"
                  >Access Level</th
                >
                <th
                  class="px-4 py-3 text-left text-[10px] font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500"
                  >Status</th
                >
                <th
                  class="px-4 py-3 text-right text-[10px] font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500"
                  >Actions</th
                >
              </tr>
            </thead>
            <tbody>
              {#each users as user (user.id)}
                <tr
                  class="border-b border-slate-100 dark:border-slate-800/50 hover:bg-slate-50 dark:hover:bg-slate-800/30 transition-colors"
                >
                  <td class="px-4 py-3 font-mono text-xs text-slate-400"
                    >{user.id?.slice(0, 8)}...</td
                  >
                  <td class="px-4 py-3">{user.name || "-"}</td>
                  <td class="px-4 py-3 text-slate-400"
                    >{user.nickname || "-"}</td
                  >
                  <td class="px-4 py-3 text-slate-400"
                    >{user.email || "-"}</td
                  >
                  <td class="px-4 py-3">
                    <span
                      class="inline-block rounded-full border px-2.5 py-0.5 text-[10px] font-bold uppercase tracking-wider {getRoleBadgeClass(
                        user
                      )}"
                    >
                      {getRoleLabel(user)}
                    </span>
                  </td>
                  <td class="px-4 py-3">
                    {#if user.isBanned}
                      <span
                        class="inline-block rounded-full border border-red-500/30 bg-red-500/20 px-2.5 py-0.5 text-[10px] font-bold uppercase tracking-wider text-red-400"
                      >
                        Banned
                      </span>
                    {:else}
                      <span
                        class="inline-block rounded-full border border-emerald-500/30 bg-emerald-500/20 px-2.5 py-0.5 text-[10px] font-bold uppercase tracking-wider text-emerald-400"
                      >
                        Active
                      </span>
                    {/if}
                  </td>
                  <td class="px-4 py-3">
                    <div class="flex items-center justify-end gap-2">
                      {#if (user.role || "player") === "admin"}
                        <span class="text-xs text-slate-500 italic"
                          >Protected</span
                        >
                      {:else if isOrphanUser(user)}
                        <button
                          class="btn btn-danger text-xs px-3 py-1"
                          on:click={() => handleDelete(user)}
                        >
                          <span class="material-symbols-outlined text-sm"
                            >delete</span
                          >
                          Delete
                        </button>
                      {:else if user.isBanned}
                        <button
                          class="btn btn-outline text-xs px-3 py-1"
                          on:click={() => handleUnban(user)}
                        >
                          <span class="material-symbols-outlined text-sm"
                            >lock_open</span
                          >
                          Unban
                        </button>
                        <button
                          class="btn btn-ghost text-xs px-2 py-1"
                          title="Delete user"
                          on:click={() => handleDelete(user)}
                        >
                          <span class="material-symbols-outlined text-sm"
                            >delete</span
                          >
                        </button>
                      {:else}
                        {#if (user.role || "player") === "player"}
                          <button
                            class="btn btn-outline text-xs px-3 py-1"
                            on:click={() => handlePromote(user)}
                          >
                            <span class="material-symbols-outlined text-sm"
                              >arrow_upward</span
                            >
                            Promote to Creator
                          </button>
                        {:else if (user.role || "player") === "creator"}
                          <button
                            class="btn btn-outline text-xs px-3 py-1"
                            on:click={() => handleDemote(user)}
                          >
                            <span class="material-symbols-outlined text-sm"
                              >arrow_downward</span
                            >
                            Demote to Player
                          </button>
                        {/if}
                        <button
                          class="btn btn-danger text-xs px-3 py-1"
                          on:click={() => openBanModal(user)}
                        >
                          <span class="material-symbols-outlined text-sm"
                            >block</span
                          >
                          Ban
                        </button>
                        <button
                          class="btn btn-ghost text-xs px-2 py-1"
                          title="Delete user"
                          on:click={() => handleDelete(user)}
                        >
                          <span class="material-symbols-outlined text-sm"
                            >delete</span
                          >
                        </button>
                      {/if}
                    </div>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>

        {#if users.length === 0}
          <div class="p-12 text-center">
            <p class="text-slate-400">No users found.</p>
          </div>
        {/if}
      </div>

      <div class="text-xs text-slate-500 dark:text-slate-500">
        Total users: {users.length}
      </div>
    {/if}
  </div>
</div>

<!-- Ban Confirmation Modal -->
{#if banModalOpen && targetUser}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <div
    class="fixed inset-0 z-[100] flex items-center justify-center bg-black/60 backdrop-blur-sm"
    on:click={closeBanModal}
  >
    <!-- svelte-ignore a11y-click-events-have-key-events -->
    <div
      class="card mx-4 w-full max-w-md p-6 space-y-4"
      on:click|stopPropagation
    >
      {#if banConfirmStep === 1}
        <div class="flex items-center gap-3">
          <div
            class="flex h-10 w-10 items-center justify-center rounded-full bg-red-500/20"
          >
            <span class="material-symbols-outlined text-red-400">warning</span>
          </div>
          <h2 class="text-lg font-bold">Ban Player</h2>
        </div>
        <p class="text-sm text-slate-400">
          Do you really want to ban player <strong class="text-slate-200"
            >{targetUser.name || targetUser.nickname || "Unknown"}</strong
          >?
        </p>
        <div class="flex justify-end gap-3">
          <button class="btn btn-outline" on:click={closeBanModal}>
            Cancel
          </button>
          <button class="btn btn-danger" on:click={advanceBanConfirm}>
            Yes, ban this player
          </button>
        </div>
      {:else}
        <div class="flex items-center gap-3">
          <div
            class="flex h-10 w-10 items-center justify-center rounded-full bg-red-500/20"
          >
            <span class="material-symbols-outlined text-red-400">gavel</span>
          </div>
          <h2 class="text-lg font-bold">Confirm Ban</h2>
        </div>
        <p class="text-sm text-slate-400">
          This will permanently ban <strong class="text-slate-200"
            >{targetUser.name || targetUser.nickname || "Unknown"}</strong
          >
          by their Reference ID and Email address. They will not be able to access
          the game.
        </p>
        <div
          class="rounded-lg border border-red-500/20 bg-red-500/5 p-3 text-xs space-y-1"
        >
          <div>
            <span class="text-slate-500">Ref ID:</span>
            <span class="font-mono text-slate-300"
              >{targetUser.refid || "-"}</span
            >
          </div>
          <div>
            <span class="text-slate-500">Email:</span>
            <span class="text-slate-300">{targetUser.email || "-"}</span>
          </div>
        </div>
        <div class="flex justify-end gap-3">
          <button class="btn btn-outline" on:click={closeBanModal}>
            Cancel
          </button>
          <button class="btn btn-danger" on:click={executeBan}>
            Confirm Ban
          </button>
        </div>
      {/if}
    </div>
  </div>
{/if}
