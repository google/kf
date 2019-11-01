provider "google-beta"{
  credentials = "${file("account.json")}"
  project = "kf-source"
  region = "us-central1"
  zone   = "us-central1-a"
}

provider "google" {
  credentials = "${file("account.json")}"
  project     = "kf-source"
  region      = "us-west1"
}

data "google_compute_network" "default" {
  name = "${var.vpc_network_name}"
}
