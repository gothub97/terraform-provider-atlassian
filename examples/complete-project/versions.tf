terraform {
  required_version = ">= 1.0"

  required_providers {
    atlassian = {
      source  = "gothub97/atlassian"
      version = ">= 0.2.0"
    }
  }
}

provider "atlassian" {
  # Configure via environment variables:
  #   ATLASSIAN_URL       - Your Jira Cloud URL (e.g., https://mysite.atlassian.net)
  #   ATLASSIAN_EMAIL     - Your Atlassian account email
  #   ATLASSIAN_API_TOKEN - Your Atlassian API token
}
