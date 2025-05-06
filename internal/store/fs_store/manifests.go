package fs_store

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func (s *FS) manifestDir(name string) string {
	return filepath.Join(s.root, manifestsBaseDir, name)
}

func (s *FS) manifestPath(name, reference string) string {
	return filepath.Join(s.manifestDir(name), reference)
}

func (s *FS) HasManifest(name, reference string) (bool, int64, string, error) {
	isDigest := strings.HasPrefix(reference, "sha256:")

	if isDigest {
		path := s.manifestPath(name, reference)
		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				return false, 0, "", nil
			}
			return false, 0, "", err
		}
		return true, info.Size(), reference, nil
	} else {
		tagPath := s.tagPath(name, reference)
		digest, err := os.ReadFile(tagPath)
		if err != nil {
			if os.IsNotExist(err) {
				return false, 0, "", nil
			}
			return false, 0, "", err
		}

		digestStr := string(digest)
		path := s.manifestPath(name, digestStr)
		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				return false, 0, digestStr, nil
			}
			return false, 0, digestStr, err
		}
		return true, info.Size(), digestStr, nil
	}
}

// GetManifest retrieves a manifest
func (s *FS) GetManifest(name, reference string) ([]byte, string, error) {
	if strings.HasPrefix(reference, "sha256:") {
		manifestDir := s.manifestDir(name)

		if err := os.MkdirAll(manifestDir, 0o755); err != nil {
			return nil, "", err
		}

		var manifestPath string
		var content []byte

		err := filepath.Walk(manifestDir, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			data, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			hasher := sha256.New()
			hasher.Write(data)
			digest := "sha256:" + hex.EncodeToString(hasher.Sum(nil))

			if digest == reference {
				manifestPath = path
				content = data
				return filepath.SkipAll
			}

			return nil
		})
		if err != nil {
			return nil, "", err
		}

		if manifestPath == "" {
			return nil, "", fmt.Errorf("manifest not found")
		}

		return content, reference, nil
	}

	// Try to resolve tag to manifest
	tagPath := s.tagPath(name, reference)
	digestBytes, err := os.ReadFile(tagPath)
	if err == nil {
		// Tag exists, resolve to manifest
		digestStr := string(digestBytes)
		manifestPath := s.manifestPath(name, digestStr)
		content, err := os.ReadFile(manifestPath)
		if err != nil {
			return nil, "", err
		}
		return content, digestStr, nil
	}

	// Try direct manifest path
	manifestPath := s.manifestPath(name, reference)
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", fmt.Errorf("manifest not found")
		}
		return nil, "", err
	}

	// Calculate digest
	hasher := sha256.New()
	hasher.Write(content)
	digest := "sha256:" + hex.EncodeToString(hasher.Sum(nil))

	return content, digest, nil
}

// PutManifest stores a manifest
func (s *FS) PutManifest(name, reference string, content []byte) (string, error) {
	// Calculate digest
	hasher := sha256.New()
	hasher.Write(content)
	digest := "sha256:" + hex.EncodeToString(hasher.Sum(nil))

	// Create manifest directory
	manifestDir := s.manifestDir(name)
	if err := os.MkdirAll(manifestDir, 0o755); err != nil {
		return "", err
	}

	// Store by digest
	digestPath := s.manifestPath(name, digest)
	if err := os.WriteFile(digestPath, content, 0o644); err != nil {
		return "", err
	}

	// If reference is a tag, create/update tag
	if !strings.HasPrefix(reference, "sha256:") {
		tagDir := s.tagDir(name)
		if err := os.MkdirAll(tagDir, 0o755); err != nil {
			return "", err
		}

		tagPath := s.tagPath(name, reference)
		if err := os.WriteFile(tagPath, []byte(digest), 0o644); err != nil {
			return "", err
		}
	}

	return digest, nil
}

func (s *FS) DeleteManifest(name, reference string) error {
	isDigest := strings.HasPrefix(reference, "sha256:")

	if isDigest {
		path := s.manifestPath(name, reference)
		if err := os.Remove(path); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("manifest not found")
			}
			return err
		}

		tagDir := s.tagDir(name)
		err := filepath.Walk(tagDir, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			if info.IsDir() {
				return nil
			}

			tagDigest, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			if string(tagDigest) == reference {
				os.Remove(path)
			}

			return nil
		})
		if err != nil {
			fmt.Printf("Error cleaning up tags: %v\n", err)
		}

		return nil
	} else {
		tagPath := s.tagPath(name, reference)
		_, err := os.ReadFile(tagPath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("tag not found")
			}
			return err
		}

		if err := os.Remove(tagPath); err != nil {
			return err
		}

		return nil
	}
}
