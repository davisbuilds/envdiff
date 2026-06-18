package paths

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

var DefaultIgnoredDirs = []string{".git", ".venv", "__pycache__", "node_modules"}

func IterRepoFiles(root string, ignoreDirs []string) ([]string, error) {
	ignored := map[string]struct{}{}
	if ignoreDirs == nil {
		ignoreDirs = DefaultIgnoredDirs
	}
	for _, dir := range ignoreDirs {
		ignored[dir] = struct{}{}
	}

	files := []string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if _, ok := ignored[entry.Name()]; ok {
				return filepath.SkipDir
			}
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(files)
	return files, nil
}

func FindNearestNamedFile(start string, root string, targetName string) (*string, error) {
	rootPath, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	current := start
	if info, statErr := os.Stat(start); statErr == nil && !info.IsDir() {
		current = filepath.Dir(start)
	}
	current, err = filepath.Abs(current)
	if err != nil {
		return nil, err
	}

	for {
		candidate := filepath.Join(current, targetName)
		if info, statErr := os.Stat(candidate); statErr == nil && !info.IsDir() {
			return &candidate, nil
		}
		if current == rootPath {
			return nil, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return nil, nil
		}
		current = parent
	}
}
