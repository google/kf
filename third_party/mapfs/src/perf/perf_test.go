package perf_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"io/ioutil"
	"os"
	"os/exec"
)

const HIGH_LATENCY_MILLIS = "80ms"
const LOW_LATENCY_MILLIS = "10ms"
const numWrites = "count=100"

var _ = Describe("Perf", func() {
	var nativeDirectory string
	var mapfsDirectory string

	BeforeEach(func() {
		var err error

		nativeDirectory, err = ioutil.TempDir(os.TempDir(), "native")
		Expect(err).NotTo(HaveOccurred())
		mapfsDirectory, err = ioutil.TempDir(os.TempDir(), "mapfs")
		Expect(err).NotTo(HaveOccurred())

	})

	AfterEach(func() {
		unshapeTraffic()

		cmd := exec.Command("umount", "-l", nativeDirectory)
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0), string(session.Out.Contents()))
	})

	Measure("MapFs should not perform writes *much* slower than writing to a native mount", func(benchmarker Benchmarker) {
		By("Natively mounting via nfs", func() {
			nfsUrl := "localhost:/"
			cmd := exec.Command("mount", "-t", "nfs", "-o", "rsize=1048576,wsize=1048576,timeo=600,retrans=2,actimeo=0", nfsUrl, nativeDirectory)
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0), string(session.Out.Contents()))
		})

		benchmarker.Time("writing data to a native mount without throttling", func() {
			writeDataToMountedDirectory(nativeDirectory)
		})

		highLatencyBenchmark := benchmarker.Time("writing data to a native mount with throttling simulating a high latency network", func() {
			shapeTraffic(HIGH_LATENCY_MILLIS)
			writeDataToMountedDirectory(nativeDirectory)
		})

		lowLatencyBenchmark := benchmarker.Time("writing data to a native mount with throttling simulating a low latency network", func() {
			shapeTraffic(LOW_LATENCY_MILLIS)
			writeDataToMountedDirectory(nativeDirectory)
		})

		By("Starting MapFS process", func() {
			cmd := exec.Command(binaryPath, "-uid", "2000", "-gid", "2000", "-debug", mapfsDirectory, nativeDirectory)
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gbytes.Say("Mounted!"))
		})

		highLatencyMapfs := benchmarker.Time("writing data to a mapfs mount with throttling simulating a high latency network", func() {
			shapeTraffic(HIGH_LATENCY_MILLIS)
			writeDataToMountedDirectory(mapfsDirectory)
		})

		lowLatencyMapfs := benchmarker.Time("writing data to a mapfs mount with throttling simulating a low latency network", func() {
			shapeTraffic(LOW_LATENCY_MILLIS)
			writeDataToMountedDirectory(mapfsDirectory)
		})

		Expect(highLatencyMapfs.Nanoseconds()).To(BeNumerically("<", int64(float64(highLatencyBenchmark.Nanoseconds())*2.0)))
		Expect(lowLatencyMapfs.Nanoseconds()).To(BeNumerically("<", int64(float64(lowLatencyBenchmark.Nanoseconds())*10.0)))
	}, 3)
})
