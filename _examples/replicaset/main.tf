terraform {
  required_providers {
    mongodb = {
      source = "registry.terraform.io/Ahton89/mongodb"
      version = "= 0.2.0"
    }
  }
}

provider "mongodb" {
  connection_string = "mongodb://admin:admin@127.0.0.1:27017/admin"
}

locals {
  replicaset_name = "example"
  members = [
    {
      id: 0,
      host = "127.0.0.1:27017"
    },
    {
      id: 1,
      host = "127.0.0.1:27018"
    },
    {
      id: 2,
      host = "127.0.0.1:27019"
    }
  ]
}

resource "mongodb_replicaset" "example" {
  name = local.replicaset_name
  members = local.members
}