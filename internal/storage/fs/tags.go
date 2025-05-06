package fs_storage

import (
	"os"
	"path/filepath"
)

func (s *FS) tagDir(name string) string {
	return filepath.Join(s.root, tagsBaseDir, name)
}

func (s *FS) tagPath(name, tag string) string {
	return filepath.Join(s.tagDir(name), tag)
}

func (s *FS) ListTags(name string) ([]string, error) {
	tagDir := s.tagDir(name)

	if err := os.MkdirAll(tagDir, 0755); err != nil {
		return nil, err
	}

	files, err := os.ReadDir(tagDir)
	if err != nil {
		return nil, err
	}

	tags := make([]string, 0, len(files))
	for _, file := range files {
		if !file.IsDir() {
			tags = append(tags, file.Name())
		}
	}

	return tags, nil
}
