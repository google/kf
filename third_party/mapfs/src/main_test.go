package main

import (
	"io"
	"io/ioutil"
	"os/exec"
	"path/filepath"

	"code.cloudfoundry.org/goshims/bufioshim"
	"code.cloudfoundry.org/goshims/osshim"
	"code.cloudfoundry.org/volumedriver/mountchecker"

	"fmt"

	"os"
	"time"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type failRunner struct {
	Command           *exec.Cmd
	Name              string
	AnsiColorCode     string
	StartCheck        string
	StartCheckTimeout time.Duration
	Cleanup           func()
	session           *gexec.Session
	sessionReady      chan struct{}
}

func (r failRunner) Run(sigChan <-chan os.Signal, ready chan<- struct{}) error {
	defer GinkgoRecover()

	allOutput := gbytes.NewBuffer()

	debugWriter := gexec.NewPrefixedWriter(
		fmt.Sprintf("\x1b[32m[d]\x1b[%s[%s]\x1b[0m ", r.AnsiColorCode, r.Name),
		GinkgoWriter,
	)

	session, err := gexec.Start(
		r.Command,
		gexec.NewPrefixedWriter(
			fmt.Sprintf("\x1b[32m[o]\x1b[%s[%s]\x1b[0m ", r.AnsiColorCode, r.Name),
			io.MultiWriter(allOutput, GinkgoWriter),
		),
		gexec.NewPrefixedWriter(
			fmt.Sprintf("\x1b[91m[e]\x1b[%s[%s]\x1b[0m ", r.AnsiColorCode, r.Name),
			io.MultiWriter(allOutput, GinkgoWriter),
		),
	)

	Ω(err).ShouldNot(HaveOccurred())

	fmt.Fprintf(debugWriter, "spawned %s (pid: %d)\n", r.Command.Path, r.Command.Process.Pid)

	r.session = session
	if r.sessionReady != nil {
		close(r.sessionReady)
	}

	startCheckDuration := r.StartCheckTimeout
	if startCheckDuration == 0 {
		startCheckDuration = 5 * time.Second
	}

	var startCheckTimeout <-chan time.Time
	if r.StartCheck != "" {
		startCheckTimeout = time.After(startCheckDuration)
	}

	detectStartCheck := allOutput.Detect(r.StartCheck)

	for {
		select {
		case <-detectStartCheck: // works even with empty string
			allOutput.CancelDetects()
			startCheckTimeout = nil
			detectStartCheck = nil
			close(ready)

		case <-startCheckTimeout:
			// clean up hanging process
			session.Kill().Wait()

			// fail to start
			return fmt.Errorf(
				"did not see %s in command's output within %s. full output:\n\n%s",
				r.StartCheck,
				startCheckDuration,
				string(allOutput.Contents()),
			)

		case signal := <-sigChan:
			session.Signal(signal)

		case <-session.Exited:
			if r.Cleanup != nil {
				r.Cleanup()
			}

			Expect(string(allOutput.Contents())).To(ContainSubstring(r.StartCheck))
			Expect(session.ExitCode()).To(Not(Equal(0)), fmt.Sprintf("Expected process to exit with non-zero, got: 0"))
			return nil
		}
	}
}

var _ = Describe("mapfs Main", func() {
	Context("Missing required args", func() {
		var process ifrit.Process
		It("shows usage", func() {
			var args []string
			driverRunner := failRunner{
				Name:       "mapfs",
				Command:    exec.Command(binaryPath, args...),
				StartCheck: "usage: " + binaryName + " -uid UID -gid GID",
			}
			process = ifrit.Invoke(driverRunner)

		})

		It("shows usage again", func() {
			var args []string = []string{"/foo", "/bar"}

			driverRunner := failRunner{
				Name:       "mapfs",
				Command:    exec.Command(binaryPath, args...),
				StartCheck: "usage: " + binaryName + " -uid UID -gid GID",
			}
			process = ifrit.Invoke(driverRunner)

		})

		It("shows usage even still", func() {
			var args []string = []string{"uid", "0", "gid", "0", "/foo", "/bar"}

			driverRunner := failRunner{
				Name:       "mapfs",
				Command:    exec.Command(binaryPath, args...),
				StartCheck: "usage: " + binaryName + " -uid UID -gid GID",
			}
			process = ifrit.Invoke(driverRunner)
		})

		AfterEach(func() {
			ginkgomon.Kill(process) // this is only if incorrect implementation leaves process running
		})
	})

	Context("when starting succesfully", func() {
		var (
			driverProcess ifrit.Process
			flockProcess  ifrit.Process
		)

		AfterEach(func() {
			ginkgomon.Kill(driverProcess)
			ginkgomon.Kill(flockProcess)
		})

		It("flock works", func() {
			srcDir, err := ioutil.TempDir("", "src")
			Expect(err).NotTo(HaveOccurred())
			targetDir, err := ioutil.TempDir("", "target")
			Expect(err).NotTo(HaveOccurred())
			Expect(os.Chmod(srcDir, os.ModePerm)).To(Succeed())

			driverRunner := ginkgomon.New(
				ginkgomon.Config{
					Name:       "mapfs",
					Command:    exec.Command(binaryPath, "-uid", "1000", "-gid", "1000", targetDir, srcDir),
					StartCheck: "Mounted!",
				},
			)

			driverProcess = ifrit.Invoke(driverRunner)

			By("ensure that mapfs has mounted correctly", func() {
				mountChecker := mountchecker.NewChecker(&bufioshim.BufioShim{}, &osshim.OsShim{})
				mounted, err := mountChecker.Exists(targetDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(mounted).To(BeTrue())
			})

			lockPath := filepath.Join(targetDir, "lockfile")
			_, err = os.OpenFile(lockPath, os.O_RDONLY|os.O_CREATE, 0666)
			Expect(err).NotTo(HaveOccurred())

			flockRunner1 := ginkgomon.New(
				ginkgomon.Config{
					Name:       "flock",
					Command:    exec.Command("flock", lockPath, "-c", "echo success1 && sleep 1"),
					StartCheck: "success1",
				},
			)
			flockProcess = ifrit.Invoke(flockRunner1)

			flockCommand := exec.Command("flock", "-n", lockPath, "-c", "echo ok")
			flockSession, err := gexec.Start(flockCommand, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			<-flockSession.Exited

			Expect(flockSession.ExitCode()).To(Equal(1))
		})
	})
})
