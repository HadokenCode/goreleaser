// Package archiveformat provides functions to get the format of given package
// based on the config
package archiveformat

import (
	"github.com/goreleaser/goreleaser/context"
)

// For return the archive format, considering overrides and all that
func For(ctx *context.Context, goos string) string {
	for _, override := range ctx.Config.Archive.FormatOverrides {
		if goos == override.Goos {
			return override.Format
		}
	}
	return ctx.Config.Archive.Format
}
