variable "project_id" {
  type = string
}

variable "repo_owner" {
  type = string
  default = "google"
}

variable "repo_name" {
  type = string 
  default = "kf"
}

variable "repo_branch" {
  type = string 
  default = ".*"
}

variable "release_channels" {
  type    = list(string)
  default = ["stable", "regular", "rapid"]
}

