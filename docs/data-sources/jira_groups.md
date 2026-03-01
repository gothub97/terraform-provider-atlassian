---
page_title: "atlassian_jira_groups Data Source - atlassian"
subcategory: "Jira"
description: |-
  Retrieves all Jira groups, optionally filtered by a name prefix.
---

# atlassian_jira_groups (Data Source)

Use this data source to retrieve Jira groups, optionally filtered by a name prefix. This is useful for discovering group IDs needed when configuring group memberships or permission schemes.

## Example Usage

### Retrieve All Groups

```hcl
data "atlassian_jira_groups" "all" {
}
```

### Retrieve Groups Matching a Prefix

```hcl
data "atlassian_jira_groups" "dev" {
  query = "dev"
}
```

## Argument Reference

* `query` - (Optional) A name prefix filter for groups. If not set, all groups are returned.

## Attributes Reference

* `groups` - The list of groups. Each group has the following attributes:
  * `group_id` - The ID of the group.
  * `name` - The name of the group.
