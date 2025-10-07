<a href="https://terraform.io">
    <img src=".github/tf.png" alt="Terraform logo" title="Terraform" align="left" height="50" />
</a>

# Terraform Provider for MongoDB

[![Release](https://img.shields.io/github/v/release/Ahton89/terraform-provider-mongodb?display_name=tag)](https://github.com/Ahton89/terraform-provider-mongodb/releases/latest)
[![License](https://img.shields.io/github/license/Ahton89/terraform-provider-mongodb)](https://github.com/Ahton89/terraform-provider-mongodb/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/Ahton89/terraform-provider-mongodb)](https://goreportcard.com/report/github.com/Ahton89/terraform-provider-mongodb)
[![Documentation](https://img.shields.io/badge/docs-terraform.io-purple)](https://registry.terraform.io/providers/Ahton89/mongodb/latest/docs)

A Terraform provider for managing MongoDB infrastructure as code. This provider enables you to manage MongoDB databases, users, and replica sets using declarative configuration.

## Features

- **Database Management** - Create, import, and delete MongoDB databases
- **User Management** - Full CRUD operations for MongoDB users with role-based access control (RBAC)
- **Replica Set Configuration** - Initialize and update MongoDB replica sets for high availability
- **Data Sources** - Query existing databases, users, and replica set configurations
- **Import Support** - Import existing MongoDB resources into Terraform state
- **Timeouts** - Configurable timeouts for all resource operations
- **Retry Logic** - Built-in retry mechanism with exponential backoff for transient errors

## Requirements

| Component | Version |
|-----------|---------|
| [Terraform](https://www.terraform.io/downloads.html) | >= 1.6.3 |
| [Go](https://golang.org/doc/install) | >= 1.21 (for building from source) |
| [MongoDB](https://www.mongodb.com/) | v6.x, v7.x, v8.x |

> **Note:** This provider supports **MongoDB v6.x, v7.x, and v8.x**.

## Supported Platforms

- **macOS** (Darwin) - AMD64, ARM64
- **Linux** - AMD64, ARM64
- **Windows** - AMD64

## Quick Start

### 1. Install the Provider

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    mongodb = {
      source  = "Ahton89/mongodb"
      version = "~> 0.2"
    }
  }
}

provider "mongodb" {
  connection_string = "mongodb://admin:password@localhost:27017/admin"
}
```

### 2. Create Your First Database

```hcl
resource "mongodb_database" "example" {
  name = "myapp"
}
```

### 3. Apply the Configuration

```bash
terraform init
terraform plan
terraform apply
```

## Usage Examples

### Database Management

Create a database with custom timeouts:

```hcl
resource "mongodb_database" "app" {
  name = "production_db"

  timeouts {
    create = "10m"
    delete = "20m"
  }
}
```

### User Management with RBAC

Create a user with multiple roles:

```hcl
resource "mongodb_user" "app_user" {
  username = "app_service"
  password = var.app_password  # Use variables for sensitive data

  roles = [
    {
      role     = "readWrite"
      database = mongodb_database.app.name
    },
    {
      role     = "read"
      database = "analytics"
    }
  ]

  depends_on = [mongodb_database.app]
}
```

### Replica Set Configuration

Initialize a 3-node replica set:

```hcl
resource "mongodb_replicaset" "main" {
  name = "rs0"

  members = [
    {
      id   = 0
      host = "mongodb-1.example.com:27017"
    },
    {
      id   = 1
      host = "mongodb-2.example.com:27017"
    },
    {
      id   = 2
      host = "mongodb-3.example.com:27017"
    }
  ]
}
```

### Data Sources

Query existing resources:

```hcl
# List all databases
data "mongodb_databases" "all" {}

output "database_names" {
  value = [for db in data.mongodb_databases.all.databases : db.name]
}

# List all users
data "mongodb_users" "all" {}

output "usernames" {
  value = [for user in data.mongodb_users.all.users : user.username]
}

# Get replica set configuration
data "mongodb_replicaset" "current" {}

output "replica_set_name" {
  value = data.mongodb_replicaset.current.name
}
```

## Provider Configuration

### Required Arguments

- **connection_string** (string) - MongoDB connection string. Format: `mongodb://username:password@host:port/authDB`

### Optional Arguments

- **retry_attempts** (number) - Number of retry attempts for transient errors. Default: `5`
- **retry_delay_sec** (number) - Delay in seconds between retries. Default: `3`
- **default_timeout** (number) - Default timeout in minutes for all operations. Default: `15`

### Configuration Example

```hcl
provider "mongodb" {
  connection_string = var.mongodb_uri
  retry_attempts    = 10
  retry_delay_sec   = 5
  default_timeout   = 30
}
```

## Authentication

The provider supports standard MongoDB authentication mechanisms:

- **SCRAM-SHA-256** (recommended for MongoDB 4.0+)
- **SCRAM-SHA-1**
- **MONGODB-CR** (legacy)

### Connection String Examples

**Local MongoDB:**
```
mongodb://admin:password@localhost:27017/admin
```

**Replica Set:**
```
mongodb://admin:password@host1:27017,host2:27017,host3:27017/admin?replicaSet=rs0
```

**With TLS:**
```
mongodb://admin:password@host:27017/admin?tls=true
```

> **Security Best Practice:** Never hardcode credentials. Use environment variables or secret management systems like HashiCorp Vault or AWS Secrets Manager.

## Resource Timeouts

All resources support custom timeouts. Default is **15 minutes** per operation.

```hcl
resource "mongodb_user" "example" {
  username = "user"
  password = var.password
  roles    = [...]

  timeouts {
    create = "5m"
    read   = "2m"
    update = "5m"
    delete = "5m"
  }
}
```

Supported timeout operations:

| Resource | Create | Read | Update | Delete |
|----------|--------|------|--------|--------|
| **mongodb_database** | ✅ | ✅ | ❌ | ✅ |
| **mongodb_user** | ✅ | ✅ | ✅ | ✅ |
| **mongodb_replicaset** | ✅ | ✅ | ✅ | ❌ |

## Importing Existing Resources

Import existing MongoDB resources into Terraform state:

### Database

```bash
terraform import mongodb_database.example database_name
```

### User

```bash
terraform import mongodb_user.example username
```

### Replica Set

```bash
terraform import mongodb_replicaset.example replicaset_name
```

## Important Limitations

### Database Resources

- ⚠️ **Renaming requires replacement** - Changing the database name deletes the old database and creates a new one
- 🚫 **No update support** - Databases cannot be modified, only recreated
- 🔒 **System databases protected** - Cannot manage `admin`, `config`, or `local` databases

### User Resources

- ⚠️ **Username changes require replacement** - Changing username deletes the old user
- 🔓 **Passwords stored in state** - Use remote state encryption ([see Security Guide](docs/guides/authentication.md))

### Replica Set Resources

- 🚫 **No delete support** - Replica sets must be removed manually
- ⚠️ **Name immutable** - Replica set name cannot be changed after creation
- ⚙️ **Incremental member updates** - Add or modify only one member at a time after initial creation

## Examples

Complete working examples are available in the [`examples/`](examples/) directory:

- [**Basic Database**](examples/database/) - Create and manage databases
- [**User Management**](examples/user/) - Create users with RBAC
- [**Replica Set**](examples/replicaset/) - Configure high-availability clusters
- [**Databases Data Source**](examples/databases_datasource/) - Query existing databases
- [**Users Data Source**](examples/users_datasource/) - List MongoDB users
- [**Replica Set Data Source**](examples/replicaset_datasource/) - Read replica set configuration

## Documentation

- 📖 **[Provider Documentation](https://registry.terraform.io/providers/Ahton89/mongodb/latest/docs)** - Complete provider reference
- 🚀 **[Getting Started Guide](docs/guides/getting-started.md)** - Step-by-step tutorial
- 🔒 **[Authentication Guide](docs/guides/authentication.md)** - Security best practices
- 📦 **[Migration Guide](docs/guides/migration.md)** - Import existing resources

## Local Development

### Prerequisites

1. Go 1.21+
2. Terraform 1.6.3+
3. MongoDB 6.x, 7.x, or 8.x running locally or accessible remotely

### Building from Source

```bash
# Clone the repository
git clone https://github.com/Ahton89/terraform-provider-mongodb.git
cd terraform-provider-mongodb

# Build and install
make install
```

By default, the provider is installed to `~/go/bin`. To customize:

```bash
make install BIN_DIR=/path/to/your/directory
```

### Dev Environment Setup

1. **Create dev overrides configuration:**

```bash
cat > ~/.terraformrc << EOF
provider_installation {
  dev_overrides {
    "registry.terraform.io/Ahton89/mongodb" = "/Users/YOUR_USER/go/bin"
  }
  direct {}
}
EOF
```

2. **Start local MongoDB:**

```bash
docker run --name mongodb \
  -e MONGO_INITDB_ROOT_USERNAME=admin \
  -e MONGO_INITDB_ROOT_PASSWORD=admin \
  -e MONGO_INITDB_DATABASE=admin \
  -p 127.0.0.1:27017:27017 \
  mongo:6.0.5
```

3. **Test the provider:**

```bash
cd examples/database
terraform init
terraform plan
terraform apply
```

### Running Tests

```bash
# Unit tests
make test

# Acceptance tests (requires MongoDB)
make testacc
```

## Documentation Development

### Generate Documentation

```bash
# Complete workflow: build + generate + validate
make docs

# Or individual steps:
make generate-schema  # Generate provider schema
make generate-docs    # Generate documentation
make validate-docs    # Validate documentation
```

For more details, see [DOCUMENTATION_WORKFLOW.md](DOCUMENTATION_WORKFLOW.md).

## Troubleshooting

### Connection Issues

**Problem:** `Error: connection to MongoDB failed`

**Solutions:**
1. Verify MongoDB is running: `mongosh "mongodb://localhost:27017"`
2. Check firewall rules
3. Verify credentials
4. Ensure authentication database exists (usually `admin`)

### Timeout Errors

**Problem:** `Error: context deadline exceeded`

**Solution:** Increase timeout in provider or resource configuration:

```hcl
provider "mongodb" {
  connection_string = var.mongodb_uri
  default_timeout   = 30  # minutes
}
```

### Version Compatibility

**Problem:** `Error: unsupported MongoDB version`

**Solution:** This provider supports MongoDB v6.x, v7.x, and v8.x. Check your version:

```bash
mongosh --eval "db.version()"
```

### State File Corruption

**Problem:** Terraform state is out of sync with MongoDB

**Solution:** Refresh state or re-import resources:

```bash
terraform refresh
# or
terraform import mongodb_database.example database_name
```

## Contributing

Contributions are welcome! Please feel free to submit pull requests, report bugs, or suggest features.

### How to Contribute

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices and conventions
- Add tests for new features
- Update documentation
- Run `make docs` before committing documentation changes
- Ensure all tests pass: `make test`

## Support

- 🐛 **Bug Reports:** [GitHub Issues](https://github.com/Ahton89/terraform-provider-mongodb/issues)
- 💡 **Feature Requests:** [GitHub Discussions](https://github.com/Ahton89/terraform-provider-mongodb/discussions)
- 📖 **Documentation:** [Terraform Registry](https://registry.terraform.io/providers/Ahton89/mongodb/latest/docs)


## License

This project is licensed under the Mozilla Public License Version 2.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework)
- MongoDB Go Driver by [MongoDB Inc](https://github.com/mongodb/mongo-go-driver)
- Inspired by the Terraform community and best practices

---

<div align="center">

**Created with ❤️ by [Ahton](https://github.com/Ahton89)**

If this project helps you, please consider giving it a ⭐️

[Report Bug](https://github.com/Ahton89/terraform-provider-mongodb/issues) · [Request Feature](https://github.com/Ahton89/terraform-provider-mongodb/issues) · [Documentation](https://registry.terraform.io/providers/Ahton89/mongodb/latest/docs)

</div>
