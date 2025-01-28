terraform {
  required_providers {
    mongodb = {
      source = "registry.terraform.io/Ahton89/mongodb"
      version = "= 0.1.0"
    }
  }
}

provider "mongodb" {
  connection_string = "mongodb://admin:admin@127.0.0.1:27017/admin"
}

locals {
  databases = [
    {
      name = "example_database_1"
    },
    {
      name = "example_database_2"
    }
  ]
}

resource "mongodb_database" "database" {
  for_each = { for database in local.databases : database.name => database }

  name = each.value.name
}
