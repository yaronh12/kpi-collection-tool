package commands

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-best-practices-for-k8s/kpi-collection-tool/internal/config"
)

var _ = Describe("setPerNodeDataIsolcpus", func() {
	It("keeps an explicit isolcpus value", func() {
		spec := config.TasksSpec{
			PerNodeData: &config.PerNodeDataTaskConfig{Isolcpus: "2-3"},
		}
		cpus := &config.CPUPlaceholders{Reserved: "0|1", Isolated: "4|5"}

		Expect(setPerNodeDataIsolcpus(&spec, cpus)).To(Succeed())
		Expect(spec.PerNodeData.Isolcpus).To(Equal("2-3"))
	})

	It("fills isolcpus from the isolated CPU set", func() {
		spec := config.TasksSpec{
			PerNodeData: &config.PerNodeDataTaskConfig{},
		}
		cpus := &config.CPUPlaceholders{Reserved: "0|1", Isolated: "2|3|4"}

		Expect(setPerNodeDataIsolcpus(&spec, cpus)).To(Succeed())
		Expect(spec.PerNodeData.Isolcpus).To(Equal("2,3,4"))
	})

	It("errors when isolcpus is unset and the isolated CPU set is empty", func() {
		spec := config.TasksSpec{
			PerNodeData: &config.PerNodeDataTaskConfig{},
		}

		err := setPerNodeDataIsolcpus(&spec, &config.CPUPlaceholders{Reserved: "0|1"})
		Expect(err).To(HaveOccurred())
	})
})
