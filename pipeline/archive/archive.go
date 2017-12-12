// Package archive implements the pipe interface with the intent of
// archiving and compressing the binaries, readme, and other artifacts. It
// also provides an Archive interface which represents an archiving format.
package archive

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/goreleaser/archive"
	"github.com/goreleaser/goreleaser/context"
	"github.com/goreleaser/goreleaser/internal/archiveformat"
	"github.com/goreleaser/goreleaser/internal/ext"
	"github.com/mattn/go-zglob"
	"golang.org/x/sync/errgroup"
)

// Pipe for archive
type Pipe struct{}

func (Pipe) String() string {
	return "creating archives"
}

// Default sets the pipe defaults
func (Pipe) Default(ctx *context.Context) error {
	if ctx.Config.Archive.NameTemplate == "" {
		ctx.Config.Archive.NameTemplate = "{{ .Binary }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}{{ if .Arm }}v{{ .Arm }}{{ end }}"
	}
	if ctx.Config.Archive.Format == "" {
		ctx.Config.Archive.Format = "tar.gz"
	}
	if len(ctx.Config.Archive.Files) == 0 {
		ctx.Config.Archive.Files = []string{
			"licence*",
			"LICENCE*",
			"license*",
			"LICENSE*",
			"readme*",
			"README*",
			"changelog*",
			"CHANGELOG*",
		}
	}
	return nil
}

// Run the pipe
func (Pipe) Run(ctx *context.Context) error {
	var g errgroup.Group
	for folder, builds := range ctx.Builds.GroupedByFolder() {
		folder := folder
		builds := builds
		g.Go(func() error {
			if ctx.Config.Archive.Format == "binary" {
				return skip(ctx, folder, builds)
			}
			return create(ctx, folder, builds)
		})
	}
	return g.Wait()
}

func create(ctx *context.Context, folder string, builds context.Builds) error {
	var format = archiveformat.For(ctx, builds[0].Goos)
	archiveName, err := nameFor(ctx, builds[0], ctx.Config.ProjectName)
	if err != nil {
		return err
	}
	archivePath := filepath.Join(ctx.Config.Dist, archiveName+"."+format)
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %s", archivePath, err.Error())
	}
	defer func() {
		if e := archiveFile.Close(); e != nil {
			log.WithField("archive", archivePath).Errorf("failed to close file: %v", e)
		}
	}()
	log.WithField("archive", archivePath).Info("creating")
	var a = archive.New(archiveFile)
	defer func() {
		if e := a.Close(); e != nil {
			log.WithField("archive", archivePath).Errorf("failed to close archive: %v", e)
		}
	}()

	files, err := findFiles(ctx)
	if err != nil {
		return fmt.Errorf("failed to find files to archive: %s", err.Error())
	}
	for _, f := range files {
		if err = a.Add(wrap(ctx, f, folder), f); err != nil {
			return fmt.Errorf("failed to add %s to the archive: %s", f, err.Error())
		}
	}
	for _, build := range builds {
		var path = filepath.Join(ctx.Config.Dist, build.Path())
		if err := a.Add(wrap(ctx, build.Name, folder), path); err != nil {
			return fmt.Errorf("failed to add %s -> %s to the archive: %s", path, build.Name, err.Error())
		}
	}
	ctx.AddArtifact(context.Artifact{
		Name:   folder + "." + format,
		Path:   archivePath,
		Type:   context.Uploadable,
		Goos:   builds[0].Goos,
		Goarch: builds[0].Goarch,
		Goarm:  builds[0].Goarm,
	})
	return nil
}

func skip(ctx *context.Context, folder string, builds context.Builds) error {
	for _, build := range builds {
		var path = filepath.Join(ctx.Config.Dist, build.Path())
		// TODO: maybe the ext should only be added here?
		name, err := nameFor(ctx, build, strings.TrimSuffix(build.Name, ".exe"))
		if err != nil {
			return err
		}
		name = name + ext.ForOS(build.Goos)
		log.WithField("binary", name).Info("skip archiving")
		ctx.AddArtifact(context.Artifact{
			Name:   name,
			Path:   path,
			Type:   context.Uploadable,
			Goos:   build.Goos,
			Goarch: build.Goarch,
			Goarm:  build.Goarm,
		})
	}
	return nil
}

func findFiles(ctx *context.Context) (result []string, err error) {
	for _, glob := range ctx.Config.Archive.Files {
		files, err := zglob.Glob(glob)
		if err != nil {
			return result, fmt.Errorf("globbing failed for pattern %s: %s", glob, err.Error())
		}
		result = append(result, files...)
	}
	return
}

// Wrap archive files with folder if set in config.
func wrap(ctx *context.Context, name, folder string) string {
	if ctx.Config.Archive.WrapInDirectory {
		return filepath.Join(folder, name)
	}
	return name
}
