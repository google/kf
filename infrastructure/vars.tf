variable "google_project" {
  type = "string"
}

variable "region" {
  type = "string"
  default = "us-central1"
}

variable "zone" {
  type = "string"
  default = "us-central1-b"
}

variable "gke_cluster_name" {
  type = "string"
}

variable "gke_initial_node_count" {
  default = 1
}