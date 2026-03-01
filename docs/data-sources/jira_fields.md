---
page_title: "atlassian_jira_fields Data Source - atlassian"
subcategory: "Jira"
description: |-
  Retrieves all Jira fields, optionally filtered by type (system or custom).
---

# atlassian_jira_fields (Data Source)

Use this data source to retrieve all Jira fields, optionally filtered by type. This is useful for discovering field IDs needed when configuring issue types, screens, or custom field contexts.

## Example Usage

### Retrieve All Fields

```hcl
data "atlassian_jira_fields" "all" {
}
```

### Retrieve Only Custom Fields

```hcl
data "atlassian_jira_fields" "custom" {
  type = "custom"
}
```

### Retrieve Only System Fields

```hcl
data "atlassian_jira_fields" "system" {
  type = "system"
}
```

## Argument Reference

* `type` - (Optional) Filter by field type. Must be `"system"` or `"custom"`. If not set, all fields are returned.

## Attributes Reference

* `fields` - The list of fields. Each field has the following attributes:
  * `id` - The ID of the field (e.g., `"summary"`, `"customfield_10001"`).
  * `name` - The display name of the field.
  * `custom` - Whether the field is a custom field (`true`) or a system field (`false`).
  * `schema_type` - The schema type of the field (e.g., `"string"`, `"array"`, `"number"`). May be null if the field has no schema.
  * `clause_names` - A list of clause names that can be used to reference this field in JQL queries.
