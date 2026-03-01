---
page_title: "atlassian_jira_security_schemes Data Source - atlassian"
subcategory: "Jira"
description: |-
  Retrieves all Jira issue security schemes.
---

# atlassian_jira_security_schemes (Data Source)

Use this data source to retrieve all Jira issue security schemes. This is useful for discovering scheme IDs needed when assigning security schemes to projects.

## Example Usage

```hcl
data "atlassian_jira_security_schemes" "all" {
}
```

## Argument Reference

This data source has no required or optional arguments.

## Attributes Reference

* `schemes` - The list of issue security schemes. Each scheme has the following attributes:
  * `id` - The ID of the security scheme.
  * `name` - The name of the security scheme.
  * `description` - The description of the security scheme.
