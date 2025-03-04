terraform {
  required_providers {
    mongodb = {
      source = "registry.terraform.io/Ahton89/mongodb"
      version = "= 0.2.4"
    }
  }
}

provider "mongodb" {
  connection_string = "mongodb://admin:admin@127.0.0.1:27017/admin"
}

locals {
  users = [
    {
      username = "example_user_1"
      password = "example_user_1_password"
      roles = [
        {
          database = "example_database_1"
          role = "readWrite"
        },
        {
          database = "example_database_2"
          role = "read"
        }
      ]
    },
    {
      username = "example_user_2"
      password = "example_user_2_password"
      roles = [
        {
          database = "example_database_2"
          role = "readWrite"
        },
        {
          database = "example_database_1"
          role = "read"
        }
      ]
    }
  ]
}

resource "mongodb_user" "user" {
  for_each = { for user in local.users : user.username => user }

  username = each.value.username
  password = each.value.password

  roles = [
    for role in each.value.roles : {
      database = role.database
      role = role.role
    }
  ]
}