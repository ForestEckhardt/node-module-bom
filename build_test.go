package nodemodulebom_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	nodemodulebom "github.com/paketo-buildpacks/node-module-bom"
	"github.com/paketo-buildpacks/node-module-bom/fakes"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/paketo-buildpacks/packit/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir         string
		cnbDir            string
		workingDir        string
		timestamp         time.Time
		dependencyManager *fakes.DependencyManager
		nodeModuleBOM     *fakes.NodeModuleBOM
		buffer            *bytes.Buffer

		build packit.BuildFunc
	)

	it.Before(func() {
		var err error
		layersDir, err = os.MkdirTemp("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = os.MkdirTemp("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		workingDir, err = os.MkdirTemp("", "workingDir")
		Expect(err).NotTo(HaveOccurred())

		timestamp = time.Now()
		clock := chronos.NewClock(func() time.Time {
			return timestamp
		})

		dependencyManager = &fakes.DependencyManager{}
		dependencyManager.ResolveCall.Returns.Dependency = postal.Dependency{
			ID:      "cyclonedx-node-module",
			Name:    "cyclonedx-node-module-dependency-name",
			SHA256:  "cyclonedx-node-module-dependency-sha",
			Stacks:  []string{"some-stack"},
			URI:     "cyclonedx-node-module-dependency-uri",
			Version: "cyclonedx-node-module-dependency-version",
		}

		dependencyManager.GenerateBillOfMaterialsCall.Returns.BOMEntrySlice = []packit.BOMEntry{
			{
				Name: "cyclonedx-node-module",
				Metadata: map[string]interface{}{
					"version": "cyclonedx-node-module-dependency-version",
					"name":    "cyclonedx-node-module-dependency-name",
					"sha256":  "cyclonedx-node-module-dependency-sha",
					"stacks":  []string{"some-stack"},
					"uri":     "cyclonedx-node-module-dependency-uri",
				},
			},
		}

		nodeModuleBOM = &fakes.NodeModuleBOM{}
		nodeModuleBOM.GenerateCall.Returns.BOMEntrySlice = []packit.BOMEntry{
			{
				Name: "leftpad",
				Metadata: map[string]interface{}{
					"version": "leftpad-dependency-version",
					"name":    "leftpad-dependency-name",
					"sha256":  "leftpad-dependency-sha",
					"stacks":  []string{"some-stack"},
					"uri":     "leftpad-dependency-uri",
				},
			},
		}

		buffer = bytes.NewBuffer(nil)
		logEmitter := scribe.NewEmitter(buffer)

		build = nodemodulebom.Build(dependencyManager, nodeModuleBOM, clock, logEmitter)
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	it("returns a result that installs cyclonedx-node-module", func() {
		result, err := build(packit.BuildContext{
			BuildpackInfo: packit.BuildpackInfo{
				Name:    "Some Buildpack",
				Version: "some-version",
			},
			CNBPath:    cnbDir,
			Platform:   packit.Platform{Path: "platform"},
			Layers:     packit.Layers{Path: layersDir},
			Stack:      "some-stack",
			WorkingDir: workingDir,
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(result).To(Equal(packit.BuildResult{
			Layers: []packit.Layer{
				{
					Name:             "cyclonedx-node-module",
					Path:             filepath.Join(layersDir, "cyclonedx-node-module"),
					SharedEnv:        packit.Environment{},
					BuildEnv:         packit.Environment{},
					LaunchEnv:        packit.Environment{},
					ProcessLaunchEnv: map[string]packit.Environment{},
					Build:            false,
					Launch:           false,
					Cache:            true,
					Metadata: map[string]interface{}{
						"dependency-sha": "cyclonedx-node-module-dependency-sha",
						"built_at":       timestamp.Format(time.RFC3339Nano),
					},
				},
			},
			Build: packit.BuildMetadata{
				BOM: []packit.BOMEntry{
					{
						Name: "cyclonedx-node-module",
						Metadata: map[string]interface{}{
							"version": "cyclonedx-node-module-dependency-version",
							"name":    "cyclonedx-node-module-dependency-name",
							"sha256":  "cyclonedx-node-module-dependency-sha",
							"stacks":  []string{"some-stack"},
							"uri":     "cyclonedx-node-module-dependency-uri",
						},
					},
					{
						Name: "leftpad",
						Metadata: map[string]interface{}{
							"version": "leftpad-dependency-version",
							"name":    "leftpad-dependency-name",
							"sha256":  "leftpad-dependency-sha",
							"stacks":  []string{"some-stack"},
							"uri":     "leftpad-dependency-uri",
						},
					},
				},
			},
			Launch: packit.LaunchMetadata{
				BOM: []packit.BOMEntry{
					{
						Name: "leftpad",
						Metadata: map[string]interface{}{
							"version": "leftpad-dependency-version",
							"name":    "leftpad-dependency-name",
							"sha256":  "leftpad-dependency-sha",
							"stacks":  []string{"some-stack"},
							"uri":     "leftpad-dependency-uri",
						},
					},
				},
			},
		}))

		Expect(dependencyManager.ResolveCall.Receives.Path).To(Equal(filepath.Join(cnbDir, "buildpack.toml")))
		Expect(dependencyManager.ResolveCall.Receives.Id).To(Equal("cyclonedx-node-module"))
		Expect(dependencyManager.ResolveCall.Receives.Version).To(Equal("*"))
		Expect(dependencyManager.ResolveCall.Receives.Stack).To(Equal("some-stack"))

		Expect(dependencyManager.DeliverCall.Receives.Dependency).To(Equal(postal.Dependency{
			ID:      "cyclonedx-node-module",
			Name:    "cyclonedx-node-module-dependency-name",
			SHA256:  "cyclonedx-node-module-dependency-sha",
			Stacks:  []string{"some-stack"},
			URI:     "cyclonedx-node-module-dependency-uri",
			Version: "cyclonedx-node-module-dependency-version",
		}))
		Expect(dependencyManager.DeliverCall.Receives.CnbPath).To(Equal(cnbDir))
		Expect(dependencyManager.DeliverCall.Receives.LayerPath).To(Equal(filepath.Join(layersDir, "cyclonedx-node-module")))
		Expect(dependencyManager.DeliverCall.Receives.PlatformPath).To(Equal("platform"))

		Expect(dependencyManager.GenerateBillOfMaterialsCall.Receives.Dependencies).To(Equal([]postal.Dependency{
			{
				ID:      "cyclonedx-node-module",
				Name:    "cyclonedx-node-module-dependency-name",
				SHA256:  "cyclonedx-node-module-dependency-sha",
				Stacks:  []string{"some-stack"},
				URI:     "cyclonedx-node-module-dependency-uri",
				Version: "cyclonedx-node-module-dependency-version",
			},
		}))

		Expect(nodeModuleBOM.GenerateCall.Receives.WorkingDir).To(Equal(workingDir))
	})
}
