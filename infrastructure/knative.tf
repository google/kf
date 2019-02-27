resource "null_resource" "deploy_knative" {
  provisioner "local-exec" {
    command = "./deploy_knative.sh ${var.gke_cluster_name} ${var.zone} ${var.google_project}"
  }
}
