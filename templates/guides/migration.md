---
page_title: "Importing Existing MongoDB Resources"
subcategory: "Guides"
description: |-
  Guide to importing existing MongoDB databases, users, and replica sets into Terraform management.
---

# Importing Existing MongoDB Resources

This guide explains how to import existing MongoDB resources (databases, users, replica sets) into Terraform state for management.

## Overview

The `terraform import` command brings existing infrastructure under Terraform management without destroying and recreating resources.

**Supported resources:**
- Databases (`mongodb_database`)
- Users (`mongodb_user`)
- Replica Sets (`mongodb_replicaset`)

## Prerequisites

Before importing:

1. **Terraform configuration** - Resource blocks must exist in `.tf` files
2. **MongoDB access** - Administrative credentials for MongoDB
3. **Resource identifiers** - Names of resources to import

## Import Workflow

1. Create resource configuration in Terraform
2. Run `terraform import` command
3. Verify imported state
4. Run `terraform plan` to check for drift
5. Update configuration if needed
6. Apply to align state with configuration

## Importing Databases

### Step 1: Create Configuration

Create a resource block in your Terraform configuration:

```terraform
# main.tf
resource "mongodb_database" "production" {
  name = "production_db"
}
```

### Step 2: Import Database

Import the existing database:

```bash
terraform import mongodb_database.production production_db
```

**Output:**
```
mongodb_database.production: Importing from ID "production_db"...
mongodb_database.production: Import prepared!
mongodb_database.production: Refreshing state...

Import successful!
```

### Step 3: Verify Import

Check the imported state:

```bash
terraform state show mongodb_database.production
```

**Output:**
```hcl
# mongodb_database.production:
resource "mongodb_database" "production" {
    name = "production_db"
}
```

### Step 4: Validate Configuration

Run plan to ensure configuration matches:

```bash
terraform plan
```

Expected output:
```
No changes. Your infrastructure matches the configuration.
```

### Multiple Databases

Import multiple databases:

```bash
# Import each database
terraform import 'mongodb_database.apps["users"]' users_db
terraform import 'mongodb_database.apps["products"]' products_db
terraform import 'mongodb_database.apps["orders"]' orders_db
```

**Configuration:**
```terraform
variable "databases" {
  default = ["users_db", "products_db", "orders_db"]
}

resource "mongodb_database" "apps" {
  for_each = toset(var.databases)
  name     = each.value
}
```

## Importing Users

### Step 1: Create Configuration

```terraform
resource "mongodb_user" "app_service" {
  username = "app_service"
  password = var.app_password  # Must be provided but cannot be read from MongoDB

  roles = [
    {
      role     = "readWrite"
      database = "production_db"
    }
  ]
}
```

~> **Important:** MongoDB does not expose passwords. You must provide the password in your configuration, but Terraform cannot verify it matches the actual password.

### Step 2: Import User

```bash
terraform import mongodb_user.app_service app_service
```

### Step 3: Handle Password

Since passwords cannot be read from MongoDB:

**Option 1: Ignore password changes (Recommended for import)**

```terraform
resource "mongodb_user" "app_service" {
  username = "app_service"
  password = "temporary_value"  # Will be ignored

  roles = [...]

  lifecycle {
    ignore_changes = [password]
  }
}
```

**Option 2: Use existing password**

If you know the password:

```terraform
variable "app_service_password" {
  type      = string
  sensitive = true
}

resource "mongodb_user" "app_service" {
  username = "app_service"
  password = var.app_service_password

  roles = [...]
}
```

### Step 4: Verify Roles

After import, verify roles match your configuration:

```bash
terraform plan
```

If roles differ, update your configuration or MongoDB:

```javascript
// Update roles in MongoDB to match Terraform
db.getSiblingDB("admin").updateUser("app_service", {
  roles: [
    { role: "readWrite", db: "production_db" }
  ]
})
```

### Multiple Users

```bash
terraform import mongodb_user.readers["analyst"] analyst
terraform import mongodb_user.readers["reporter"] reporter
```

**Configuration:**
```terraform
locals {
  readers = {
    analyst  = { password = var.analyst_password }
    reporter = { password = var.reporter_password }
  }
}

resource "mongodb_user" "readers" {
  for_each = local.readers

  username = each.key
  password = each.value.password

  roles = [
    {
      role     = "read"
      database = "analytics"
    }
  ]

  lifecycle {
    ignore_changes = [password]
  }
}
```

## Importing Replica Sets

### Prerequisites

- Replica set must be already initialized
- Connection string must include replica set members
- Provider must use replica set connection (not direct)

### Step 1: Configure Provider

```terraform
provider "mongodb" {
  connection_string = "mongodb://admin:pass@host1:27017,host2:27017,host3:27017/admin?replicaSet=rs0"
}
```

### Step 2: Create Configuration

```terraform
resource "mongodb_replicaset" "main" {
  name = "rs0"

  members = [
    {
      id   = 0
      host = "host1:27017"
    },
    {
      id   = 1
      host = "host2:27017"
    },
    {
      id   = 2
      host = "host3:27017"
    }
  ]
}
```

### Step 3: Import Replica Set

```bash
terraform import mongodb_replicaset.main rs0
```

### Step 4: Verify Configuration

```bash
terraform state show mongodb_replicaset.main
```

The imported state will include all current configuration including optional settings with default values.

### Step 5: Clean Up Configuration

Remove default values from configuration:

**Before:**
```terraform
resource "mongodb_replicaset" "main" {
  name             = "rs0"
  protocol_version = 1              # Default value

  members = [
    {
      id           = 0
      host         = "host1:27017"
      priority     = 1              # Default value
      votes        = 1              # Default value
      arbiter_only = false          # Default value
    },
    # ...
  ]
}
```

**After (cleaned):**
```terraform
resource "mongodb_replicaset" "main" {
  name = "rs0"

  members = [
    {
      id   = 0
      host = "host1:27017"
    },
    {
      id   = 1
      host = "host2:27017"
    },
    {
      id   = 2
      host = "host3:27017"
    }
  ]
}
```

## Migration Strategies

### Strategy 1: Gradual Migration

Import resources gradually:

1. **Week 1:** Import databases
2. **Week 2:** Import users
3. **Week 3:** Import replica set
4. **Week 4:** Full management

### Strategy 2: Environment-by-Environment

1. **Development:** Import and test
2. **Staging:** Apply learnings
3. **Production:** Final migration

### Strategy 3: Resource Type by Type

1. **Databases first:** Lower risk
2. **Users second:** Test password handling
3. **Replica set last:** Most critical

## Complete Migration Example

### Scenario

Existing MongoDB setup:
- Databases: `users_db`, `products_db`, `orders_db`
- Users: `app_service`, `readonly`
- Replica set: `rs0` with 3 members

### Step 1: Inventory Resources

List existing resources:

```bash
# List databases
mongosh "mongodb://admin:pass@host:27017/admin" --eval "db.adminCommand('listDatabases').databases.map(d => d.name)"

# List users
mongosh "mongodb://admin:pass@host:27017/admin" --eval "db.getUsers().map(u => u.user)"

# Check replica set
mongosh "mongodb://admin:pass@host1:27017,host2:27017/admin?replicaSet=rs0" --eval "rs.conf()"
```

### Step 2: Create Terraform Configuration

```terraform
# versions.tf
terraform {
  required_version = ">= 1.6.0"

  required_providers {
    mongodb = {
      source  = "registry.terraform.io/Ahton89/mongodb"
      version = "~> 0.2"
    }
  }
}

# variables.tf
variable "mongodb_uri" {
  type      = string
  sensitive = true
}

variable "app_password" {
  type      = string
  sensitive = true
}

variable "readonly_password" {
  type      = string
  sensitive = true
}

# main.tf
provider "mongodb" {
  connection_string = var.mongodb_uri
}

# Databases
resource "mongodb_database" "users" {
  name = "users_db"
}

resource "mongodb_database" "products" {
  name = "products_db"
}

resource "mongodb_database" "orders" {
  name = "orders_db"
}

# Users
resource "mongodb_user" "app_service" {
  username = "app_service"
  password = var.app_password

  roles = [
    { role = "readWrite", database = "users_db" },
    { role = "readWrite", database = "products_db" },
    { role = "readWrite", database = "orders_db" }
  ]

  lifecycle {
    ignore_changes = [password]
  }
}

resource "mongodb_user" "readonly" {
  username = "readonly"
  password = var.readonly_password

  roles = [
    { role = "read", database = "users_db" },
    { role = "read", database = "products_db" },
    { role = "read", database = "orders_db" }
  ]

  lifecycle {
    ignore_changes = [password]
  }
}

# Replica Set
resource "mongodb_replicaset" "main" {
  name = "rs0"

  members = [
    { id = 0, host = "mongodb-1.internal:27017" },
    { id = 1, host = "mongodb-2.internal:27017" },
    { id = 2, host = "mongodb-3.internal:27017" }
  ]
}
```

### Step 3: Import Resources

Create import script:

```bash
#!/bin/bash
# import_mongodb.sh

set -e

echo "Importing MongoDB resources..."

# Import databases
echo "Importing databases..."
terraform import mongodb_database.users users_db
terraform import mongodb_database.products products_db
terraform import mongodb_database.orders orders_db

# Import users
echo "Importing users..."
terraform import mongodb_user.app_service app_service
terraform import mongodb_user.readonly readonly

# Import replica set
echo "Importing replica set..."
terraform import mongodb_replicaset.main rs0

echo "Import complete!"
```

Run import:

```bash
chmod +x import_mongodb.sh
./import_mongodb.sh
```

### Step 4: Verify Import

```bash
# Check state
terraform state list

# Verify each resource
terraform state show mongodb_database.users
terraform state show mongodb_user.app_service
terraform state show mongodb_replicaset.main

# Run plan
terraform plan
```

Expected output:
```
No changes. Your infrastructure matches the configuration.
```

## Handling Import Errors

### Error: Resource Already Exists

```
Error: resource already exists in state
```

**Solution:**
Remove from state first:

```bash
terraform state rm mongodb_database.production
terraform import mongodb_database.production production_db
```

### Error: Resource Not Found

```
Error: database production_db does not exist
```

**Solution:**
Verify resource exists in MongoDB:

```bash
mongosh "mongodb://..." --eval "db.adminCommand('listDatabases')"
```

### Error: Authentication Failed During Import

```
Error: auth error: Authentication failed
```

**Solution:**
1. Verify credentials
2. Ensure user has required permissions
3. Check authentication database

```bash
# Test connection
mongosh "mongodb://admin:pass@host:27017/admin" --eval "db.runCommand({connectionStatus: 1})"
```

### Error: Cannot Import System Database

```
Error: database admin is a default database and cannot be imported
```

**Solution:**
System databases (`admin`, `config`, `local`) cannot be managed by Terraform.

## Post-Import Checklist

After successful import:

- [ ] All resources imported successfully
- [ ] `terraform plan` shows no changes
- [ ] State file backed up
- [ ] Configuration committed to version control
- [ ] Team notified of Terraform management
- [ ] Documentation updated
- [ ] Monitoring adjusted for Terraform operations
- [ ] Backup strategy updated
- [ ] Disaster recovery plan updated

## State Management

### Backup State Before Import

```bash
# Local state
cp terraform.tfstate terraform.tfstate.backup.$(date +%Y%m%d_%H%M%S)

# Remote state (S3)
aws s3 cp s3://bucket/path/terraform.tfstate ./terraform.tfstate.backup.$(date +%Y%m%d_%H%M%S)
```

### Move Resources Between States

```bash
# Pull from remote state
terraform state pull > old_state.tfstate

# Move resource
terraform state mv -state=old_state.tfstate mongodb_database.old mongodb_database.new

# Push to remote state
terraform state push old_state.tfstate
```

## Rollback Strategy

If import causes issues:

### Option 1: Remove from Terraform

```bash
# Remove from state
terraform state rm mongodb_database.production
terraform state rm mongodb_user.app
terraform state rm mongodb_replicaset.main

# Resource continues to exist in MongoDB
```

### Option 2: Restore State Backup

```bash
# Restore from backup
cp terraform.tfstate.backup.20240101_120000 terraform.tfstate

# Verify
terraform plan
```

## Best Practices

### 1. Test in Development First

Import in dev environment before production:

```bash
# Development
export TF_VAR_mongodb_uri=$DEV_MONGODB_URI
terraform import mongodb_database.test test_db

# After successful test, proceed to production
export TF_VAR_mongodb_uri=$PROD_MONGODB_URI
terraform import mongodb_database.production production_db
```

### 2. Import During Maintenance Window

Schedule imports during low-traffic periods:
- Lower risk of concurrent changes
- Easier to rollback if needed
- Less impact on applications

### 3. Document Imports

Maintain import log:

```bash
# import_log.md
## 2024-01-15: Database Import

### Resources Imported
- mongodb_database.users (users_db)
- mongodb_database.products (products_db)

### Issues Encountered
- None

### Rollback Plan
- State backup: terraform.tfstate.backup.20240115_100000
```

### 4. Use Automation

Create reusable import scripts:

```bash
#!/bin/bash
# import_databases.sh

DATABASES=("users_db" "products_db" "orders_db")

for db in "${DATABASES[@]}"; do
  echo "Importing $db..."
  terraform import "mongodb_database.${db}" "$db"
done
```

### 5. Validate After Import

Comprehensive validation:

```bash
# Validate configuration
terraform validate

# Check for drift
terraform plan -detailed-exitcode

# Test operations
terraform plan -target=mongodb_database.test
```

## Advanced: Bulk Import

For many resources:

```python
#!/usr/bin/env python3
# bulk_import.py

import subprocess
import json

# Get list of databases from MongoDB
result = subprocess.run([
    'mongosh', 'mongodb://admin:pass@host:27017/admin',
    '--quiet', '--eval', 'JSON.stringify(db.adminCommand({listDatabases: 1}).databases)'
], capture_output=True, text=True)

databases = json.loads(result.stdout)

# Import each database
for db in databases:
    if db['name'] not in ['admin', 'config', 'local']:
        resource_name = db['name'].replace('-', '_')
        print(f"Importing {db['name']}...")
        subprocess.run([
            'terraform', 'import',
            f'mongodb_database.{resource_name}',
            db['name']
        ])
```

## Further Reading

- [Terraform Import Command](https://www.terraform.io/cli/commands/import)
- [State Management](https://www.terraform.io/cli/state)
- [Resource Addressing](https://www.terraform.io/cli/state/resource-addressing)
