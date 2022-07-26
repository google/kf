package integration

//go:generate mockgen --package=integration --copyright_file ../../internal/tools/option-builder/LICENSE_HEADER --destination=fake_metric_logger.go --mock_names=MetricLogger=FakeMetricLogger github.com/google/kf/v2/pkg/kf/testutil/integration MetricLogger
