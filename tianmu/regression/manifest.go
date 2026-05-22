package regression

import (
	"encoding/json"
	"errors"
	"os"
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
