variable "project_key" {
  type        = string
  default     = "TFDEMO"
  description = "The Jira project key (2-10 uppercase alphanumeric characters, starting with a letter)."

  validation {
    condition     = can(regex("^[A-Z][A-Z0-9]{1,9}$", var.project_key))
    error_message = "Project key must be 2-10 uppercase alphanumeric characters, starting with a letter."
  }
}

variable "project_name" {
  type        = string
  default     = "Terraform Demo Project"
  description = "The display name for the Jira project."
}

variable "team_name" {
  type        = string
  default     = "tf-demo-team"
  description = "The name of the team group to create."
}
