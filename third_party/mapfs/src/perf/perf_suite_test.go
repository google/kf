package perf_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"io/ioutil"
	"os/exec"
	"time"

	"testing"
)

func TestPerf(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Perf Suite")
}

var binaryPath string

var nfsServerSession *gexec.Session
var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(10 * time.Second)

	var err error
	binaryPath, err = gexec.Build("code.cloudfoundry.org/mapfs", "-race")
	Expect(err).NotTo(HaveOccurred())

	startNfsCmd := exec.Command("/start.sh")
	nfsServerSession, err = gexec.Start(startNfsCmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(nfsServerSession, 5*time.Second).Should(gbytes.Say("NFS SERVER INITIALIZED"))
})

var _ = AfterSuite(func() {
	nfsServerSession.Kill()
	Eventually(nfsServerSession).Should(gexec.Exit())
})

func writeDataToMountedDirectory(directory string) {
	nativeWriteFile, err := ioutil.TempFile(directory, "perf_file")
	Expect(err).NotTo(HaveOccurred())

	defer func() {
		err := nativeWriteFile.Close()
		Expect(err).NotTo(HaveOccurred())
	}()

	cmd := exec.Command("dd", "if=/dev/zero", "bs=16k", numWrites, fmt.Sprintf("of=%s", nativeWriteFile.Name()))
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session, 10*time.Second).Should(gexec.Exit(0), string(session.Out.Contents()))
}

func shapeTraffic(delayInMs string) {
	unshapeTraffic()
	tcHighLatencyCmd := exec.Command("tc", "qdisc", "add", "dev", "lo", "root", "netem", "delay", delayInMs)
	session, err := gexec.Start(tcHighLatencyCmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
}

func unshapeTraffic() {
	tcHighLatencyCmd := exec.Command("tc", "qdisc", "del", "dev", "lo", "root", "netem")
	session, err := gexec.Start(tcHighLatencyCmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit())

	tcCheckCmd := exec.Command("tc", "qdisc")
	session, err = gexec.Start(tcCheckCmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit())

	GinkgoWriter.Write(session.Out.Contents())
}
