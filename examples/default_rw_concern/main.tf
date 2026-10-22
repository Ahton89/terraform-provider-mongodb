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

resource "mongodb_default_rw_concern" "example" {
  default_read_concern = {
    level = "majority"
  }

  default_write_concern = {
    w        = "majority"
    journal  = true
    wtimeout = 5000
  }

  timeouts = {
    create = "5m"
    read   = "2m"
    update = "5m"
  }
}
