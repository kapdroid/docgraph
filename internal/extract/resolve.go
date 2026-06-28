package extract

import (
	"fmt"
	"path"
	"strings"
)

// resolveRef turns a reference target found inside fromID into a root-relative, slash-cleaned node
// ID. A target beginning with "/" is treated as already root-relative; otherwise it resolves
// relative to the referencing file's directory. This is the single place reference→node-ID
// resolution lives, shared by every extractor, so all edge targets use discovery's identity scheme.
func resolveRef(fromID, target string) string {
	if strings.HasPrefix(target, "/") {
		return path.Clean(strings.TrimPrefix(target, "/"))
	}
	return path.Clean(path.Join(path.Dir(fromID), target))
}

// locOf formats a source location "id:line" for an edge's Loc field.
func locOf(id string, line int) string {
	return fmt.Sprintf("%s:%d", id, line)
}
