---
page_title: "Authentication and Security"
subcategory: "Guides"
description: |-
  Comprehensive guide to authentication, security, and credential management for the MongoDB Terraform provider.
---

# Authentication and Security

This guide covers authentication mechanisms, security best practices, and credential management when using the MongoDB Terraform provider.

## Authentication Mechanisms

The MongoDB provider supports standard MongoDB authentication mechanisms via the connection string.

### SCRAM-SHA-256 (Default for MongoDB 4.0+)

```terraform
provider "mongodb" {
  connection_string = "mongodb://username:password@host:27017/authSource"
}
```

### SCRAM-SHA-1 (Legacy)

```terraform
provider "mongodb" {
  connection_string = "mongodb://username:password@host:27017/authSource?authMechanism=SCRAM-SHA-1"
}
```

### Connection String Components

```
mongodb://[username:password@]host[:port][/authDB][?options]
```

**Components:**
- `username` - MongoDB username
- `password` - User password (URL-encoded if contains special characters)
- `host` - MongoDB host (hostname or IP)
- `port` - MongoDB port (default: 27017)
- `authDB` - Authentication database (typically `admin`)
- `options` - Additional connection options

## Credential Management

~> **Critical:** Never commit credentials to version control!

### Method 1: Environment Variables (Recommended)

**Setup:**

`variables.tf`:
```terraform
variable "mongodb_uri" {
  description = "MongoDB connection string"
  type        = string
  sensitive   = true
}
```

`main.tf`:
```terraform
provider "mongodb" {
  connection_string = var.mongodb_uri
}
```

**Usage:**

```bash
# Set environment variable
export TF_VAR_mongodb_uri="mongodb://admin:password@localhost:27017/admin"

# Run Terraform
terraform plan
terraform apply
```

**CI/CD Integration:**

```yaml
# GitHub Actions
env:
  TF_VAR_mongodb_uri: ${{ secrets.MONGODB_URI }}

# GitLab CI
variables:
  TF_VAR_mongodb_uri: $MONGODB_URI
```

### Method 2: Terraform Cloud/Enterprise

1. Navigate to workspace settings
2. Add variable `mongodb_uri`
3. Mark as **Sensitive**
4. Mark as **Environment variable** or **Terraform variable**

### Method 3: HashiCorp Vault

```terraform
# Configure Vault provider
provider "vault" {
  address = "https://vault.example.com"
  token   = var.vault_token
}

# Read MongoDB credentials from Vault
data "vault_generic_secret" "mongodb" {
  path = "secret/mongodb/connection"
}

# Use in MongoDB provider
provider "mongodb" {
  connection_string = data.vault_generic_secret.mongodb.data["uri"]
}
```

### Method 4: AWS Secrets Manager

```terraform
# Configure AWS provider
provider "aws" {
  region = "us-west-2"
}

# Read secret from AWS Secrets Manager
data "aws_secretsmanager_secret_version" "mongodb" {
  secret_id = "prod/mongodb/connection"
}

# Parse and use
locals {
  mongodb_secrets = jsondecode(data.aws_secretsmanager_secret_version.mongodb.secret_string)
}

provider "mongodb" {
  connection_string = local.mongodb_secrets["uri"]
}
```

### Method 5: Azure Key Vault

```terraform
# Configure Azure provider
provider "azurerm" {
  features {}
}

# Read from Key Vault
data "azurerm_key_vault_secret" "mongodb_uri" {
  name         = "mongodb-connection-string"
  key_vault_id = data.azurerm_key_vault.main.id
}

provider "mongodb" {
  connection_string = data.azurerm_key_vault_secret.mongodb_uri.value
}
```

### Method 6: GCP Secret Manager

```terraform
# Configure GCP provider
provider "google" {
  project = "my-project"
  region  = "us-central1"
}

# Read from Secret Manager
data "google_secret_manager_secret_version" "mongodb_uri" {
  secret = "mongodb-connection-string"
}

provider "mongodb" {
  connection_string = data.google_secret_manager_secret_version.mongodb_uri.secret_data
}
```

## Password Security

### URL Encoding Special Characters

Passwords with special characters must be URL-encoded:

```bash
# Password with special characters: p@ssw0rd!#
# Encoded: p%40ssw0rd%21%23

mongodb://user:p%40ssw0rd%21%23@host:27017/admin
```

**Common encodings:**
- `@` → `%40`
- `:` → `%3A`
- `/` → `%2F`
- `?` → `%3F`
- `#` → `%23`
- `[` → `%5B`
- `]` → `%5D`
- `!` → `%21`
- `$` → `%24`
- `&` → `%26`
- `'` → `%27`
- `(` → `%28`
- `)` → `%29`
- `*` → `%2A`
- `+` → `%2B`
- `,` → `%2C`
- `;` → `%3B`
- `=` → `%3D`

**Python helper:**
```python
from urllib.parse import quote_plus
password = "p@ssw0rd!#"
encoded = quote_plus(password)
print(f"mongodb://user:{encoded}@host:27017/admin")
```

### Generating Secure Passwords

Use Terraform's `random_password` resource:

```terraform
resource "random_password" "mongodb_user" {
  length  = 32
  special = true

  # Exclude problematic characters
  override_special = "!#$%&*()-_=+[]{}<>?"
}

resource "mongodb_user" "app" {
  username = "app"
  password = random_password.mongodb_user.result

  roles = [...]
}

# Store in AWS Secrets Manager
resource "aws_secretsmanager_secret" "mongodb_password" {
  name = "mongodb/${mongodb_user.app.username}/password"
}

resource "aws_secretsmanager_secret_version" "mongodb_password" {
  secret_id     = aws_secretsmanager_secret.mongodb_password.id
  secret_string = random_password.mongodb_user.result
}
```

## State File Security

~> **Warning:** Terraform state files contain sensitive data in plaintext!

### Remote State with Encryption

**S3 Backend (AWS):**

```terraform
terraform {
  backend "s3" {
    bucket         = "terraform-state-bucket"
    key            = "mongodb/terraform.tfstate"
    region         = "us-west-2"
    encrypt        = true  # Enable server-side encryption
    kms_key_id     = "arn:aws:kms:us-west-2:123456789:key/12345678-1234-1234-1234-123456789012"
    dynamodb_table = "terraform-locks"
  }
}
```

**Azure Backend:**

```terraform
terraform {
  backend "azurerm" {
    resource_group_name  = "terraform-state-rg"
    storage_account_name = "tfstatestorage"
    container_name       = "tfstate"
    key                  = "mongodb.terraform.tfstate"
    use_azuread_auth     = true
  }
}
```

**GCS Backend:**

```terraform
terraform {
  backend "gcs" {
    bucket  = "terraform-state-bucket"
    prefix  = "mongodb"
    encryption_key = var.gcs_encryption_key
  }
}
```

### State File Access Control

**S3 Bucket Policy:**

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Deny",
      "Principal": "*",
      "Action": "s3:GetObject",
      "Resource": "arn:aws:s3:::terraform-state-bucket/*",
      "Condition": {
        "StringNotLike": {
          "aws:userId": [
            "AIDAI*:terraform-user",
            "AROAI*"
          ]
        }
      }
    }
  ]
}
```

### Enable Versioning and MFA Delete

**S3:**
```bash
aws s3api put-bucket-versioning \
  --bucket terraform-state-bucket \
  --versioning-configuration Status=Enabled,MFADelete=Enabled \
  --mfa "arn:aws:iam::123456789:mfa/root-account-mfa-device 123456"
```

## MongoDB User Permissions

### Terraform User Requirements

Create a dedicated MongoDB user for Terraform with minimal required permissions:

```javascript
// Connect to MongoDB
use admin

// Create Terraform user
db.createUser({
  user: "terraform",
  pwd: "secure_password",
  roles: [
    { role: "userAdminAnyDatabase", db: "admin" },  // Manage users
    { role: "dbAdminAnyDatabase", db: "admin" },    // Manage databases
    { role: "readWriteAnyDatabase", db: "admin" },  // Read/write data
    { role: "clusterMonitor", db: "admin" }         // Monitor cluster
  ]
})
```

### Minimal Permissions

For production, restrict permissions to specific databases:

```javascript
db.createUser({
  user: "terraform_limited",
  pwd: "secure_password",
  roles: [
    { role: "userAdmin", db: "app_db" },      // Manage users in app_db only
    { role: "dbAdmin", db: "app_db" },        // Manage app_db only
    { role: "readWrite", db: "app_db" },      // Read/write app_db
    { role: "clusterMonitor", db: "admin" }   // Monitor replica set
  ]
})
```

### Custom Role for Terraform

```javascript
db.createRole({
  role: "terraformOperator",
  privileges: [
    {
      resource: { db: "", collection: "" },
      actions: ["createCollection", "dropCollection", "createIndex", "dropIndex"]
    },
    {
      resource: { db: "admin", collection: "system.users" },
      actions: ["find", "insert", "update", "remove"]
    }
  ],
  roles: [
    { role: "clusterMonitor", db: "admin" }
  ]
})

// Assign to user
db.grantRolesToUser("terraform", ["terraformOperator"])
```

## TLS/SSL Configuration

### Enable TLS

```terraform
provider "mongodb" {
  connection_string = "mongodb://user:pass@host:27017/admin?tls=true&tlsCAFile=/path/to/ca.pem"
}
```

### With Client Certificates

```terraform
provider "mongodb" {
  connection_string = "mongodb://host:27017/admin?tls=true&tlsCAFile=/path/to/ca.pem&tlsCertificateKeyFile=/path/to/client.pem"
}
```

### Self-Signed Certificates

```terraform
provider "mongodb" {
  connection_string = "mongodb://user:pass@host:27017/admin?tls=true&tlsAllowInvalidCertificates=true"
}
```

~> **Warning:** Only use `tlsAllowInvalidCertificates=true` in development. Never in production!

## Network Security

### SSH Tunnel

For enhanced security, use SSH tunnel:

```bash
# Create SSH tunnel
ssh -L 27017:mongodb-host:27017 bastion-host -N &

# Use in Terraform
export TF_VAR_mongodb_uri="mongodb://user:pass@localhost:27017/admin"
```

### VPN Connection

Require VPN for MongoDB access:

```terraform
# Only works when connected to VPN
provider "mongodb" {
  connection_string = "mongodb://user:pass@internal-mongodb.vpc:27017/admin"
}
```

### IP Whitelisting

Configure MongoDB to accept connections only from Terraform runner IPs.

**MongoDB config (`/etc/mongod.conf`):**
```yaml
net:
  bindIp: 127.0.0.1,10.0.0.5
  port: 27017
```

## Audit Logging

### MongoDB Audit Log

Enable MongoDB audit log to track Terraform actions:

```yaml
# /etc/mongod.conf
auditLog:
  destination: file
  format: JSON
  path: /var/log/mongodb/audit.log
```

### Terraform Audit with Sentinel (Terraform Enterprise)

```hcl
# sentinel.hcl
policy "audit-mongodb-changes" {
  enforcement_level = "advisory"
}
```

## Password Rotation

### Automated Rotation with Terraform

```terraform
resource "time_rotating" "monthly" {
  rotation_days = 30
}

resource "random_password" "rotated" {
  length  = 32
  special = true

  keepers = {
    rotation_time = time_rotating.monthly.id
  }
}

resource "mongodb_user" "app" {
  username = "app"
  password = random_password.rotated.result

  roles = [...]

  lifecycle {
    create_before_destroy = true
  }
}

# Update secret in AWS Secrets Manager
resource "aws_secretsmanager_secret_version" "current" {
  secret_id     = aws_secretsmanager_secret.mongodb.id
  secret_string = jsonencode({
    username = mongodb_user.app.username
    password = random_password.rotated.result
  })
}
```

### Manual Rotation Process

1. Generate new password
2. Update Terraform configuration
3. Run `terraform apply`
4. Update application configuration
5. Verify application connectivity
6. Commit changes (excluding sensitive data)

## Security Checklist

- [ ] Credentials stored in secure vault (not in code)
- [ ] Connection strings use environment variables
- [ ] Remote state backend configured with encryption
- [ ] State file access restricted with IAM policies
- [ ] Versioning enabled on state storage
- [ ] TLS/SSL enabled for MongoDB connections
- [ ] MongoDB user has minimal required permissions
- [ ] Audit logging enabled
- [ ] Password rotation policy in place
- [ ] Secrets never committed to version control
- [ ] CI/CD pipeline uses secure secret injection
- [ ] Network access restricted (firewall/VPN)
- [ ] `.tfvars` files in `.gitignore`
- [ ] Regular security audits scheduled

## Best Practices Summary

1. **Never commit credentials** - Use secure secret management
2. **Encrypt state files** - Use remote backend with encryption
3. **Minimal permissions** - Grant only required MongoDB privileges
4. **Use TLS/SSL** - Encrypt connections to MongoDB
5. **Rotate passwords** - Implement regular rotation policy
6. **Audit access** - Enable MongoDB and Terraform audit logs
7. **Network isolation** - Use VPN, SSH tunnels, or private networks
8. **Separate environments** - Use different credentials for dev/staging/prod
9. **Review access** - Regular audit of who has access to secrets
10. **Backup state** - Enable versioning and maintain backups

## Example: Complete Secure Setup

```terraform
terraform {
  required_version = ">= 1.6.0"

  # Encrypted remote state
  backend "s3" {
    bucket         = "terraform-state"
    key            = "mongodb/prod.tfstate"
    region         = "us-west-2"
    encrypt        = true
    kms_key_id     = "arn:aws:kms:..."
    dynamodb_table = "terraform-locks"
  }

  required_providers {
    mongodb = {
      source  = "registry.terraform.io/Ahton89/mongodb"
      version = "~> 0.2"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# Read credentials from AWS Secrets Manager
data "aws_secretsmanager_secret_version" "mongodb" {
  secret_id = "prod/mongodb/terraform"
}

locals {
  mongodb_creds = jsondecode(data.aws_secretsmanager_secret_version.mongodb.secret_string)
}

# MongoDB provider with TLS
provider "mongodb" {
  connection_string = "mongodb://${local.mongodb_creds.username}:${urlencode(local.mongodb_creds.password)}@${var.mongodb_host}:27017/admin?tls=true&tlsCAFile=/etc/ssl/mongodb-ca.pem"
  retry_attempts    = 10
  retry_delay_sec   = 5
}

# Sensitive variables
variable "mongodb_host" {
  type = string
}

# Resources with secure password generation
resource "random_password" "app_user" {
  length  = 32
  special = true

  keepers = {
    rotation = time_rotating.monthly.id
  }
}

resource "mongodb_user" "app" {
  username = "app_service"
  password = random_password.app_user.result

  roles = [
    {
      role     = "readWrite"
      database = "production"
    }
  ]

  lifecycle {
    create_before_destroy = true
  }
}

# Store generated password
resource "aws_secretsmanager_secret_version" "app_password" {
  secret_id = aws_secretsmanager_secret.app.id
  secret_string = jsonencode({
    username = mongodb_user.app.username
    password = random_password.app_user.result
  })
}
```

## Troubleshooting

### Authentication Failed

```
Error: auth error: Authentication failed
```

**Solutions:**
1. Verify credentials are correct
2. Check authentication database (usually `admin`)
3. Ensure user has required permissions
4. URL-encode special characters in password
5. Check MongoDB user exists: `db.getUser("username")`

### SSL/TLS Errors

```
Error: x509: certificate signed by unknown authority
```

**Solutions:**
1. Specify CA certificate: `?tlsCAFile=/path/to/ca.pem`
2. For development only: `?tlsAllowInvalidCertificates=true`
3. Verify certificate chain is complete
4. Check certificate hasn't expired

### Permission Denied

```
Error: not authorized on admin to execute command
```

**Solutions:**
1. Grant required role: `db.grantRolesToUser("user", [{role: "role", db: "db"}])`
2. Verify user has admin privileges
3. Check authentication database matches user location

## Further Reading

- [MongoDB Security Checklist](https://docs.mongodb.com/manual/administration/security-checklist/)
- [MongoDB Authentication](https://docs.mongodb.com/manual/core/authentication/)
- [MongoDB Encryption](https://docs.mongodb.com/manual/core/security-encryption/)
- [Terraform Sensitive Data](https://www.terraform.io/docs/language/values/variables.html#suppressing-values-in-cli-output)
