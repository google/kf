variable "project_id" {
  description = "The Google Cloud Project ID"
  type        = string
}

variable "deployment_name" {
  description = "Name prefix for resources (formerly deployment name)"
  type        = string
}

variable "zone" {
  description = "The zone to deploy the cluster into"
  type        = string
}

variable "network" {
  description = "The VPC network to use"
  type        = string
}

variable "cluster_auto_scaling" {
  description = "Enable node pool autoscaling"
  type        = bool
  default     = false
}

variable "cluster_auto_scaling_min_node_count" {
  description = "Minimum nodes if autoscaling is enabled"
  type        = number
  default     = 0
}

variable "cluster_auto_scaling_max_node_count" {
  description = "Maximum nodes if autoscaling is enabled"
  type        = number
  default     = 3
}

variable "initial_node_count" {
  description = "Initial number of nodes"
  type        = number
  default     = 1
}

variable "machine_type" {
  description = "Machine type for the nodes"
  type        = string
  default     = "e2-standard-2"
}

variable "image_type" {
  description = "OS Image type"
  type        = string
  default     = "COS_CONTAINERD"
}

variable "disk_type" {
  description = "Disk type for nodes"
  type        = string
  default     = "pd-standard"
}

variable "disk_size_gb" {
  description = "Disk size in GB"
  type        = number
  default     = 100
}

variable "release_channel" {
  description = "GKE Release Channel (REGULAR, STABLE, RAPID)"
  type        = string
  default     = "REGULAR"
}


provider "google" {
  project = var.project_id
  region  = substr(var.zone, 0, length(var.zone) - 2) # e.g. us-central1-a -> us-central1
  zone    = var.zone
}

# ------------------------------------------------------------------------------
# 1. Service Account
# ------------------------------------------------------------------------------
resource "google_service_account" "kf_sa" {
  account_id   = "${var.deployment_name}-sa"
  display_name = "Kf Cluster ${var.deployment_name}"
}

# ------------------------------------------------------------------------------
# 2. IAM Roles
# ------------------------------------------------------------------------------
locals {
  required_roles = [
    "roles/storage.admin",
    "roles/logging.logWriter",
    "roles/monitoring.metricWriter",
    "roles/iam.serviceAccountAdmin"
  ]
}

resource "google_project_iam_member" "sa_permissions" {
  for_each = toset(local.required_roles)

  project = var.project_id
  role    = each.value
  member  = "serviceAccount:${google_service_account.kf_sa.email}"
}

# ------------------------------------------------------------------------------
# 3. GKE Cluster
# ------------------------------------------------------------------------------
resource "google_container_cluster" "primary" {
  name     = var.deployment_name
  location = var.zone
  
  remove_default_node_pool = true
  initial_node_count       = 1
  deletion_protection      = false

  network = var.network

  # Networking
  ip_allocation_policy {
    # intentionally left blank to enable VPC-native (alias IP) networking
  }

  network_policy {
    provider = "CALICO"
    enabled  = true
  }

  addons_config {
    http_load_balancing {
      disabled = false
    }
    horizontal_pod_autoscaling {
      disabled = false
    }
    network_policy_config {
      disabled = false
    }
  }

  maintenance_policy {
    daily_maintenance_window {
      start_time = "08:00"
    }
  }

  logging_service    = "logging.googleapis.com/kubernetes"
  monitoring_service = "monitoring.googleapis.com/kubernetes"

  release_channel {
    channel = var.release_channel
  }

  workload_identity_config {
    workload_pool = "${var.project_id}.svc.id.goog"
  }
}

# ------------------------------------------------------------------------------
# 4. Node Pool
# ------------------------------------------------------------------------------
resource "google_container_node_pool" "primary_nodes" {
  name       = var.deployment_name
  location   = var.zone
  cluster    = google_container_cluster.primary.name
  
  initial_node_count = var.initial_node_count

  dynamic "autoscaling" {
    for_each = var.cluster_auto_scaling ? [1] : []
    content {
      min_node_count = var.cluster_auto_scaling_min_node_count
      max_node_count = var.cluster_auto_scaling_max_node_count
    }
  }

  management {
    auto_repair  = true
    auto_upgrade = true
  }

  node_config {
    machine_type = var.machine_type
    image_type   = var.image_type
    disk_type    = var.disk_type
    disk_size_gb = var.disk_size_gb
    
    service_account = google_service_account.kf_sa.email

    oauth_scopes = [
      "https://www.googleapis.com/auth/compute",
      "https://www.googleapis.com/auth/devstorage.read_only",
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring"
    ]

    metadata = {
      disable-legacy-endpoints = "true"
    }
    
    workload_metadata_config {
      mode = "GKE_METADATA"
    }
  }
}

# ------------------------------------------------------------------------------
# 5. Legacy Service Account Permissions
#    This ensures the 'Google APIs Service Agent' has Owner permissions.
# ------------------------------------------------------------------------------
data "google_project" "current" {
  project_id = var.project_id
}

resource "google_project_iam_member" "google_apis_sa_owner" {
  project = var.project_id
  role    = "roles/owner"
  # The Python script used regex to find '@cloudservices.gserviceaccount.com'.
  # This is the standard Google APIs Service Agent email format.
  member  = "serviceAccount:${data.google_project.current.number}@cloudservices.gserviceaccount.com"
}