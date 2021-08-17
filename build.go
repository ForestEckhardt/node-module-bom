package nodemodulebom

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/paketo-buildpacks/packit/scribe"
)

//go:generate faux --interface DependencyManager --output fakes/dependency_manager.go
type DependencyManager interface {
	Resolve(path, id, version, stack string) (postal.Dependency, error)
	Deliver(dependency postal.Dependency, cnbPath, layerPath, platformPath string) error
	GenerateBillOfMaterials(dependencies ...postal.Dependency) []packit.BOMEntry
}

//go:generate faux --interface NodeModuleBOM --output fakes/node_module_bom.go
type NodeModuleBOM interface {
	Generate(workingDir string) ([]packit.BOMEntry, error)
}

func Build(dependencyManager DependencyManager, nodeModuleBOM NodeModuleBOM, clock chronos.Clock, logger scribe.Emitter) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		dependency, err := dependencyManager.Resolve(
			filepath.Join(context.CNBPath, "buildpack.toml"),
			"cyclonedx-node-module",
			"*",
			context.Stack,
		)
		if err != nil {
			return packit.BuildResult{}, err
		}

		cycloneDXNodeModuleLayer, err := context.Layers.Get("cyclonedx-node-module")
		if err != nil {
			return packit.BuildResult{}, err
		}

		cachedSHA, ok := cycloneDXNodeModuleLayer.Metadata["dependency-sha"].(string)
		if !ok || cachedSHA != dependency.SHA256 {
			cycloneDXNodeModuleLayer, err = cycloneDXNodeModuleLayer.Reset()
			if err != nil {
				return packit.BuildResult{}, err
			}

			_, err = clock.Measure(func() error {
				return dependencyManager.Deliver(dependency, context.CNBPath, cycloneDXNodeModuleLayer.Path, context.Platform.Path)
			})
			if err != nil {
				return packit.BuildResult{}, err
			}

			cycloneDXNodeModuleLayer.Metadata = map[string]interface{}{
				"dependency-sha": dependency.SHA256,
				"built_at":       clock.Now().Format(time.RFC3339Nano),
			}
		}

		cycloneDXNodeModuleLayer.Cache = true

		logger.Process("Configuring environment")
		os.Setenv("PATH", fmt.Sprint(os.Getenv("PATH"), string(os.PathListSeparator), filepath.Join(cycloneDXNodeModuleLayer.Path, "bin")))

		toolBOM := dependencyManager.GenerateBillOfMaterials(dependency)
		moduleBOM, err := nodeModuleBOM.Generate(context.WorkingDir)
		if err != nil {
			return packit.BuildResult{}, err
		}

		return packit.BuildResult{
			Layers: []packit.Layer{cycloneDXNodeModuleLayer},
			Build: packit.BuildMetadata{
				BOM: append(toolBOM, moduleBOM...),
			},
			Launch: packit.LaunchMetadata{
				BOM: moduleBOM,
			},
		}, nil
	}
}
