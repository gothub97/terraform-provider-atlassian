---
page_title: "atlassian_jira_field Resource - atlassian"
subcategory: "Jira"
description: |-
  Manages a Jira custom field.
---

# atlassian_jira_field (Resource)

Manages a Jira custom field. This resource allows you to create, update, and delete custom fields in your Jira instance.

Custom fields extend the default set of fields available on Jira issues, enabling you to capture additional information specific to your workflows.

## Example Usage

```hcl
resource "atlassian_jira_field" "story_points" {
  name         = "Story Points"
  type         = "com.atlassian.jira.plugin.system.customfieldtypes:float"
  description  = "Estimated story points for the issue"
  searcher_key = "com.atlassian.jira.plugin.system.customfieldtypes:exactnumber"
}

resource "atlassian_jira_field" "team_name" {
  name = "Team Name"
  type = "com.atlassian.jira.plugin.system.customfieldtypes:textfield"
}
```

## Argument Reference

* `name` - (Required) The name of the custom field.
* `type` - (Required, Forces new resource) The type of the custom field (e.g., `com.atlassian.jira.plugin.system.customfieldtypes:float`). Cannot be changed after creation.
* `description` - (Optional) A description of the custom field.
* `searcher_key` - (Optional) The searcher key for the custom field (e.g., `com.atlassian.jira.plugin.system.customfieldtypes:exactnumber`).

## Attributes Reference

* `id` - The ID of the field (e.g., `customfield_10001`).

## Import

Import using the custom field ID:

```shell
terraform import atlassian_jira_field.example customfield_10001
```
