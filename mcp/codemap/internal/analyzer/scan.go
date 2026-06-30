package analyzer

import "os"

var skippedDirNames = map[string]bool{
	".git":         true,
	".codemap":     true,
	".idea":        true,
	"vendor":       true,
	"node_modules": true,
	"target":       true,
	"build":        true,
	"dist":         true,
	"out":          true,
}

func ShouldSkipDir(info os.FileInfo) bool {
	return info != nil && info.IsDir() && skippedDirNames[info.Name()]
}
