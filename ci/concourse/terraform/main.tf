variable "project" {
  type = string
}

variable "gke_version" {
  type = string
}

provider "google" {
  project = var.project
  region  = "us-central1"
}

provider "google-beta" {
  project = var.project
  region  = "us-central1"
}

resource "random_pet" "kf_test" {
}

resource "google_compute_network" "k8s_network" {
  name                    = "kf-test-${random_pet.kf_test.id}"
  description             = "Managed by Terraform in Concourse"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "kf_apps" {
  name                     = "kf-apps-${random_pet.kf_test.id}"
  description              = "Managed by Terraform in Concourse"
  ip_cidr_range            = "10.0.0.0/16"
  region                   = "us-central1"
  private_ip_google_access = true
  network                  = "${google_compute_network.k8s_network.self_link}"
}

resource "google_service_account" "kf_test" {
  account_id   = "kf-test-${random_pet.kf_test.id}"
  display_name = "Managed by Terraform in Concourse"
}

resource "google_project_iam_member" "kf_test" {
  role   = "roles/storage.admin"
  member = "serviceAccount:${google_service_account.kf_test.email}"
}

resource "google_service_account_key" "kf_test" {
  service_account_id = "${google_service_account.kf_test.name}"
}

resource "google_container_cluster" "kf_test" {
  provider = "google-beta"
  name     = "kf-test-${random_pet.kf_test.id}"
  location = "us-central1"

  min_master_version = var.gke_version

  initial_node_count = 1

  master_auth {
    username = ""
    password = ""

    client_certificate_config {
      issue_client_certificate = false
    }
  }

  addons_config {
    istio_config {
      disabled = false
    }
    cloudrun_config {
      disabled = false
    }
    http_load_balancing {
      disabled = false
    }
  }

  # These services must be set for Cloud Run to work correctly.
  logging_service    = "logging.googleapis.com/kubernetes"
  monitoring_service = "monitoring.googleapis.com/kubernetes"


  node_config {
    machine_type = "n1-standard-4"

    metadata = {
      disable-legacy-endpoints = "true"
    }

    service_account = "${google_service_account.kf_test.email}"

    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform",
      "https://www.googleapis.com/auth/userinfo.email",
    ]
  }

  network    = "${google_compute_network.k8s_network.self_link}"
  subnetwork = "${google_compute_subnetwork.kf_apps.self_link}"
}

output "cluster_name" {
  value = google_container_cluster.kf_test.name
}

output "cluster_region" {
  value = google_container_cluster.kf_test.location
}

output "cluster_project" {
  value = var.project
}

output "cluster_version" {
  value = google_container_cluster.kf_test.master_version
}

output "cluster_network" {
  value = google_compute_network.k8s_network.name
}

output "gcr_key" {
  value     = "${base64decode(google_service_account_key.kf_test.private_key)}"
  sensitive = true
}
