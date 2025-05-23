package fetchers

import (
	"context"
	"testing"
)

func TestInstanceSpecCollector_ParseJson(t *testing.T) {
	data := []byte(`{
	"deployment": "test-dev",
	"name": "exporters",
	"index": 0,
	"id": "b36ca4c9-80ce-426a-998e-8b23fd50efe3",
	"az": "z2",
	"other_field": "x"
}`)
	fetcher := NewInstanceSpecFetcher("fake-path")
	spec, err := fetcher.FetchData(data)
	if err != nil {
		t.Fatalf("Fetch() returned error: %v", err)
	}

	// verify fields
	if spec.Deployment != "test-dev" {
		t.Errorf("expected Deployment 'test-dev', got '%s'", spec.Deployment)
	}
	if spec.Name != "exporters" {
		t.Errorf("expected Name 'exporters', got '%s'", spec.Name)
	}
	if spec.Index != 0 {
		t.Errorf("expected Index 0, got %d", spec.Index)
	}
	if spec.ID != "b36ca4c9-80ce-426a-998e-8b23fd50efe3" {
		t.Errorf("expected ID 'b36ca4c9-80ce-426a-998e-8b23fd50efe3', got '%s'", spec.ID)
	}
	if spec.AZ != "z2" {
		t.Errorf("expected AZ 'z2', got '%s'", spec.AZ)
	}
}

func TestInstanceSpecCollector_FileNotFound(t *testing.T) {
	fetcher := NewInstanceSpecFetcher("/does/not/exist.json")
	_, err := fetcher.Fetch(context.Background())
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}
