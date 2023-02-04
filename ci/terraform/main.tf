
resource "google_storage_bucket" "test_results" {
  project                     = var.project_id
  name                        = "${var.project_id}-test-results"
  location                    = "US"
  uniform_bucket_level_access = true
}
# Create a storage bucket to hold build artifacts
resource "google_storage_bucket" "test_results" {
  project                     = var.project_id
  name                        = "${var.project_id}-test-results"
  location                    = "US"
  uniform_bucket_level_access = true
}

# Create scheduled cloudbuild trigger to reap gke clusters
resource "google_pubsub_topic" "cloudbuild_cron_cleanup" {
  name = "cloudbuild_cron_cleanup"
  project  = var.project_id
}

resource "google_cloud_scheduler_job" "cloudbuild_cron_cleanup" {
  name        = "cloudbuild_cron_cleanup"
  project     = var.project_id
  region      = "us-central1"
  # daily at 1 AM
  schedule    = "0 1 * * *"
  time_zone   = "America/Los_Angeles"

  pubsub_target {
    topic_name = google_pubsub_topic.cloudbuild_cron_cleanup.id
    data       = base64encode("trigger cloudbuild")
  }
}

resource "google_cloudbuild_trigger" "reap_gke_clusters" {
  location    = "global"
  project     = var.project_id
  name        = "kf-reap-gke-clusters"
  description = "Cleanup gke clusters, triggered by pubsub nightly at 1:00 AM"

  pubsub_config {
    topic = google_pubsub_topic.cloudbuild_cron_cleanup.id
  }

  source_to_build {
    uri       = "https://github.com/${var.repo_owner}/${var.repo_name}"
    ref       = "refs/heads/${var.daily_tests_branch}"
    repo_type = "GITHUB"
  }

  git_file_source {
    path      = "ci/cloudbuild/scheduled/reap-gke-clusters/cleanup-cron.yaml"
    uri       = "https://github.com/${var.repo_owner}/${var.repo_name}"
    revision  = "refs/heads/${var.daily_tests_branch}"
    repo_type = "GITHUB"
  }
}

# Create scheduled cloudbuild trigger to run integration tests
resource "google_pubsub_topic" "cloudbuild_cron_integ_tests" {
  name = "cloudbuild_cron_integ_tests"
  project  = var.project_id
}

resource "google_cloud_scheduler_job" "cloudbuild_cron_integ_tests" {
  name        = "cloudbuild_cron_integ_tests"
  project     = var.project_id
  region      = "us-central1"
  # daily at 9 AM
  schedule    = "0 9 * * *"
  time_zone   = "America/Los_Angeles"

  pubsub_target {
    topic_name = google_pubsub_topic.cloudbuild_cron_integ_tests.id
    data       = base64encode("trigger cloudbuild")
  }
}

resource "google_cloudbuild_trigger" "integ_tests_daily" {
  count = length(var.release_channels)
  location    = "global"
  project     = var.project_id
  name        = "kf-integ-tests-daily-${var.release_channels[count.index]}"
  description = "Run integ tests daily at 9:00 AM on ${var.release_channels[count.index]} release channel."

  pubsub_config {
    topic = google_pubsub_topic.cloudbuild_cron_integ_tests.id
  }

  source_to_build {
    uri       = "https://github.com/${var.repo_owner}/${var.repo_name}"
    ref       = "refs/heads/${var.daily_tests_branch}"
    repo_type = "GITHUB"
  }

  git_file_source {
    path      = "ci/cloudbuild/release-and-test.yaml"
    uri       = "https://github.com/${var.repo_owner}/${var.repo_name}"
    revision       = "refs/heads/${var.daily_tests_branch}"
    repo_type = "GITHUB"
  }

  substitutions = {
    _RELEASE_BUCKET  = google_storage_bucket.build_artifacts.name
    _RELEASE_CHANNEL = "${var.release_channels[count.index]}"
    _EXPORT_BUCKET = "${google_storage_bucket.test_results.name}"
    _EXPORT_JOB_NAME = "integ-test-${var.release_channels[count.index]}"
  }
}

# Create on push cloudbuild triggers to run unit tests
resource "google_cloudbuild_trigger" "unit_tests_on_push" {
  name     = "kf-unit-tests-on-push"
  provider = google-beta
  location = "global"
  project  = var.project_id

  filename = "ci/cloudbuild/unit-test.yaml"

  github {
    owner = var.repo_owner
    name  = var.repo_name
    push {
      branch = var.unit_tests_branch
    }
  }

  include_build_logs = "INCLUDE_BUILD_LOGS_WITH_STATUS"
}

# Create on push cloudbuild triggers to run integration tests.
resource "google_cloudbuild_trigger" "integ_tests_on_push" {
  count = length(var.release_channels)

  name     = "kf-integ-tests-on-push-${var.release_channels[count.index]}"
  provider = google-beta
  location = "global"
  project  = var.project_id

  filename = "ci/cloudbuild/release-and-test.yaml"

  github {
    owner = var.repo_owner
    name  = var.repo_name
    push {
      branch = var.integ_tests_branch_regex
    }
  }

  substitutions = {
    _RELEASE_BUCKET  = google_storage_bucket.build_artifacts.name
    _RELEASE_CHANNEL = "${var.release_channels[count.index]}"
  }

  include_build_logs = "INCLUDE_BUILD_LOGS_WITH_STATUS"
}
