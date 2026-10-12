package game

import (
	"testing"
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
)

func TestSpawnerRespawnDelayed(t *testing.T) {
	now := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	if spawnerRespawnDelayed(0, now.Add(-time.Second), now) {
		t.Fatal("no delay configured must never block")
	}
	if spawnerRespawnDelayed(30*time.Minute, time.Time{}, now) {
		t.Fatal("a spawner that never lost an instance must not be blocked")
	}
	if !spawnerRespawnDelayed(30*time.Minute, now.Add(-29*time.Minute), now) {
		t.Fatal("inside the delay window the boss must stay gone")
	}
	if spawnerRespawnDelayed(30*time.Minute, now.Add(-30*time.Minute), now) {
		t.Fatal("after the delay the boss must come back")
	}
}

func newTestManagerWithInstance(spawnerID string, inst *npc.NPC) *NPCInstanceManager {
	m := NewNPCInstanceManager(nil)
	m.instances[inst.Entity.ID] = inst
	m.spawnerState[spawnerID] = &SpawnerState{ActiveInstances: []string{inst.Entity.ID}, LastSpawnTime: time.Now()}
	return m
}

func TestSpawnerRecordsDeathOnRemove(t *testing.T) {
	died := time.Now().Add(-time.Minute).Truncate(time.Second)
	inst := &npc.NPC{Entity: &entities.Entity{ID: "boss-1"}, IsDead: true, DeathTime: died}
	m := newTestManagerWithInstance("SPW_BOSS", inst)
	m.RemoveInstance("boss-1")
	st := m.spawnerState["SPW_BOSS"]
	if !st.LastDeathTime.Equal(died) {
		t.Fatalf("LastDeathTime = %v, want %v", st.LastDeathTime, died)
	}
	if len(st.ActiveInstances) != 0 {
		t.Fatal("dead instance should leave the active list")
	}
}

func TestSpawnerRecordsDeathOnCleanup(t *testing.T) {
	died := time.Now().Add(-2 * time.Minute).Truncate(time.Second)
	inst := &npc.NPC{Entity: &entities.Entity{ID: "boss-2"}, IsDead: true, DeathTime: died}
	m := newTestManagerWithInstance("SPW_BOSS", inst)
	if removed := m.CleanupDeadFromSpawner("SPW_BOSS"); removed != 1 {
		t.Fatalf("removed = %d", removed)
	}
	if !m.spawnerState["SPW_BOSS"].LastDeathTime.Equal(died) {
		t.Fatal("cleanup must stamp the death time")
	}
}

func TestSpawnerRemoveAliveDoesNotStampDeath(t *testing.T) {
	inst := &npc.NPC{Entity: &entities.Entity{ID: "add-1"}}
	m := newTestManagerWithInstance("SPW_X", inst)
	m.RemoveInstance("add-1")
	if !m.spawnerState["SPW_X"].LastDeathTime.IsZero() {
		t.Fatal("removing a living instance is not a death")
	}
}

func TestBossNeverDoubleSpawnsWhileAliveOrInCombat(t *testing.T) {
	inst := &npc.NPC{Entity: &entities.Entity{ID: "boss-3"}, InCombat: true, State: "combat"}
	m := newTestManagerWithInstance("SPW_BOSS", inst)
	alive := m.CountAliveForSpawner("SPW_BOSS")
	if alive != 1 {
		t.Fatalf("alive = %d", alive)
	}
	if spawnerShouldSpawn(alive, 1, 1, 30*time.Minute, time.Now().Add(-time.Hour), time.Now()) {
		t.Fatal("a boss in combat must not be spawned again")
	}
}

func boolPtr(b bool) *bool { return &b }

func TestResetNPCAfterDisengage(t *testing.T) {
	boss := &npc.NPC{Entity: &entities.Entity{ID: "b"}, MaxHitPoints: 150, CurrentHitPoints: 20,
		EnemyTrait: &npc.EnemyTrait{Difficulty: "boss"}}
	if !resetNPCAfterDisengage(boss) || boss.CurrentHitPoints != 150 {
		t.Fatalf("boss should reset to full, got %d", boss.CurrentHitPoints)
	}
	trash := &npc.NPC{Entity: &entities.Entity{ID: "t"}, MaxHitPoints: 30, CurrentHitPoints: 5,
		EnemyTrait: &npc.EnemyTrait{Difficulty: "normal"}}
	if resetNPCAfterDisengage(trash) || trash.CurrentHitPoints != 5 {
		t.Fatal("normal NPCs keep their damage by default")
	}
	optIn := &npc.NPC{Entity: &entities.Entity{ID: "o"}, MaxHitPoints: 90, CurrentHitPoints: 10,
		EnemyTrait: &npc.EnemyTrait{Difficulty: "hard", ResetOnDisengage: boolPtr(true)}}
	if !resetNPCAfterDisengage(optIn) || optIn.CurrentHitPoints != 90 {
		t.Fatal("content opt-in must reset")
	}
	optOut := &npc.NPC{Entity: &entities.Entity{ID: "x"}, MaxHitPoints: 90, CurrentHitPoints: 10,
		EnemyTrait: &npc.EnemyTrait{Difficulty: "boss", ResetOnDisengage: boolPtr(false)}}
	if resetNPCAfterDisengage(optOut) || optOut.CurrentHitPoints != 10 {
		t.Fatal("content opt-out must keep HP")
	}
	dead := &npc.NPC{Entity: &entities.Entity{ID: "d"}, MaxHitPoints: 90, IsDead: true,
		EnemyTrait: &npc.EnemyTrait{Difficulty: "boss"}}
	if resetNPCAfterDisengage(dead) {
		t.Fatal("a dead boss is not reset")
	}
}
