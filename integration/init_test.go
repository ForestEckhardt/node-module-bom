package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var (
	nodeModuleBOMBuildpack        string
	offlineNodeModuleBOMBuildpack string
	nodeEngineBuildpack           string
	offlineNodeEngineBuildpack    string
	nodeStartBuildpack            string
	root                          string

	config struct {
		Buildpack struct {
			ID   string
			Name string
		}
	}

	integrationjson struct {
		NodeEngine string `json:"node-engine"`
		NodeStart  string `json:"node-start"`
	}
)

func TestIntegration(t *testing.T) {
	Expect := NewWithT(t).Expect

	var err error
	root, err = filepath.Abs("./..")
	Expect(err).ToNot(HaveOccurred())

	file, err := os.Open("../buildpack.toml")
	Expect(err).NotTo(HaveOccurred())
	defer file.Close()

	_, err = toml.DecodeReader(file, &config)
	Expect(err).NotTo(HaveOccurred())

	file, err = os.Open("../integration.json")
	Expect(err).NotTo(HaveOccurred())

	Expect(json.NewDecoder(file).Decode(&integrationjson)).To(Succeed())
	Expect(file.Close()).To(Succeed())

	buildpackStore := occam.NewBuildpackStore()

	nodeModuleBOMBuildpack, err = buildpackStore.Get.
		WithVersion("1.2.3").
		Execute(root)
	Expect(err).NotTo(HaveOccurred())

	offlineNodeModuleBOMBuildpack, err = buildpackStore.Get.
		WithOfflineDependencies().
		WithVersion("1.2.3").
		Execute(root)
	Expect(err).NotTo(HaveOccurred())

	nodeEngineBuildpack, err = buildpackStore.Get.
		Execute(integrationjson.NodeEngine)
	Expect(err).NotTo(HaveOccurred())

	nodeEngineBuildpack, err = buildpackStore.Get.
		WithOfflineDependencies().
		Execute(integrationjson.NodeEngine)
	Expect(err).NotTo(HaveOccurred())

	nodeStartBuildpack, err = buildpackStore.Get.
		Execute(integrationjson.NodeStart)
	Expect(err).NotTo(HaveOccurred())

	SetDefaultEventuallyTimeout(5 * time.Second)

	suite := spec.New("Integration", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Default", testDefault)
	suite.Run(t)
}
