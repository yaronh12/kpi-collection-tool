package kubernetes

import (
	"context"
	"os"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ListNodeNames", func() {
	It("returns node names in list order", func() {
		n1 := &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "worker-0"}}
		n2 := &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "worker-1"}}
		client := fake.NewClientset(n1, n2)

		names, err := ListNodeNames(context.Background(), client)
		Expect(err).NotTo(HaveOccurred())
		Expect(names).To(Equal([]string{"worker-0", "worker-1"}))
	})

	It("fails when the cluster has no nodes", func() {
		client := fake.NewClientset()

		names, err := ListNodeNames(context.Background(), client)
		Expect(err).To(MatchError(ContainSubstring("no nodes found")))
		Expect(names).To(BeNil())
	})
})

var _ = Describe("WriteNodesYAML", func() {
	var tmpDir string

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "nodes-yaml-*")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
	})

	It("writes a v1 NodeList manifest", func() {
		node := &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node-a"}}
		client := fake.NewClientset(node)
		path := filepath.Join(tmpDir, "nodes.yaml")

		err := WriteNodesYAML(context.Background(), client, path)
		Expect(err).NotTo(HaveOccurred())

		body, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(body)).To(ContainSubstring("kind: NodeList"))
		Expect(string(body)).To(ContainSubstring("name: node-a"))
	})
})

var _ = Describe("WritePodsYAML", func() {
	var tmpDir string

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "pods-yaml-*")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
	})

	It("writes an empty PodList when the namespace has no pods", func() {
		client := fake.NewClientset()
		path := filepath.Join(tmpDir, "pods.yaml")

		err := WritePodsYAML(context.Background(), client, "workload", path)
		Expect(err).NotTo(HaveOccurred())

		body, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(body)).To(ContainSubstring("kind: PodList"))
		Expect(string(body)).NotTo(ContainSubstring("name:"))
	})

	It("includes pods from the requested namespace", func() {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "workload"},
		}
		client := fake.NewClientset(pod)
		path := filepath.Join(tmpDir, "pods.yaml")

		err := WritePodsYAML(context.Background(), client, "workload", path)
		Expect(err).NotTo(HaveOccurred())

		body, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(body)).To(ContainSubstring("name: app"))
		Expect(string(body)).To(ContainSubstring("namespace: workload"))
	})
})
