terraform {
  required_providers {
    mongodb = {
      source = "registry.terraform.io/Ahton89/mongodb"
      version = "= 0.2.7"
    }
  }
}

provider "mongodb" {
  connection_string = "mongodb://admin:admin@127.0.0.1:27017/admin"
}

data "mongodb_databases" "example_mongodb" {}

output "example_mongodb" {
  value = data.mongodb_databases.example_mongodb
}