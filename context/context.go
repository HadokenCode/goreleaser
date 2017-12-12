// Package context provides gorelease context which is passed through the
// pipeline.
//
// The context extends the standard library context and add a few more
// fields and other things, so pipes can gather data provided by previous
// pipes without really knowing each other.
package context

import (
	ctx "context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/apex/log"
	"github.com/goreleaser/goreleaser/config"
)

// GitInfo includes tags and diffs used in some point
type GitInfo struct {
	CurrentTag string
	Commit     string
}

// Context carries along some data through the pipes
type Context struct {
	ctx.Context
	Config       config.Project
	Env          map[string]string
	Token        string
	Git          GitInfo
	ReleaseNotes string
	Version      string
	Validate     bool
	Publish      bool
	Snapshot     bool
	RmDist       bool
	Debug        bool
	Parallelism  int

	buildsLock sync.Mutex
	Builds     Builds

	artifactsLock sync.Mutex
	Artifacts     Artifacts
}

func (builds Builds) GroupedByFolder() map[string]Builds {
	result := map[string]Builds{}
	for _, build := range builds {
		if result[build.Folder] == nil {
			result[build.Folder] = Builds{}
		}
		result[build.Folder] = append(result[build.Folder], build)
	}
	return result
}

type Build struct {
	Name   string
	Folder string
	Goos   string
	Goarch string
	Goarm  string
}

func (b Build) Path() string {
	return filepath.Join(b.Folder, b.Name)
}

type Builds []Build

func (builds Builds) ByGoos(goos string) Builds {
	var result Builds
	for _, build := range builds {
		if build.Goos == goos {
			result = append(result, build)
		}
	}
	return result
}

func (builds Builds) ByGoarch(goarch string) Builds {
	var result Builds
	for _, build := range builds {
		if build.Goarch == goarch {
			result = append(result, build)
		}
	}
	return result
}

func (builds Builds) ByGoarm(goarm string) Builds {
	var result Builds
	for _, build := range builds {
		if build.Goarm == goarm {
			result = append(result, build)
		}
	}
	return result
}

type ArtifactType int

const (
	Uploadable ArtifactType = iota
	DockerImage
	Checksum
)

type Artifact struct {
	Name   string
	Path   string
	Goos   string
	Goarch string
	Goarm  string
	Type   ArtifactType
}

func (a Artifact) IsUploadable() bool {
	return a.Type == Uploadable
}

func (a Artifact) String() string {
	return fmt.Sprintf("[%v] %s (%s)", a.Type, a.Name, a.Path)
}

type Artifacts []Artifact

func (artifacts Artifacts) ByType(t ArtifactType) Artifacts {
	var result Artifacts
	for _, artifact := range artifacts {
		if artifact.Type == t {
			result = append(result, artifact)
		}
	}
	return result
}

func (artifacts Artifacts) ByGoos(goos string) Artifacts {
	var result Artifacts
	for _, artifact := range artifacts {
		if artifact.Goos == goos {
			result = append(result, artifact)
		}
	}
	return result
}

func (artifacts Artifacts) ByGoarch(goarch string) Artifacts {
	var result Artifacts
	for _, artifact := range artifacts {
		if artifact.Goarch == goarch {
			result = append(result, artifact)
		}
	}
	return result
}

func (artifacts Artifacts) ByGoarm(goarm string) Artifacts {
	var result Artifacts
	for _, artifact := range artifacts {
		if artifact.Goarm == goarm {
			result = append(result, artifact)
		}
	}
	return result
}

// AddArtifact adds a file to upload list
func (ctx *Context) AddArtifact(artifact Artifact) {
	ctx.artifactsLock.Lock()
	defer ctx.artifactsLock.Unlock()
	ctx.Artifacts = append(ctx.Artifacts, artifact)
	log.WithField("artifact", artifact).Info("new release artifact")
}

// AddBinary adds a built binary to the current context
func (ctx *Context) AddBuild(build Build) {
	ctx.buildsLock.Lock()
	defer ctx.buildsLock.Unlock()
	ctx.Builds = append(ctx.Builds, build)
	log.WithField("build", build).Debug("new binary")
}

// New context
func New(config config.Project) *Context {
	return &Context{
		Context:     ctx.Background(),
		Config:      config,
		Env:         splitEnv(os.Environ()),
		Parallelism: 4,
	}
}

func splitEnv(env []string) map[string]string {
	r := map[string]string{}
	for _, e := range env {
		p := strings.SplitN(e, "=", 2)
		r[p[0]] = p[1]
	}
	return r
}
