package shard

import (
	"os"
	"testing"
	"time"
)

type ComplexUser struct {
	ID     int
	Name   string
	Active bool
}

func TestAOF_GenericTypes(t *testing.T) {
	aofPath := "test_generic.aof"
	defer os.Remove(aofPath) // Clean up after test

	// 1. Create a cache for structs
	mgr := NewCacheManager[string, ComplexUser](4, 100, 3, aofPath)

	user := ComplexUser{ID: 1, Name: "Bruce Wayne", Active: true}
	mgr.Set("user_1", user, 1*time.Hour)

	// Manually flush to disk
	mgr.writer.Flush()
	mgr.aof.Sync()
	mgr.aof.Close()

	// 2. Create a NEW manager to simulate restart
	newMgr := NewCacheManager[string, ComplexUser](4, 100, 3, aofPath)
	err := newMgr.LoadAOF()
	if err != nil {
		t.Fatalf("Failed to load AOF: %v", err)
	}

	// 3. Verify the struct is exactly the same
	recovered, found := newMgr.Get("user_1")
	if !found {
		t.Fatal("User not found after recovery")
	}

	if recovered.Name != user.Name || recovered.ID != user.ID {
		t.Errorf("Recovered data mismatch. Got %+v, want %+v", recovered, user)
	}
}
