package archive

import (
	"bytes"
	"text/template"

	"github.com/goreleaser/goreleaser/context"
)

func nameFor(ctx *context.Context, build context.Build, name string) (string, error) {
	var out bytes.Buffer
	t, err := template.New(name).Parse(ctx.Config.Archive.NameTemplate)
	if err != nil {
		return "", err
	}
	data := struct {
		Os, Arch, Arm, Version, Tag, Binary, ProjectName string
		Env                                              map[string]string
	}{
		Os:          replace(ctx.Config.Archive.Replacements, build.Goos),
		Arch:        replace(ctx.Config.Archive.Replacements, build.Goarch),
		Arm:         replace(ctx.Config.Archive.Replacements, build.Goarm),
		Version:     ctx.Version,
		Tag:         ctx.Git.CurrentTag,
		Binary:      name, // TODO: deprecated: remove this sometime
		ProjectName: name,
		Env:         ctx.Env,
	}
	err = t.Execute(&out, data)
	return out.String(), err
}

func replace(replacements map[string]string, original string) string {
	result := replacements[original]
	if result == "" {
		return original
	}
	return result
}
