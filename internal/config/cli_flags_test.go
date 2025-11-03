package config

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	validClusterName  = "test-cluster"
	validBearerToken  = "test-token"
	validThanosURL    = "https://thanos.example.com"
	validKubeconfig   = "/path/to/kubeconfig"
	validOutputFile   = "output.json"
	validLogFile      = "app.log"
	validSamplingFreq = 60
	validDuration     = 45 * time.Minute

	errClusterNameRequiredMsg = "cluster name is required: use --cluster-name flag"
	errInvalidFlagComboMsg    = "invalid flag combination: either provide --token and --thanos-url, or provide --kubeconfig"
	errSamplingFreqMsg        = "sampling frequency must be greater than 0"
	errDurationMsg            = "duration must be greater than 0"
	errOutputFileMsg          = "output file must be specified"
	errLogFileMsg             = "log file must be specified"
)

var _ = Describe("validateFlags test", func() {

	DescribeTable("flag validation scenarios",
		func(flags InputFlags, expectedErr string) {
			err := validateFlags(flags)

			if expectedErr != "" {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(expectedErr))
			} else {
				Expect(err).ToNot(HaveOccurred())
			}
		},

		// Valid cases
		Entry("valid token and thanos-url",
			InputFlags{
				ClusterName:  validClusterName,
				BearerToken:  validBearerToken,
				ThanosURL:    validThanosURL,
				SamplingFreq: validSamplingFreq,
				Duration:     validDuration,
				OutputFile:   validOutputFile,
				LogFile:      validLogFile,
			},
			"", // no error expected
		),
		Entry("valid kubeconfig",
			InputFlags{
				ClusterName:  validClusterName,
				Kubeconfig:   validKubeconfig,
				SamplingFreq: validSamplingFreq,
				Duration:     validDuration,
				OutputFile:   validOutputFile,
				LogFile:      validLogFile,
			},
			"",
		),
		// Error cases - missing cluster name
		Entry("missing cluster name",
			InputFlags{
				BearerToken: validBearerToken,
				ThanosURL:   validThanosURL,
			},
			errClusterNameRequiredMsg,
		),
		// Error cases - invalid flag combinations
		Entry("only token provided",
			InputFlags{
				ClusterName: validClusterName,
				BearerToken: validBearerToken,
			},
			errInvalidFlagComboMsg,
		),
		Entry("only thanos-url provided",
			InputFlags{
				ClusterName: validClusterName,
				ThanosURL:   validThanosURL,
			},
			errInvalidFlagComboMsg,
		),
		Entry("all three auth methods provided",
			InputFlags{
				ClusterName: validClusterName,
				BearerToken: validBearerToken,
				ThanosURL:   validThanosURL,
				Kubeconfig:  validKubeconfig,
			},
			errInvalidFlagComboMsg,
		),
		Entry("no authentication method",
			InputFlags{
				ClusterName: validClusterName,
			},
			errInvalidFlagComboMsg,
		),
		Entry("token and kubeconfig without thanos-url",
			InputFlags{
				ClusterName: validClusterName,
				BearerToken: validBearerToken,
				Kubeconfig:  validKubeconfig,
			},
			errInvalidFlagComboMsg,
		),
		Entry("thanos-url and kubeconfig without token",
			InputFlags{
				ClusterName: validClusterName,
				ThanosURL:   validThanosURL,
				Kubeconfig:  validKubeconfig,
			},
			errInvalidFlagComboMsg,
		),
		// Error cases - invalid sampling frequency
		Entry("zero sampling frequency",
			InputFlags{
				ClusterName:  validClusterName,
				BearerToken:  validBearerToken,
				ThanosURL:    validThanosURL,
				SamplingFreq: 0,
				Duration:     validDuration,
				OutputFile:   validOutputFile,
				LogFile:      validLogFile,
			},
			errSamplingFreqMsg,
		),
		Entry("negative sampling frequency",
			InputFlags{
				ClusterName:  validClusterName,
				BearerToken:  validBearerToken,
				ThanosURL:    validThanosURL,
				SamplingFreq: -10,
				Duration:     validDuration,
				OutputFile:   validOutputFile,
				LogFile:      validLogFile,
			},
			errSamplingFreqMsg,
		),
		// Error cases - invalid duration
		Entry("zero duration",
			InputFlags{
				ClusterName:  validClusterName,
				BearerToken:  validBearerToken,
				ThanosURL:    validThanosURL,
				SamplingFreq: validSamplingFreq,
				Duration:     0,
				OutputFile:   validOutputFile,
				LogFile:      validLogFile,
			},
			errDurationMsg,
		),
		Entry("negative duration",
			InputFlags{
				ClusterName:  validClusterName,
				BearerToken:  validBearerToken,
				ThanosURL:    validThanosURL,
				SamplingFreq: validSamplingFreq,
				Duration:     -10 * time.Minute,
				OutputFile:   validOutputFile,
				LogFile:      validLogFile,
			},
			errDurationMsg,
		),
		// Error cases - missing output file
		Entry("empty output file",
			InputFlags{
				ClusterName:  validClusterName,
				BearerToken:  validBearerToken,
				ThanosURL:    validThanosURL,
				SamplingFreq: validSamplingFreq,
				Duration:     validDuration,
				OutputFile:   "",
				LogFile:      validLogFile,
			},
			errOutputFileMsg,
		),
		// Error cases - missing log file
		Entry("empty log file",
			InputFlags{
				ClusterName:  validClusterName,
				BearerToken:  validBearerToken,
				ThanosURL:    validThanosURL,
				SamplingFreq: validSamplingFreq,
				Duration:     validDuration,
				OutputFile:   validOutputFile,
				LogFile:      "",
			},
			errLogFileMsg,
		),
	)
})
