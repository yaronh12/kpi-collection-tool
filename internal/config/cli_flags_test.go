package config

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("validateFlags test", func() {

	DescribeTable("flag validation scenarios",
		func(flags InputFlags, shouldError bool) {
			err := validateFlags(flags)

			if shouldError {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).ToNot(HaveOccurred())
			}
		},

		// Valid cases
		Entry("valid token and thanos-url",
			InputFlags{
				ClusterName:  "test-cluster",
				BearerToken:  "test-token",
				ThanosURL:    "https://thanos.example.com",
				SamplingFreq: 60,
				Duration:     45 * time.Minute,
				OutputFile:   "output.json",
				LogFile:      "app.log",
			},
			false, // shouldError
		),
		Entry("valid kubeconfig",
			InputFlags{
				ClusterName:  "test-cluster",
				Kubeconfig:   "/path/to/kubeconfig",
				SamplingFreq: 60,
				Duration:     45 * time.Minute,
				OutputFile:   "output.json",
				LogFile:      "app.log",
			},
			false,
		),
		// Error cases - missing cluster name
		Entry("missing cluster name",
			InputFlags{
				BearerToken: "test-token",
				ThanosURL:   "https://thanos.example.com",
			},
			true,
		),
		// Error cases - invalid flag combinations
		Entry("only token provided",
			InputFlags{
				ClusterName: "test-cluster",
				BearerToken: "test-token",
			},
			true,
		),
		Entry("only thanos-url provided",
			InputFlags{
				ClusterName: "test-cluster",
				ThanosURL:   "https://thanos.example.com",
			},
			true,
		),
		Entry("all three auth methods provided",
			InputFlags{
				ClusterName: "test-cluster",
				BearerToken: "test-token",
				ThanosURL:   "https://thanos.example.com",
				Kubeconfig:  "/path/to/kubeconfig",
			},
			true,
		),
		Entry("no authentication method",
			InputFlags{
				ClusterName: "test-cluster",
			},
			true,
		),
		Entry("token and kubeconfig without thanos-url",
			InputFlags{
				ClusterName: "test-cluster",
				BearerToken: "test-token",
				Kubeconfig:  "/path/to/kubeconfig",
			},
			true,
		),
		Entry("thanos-url and kubeconfig without token",
			InputFlags{
				ClusterName: "test-cluster",
				ThanosURL:   "https://thanos.example.com",
				Kubeconfig:  "/path/to/kubeconfig",
			},
			true,
		),
		// Error cases - invalid sampling frequency
		Entry("zero sampling frequency",
			InputFlags{
				ClusterName:  "test-cluster",
				BearerToken:  "test-token",
				ThanosURL:    "https://thanos.example.com",
				SamplingFreq: 0,
				Duration:     45 * time.Minute,
				OutputFile:   "output.json",
				LogFile:      "app.log",
			},
			true,
		),
		Entry("negative sampling frequency",
			InputFlags{
				ClusterName:  "test-cluster",
				BearerToken:  "test-token",
				ThanosURL:    "https://thanos.example.com",
				SamplingFreq: -10,
				Duration:     45 * time.Minute,
				OutputFile:   "output.json",
				LogFile:      "app.log",
			},
			true,
		),
		// Error cases - invalid duration
		Entry("zero duration",
			InputFlags{
				ClusterName:  "test-cluster",
				BearerToken:  "test-token",
				ThanosURL:    "https://thanos.example.com",
				SamplingFreq: 60,
				Duration:     0,
				OutputFile:   "output.json",
				LogFile:      "app.log",
			},
			true,
		),
		Entry("negative duration",
			InputFlags{
				ClusterName:  "test-cluster",
				BearerToken:  "test-token",
				ThanosURL:    "https://thanos.example.com",
				SamplingFreq: 60,
				Duration:     -10 * time.Minute,
				OutputFile:   "output.json",
				LogFile:      "app.log",
			},
			true,
		),
		// Error cases - missing output file
		Entry("empty output file",
			InputFlags{
				ClusterName:  "test-cluster",
				BearerToken:  "test-token",
				ThanosURL:    "https://thanos.example.com",
				SamplingFreq: 60,
				Duration:     45 * time.Minute,
				OutputFile:   "",
				LogFile:      "app.log",
			},
			true,
		),
		// Error cases - missing log file
		Entry("empty log file",
			InputFlags{
				ClusterName:  "test-cluster",
				BearerToken:  "test-token",
				ThanosURL:    "https://thanos.example.com",
				SamplingFreq: 60,
				Duration:     45 * time.Minute,
				OutputFile:   "output.json",
				LogFile:      "",
			},
			true,
		),
	)
})
