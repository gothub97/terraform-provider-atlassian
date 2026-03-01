---
page_title: "atlassian_jira_project_role_actor Resource - atlassian"
subcategory: "Jira"
description: |-
  Manages a Jira project role actor (user or group membership).
---

# atlassian_jira_project_role_actor (Resource)

Manages a Jira project role actor, which assigns a user or group to a project role within a specific project. All attributes are immutable; changing any attribute will destroy and recreate the actor assignment.

## Example Usage

```hcl
# Assign a user to a project role
resource "atlassian_jira_project_role_actor" "lead_user" {
  project_key = "PROJ"
  role_id     = "10360"
  actor_type  = "user"
  actor_value = "5b10ac8d82e05b22cc7d4ef5"
}

# Assign a group to a project role
resource "atlassian_jira_project_role_actor" "dev_group" {
  project_key = "PROJ"
  role_id     = "10360"
  actor_type  = "group"
  actor_value = atlassian_jira_group.developers.id
}
```

## Argument Reference

* `project_key` - (Required, Forces new resource) The project key.
* `role_id` - (Required, Forces new resource) The numeric ID of the project role.
* `actor_type` - (Required, Forces new resource) The type of actor. Valid values are `user` (maps to `atlassian-user-role-actor`) or `group` (maps to `atlassian-group-role-actor`).
* `actor_value` - (Required, Forces new resource) The account ID (for user) or group ID (for group).

## Attributes Reference

* `id` - Composite ID in the format `projectKey/roleId/actorType/actorValue`.

## Import

Import using the format `projectKey/roleId/actorType/actorValue`:

```shell
terraform import atlassian_jira_project_role_actor.example PROJ/10360/user/5b10ac8d82e05b22cc7d4ef5
```
