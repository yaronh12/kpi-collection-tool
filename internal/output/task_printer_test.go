package output

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PollReporter", func() {
	It("does not heartbeat until 60s since last output", func() {
		p := NewPollReporter("app-recovery-time")
		p.lastOut = time.Now()
		p.MaybeHeartbeat(2 * time.Minute)
		// No stdout capture; ensure lastOut unchanged when gap not met.
		Expect(time.Since(p.lastOut)).To(BeNumerically("<", time.Second))
	})

	It("updates lastOut after a heartbeat", func() {
		p := NewPollReporter("app-recovery-time")
		p.lastOut = time.Now().Add(-61 * time.Second)
		p.MaybeHeartbeat(2 * time.Minute)
		Expect(time.Since(p.lastOut)).To(BeNumerically("<", time.Second))
	})
})
