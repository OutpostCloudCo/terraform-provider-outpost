# Terraform Provider Outpost

The Outpost Terraform Provider provides utility functions for Terraform configurations, with a focus on generating clean YAML output for Helm values files.

## Features

- **`helm_values_encode` function**: Converts Terraform objects to YAML format while recursively removing null values and empty containers

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.4
- [Go](https://golang.org/doc/install) >= 1.22 (for development)

## Installation

### Using the Terraform Registry

Add this to your Terraform configuration:

```hcl
terraform {
  required_providers {
    outpost = {
      source  = "outpost/outpost"
      version = "~> 0.1"
    }
  }
}
```

### Using GitHub Releases (Development)

For development or custom builds, you can reference the provider from GitHub:

```hcl
terraform {
  required_providers {
    outpost = {
      source  = "github.com/outpost/terraform-provider-outpost"
      version = "~> 0.1"
    }
  }
}
```

## Usage

### helm_values_encode Function

The `helm_values_encode` function converts any Terraform object to YAML format, automatically removing:
- Null values
- Empty maps/objects
- Empty lists/arrays that become empty after null removal

This is particularly useful for generating Helm values files where you want to omit unset values rather than explicitly setting them to null.

#### Example

```hcl
terraform {
  required_providers {
    outpost = {
      source  = "outpost/outpost"
      version = "~> 0.1"
    }
  }
}

locals {
  helm_values = {
    replicaCount = 3
    image = {
      repository = "nginx"
      tag        = "1.21.0"
      pullPolicy = null  # This will be omitted
    }
    service = {
      type = "ClusterIP"
      port = 80
      annotations = {}  # This will be omitted (empty map)
    }
    resources = null  # This will be omitted
    autoscaling = {
      enabled     = true
      minReplicas = 2
      maxReplicas = 10
      targetCPU   = null  # This will be omitted
    }
  }
}

output "helm_values_yaml" {
  value = provider::outpost::helm_values_encode(local.helm_values)
}
```

The output will be:

```yaml
replicaCount: 3
image:
  repository: nginx
  tag: 1.21.0
service:
  type: ClusterIP
  port: 80
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
```

Notice that:
- `image.pullPolicy` (null) was omitted
- `service.annotations` (empty map) was omitted
- `resources` (null) was omitted
- `autoscaling.targetCPU` (null) was omitted

#### Using with kubernetes_manifest or helm_release

```hcl
resource "kubernetes_config_map" "helm_values" {
  metadata {
    name      = "app-values"
    namespace = "default"
  }

  data = {
    "values.yaml" = provider::outpost::helm_values_encode(local.helm_values)
  }
}

resource "helm_release" "app" {
  name       = "my-app"
  repository = "https://charts.example.com"
  chart      = "app"
  
  values = [
    provider::outpost::helm_values_encode(local.helm_values)
  ]
}
```

## Development

### Building the Provider

```bash
go build -o terraform-provider-outpost
```

### Testing

```bash
go test -v ./...
```

### Installing Locally

For local development, you can build and install the provider to your local Terraform plugin directory:

```bash
# Build the provider
go build -o terraform-provider-outpost

# Create the local plugin directory
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/outpost/outpost/0.1.0/$(go env GOOS)_$(go env GOARCH)/

# Copy the binary
cp terraform-provider-outpost ~/.terraform.d/plugins/registry.terraform.io/outpost/outpost/0.1.0/$(go env GOOS)_$(go env GOARCH)/
```

Then use this configuration in your Terraform files:

```hcl
terraform {
  required_providers {
    outpost = {
      source  = "registry.terraform.io/outpost/outpost"
      version = "0.1.0"
    }
  }
}
```

### Releasing

This provider uses [GoReleaser](https://goreleaser.com/) for releases. To create a new release:

1. Tag your commit:
   ```bash
   git tag -a v0.1.0 -m "Release v0.1.0"
   git push origin v0.1.0
   ```

2. GoReleaser will automatically build and publish the release to GitHub

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This provider is released under the MIT License.

