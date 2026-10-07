package builtin_patterns

import (
	"bytes"
	"embed"
	"github.com/marrow16/gogol/patterns"
	"io/fs"
)

//go:embed *.rle
var files embed.FS

func LoadBuiltInPatterns() {
	entries, err := fs.ReadDir(files, ".")
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if data, err := files.ReadFile(name); err == nil {
			if pattern, err := patterns.PatternRleDecoder(bytes.NewReader(data)); err == nil {
				if len(pattern.Name) == 0 {
					pattern.Name = name
				}
				patterns.Library.Register(pattern)
			}
		}
	}
}
