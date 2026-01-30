---
page_title: "Data Source mongodb_users - mongodb"
subcategory: ""
description: |-
  Retrieves a list of all MongoDB users with their roles and permissions, excluding system users.
---

# Data Source: mongodb_users

Retrieves a list of all MongoDB users with their roles and permissions, excluding system users.

Retrieves a list of all MongoDB users with their roles and permissions, excluding system users.

Use this data source to audit user permissions, verify user creation, or configure dependent resources.

~> **Security Note:** This data source does not return user passwords. Passwords are write-only in MongoDB.

## Example Usage

### List All Users

```terraform
data "mongodb_users" "all" {}

output "user_list" {
  value = [for user in data.mongodb_users.all.users : user.username]
}
```

### List Users from Specific Authentication Source

```terraform
data "mongodb_users" "custom_db" {
  auth_source = "users_db"
}

output "custom_db_usernames" {
  value = [for user in data.mongodb_users.custom_db.users : user.username]
}
```

### Find User by Username

```terraform
data "mongodb_users" "all" {}

locals {
  target_user = "app_service"
  user_exists = contains(
    [for user in data.mongodb_users.all.users : user.username],
    local.target_user
  )

  user_details = [
    for user in data.mongodb_users.all.users :
    user if user.username == local.target_user
  ]
}

output "app_service_user" {
  value = local.user_exists ? local.user_details[0] : null
}
```

### Audit User Permissions

```terraform
data "mongodb_users" "all" {}

locals {
  # Find users with admin access
  admin_users = [
    for user in data.mongodb_users.all.users :
    user.username if contains(
      [for role in user.roles : role.role],
      "root"
    )
  ]

  # Find users by database
  production_users = [
    for user in data.mongodb_users.all.users :
    {
      username = user.username
      roles    = [
        for role in user.roles :
        role.role if role.database == "production"
      ]
    } if length([for role in user.roles : role if role.database == "production"]) > 0
  ]
}

output "admin_users" {
  value = local.admin_users
}

output "production_db_users" {
  value = local.production_users
}
```

### Generate User Report

```terraform
data "mongodb_users" "all" {}

locals {
  user_report = {
    for user in data.mongodb_users.all.users :
    user.username => {
      databases = distinct([for role in user.roles : role.database])
      roles     = distinct([for role in user.roles : role.role])
      role_count = length(user.roles)
    }
  }
}

output "user_audit_report" {
  value = local.user_report
}
```

### Verify User Creation

```terraform
resource "mongodb_user" "app" {
  username = "application"
  password = var.app_password

  roles = [
    {
      role     = "readWrite"
      database = "myapp"
    }
  ]
}

data "mongodb_users" "verify" {
  depends_on = [mongodb_user.app]
}

output "user_created" {
  value = contains(
    [for user in data.mongodb_users.verify.users : user.username],
    mongodb_user.app.username
  )
}
```

### Find Users with Specific Permissions

```terraform
data "mongodb_users" "all" {}

locals {
  # Find all users with write access to any database
  write_users = [
    for user in data.mongodb_users.all.users :
    user.username if length([
      for role in user.roles :
      role if contains(["readWrite", "dbOwner", "root"], role.role)
    ]) > 0
  ]

  # Find users with read-only access
  readonly_users = [
    for user in data.mongodb_users.all.users :
    user.username if length([
      for role in user.roles :
      role if role.role == "read"
    ]) > 0 && length([
      for role in user.roles :
      role if contains(["readWrite", "dbOwner", "root"], role.role)
    ]) == 0
  ]
}

output "users_with_write_access" {
  value = local.write_users
}

output "readonly_users" {
  value = local.readonly_users
}
```

### Security Compliance Check

```terraform
data "mongodb_users" "all" {}

locals {
  # Check for overprivileged users
  overprivileged_users = [
    for user in data.mongodb_users.all.users :
    user.username if contains(
      [for role in user.roles : role.role],
      "root"
    ) || contains(
      [for role in user.roles : role.role],
      "dbAdminAnyDatabase"
    )
  ]

  # Check for users without database restrictions
  unrestricted_users = [
    for user in data.mongodb_users.all.users :
    user.username if contains(
      [for role in user.roles : role.database],
      "admin"
    ) && length([
      for role in user.roles :
      role if contains(["root", "readWriteAnyDatabase", "dbAdminAnyDatabase"], role.role)
    ]) > 0
  ]

  compliance_issues = {
    overprivileged_count = length(local.overprivileged_users)
    overprivileged_users = local.overprivileged_users
    unrestricted_count   = length(local.unrestricted_users)
    unrestricted_users   = local.unrestricted_users
  }
}

output "security_compliance" {
  value = local.compliance_issues
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `auth_source` (String) The authentication source to use for the users. Default is 'admin'.

### Read-Only

- `users` (Attributes List) (see [below for nested schema](#nestedatt--users))

<a id="nestedatt--users"></a>
### Nested Schema for `users`

Optional:

- `timeouts` (Attributes) (see [below for nested schema](#nestedatt--users--timeouts))

Read-Only:

- `auth_source` (String) The authentication source of the user.
- `password` (String, Sensitive) The password of the user.
- `roles` (Attributes List) (see [below for nested schema](#nestedatt--users--roles))
- `username` (String) The username of the user.

<a id="nestedatt--users--timeouts"></a>
### Nested Schema for `users.timeouts`


<a id="nestedatt--users--roles"></a>
### Nested Schema for `users.roles`

Read-Only:

- `database` (String) The database of the user.
- `role` (String) The role of the user.

## Attribute Reference

The following attributes are exported:

- `users` - (List of Objects) List of MongoDB users. Each object contains:
  - `username` - (String) The username.
  - `password` - (String) Empty string (passwords are not readable from MongoDB).
  - `roles` - (List of Objects) List of roles assigned to the user:
    - `role` - (String) The role name (e.g., "read", "readWrite", "dbAdmin").
    - `database` - (String) The database the role applies to.

-> **Note:** The `password` field is always empty. MongoDB does not expose passwords via API.

## Behavior

### Excluded Users

System users are automatically excluded from results (e.g., `admin` user if it's a default system user).

### Role Information

Each user object includes complete role information:
- Role name (e.g., `read`, `readWrite`, `dbAdmin`)
- Database scope for each role
- Multiple roles per user are supported

### Refresh Behavior

This data source refreshes on every `terraform plan` and `terraform apply`. Use `terraform refresh` to update data source state.

## Common Use Cases

### 1. User Access Audit

Generate comprehensive access report:

```terraform
data "mongodb_users" "all" {}

resource "local_file" "access_audit" {
  filename = "${path.module}/user_access_audit.json"
  content = jsonencode({
    audit_date = timestamp()
    total_users = length(data.mongodb_users.all.users)
    users = [
      for user in data.mongodb_users.all.users : {
        username  = user.username
        databases = distinct([for role in user.roles : role.database])
        roles     = [
          for role in user.roles : {
            role     = role.role
            database = role.database
          }
        ]
      }
    ]
  })
}
```

### 2. Least Privilege Verification

Check for overprivileged accounts:

```terraform
data "mongodb_users" "all" {}

locals {
  dangerous_roles = ["root", "dbAdminAnyDatabase", "userAdminAnyDatabase"]

  users_to_review = [
    for user in data.mongodb_users.all.users : {
      username = user.username
      dangerous_permissions = [
        for role in user.roles :
        role if contains(local.dangerous_roles, role.role)
      ]
    } if length([
      for role in user.roles :
      role if contains(local.dangerous_roles, role.role)
    ]) > 0
  ]
}

output "security_review_required" {
  value = length(local.users_to_review) > 0 ? local.users_to_review : "No overprivileged users found"
}
```

### 3. Drift Detection

Identify users created outside Terraform:

```terraform
data "mongodb_users" "current" {}

locals {
  managed_users = [
    for user in mongodb_user.managed : user.username
  ]

  unmanaged_users = [
    for user in data.mongodb_users.current.users :
    user.username if !contains(local.managed_users, user.username)
  ]
}

output "unmanaged_users_warning" {
  value = length(local.unmanaged_users) > 0 ? "Warning: Found ${length(local.unmanaged_users)} unmanaged users: ${join(", ", local.unmanaged_users)}" : "All users are managed by Terraform"
}
```

### 4. Database Access Matrix

Create access matrix for all users and databases:

```terraform
data "mongodb_users" "all" {}
data "mongodb_databases" "all" {}

locals {
  access_matrix = {
    for db in data.mongodb_databases.all.databases :
    db.name => {
      read_users = [
        for user in data.mongodb_users.all.users :
        user.username if contains(
          [for role in user.roles : "${role.database}:${role.role}"],
          "${db.name}:read"
        )
      ]
      write_users = [
        for user in data.mongodb_users.all.users :
        user.username if contains(
          [for role in user.roles : "${role.database}:${role.role}"],
          "${db.name}:readWrite"
        )
      ]
      admin_users = [
        for user in data.mongodb_users.all.users :
        user.username if contains(
          [for role in user.roles : "${role.database}:${role.role}"],
          "${db.name}:dbOwner"
        )
      ]
    }
  }
}

output "database_access_matrix" {
  value = local.access_matrix
}
```

### 5. Compliance Reporting

Generate compliance report:

```terraform
data "mongodb_users" "all" {}

locals {
  compliance_report = {
    report_date = timestamp()
    metrics = {
      total_users               = length(data.mongodb_users.all.users)
      admin_users              = length([for u in data.mongodb_users.all.users : u if contains([for r in u.roles : r.role], "root")])
      readonly_users           = length([for u in data.mongodb_users.all.users : u if contains([for r in u.roles : r.role], "read")])
      multi_role_users         = length([for u in data.mongodb_users.all.users : u if length(u.roles) > 1])
      single_database_users    = length([for u in data.mongodb_users.all.users : u if length(distinct([for r in u.roles : r.database])) == 1])
      multi_database_users     = length([for u in data.mongodb_users.all.users : u if length(distinct([for r in u.roles : r.database])) > 1])
    }
  }
}

output "compliance_metrics" {
  value = local.compliance_report
}
```

## Best Practices

### 1. Regular Audits

Schedule regular user audits:

```terraform
data "mongodb_users" "audit" {}

resource "null_resource" "monthly_audit" {
  triggers = {
    # Run monthly
    month = formatdate("YYYY-MM", timestamp())
  }

  provisioner "local-exec" {
    command = "echo 'User count: ${length(data.mongodb_users.audit.users)}' >> audit.log"
  }
}
```

### 2. Combine with Resources

Use data source output to configure related resources:

```terraform
data "mongodb_users" "all" {}

# Grant read access to all existing databases for a reporting user
resource "mongodb_user" "reporter" {
  username = "global_reporter"
  password = var.reporter_password

  roles = [
    for db in distinct(flatten([
      for user in data.mongodb_users.all.users : [
        for role in user.roles : role.database
      ]
    ])) : {
      role     = "read"
      database = db
    }
  ]
}
```

### 3. Cache Results

Cache frequently accessed data:

```terraform
data "mongodb_users" "all" {}

locals {
  usernames      = [for user in data.mongodb_users.all.users : user.username]
  user_role_map  = {
    for user in data.mongodb_users.all.users :
    user.username => user.roles
  }
}
```

## Troubleshooting

### Error: Not Authorized

```
Error: not authorized on admin to execute command
```

**Solution:** Grant required permissions to Terraform user:

```javascript
// In MongoDB shell
db.getSiblingDB("admin").grantRolesToUser("terraform_user", [
  { role: "userAdmin", db: "admin" }
]);
```

### Empty User List

If no users are returned but you expect some:

1. **Check permissions:** User needs `usersInfo` privilege
2. **Verify connection:** Ensure connection to correct database
3. **Check MongoDB:** Run `db.getUsers()` in MongoDB shell

```bash
# Verify users exist
mongosh "mongodb://admin:pass@host:27017/admin" --eval "db.getUsers()"
```

### Password Field Always Empty

This is expected behavior. MongoDB does not expose passwords:

```terraform
# ✅ Expected
data "mongodb_users" "all" {}

output "user_details" {
  value = data.mongodb_users.all.users[0]
  # password will be empty string
}
```

To manage passwords, use `mongodb_user` resource.

### Slow Data Source Reads

If reads take too long:

1. Large number of users may slow down reads
2. Network latency issues
3. MongoDB performance issues

**Solution:** Increase timeout:

```terraform
provider "mongodb" {
  connection_string = var.mongodb_uri
  default_timeout   = 20
}
```

## Related Resources

- [mongodb_user](../resources/user.md) - Create and manage users
- [mongodb_databases](databases.md) - List all databases
- [mongodb_replicaset](replicaset.md) - Read replica set configuration

## MongoDB Documentation

- [MongoDB Users](https://docs.mongodb.com/manual/core/security-users/)
- [MongoDB Built-In Roles](https://docs.mongodb.com/manual/reference/built-in-roles/)
- [usersInfo Command](https://docs.mongodb.com/manual/reference/command/usersInfo/)
