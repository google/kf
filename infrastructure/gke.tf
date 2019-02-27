provider "google" {
  project = "${var.google_project}"
  region  = "${var.region}"
}

data "google_container_engine_versions" "default" {
  zone = "${var.zone}"
}

data "google_client_config" "current" {}

# https://github.com/knative/docs/blob/master/install/Knative-with-GKE.md#creating-a-kubernetes-cluster
resource "google_container_cluster" "default" {
  name = "${var.gke_cluster_name}"
  zone = "${var.zone}"
  initial_node_count = "${var.gke_initial_node_count}"
  min_master_version = "${data.google_container_engine_versions.default.latest_master_version}"

  node_config {
    image_type = "COS"
    machine_type = "n1-standard-16"
  }

  // Wait for the GCE LB controller to cleanup the resources.
  provisioner "local-exec" {
    when    = "destroy"
    command = "sleep 90"
  }
}

data "google_compute_instance_group" "default" {
  self_link = "${google_container_cluster.default.instance_group_urls.0}"
}

data "google_compute_instance" "i0" {
  name = "${replace(data.google_compute_instance_group.default.instances[0], "/.*//", "")}"
  zone = "${var.zone}"
}
