package task

import (
	"context"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/redhat-best-practices-for-k8s/kpi-collection-tool/internal/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("podReadyFields", func() {
	It("returns Ready status and last transition time", func() {
		when := metav1.NewTime(time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC))
		pod := &corev1.Pod{
			Status: corev1.PodStatus{
				Conditions: []corev1.PodCondition{{
					Type:               corev1.PodReady,
					Status:             corev1.ConditionTrue,
					LastTransitionTime: when,
				}},
			},
		}

		status, ts := podReadyFields(pod)
		Expect(status).To(Equal("True"))
		Expect(ts).To(Equal("2026-09-15T12:00:00Z"))
	})

	It("returns empty strings when PodReady is missing", func() {
		status, ts := podReadyFields(&corev1.Pod{})
		Expect(status).To(BeEmpty())
		Expect(ts).To(BeEmpty())
	})
})

var _ = Describe("AppRecoveryTimeTask writePollBlock", func() {
	It("writes pod rows and list errors to the output file", func() {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "workload"},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
				Conditions: []corev1.PodCondition{{
					Type:   corev1.PodReady,
					Status: corev1.ConditionTrue,
				}},
			},
		}
		client := fake.NewClientset(pod)

		f, err := os.CreateTemp("", "pod-status-*")
		Expect(err).NotTo(HaveOccurred())
		defer func() {
			Expect(f.Close()).To(Succeed())
			Expect(os.Remove(f.Name())).To(Succeed())
		}()

		t := &AppRecoveryTimeTask{
			cfg: config.AppRecoveryTimeTaskConfig{
				WorkloadNamespaces: []string{"workload"},
			},
		}

		err = t.writePollBlock(context.Background(), client, f, false)
		Expect(err).NotTo(HaveOccurred())

		body, err := os.ReadFile(f.Name())
		Expect(err).NotTo(HaveOccurred())
		Expect(string(body)).To(ContainSubstring("NAMESPACE  NAME  PHASE  READY  READYTIME"))
		Expect(string(body)).To(ContainSubstring("workload  app  Running  True"))
	})
})
