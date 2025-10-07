---
page_title: "Getting Started with MongoDB Provider"
subcategory: "Guides"
description: |-
  Learn how to set up and use the MongoDB Terraform provider to manage your MongoDB infrastructure.
---

# Getting Started with MongoDB Provider

This guide will walk you through setting up the MongoDB Terraform provider and creating your first MongoDB resources.

## Prerequisites

Before you begin, ensure you have:

1. **Terraform** (version 1.6.3 or later)
   ```bash
   terraform --version
   ```

2. **MongoDB** (version 6.x, 7.x, or 8.x)
   ```bash
   mongosh --version
   ```

3. **MongoDB running** and accessible
   ```bash
   # Test connection
   mongosh "mongodb://localhost:27017"
   ```

4. **Administrative credentials** for MongoDB
   - User with `root` or appropriate permissions
   - Password stored securely

## Step 1: Install the Provider

Create a new directory for your Terraform configuration:

```bash
mkdir terraform-mongodb-example
cd terraform-mongodb-example
```

Create a `main.tf` file with provider configuration:

```terraform
terraform {
  required_version = ">= 1.6.0"

  required_providers {
    mongodb = {
      source  = "registry.terraform.io/Ahton89/mongodb"
      version = "~> 0.2"
    }
  }
}

provider "mongodb" {
  connection_string = "mongodb://admin:password@localhost:27017/admin"
}
```

Initialize Terraform:

```bash
terraform init
```

You should see output similar to:

```
Initializing the backend...
Initializing provider plugins...
- Finding ahton89/mongodb versions matching "~> 0.2"...
- Installing ahton89/mongodb v0.2.x...
- Installed ahton89/mongodb v0.2.x
Terraform has been successfully initialized!
```

## Step 2: Secure Your Credentials

~> **Warning:** Never commit credentials to version control!

### Option 1: Environment Variables (Recommended)

Create a `variables.tf` file:

```terraform
variable "mongodb_uri" {
  description = "MongoDB connection string"
  type        = string
  sensitive   = true
}
```

Update your `main.tf`:

```terraform
provider "mongodb" {
  connection_string = var.mongodb_uri
}
```

Create a `.tfvars` file (add to `.gitignore`):

```bash
# terraform.tfvars
mongodb_uri = "mongodb://admin:password@localhost:27017/admin"
```

Or set environment variable:

```bash
export TF_VAR_mongodb_uri="mongodb://admin:password@localhost:27017/admin"
```

### Option 2: Terraform Cloud/Enterprise

Store the connection string as a sensitive workspace variable in Terraform Cloud.

## Step 3: Create Your First Database

Add database configuration to `main.tf`:

```terraform
# Create a database
resource "mongodb_database" "example" {
  name = "myapp"
}

# Output the database name
output "database_name" {
  value = mongodb_database.example.name
}
```

Plan and apply the configuration:

```bash
# See what will be created
terraform plan

# Create the database
terraform apply
```

Verify the database was created:

```bash
mongosh "mongodb://admin:password@localhost:27017/admin" --eval "db.adminCommand('listDatabases')"
```

## Step 4: Create a User

Add user configuration to `main.tf`:

```terraform
# Create a user with read-write access
resource "mongodb_user" "app_user" {
  username = "app_service"
  password = var.app_password

  roles = [
    {
      role     = "readWrite"
      database = mongodb_database.example.name
    }
  ]

  depends_on = [mongodb_database.example]
}
```

Add password variable to `variables.tf`:

```terraform
variable "app_password" {
  description = "Application user password"
  type        = string
  sensitive   = true
}
```

Update your `.tfvars`:

```bash
mongodb_uri  = "mongodb://admin:password@localhost:27017/admin"
app_password = "secure_app_password"
```

Apply the changes:

```bash
terraform apply
```

## Step 5: Read Data Sources

Query existing resources using data sources:

```terraform
# List all databases
data "mongodb_databases" "all" {
  depends_on = [mongodb_database.example]
}

# List all users
data "mongodb_users" "all" {
  depends_on = [mongodb_user.app_user]
}

# Output the data
output "all_databases" {
  value = [for db in data.mongodb_databases.all.databases : db.name]
}

output "all_users" {
  value = [for user in data.mongodb_users.all.users : user.username]
}
```

Apply and view outputs:

```bash
terraform apply
terraform output
```

## Complete Example

Here's a complete working example:

```terraform
terraform {
  required_version = ">= 1.6.0"

  required_providers {
    mongodb = {
      source  = "registry.terraform.io/Ahton89/mongodb"
      version = "~> 0.2"
    }
  }
}

# Variables
variable "mongodb_uri" {
  type      = string
  sensitive = true
}

variable "environment" {
  type    = string
  default = "development"
}

# Provider configuration
provider "mongodb" {
  connection_string = var.mongodb_uri
  retry_attempts    = 5
  retry_delay_sec   = 3
}

# Create database
resource "mongodb_database" "app" {
  name = "myapp_${var.environment}"

  timeouts {
    create = "10m"
    delete = "10m"
  }
}

# Create application user
resource "mongodb_user" "app" {
  username = "app_${var.environment}"
  password = var.app_password

  roles = [
    {
      role     = "readWrite"
      database = mongodb_database.app.name
    }
  ]

  timeouts {
    create = "5m"
    update = "5m"
    delete = "5m"
  }
}

# Create read-only user
resource "mongodb_user" "readonly" {
  username = "reader_${var.environment}"
  password = var.reader_password

  roles = [
    {
      role     = "read"
      database = mongodb_database.app.name
    }
  ]
}

# Data sources
data "mongodb_databases" "current" {
  depends_on = [mongodb_database.app]
}

data "mongodb_users" "current" {
  depends_on = [
    mongodb_user.app,
    mongodb_user.readonly
  ]
}

# Outputs
output "database_name" {
  value = mongodb_database.app.name
}

output "app_username" {
  value = mongodb_user.app.username
}

output "connection_string" {
  value     = "mongodb://${mongodb_user.app.username}:${var.app_password}@localhost:27017/${mongodb_database.app.name}"
  sensitive = true
}

output "all_databases" {
  value = [for db in data.mongodb_databases.current.databases : db.name]
}

output "all_users" {
  value = [for user in data.mongodb_users.current.users : user.username]
}
```

Create `terraform.tfvars`:

```bash
mongodb_uri     = "mongodb://admin:password@localhost:27017/admin"
environment     = "development"
app_password    = "secure_app_password"
reader_password = "secure_reader_password"
```

## Common Patterns

### Multiple Databases

```terraform
locals {
  databases = ["users", "products", "orders"]
}

resource "mongodb_database" "apps" {
  for_each = toset(local.databases)
  name     = "${each.value}_${var.environment}"
}
```

### User Per Database

```terraform
resource "mongodb_database" "apps" {
  for_each = toset(["users", "products"])
  name     = each.key
}

resource "mongodb_user" "app_users" {
  for_each = mongodb_database.apps

  username = "service_${each.key}"
  password = var.passwords[each.key]

  roles = [
    {
      role     = "readWrite"
      database = each.value.name
    }
  ]
}
```

### Environment-Specific Configuration

```terraform
locals {
  config = {
    production = {
      timeout = 30
      retry   = 10
    }
    development = {
      timeout = 15
      retry   = 5
    }
  }

  current_config = local.config[var.environment]
}

provider "mongodb" {
  connection_string = var.mongodb_uri
  default_timeout   = local.current_config.timeout
  retry_attempts    = local.current_config.retry
}
```

## Best Practices

### 1. Use Remote State

Store state remotely with encryption:

```terraform
terraform {
  backend "s3" {
    bucket         = "my-terraform-state"
    key            = "mongodb/terraform.tfstate"
    region         = "us-west-2"
    encrypt        = true
    kms_key_id     = "arn:aws:kms:..."
    dynamodb_table = "terraform-locks"
  }
}
```

### 2. Use Workspaces for Environments

```bash
# Create workspaces
terraform workspace new development
terraform workspace new staging
terraform workspace new production

# Switch between environments
terraform workspace select development
terraform apply

terraform workspace select production
terraform apply
```

### 3. Organize Your Code

Structure your project:

```
terraform-mongodb/
├── main.tf           # Provider and resources
├── variables.tf      # Variable definitions
├── outputs.tf        # Output definitions
├── versions.tf       # Terraform and provider versions
├── terraform.tfvars  # Variable values (gitignored)
└── environments/
    ├── dev.tfvars
    ├── staging.tfvars
    └── prod.tfvars
```

### 4. Use Modules

Create reusable modules:

```terraform
# modules/mongodb-app/main.tf
variable "app_name" {}
variable "environment" {}
variable "password" { sensitive = true }

resource "mongodb_database" "app" {
  name = "${var.app_name}_${var.environment}"
}

resource "mongodb_user" "app" {
  username = var.app_name
  password = var.password
  roles = [{
    role     = "readWrite"
    database = mongodb_database.app.name
  }]
}

output "database_name" { value = mongodb_database.app.name }
output "username" { value = mongodb_user.app.username }
```

Use the module:

```terraform
module "api_app" {
  source = "./modules/mongodb-app"

  app_name    = "api"
  environment = var.environment
  password    = var.api_password
}

module "worker_app" {
  source = "./modules/mongodb-app"

  app_name    = "worker"
  environment = var.environment
  password    = var.worker_password
}
```

## Next Steps

Now that you have the basics:

1. [Learn about Authentication](authentication.md) - Secure your MongoDB connections
2. [Import Existing Resources](migration.md) - Bring existing MongoDB resources under Terraform management
3. [Configure Replica Sets](../resources/replicaset.md) - Set up high availability

## Troubleshooting

### Provider Initialization Fails

```
Error: Failed to query available provider packages
```

**Solution:**
1. Check internet connection
2. Verify Terraform version: `terraform version`
3. Clear provider cache: `rm -rf .terraform/`
4. Re-initialize: `terraform init`

### Connection Refused

```
Error: connection to MongoDB failed with error: connection refused
```

**Solution:**
1. Verify MongoDB is running: `sudo systemctl status mongod`
2. Check MongoDB is listening: `netstat -an | grep 27017`
3. Verify firewall rules allow connection
4. Test connection: `mongosh "mongodb://localhost:27017"`

### Authentication Failed

```
Error: failed to ping MongoDB: auth error
```

**Solution:**
1. Verify credentials are correct
2. Ensure user exists and has required permissions
3. Check authentication database (usually `admin`)
4. URL-encode special characters in password

### Version Mismatch

```
Error: unsupported MongoDB version
```

**Solution:**
This provider supports MongoDB 6.x, 7.x, and 8.x. Check your MongoDB version:

```bash
mongosh --eval "db.version()"
```

## Getting Help

- **Documentation:** Review the [provider documentation](../index.md)
- **Examples:** Check the `_examples/` directory in the repository
- **Issues:** Report bugs at [GitHub Issues](https://github.com/Ahton89/terraform-provider-mongodb/issues)
- **Community:** Join Terraform community forums

## Useful Commands

```bash
# Terraform commands
terraform init          # Initialize provider
terraform plan          # Preview changes
terraform apply         # Apply changes
terraform destroy       # Destroy resources
terraform fmt           # Format code
terraform validate      # Validate configuration
terraform state list    # List resources in state
terraform state show    # Show resource details
terraform refresh       # Update state with current infrastructure

# MongoDB verification commands
mongosh "mongodb://..." --eval "show dbs"
mongosh "mongodb://..." --eval "show users"
mongosh "mongodb://..." --eval "rs.status()"
```
