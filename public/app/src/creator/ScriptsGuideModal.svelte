<script>
  import { createEventDispatcher } from "svelte";

  export let open = false;

  const dispatch = createEventDispatcher();

  const close = () => dispatch("close");

  const handleKey = (event) => {
    if (!open) return;
    if (event.key === "Escape") {
      close();
    }
  };
</script>

<svelte:window on:keydown={handleKey} />
{#if open}
  <div
    class="fixed inset-0 z-50 flex items-start justify-center bg-black/60 backdrop-blur-sm p-6"
    on:click|self={close}
  >
    <div class="w-full max-w-4xl bg-slate-950 text-slate-100 border border-slate-800 rounded-2xl shadow-2xl overflow-hidden">
      <div class="flex items-center justify-between px-6 py-4 border-b border-slate-800">
        <div class="space-y-1">
          <h2 class="text-lg font-semibold">Lua Scripting Guide</h2>
          <p class="text-xs text-slate-400">
            How scripts run, what data you get, and the available APIs.
          </p>
        </div>
        <button class="btn btn-ghost" type="button" on:click={close}>
          <span class="material-symbols-outlined text-sm">close</span>
          Close
        </button>
      </div>

      <div class="max-h-[70vh] overflow-y-auto px-6 py-5 space-y-6 text-sm text-slate-300">
        <section class="space-y-2">
          <h3 class="text-sm font-semibold text-slate-100">Execution model</h3>
          <p>
            Scripts run on the server using the embedded Lua runtime. Each run has a
            sandboxed environment and a 5-second execution time limit.
          </p>
        </section>

        <section class="space-y-2">
          <h3 class="text-sm font-semibold text-slate-100">Context data</h3>
          <p>
            The global <code class="text-slate-200">ctx</code> table is always present.
            The actual payload is stored under <code class="text-slate-200">ctx.ctx</code>.
            In the test runner, <code class="text-slate-200">ctx.ctx</code> is the JSON you submit.
            When triggered by gameplay, <code class="text-slate-200">ctx.ctx</code> may be an item,
            room, character, or other object.
          </p>
          <pre class="bg-black/40 rounded p-3 text-xs text-slate-200 border border-slate-800">local input = ctx.ctx
if input then
  tales.game.log("info", "Got input from test runner")
end</pre>
        </section>

        <section class="space-y-2">
          <h3 class="text-sm font-semibold text-slate-100">Return values</h3>
          <p>
            Prefer returning a value with <code class="text-slate-200">return</code>. If you
            do not return anything, the runtime will return the entire
            <code class="text-slate-200">ctx</code> table instead.
          </p>
          <pre class="bg-black/40 rounded p-3 text-xs text-slate-200 border border-slate-800">local data = ctx.ctx or &#123;&#125;
data.roll = tales.utils.roll("1d20+3")
return data</pre>
        </section>

        <section class="space-y-3">
          <h3 class="text-sm font-semibold text-slate-100">Available APIs</h3>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-3 text-xs">
            <div class="space-y-1">
              <div class="text-slate-200 font-semibold">tales.items</div>
              <div>get(id)</div>
              <div>findByName(name)</div>
              <div>getTemplate(id)</div>
              <div>findTemplates(name)</div>
              <div>createFromTemplate(templateId)</div>
              <div>delete(id)</div>
            </div>
            <div class="space-y-1">
              <div class="text-slate-200 font-semibold">tales.rooms</div>
              <div>get(id)</div>
              <div>findByName(name)</div>
              <div>findByArea(area)</div>
              <div>getAll()</div>
              <div>getCharacters(roomId)</div>
              <div>getNPCs(roomId)</div>
              <div>getItems(roomId)</div>
            </div>
            <div class="space-y-1">
              <div class="text-slate-200 font-semibold">tales.characters</div>
              <div>get(id)</div>
              <div>findByName(name)</div>
              <div>getAll()</div>
              <div>getRoom(characterId)</div>
              <div>damage(id, amount)</div>
              <div>heal(id, amount)</div>
              <div>teleport(id, roomId)</div>
              <div>giveXP(id, amount)</div>
            </div>
            <div class="space-y-1">
              <div class="text-slate-200 font-semibold">tales.npcs</div>
              <div>get(id)</div>
              <div>findByName(name)</div>
              <div>findInRoom(roomId)</div>
              <div>getAll()</div>
              <div>damage(id, amount)</div>
              <div>heal(id, amount)</div>
              <div>moveTo(id, roomId)</div>
              <div>isDead(id)</div>
              <div>isEnemy(id)</div>
              <div>isMerchant(id)</div>
              <div>delete(id)</div>
            </div>
            <div class="space-y-1">
              <div class="text-slate-200 font-semibold">tales.dialogs</div>
              <div>get(id)</div>
              <div>findByName(name)</div>
              <div>getAll()</div>
              <div>getConversation(characterId, targetId)</div>
              <div>setContext(conversationId, key, value)</div>
              <div>getContext(conversationId, key)</div>
              <div>hasVisited(conversationId, nodeId)</div>
              <div>getVisitCount(conversationId, nodeId)</div>
            </div>
            <div class="space-y-1">
              <div class="text-slate-200 font-semibold">tales.game</div>
              <div>msgToRoom(roomId, message)</div>
              <div>msgToRoomExcept(roomId, message, excludeId)</div>
              <div>msgToCharacter(characterId, message)</div>
              <div>msgToUser(userId, message)</div>
              <div>broadcast(message)</div>
              <div>log(level, message)</div>
            </div>
            <div class="space-y-1">
              <div class="text-slate-200 font-semibold">tales.utils</div>
              <div>random(min, max)</div>
              <div>randomFloat()</div>
              <div>uuid()</div>
              <div>now()</div>
              <div>nowMs()</div>
              <div>formatTime(timestamp)</div>
              <div>roll(dice)</div>
              <div>chance(percentage)</div>
              <div>pick(array)</div>
              <div>shuffle(array)</div>
              <div>clamp(value, min, max)</div>
              <div>lerp(a, b, t)</div>
            </div>
          </div>
        </section>

        <section class="space-y-2">
          <h3 class="text-sm font-semibold text-slate-100">Examples</h3>
          <pre class="bg-black/40 rounded p-3 text-xs text-slate-200 border border-slate-800">-- 1) Announce to a room
local input = ctx.ctx or &#123;&#125;
if input.roomId then
  -- NOTE: roomId must be the room's stored ID (copy it from Creator > Rooms), not the background name.
  return tales.game.msgToRoom(input.roomId, "The air crackles with energy.")
end

-- 2) Roll a die and return the result
local roll = tales.utils.roll("2d6+1")
return &#123; roll = roll &#125;

-- 3) Heal an NPC
if input.npcId then
  tales.npcs.heal(input.npcId, 10)
end</pre>
        </section>

        <section class="space-y-2">
          <h3 class="text-sm font-semibold text-slate-100">Sandbox limits</h3>
          <p>
            Dangerous globals are removed (os, io, debug, loadfile, dofile, load,
            loadstring, package). The Lua runtime enforces a 5-second execution timeout.
          </p>
        </section>
      </div>
    </div>
  </div>
{/if}
