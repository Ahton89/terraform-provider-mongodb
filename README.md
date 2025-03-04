<a href="https://terraform.io">
    <img src=".github/tf.png" alt="Terraform logo" title="Terraform" align="left" height="50" />
</a>

# Terraform Provider for MongoDB
This is a Terraform provider for MongoDB. It allows you to manage MongoDB resources using Terraform.

## At the moment, the provider can:
- `Create/Delete` databases (collections)
> [!CAUTION]
> When changing the database name - the old database will be deleted. Carefully read the plan output, the "stringplanmodifier.RequiresReplace()" option was made specifically for this

- `Create/Delete/Modify` users and their permissions.
> [!CAUTION]
> When changing the username - the old user will be deleted. Carefully read the plan output, the "stringplanmodifier.RequiresReplace()" option was made specifically for this

- `Create/Modify` replica set.
> [!CAUTION]
> Delete replica set and change replica set name is not supported. This requires additional steps, so you will need to do it manually.

- `Return information` on all available databases (collections), users, replica set `[DATA SOURCES]`

- `Import` user, database and replica set data from an existing MongoDB instance.

> [!CAUTION]
> Support only MongoDB v6

## Examples
⚠️ See the terraform examples in the [examples](_examples) directory. 

#### Supported examples:
- [Create/Delete database](_examples/database)
- [Create/Delete/Modify user](_examples/user)
- [Create/Modify replica set](_examples/replicaset)
- [Databases data source](_examples/databases_datasource)
- [Users data source](_examples/users_datasource)
- [Replica set data source](_examples/replicaset_datasource)

## Local development
1. Clone the repository
2. Run `make install` to build the provider plugin
> [!TIP]
> By default, the provider will be installed in the `~/go/bin` directory. If you want change the installation directory, you can use the `make install BIN_DIR=/path/to/your/directory` command
3. Create `~/.terraformrc` file with the following content:
``` hcl
provider_installation {

  dev_overrides {
      "registry.terraform.io/Ahton89/mongodb" = "/Users/YOUR_USER/go/bin"
  }

  direct {}
}
```
> [!TIP]
> Replace `/Users/YOUR_USER/go/bin` with the path to the directory where the provider was installed in the previous step
4. Run your local docker mongodb instance 
```shell
  docker run --name mongodb -e MONGO_INITDB_ROOT_USERNAME=admin -e MONGO_INITDB_ROOT_PASSWORD=admin -e MONGO_INITDB_DATABASE=admin -p 127.0.0.1:27017:27017 mongo:6.0.5
```
5. Go to the [examples](_examples) directory and run `terraform plan` and `terraform apply` to test the provider

## Import

The provider supports importing existing resources. The import command is as follows:
```shell
  terraform import 'mongodb_database.database["example_database_1"]' example_database_1
```

## Requirements
-	[Terraform](https://www.terraform.io/downloads.html) 1.6.3+ (everything was tested on this version)
-	[Go](https://golang.org/doc/install) 1.23.4 (to build the provider plugin)


## Contributing to the provider
I would be very glad for your help. 🤗

---
Created with ❤️ by Ahton