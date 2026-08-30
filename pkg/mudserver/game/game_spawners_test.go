package game

import (
	"testing"
	"time"
)

func TestSpawnerRefillsInitialCountWithoutWaitingInterval(t *testing.T) {
	last := time.Now()
	if !spawnerShouldSpawn(0, 1, 3, 5*time.Minute, last, last.Add(time.Second)) {
		t.Fatal("expected immediate refill below initialCount")
	}
	if spawnerShouldSpawn(1, 1, 3, 5*time.Minute, last, last.Add(time.Second)) {
		t.Fatal("should wait spawnInterval once initialCount is met")
	}
	if !spawnerShouldSpawn(1, 1, 3, 5*time.Minute, last, last.Add(5*time.Minute)) {
		t.Fatal("expected extra spawn after interval")
	}
	if spawnerShouldSpawn(3, 1, 3, time.Second, last, last.Add(time.Hour)) {
		t.Fatal("should not exceed maxInstances")
	}
}
