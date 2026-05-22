package regression

import (
	"errors"
	"os"
	"testing"
)

func TestLoadDatasetManifest(t *testing.T) {
	manifest, err := LoadDatasetManifest("../../datasets/tc260/dataset_v6/manifest.json")
	if errors.Is(err, os.ErrNotExist) {
		t.Skip("tc260 manifest is not available")
	}
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}

	if manifest.ManifestFormatVersion != "manifest-v1" {
		t.Fatalf("manifest format = %q, want manifest-v1", manifest.ManifestFormatVersion)
	}
	if _, ok := manifest.FindFile("dataset_tiny.jsonl"); !ok {
		t.Fatal("manifest must describe dataset_tiny.jsonl")
	}
}

func TestVerifyManifestFile(t *testing.T) {
	manifest, err := LoadDatasetManifest("../../datasets/tc260/dataset_v6/manifest.json")
	if errors.Is(err, os.ErrNotExist) {
		t.Skip("tc260 manifest is not available")
	}
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}

	file, err := VerifyManifestFile(*manifest, "../../datasets/tc260/dataset_v6/dataset_tiny.jsonl")
	if errors.Is(err, os.ErrNotExist) {
		t.Skip("tc260 dataset is not available")
	}
	if err != nil {
		t.Fatalf("verify manifest file: %v", err)
	}
	if file.Name != "dataset_tiny.jsonl" {
		t.Fatalf("verified file = %q, want dataset_tiny.jsonl", file.Name)
	}
}
