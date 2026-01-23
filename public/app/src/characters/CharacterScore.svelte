<script>
  export let name;
  export let level;
  export let cclass;
  export let xp;

  function avatar() {
    let num = 1 + Math.abs(name.hashCode() % 12);
    return "/img/avatars/" + num + "p.png";
  }
</script>

<div class="card p-4 text-center">
  <img
    src={avatar()}
    alt=""
    class="w-16 h-16 rounded-full mx-auto -mt-10 border border-slate-700 shadow-lg"
  />
  <div class="mt-3 text-sm font-semibold">{name}</div>
  <div class="text-xs text-slate-400">
    LVL {level} · {cclass?.name || cclass}
  </div>
  <div class="text-xs text-primary mt-1">{xp} XP</div>
</div>
