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

variable "unit_tests_branch" {
  type = string
  default = ".*"
}

variable "integ_tests_branch_regex" {
  type = string
  default = "^main$"
}

variable "daily_tests_branch" {
  type = string
  default = "main"
}

variable "release_channels" {
  type    = list(string)
  default = ["stable", "regular", "rapid"]
}

variable "revisions_to_test" {
  type    = map
  default = {
    "head" = 0
    "current-release" = 1
    "prev-release" = 2
  }
}

