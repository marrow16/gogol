package help

import (
	"os"
	"strings"
	"testing"
)

func TestGenerateHelpHtmls(t *testing.T) {
	files := make([]*os.File, 0)
	defer func() {
		for _, f := range files {
			_ = f.Close()
		}
	}()
	if f, err := os.Create("../../../_help/index.html"); err == nil {
		files = append(files, f)
		var sb strings.Builder
		copyIndexHtml(&sb)
		_, _ = f.WriteString(sb.String())
	}
	for tp, c := range contents {
		if tp != Index {
			if f, err := os.Create("../../../_help/" + strings.TrimSuffix(strings.ReplaceAll(tp.String(), "/", " "), " Help") + ".html"); err == nil {
				files = append(files, f)
				var sb strings.Builder
				copyHtml(&sb, tp, c)
				_, _ = f.WriteString(sb.String())
			}
		}
	}
}
