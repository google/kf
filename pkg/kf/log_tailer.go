package kf

import (
	"context"
	"fmt"
	"io"

	build "github.com/knative/build/pkg/apis/build/v1alpha1"
	cbuild "github.com/knative/build/pkg/client/clientset/versioned/typed/build/v1alpha1"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	k8smeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

// BuildFactory returns a bclient for build.
type BuildFactory func() (cbuild.BuildV1alpha1Interface, error)

// BuildTail writes the build logs to out.
type BuildTail func(ctx context.Context, out io.Writer, buildName, namespace string) error

// LogTailer tails logs for a service. This includes the build and deploy
// step. It should be created via NewLogTailer.
type LogTailer struct {
	f  ServingFactory
	bf BuildFactory
	t  BuildTail
}

// NewLogTailer creates a new LogTailer.
func NewLogTailer(bf BuildFactory, f ServingFactory, t BuildTail) *LogTailer {
	return &LogTailer{
		bf: bf,
		f:  f,
		t:  t,
	}
}

// Tail writes the logs for the build and deploy step for the resourceVersion
// to out. It blocks until the operation has completed.
func (t LogTailer) Tail(out io.Writer, resourceVersion, namespace string) error {
	errs := make(chan error, 2)

	ctx, cancel := context.WithCancel(context.Background())

	go func() { errs <- t.buildLogs(ctx, out, resourceVersion, namespace) }()
	go func() { errs <- t.serviceLogs(cancel, out, resourceVersion, namespace) }()

	err1, err2 := <-errs, <-errs
	for _, err := range []error{err1, err2} {
		if err != nil {
			return err
		}
	}

	return nil
}

func (t LogTailer) buildLogs(ctx context.Context, out io.Writer, resourceVersion, namespace string) error {
	bclient, err := t.bf()
	if err != nil {
		return err
	}

	wb, err := bclient.Builds(namespace).Watch(k8smeta.ListOptions{
		ResourceVersion: resourceVersion,
	})
	if err != nil {
		return err
	}
	defer wb.Stop()

	go func() {
		<-ctx.Done()
		wb.Stop()
	}()

	for e := range wb.ResultChan() {
		obj := e.Object.(*build.Build)
		if e.Type == watch.Added {
			// This blocks until the build is finished.
			if err := t.t(ctx, out, obj.Name, namespace); err != nil {
				return err
			}
		}

		for _, condition := range obj.Status.Conditions {
			if condition.Type == "Succeeded" && condition.Status == "False" {
				return fmt.Errorf("build failed: %s", condition.Message)
			}
		}
	}

	return nil
}

func (t LogTailer) serviceLogs(done func(), out io.Writer, resourceVersion, namespace string) error {
	defer done()
	sclient, err := t.f()
	if err != nil {
		return err
	}

	ws, err := sclient.Services(namespace).Watch(k8smeta.ListOptions{
		ResourceVersion: resourceVersion,
	})
	if err != nil {
		return err
	}
	defer ws.Stop()

	for e := range ws.ResultChan() {
		for _, condition := range e.Object.(*serving.Service).Status.Conditions {
			if condition.Reason == "RevisionFailed" {
				return fmt.Errorf("deployment failed: %s", condition.Message)
			}

			if condition.Message != "" {
				fmt.Fprintf(out, "\033[32m[deploy-revision]\033[0m %s\n", condition.Message)
			}
		}
	}
	// Success
	return nil
}
