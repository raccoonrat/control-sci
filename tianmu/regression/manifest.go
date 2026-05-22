package regression

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type DatasetManifest struct {
	ManifestFormatVersion string         `json:"manifest_format_version"`
	DatasetLayoutVersion  string         `json:"dataset_layout_version"`
	CreatedAt             string         `json:"created_at"`
	Files                 []ManifestFile `json:"files"`
	CLICommand            string         `json:"cli_command"`
	TinySampling          map[string]any `json:"tiny_sampling"`
}

type ManifestFile struct {
	Name      string `json:"name"`
	SHA256    string `json:"sha256"`
	Bytes     int64  `json:"bytes"`
	LineCount int    `json:"line_count"`
}

func LoadDatasetManifest(path string) (*DatasetManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var manifest DatasetManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}
	if manifest.ManifestFormatVersion == "" {
		return nil, errors.New("manifest format version is required")
	}
	if manifest.DatasetLayoutVersion == "" {
		return nil, errors.New("dataset layout version is required")
	}
	if len(manifest.Files) == 0 {
		return nil, errors.New("manifest files are required")
	}

	return &manifest, nil
}

func (m DatasetManifest) FindFile(name string) (ManifestFile, bool) {
	for _, file := range m.Files {
		if file.Name == name {
			return file, true
		}
	}

	return ManifestFile{}, false
}

func VerifyManifestFile(manifest DatasetManifest, datasetPath string) (ManifestFile, error) {
	fileName := filepath.Base(datasetPath)
	expected, ok := manifest.FindFile(fileName)
	if !ok {
		return ManifestFile{}, fmt.Errorf("manifest does not describe %q", fileName)
	}

	actual, err := fingerprintFile(datasetPath)
	if err != nil {
		return ManifestFile{}, err
	}
	if actual.SHA256 != expected.SHA256 {
		return ManifestFile{}, fmt.Errorf("sha256 mismatch for %q: got %s, want %s", fileName, actual.SHA256, expected.SHA256)
	}
	if actual.Bytes != expected.Bytes {
		return ManifestFile{}, fmt.Errorf("byte size mismatch for %q: got %d, want %d", fileName, actual.Bytes, expected.Bytes)
	}
	if actual.LineCount != expected.LineCount {
		return ManifestFile{}, fmt.Errorf("line count mismatch for %q: got %d, want %d", fileName, actual.LineCount, expected.LineCount)
	}

	return expected, nil
}

func fingerprintFile(path string) (ManifestFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ManifestFile{}, err
	}
	sum := sha256.Sum256(data)

	return ManifestFile{
		Name:      filepath.Base(path),
		SHA256:    fmt.Sprintf("%x", sum[:]),
		Bytes:     int64(len(data)),
		LineCount: countLines(data),
	}, nil
}

func countLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	lines := bytes.Count(data, []byte{'\n'})
	if data[len(data)-1] != '\n' {
		lines++
	}

	return lines
}
