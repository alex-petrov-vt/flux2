terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "2.65.0"
    }
  }
}

provider "azurerm" {
  features {}
}

resource "random_pet" "prefix" {}

locals {
  name_prefix = "${random_pet.prefix.id}-e2e"
}

resource "azurerm_resource_group" "this" {
  name     = "${local.name_prefix}-rg"
  location = "West Europe"

  tags = {
    environment = "e2e"
  }
}

resource "azurerm_kubernetes_cluster" "this" {
  name                = "${local.name_prefix}-aks"
  location            = azurerm_resource_group.this.location
  resource_group_name = azurerm_resource_group.this.name
  dns_prefix          = "${local.name_prefix}-k8s"

  default_node_pool {
    name            = "default"
    node_count      = 2
    vm_size         = "Standard_B2s"
    os_disk_size_gb = 30
  }

  identity {
    type = "SystemAssigned"
  }

  role_based_access_control {
    enabled = true
  }

  network_profile {
    network_plugin = "kubenet"
    network_policy = "calico"
  }

  tags = {
    environment = "e2e"
  }
}
