terraform {
  required_providers {
    mongodb = {
      source = "registry.terraform.io/Ahton89/mongodb"
      version = "= 0.3.0"
    }
  }
}

provider "mongodb" {
  connection_string = "mongodb://admin:admin@127.0.0.1:27017,127.0.0.1:27018,127.0.0.1:27019/admin?replicaSet=example"
}

data "mongodb_default_rw_concern" "current" {}

output "default_read_concern" {
  value = data.mongodb_default_rw_concern.current.default_read_concern
}

output "default_write_concern" {
  value = data.mongodb_default_rw_concern.current.default_write_concern
}
