terraform {
  required_providers {
    azurerm-quota = {
      source = "hashicorp.com/tnewman/azurerm-quota"
    }
  }
}

provider "azurerm-quota" {}
