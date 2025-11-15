terraform {
  required_providers {
    outpost = {
      source  = "outpost/outpost"
      version = "~> 0.1"
    }
  }
}

# Example 1: Basic usage with Helm values
locals {
  helm_values = {
    replicaCount = 3
    image = {
      repository = "nginx"
      tag        = "1.21.0"
      pullPolicy = null # This will be omitted
    }
    service = {
      type        = "ClusterIP"
      port        = 80
      annotations = {} # This will be omitted (empty map)
    }
    resources = null # This will be omitted
    autoscaling = {
      enabled     = true
      minReplicas = 2
      maxReplicas = 10
      targetCPU   = null # This will be omitted
    }
  }
}

output "helm_values_yaml" {
  value       = provider::outpost::helm_values_encode(local.helm_values)
  description = "Clean YAML output with null values omitted"
}

# Example 2: Complex nested structure
locals {
  complex_values = {
    global = {
      storageClass = "fast"
      imageRegistry = "ghcr.io"
    }
    postgresql = {
      enabled = true
      primary = {
        persistence = {
          enabled      = true
          size         = "10Gi"
          storageClass = null # Will use global default
        }
      }
      auth = {
        username = "appuser"
        password = null # Should be set separately
        database = "appdb"
      }
    }
    redis = {
      enabled = false
      # All nested values will be included even though enabled is false
      master = null
      replica = null
    }
  }
}

output "complex_values_yaml" {
  value       = provider::outpost::helm_values_encode(local.complex_values)
  description = "Complex nested structure with selective null omission"
}

# Example 3: List with null values
locals {
  list_values = {
    ingress = {
      enabled = true
      hosts = [
        {
          host = "example.com"
          paths = ["/"]
        },
        null, # This will be omitted
        {
          host = "www.example.com"
          paths = ["/", "/api"]
        }
      ]
      tls = [] # Empty list will be omitted
    }
  }
}

output "list_values_yaml" {
  value       = provider::outpost::helm_values_encode(local.list_values)
  description = "List structure with null values removed"
}

