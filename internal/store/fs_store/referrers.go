package fs_store

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func (s *FS) referrerDir(name string) string {
	return filepath.Join(s.root, referrersBaseDir, name)
}

func (s *FS) referrerPath(name, digest string) string {
	return filepath.Join(s.referrerDir(name), digest)
}

func (s *FS) GetReferrers(name, digest string, artifactType string) ([]byte, error) {
	referrerDir := s.referrerDir(name)
	if err := os.MkdirAll(referrerDir, 0755); err != nil {
		return nil, err
	}

	referrerPath := s.referrerPath(name, digest)
	content, err := os.ReadFile(referrerPath)

	if err != nil {
		index := map[string]any{
			"schemaVersion": 2,
			"mediaType":     "application/vnd.oci.image.index.v1+json",
			"manifests":     []any{},
		}

		manifestDir := s.manifestDir(name)
		manifests := []any{}

		err := filepath.Walk(manifestDir, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			if info.IsDir() {
				return nil
			}

			data, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			var manifest map[string]any
			if err := json.Unmarshal(data, &manifest); err != nil {
				return nil
			}

			subject, ok := manifest["subject"].(map[string]any)
			if !ok {
				return nil
			}

			subjectDigest, ok := subject["digest"].(string)
			if !ok || subjectDigest != digest {
				return nil
			}

			hasher := sha256.New()
			hasher.Write(data)
			manifestDigest := "sha256:" + hex.EncodeToString(hasher.Sum(nil))

			mediaType, _ := manifest["mediaType"].(string)
			if mediaType == "" {
				mediaType = "application/vnd.oci.image.manifest.v1+json"
			}

			artifactType := ""
			if m, ok := manifest["artifactType"].(string); ok && m != "" {
				artifactType = m
			} else if config, ok := manifest["config"].(map[string]any); ok {
				if configMediaType, ok := config["mediaType"].(string); ok {
					artifactType = configMediaType
				}
			}

			annotations := map[string]string{}
			if annots, ok := manifest["annotations"].(map[string]any); ok {
				for k, v := range annots {
					if strVal, ok := v.(string); ok {
						annotations[k] = strVal
					}
				}
			}

			descriptor := map[string]any{
				"mediaType": mediaType,
				"size":      len(data),
				"digest":    manifestDigest,
			}

			if artifactType != "" {
				descriptor["artifactType"] = artifactType
			}

			if len(annotations) > 0 {
				descriptor["annotations"] = annotations
			}

			manifests = append(manifests, descriptor)

			return nil
		})

		if err != nil {
			return nil, err
		}

		if artifactType != "" {
			filtered := []any{}
			for _, m := range manifests {
				descriptor := m.(map[string]any)
				if at, ok := descriptor["artifactType"].(string); ok && at == artifactType {
					filtered = append(filtered, descriptor)
				}
			}
			manifests = filtered
		}

		index["manifests"] = manifests

		content, err = json.Marshal(index)
		if err != nil {
			return nil, err
		}

		if err := os.WriteFile(referrerPath, content, 0644); err != nil {
			fmt.Printf("Failed to cache referrers: %v\n", err)
		}
	} else if artifactType != "" {
		var index map[string]any
		if err := json.Unmarshal(content, &index); err != nil {
			return nil, err
		}

		manifests, ok := index["manifests"].([]any)
		if !ok {
			return nil, fmt.Errorf("invalid referrers index format")
		}

		filtered := []any{}
		for _, m := range manifests {
			descriptor, ok := m.(map[string]any)
			if !ok {
				continue
			}

			if at, ok := descriptor["artifactType"].(string); ok && at == artifactType {
				filtered = append(filtered, descriptor)
			}
		}

		index["manifests"] = filtered

		content, err = json.Marshal(index)
		if err != nil {
			return nil, err
		}
	}

	return content, nil
}

func (s *FS) UpdateReferrers(name, digest string, manifest []byte) error {
	var m map[string]any
	if err := json.Unmarshal(manifest, &m); err != nil {
		return err
	}

	subject, ok := m["subject"].(map[string]any)
	if !ok {
		return nil
	}

	subjectDigest, ok := subject["digest"].(string)
	if !ok {
		return nil
	}

	hasher := sha256.New()
	hasher.Write(manifest)
	manifestDigest := "sha256:" + hex.EncodeToString(hasher.Sum(nil))

	referrerDir := s.referrerDir(name)
	if err := os.MkdirAll(referrerDir, 0755); err != nil {
		return err
	}

	referrerPath := s.referrerPath(name, subjectDigest)

	var index map[string]any
	content, err := os.ReadFile(referrerPath)
	if err != nil {
		if os.IsNotExist(err) {
			index = map[string]any{
				"schemaVersion": 2,
				"mediaType":     "application/vnd.oci.image.index.v1+json",
				"manifests":     []any{},
			}
		} else {
			return err
		}
	} else {
		if err := json.Unmarshal(content, &index); err != nil {
			return err
		}
	}

	manifests, ok := index["manifests"].([]any)
	if !ok {
		manifests = []any{}
	}

	mediaType, _ := m["mediaType"].(string)
	if mediaType == "" {
		mediaType = "application/vnd.oci.image.manifest.v1+json"
	}

	artifactType := ""
	if at, ok := m["artifactType"].(string); ok && at != "" {
		artifactType = at
	} else if config, ok := m["config"].(map[string]any); ok {
		if configMediaType, ok := config["mediaType"].(string); ok {
			artifactType = configMediaType
		}
	}

	annotations := map[string]string{}
	if annots, ok := m["annotations"].(map[string]any); ok {
		for k, v := range annots {
			if strVal, ok := v.(string); ok {
				annotations[k] = strVal
			}
		}
	}

	descriptor := map[string]any{
		"mediaType": mediaType,
		"size":      len(manifest),
		"digest":    manifestDigest,
	}

	if artifactType != "" {
		descriptor["artifactType"] = artifactType
	}

	if len(annotations) > 0 {
		descriptor["annotations"] = annotations
	}

	found := false
	for i, m := range manifests {
		desc, ok := m.(map[string]any)
		if !ok {
			continue
		}

		if desc["digest"] == manifestDigest {
			manifests[i] = descriptor
			found = true
			break
		}
	}

	if !found {
		manifests = append(manifests, descriptor)
	}

	index["manifests"] = manifests

	updatedContent, err := json.Marshal(index)
	if err != nil {
		return err
	}

	return os.WriteFile(referrerPath, updatedContent, 0644)
}

func (s *FS) RemoveReferrer(name, digest, manifestDigest string) error {
	referrerPath := s.referrerPath(name, digest)

	content, err := os.ReadFile(referrerPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var index map[string]any
	if err := json.Unmarshal(content, &index); err != nil {
		return err
	}

	manifests, ok := index["manifests"].([]any)
	if !ok {
		return nil
	}

	newManifests := []any{}
	for _, m := range manifests {
		desc, ok := m.(map[string]any)
		if !ok {
			continue
		}

		if digest, ok := desc["digest"].(string); ok && digest != manifestDigest {
			newManifests = append(newManifests, desc)
		}
	}

	index["manifests"] = newManifests

	updatedContent, err := json.Marshal(index)
	if err != nil {
		return err
	}

	return os.WriteFile(referrerPath, updatedContent, 0644)
}
