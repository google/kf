resource "google_storage_bucket" "build-artifacts" {
  project                     = var.project_id
  name                        = "${var.project_id}-build-artifacts"
  location                    = "US"
  force_destroy               = true
  uniform_bucket_level_access = true
}

resource "google_cloudbuild_trigger" "unit-tests-on-push" {
  name     = "kf-unit-tests-on-push"
  provider = google-beta
  location = "global"
  project  = var.project_id

  filename = "ci/cloudbuild/unit-test.yaml"

  github {
    owner = var.repo_owner
    name  = var.repo_name
    push {
      branch = var.repo_branch
    }
  }
}

resource "google_cloudbuild_trigger" "integ-tests-on-push" {
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
      branch = var.repo_branch
    }
  }

  substitutions = {
    _RELEASE_BUCKET  = google_storage_bucket.build-artifacts.name
    _RELEASE_CHANNEL = "${var.release_channels[count.index]}"
  }

  # Avoid wasting resources each time the PR is updated by 
  # only running builds that are explicitly approved.
  approval_config {
     approval_required = true 
  }
}
