# Create a storage bucket to hold build artifacts
resource "google_storage_bucket" "build_artifacts" {
  project                     = var.project_id
  name                        = "${var.project_id}-build-artifacts"
  location                    = "US"
  uniform_bucket_level_access = true
}

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
  for_each = toset(var.release_channels)
  name = "cloudbuild_cron_integ_tests_${each.value}"
  project  = var.project_id
}

resource "google_cloud_scheduler_job" "cloudbuild_cron_integ_tests" {
  for_each = toset(var.release_channels)
  name = "cloudbuild_cron_integ_tests_${each.value}"

  project     = var.project_id
  region      = "us-central1"
  # Daily starting at 2 AM staggered by 2 hours to avoid thundering herd.
  schedule    = "0 ${index(var.release_channels, each.value) * 2 + 2} * * *"
  time_zone   = "America/Los_Angeles"

  pubsub_target {
    topic_name = google_pubsub_topic.cloudbuild_cron_integ_tests[each.value].id
    data       = base64encode("trigger cloudbuild")
  }
}

resource "google_cloudbuild_trigger" "integ_tests_daily" {
  for_each = {
    for v in setproduct(var.release_channels, [for k,v in var.revisions_to_test: [k, v]]): 
    "${v[0]}-${v[1][0]}" => {
      "channel" = v[0]
      "revision_idx" = v[1][1]
    }
  }
  location    = "global"
  project     = var.project_id
  name        = "kf-integ-tests-daily-${each.key}"
  description = "Run integ tests daily for configuration ${each.key}"

  pubsub_config {
    topic = google_pubsub_topic.cloudbuild_cron_integ_tests[each.value.channel].id
  }

  source_to_build {
    uri       = "https://github.com/${var.repo_owner}/${var.repo_name}"
    ref       = "refs/heads/${var.daily_tests_branch}"
    repo_type = "GITHUB"
  }

  git_file_source {
    path      = "ci/cloudbuild/release-and-test.yaml"
    uri       = "https://github.com/${var.repo_owner}/${var.repo_name}"
    revision  = "refs/heads/${var.daily_tests_branch}"
    repo_type = "GITHUB"
  }

  substitutions = {
    _RELEASE_BUCKET  = google_storage_bucket.build_artifacts.name
    _RELEASE_CHANNEL = "${each.value.channel}"
    _EXPORT_BUCKET = "${google_storage_bucket.test_results.name}"
    _EXPORT_JOB_NAME = "integ-test-${each.key}"
    _TAGGED_RELEASE_INDEX = "${each.value.revision_idx}"
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
