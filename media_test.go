package main

import "testing"

func TestMediaStoreQueuesCreatorAsset(t *testing.T) {
	s := &MediaStore{jobs: map[string]MediaJob{}}
	got := s.create(MediaJob{ID: "clip-7", Creator: "creator-1", Asset: "episode.mp4", Status: "queued"})
	if got.Status != "queued" || got.Creator != "creator-1" {
		t.Fatalf("job decision = %+v", got)
	}
	loaded, ok := s.get("clip-7")
	if !ok || loaded.Asset != "episode.mp4" {
		t.Fatalf("stored job = %+v, found=%v", loaded, ok)
	}
}
