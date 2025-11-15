# Quick Start Guide

## 🎯 What You Have

A complete Terraform provider with the `helm_values_encode` function that:
- Converts Terraform objects to YAML
- Removes null values recursively
- Removes empty containers (maps/lists) that become empty after null removal
- Perfect for generating clean Helm values files

## 📁 Project Structure

```
terraform-provider-outpost/
├── .github/
│   └── workflows/
│       ├── release.yml       # Automated releases on git tags
│       └── test.yml          # CI testing on PRs and pushes
├── examples/
│   └── functions/
│       └── helm_values_encode/
│           └── main.tf       # Usage examples
├── function_helm_values_encode.go      # Your function implementation
├── function_helm_values_encode_test.go # Unit tests
├── go.mod                              # Go module definition
├── main.go                             # Provider entry point
├── provider.go                         # Provider definition
├── LICENSE                             # MIT License
├── Makefile                            # Build commands
├── README.md                           # Full documentation
├── SETUP.md                            # Detailed setup instructions
├── .gitignore                          # Git ignore rules
├── .golangci.yml                       # Linter configuration
└── .goreleaser.yml                     # Release automation config
```

## 🚀 Quick Setup

### 1. Install Go (if not already installed)

```bash
# Ubuntu/Debian
sudo apt update && sudo apt install -y golang-go

# Or download from https://go.dev/dl/
```

### 2. Build the Provider

```bash
cd /home/umbrel/outpost/terraform-provider-outpost
go mod tidy      # Download dependencies
go build         # Build the provider
```

### 3. Install Locally

```bash
make install
```

This installs the provider to `~/.terraform.d/plugins/` so Terraform can find it.

### 4. Test It

```bash
cd examples/functions/helm_values_encode
terraform init
terraform apply
```

## 📝 Usage Example

```hcl
terraform {
  required_providers {
    outpost = {
      source  = "registry.terraform.io/outpost/outpost"
      version = "0.1.0"
    }
  }
}

locals {
  helm_values = {
    replicaCount = 3
    image = {
      repository = "nginx"
      tag        = "1.21.0"
      pullPolicy = null  # Will be omitted
    }
    resources = null     # Will be omitted
    annotations = {}     # Will be omitted (empty)
  }
}

output "yaml" {
  value = provider::outpost::helm_values_encode(local.helm_values)
}
```

Output:
```yaml
replicaCount: 3
image:
  repository: nginx
  tag: 1.21.0
```

## 🔄 Integration with Your Existing Code

Your existing configuration at:
`/home/umbrel/outpost/outpost-service-management/terraform/terraform-argocd-outpost-service/terraform.tf`

**Currently has:**
```hcl
helmvaluesencoding = {
  source  = "spacelift.io/outpostcloud25/helmvaluesencoding"
  version = "0.0.0-SNAPSHOT-none"
}
```

**Change to:**
```hcl
outpost = {
  source  = "registry.terraform.io/outpost/outpost"
  version = "0.1.0"
}
```

**Update function calls from:**
```hcl
provider::helmvaluesencoding::helm_values_encode(...)
```

**To:**
```hcl
provider::outpost::helm_values_encode(...)
```

## 🌐 Publishing to GitHub

### 1. Create GitHub Repository

```bash
cd /home/umbrel/outpost/terraform-provider-outpost
git init
git add .
git commit -m "Initial commit: Terraform Provider Outpost"
```

### 2. Push to GitHub

```bash
# Replace YOUR_ORG with your GitHub organization/username
git remote add origin https://github.com/YOUR_ORG/terraform-provider-outpost.git
git branch -M main
git push -u origin main
```

### 3. Create First Release

```bash
git tag -a v0.1.0 -m "Initial release"
git push origin v0.1.0
```

GitHub Actions will automatically:
- Build binaries for all platforms
- Create checksums
- Sign the release
- Publish to GitHub Releases

### 4. Configure GitHub Secrets

For automatic releases to work, add these secrets in your GitHub repo settings:

- `GPG_PRIVATE_KEY`: Your GPG private key for signing
- `PASSPHRASE`: Your GPG key passphrase

See `SETUP.md` for details on generating GPG keys.

## 🔍 Using the Provider

### Option 1: Local Development (after `make install`)

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

### Option 2: From GitHub (after first release)

```hcl
terraform {
  required_providers {
    outpost = {
      source  = "github.com/YOUR_ORG/terraform-provider-outpost"
      version = "~> 0.1"
    }
  }
}
```

### Option 3: Terraform Registry (after publishing)

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

## 🧪 Testing

```bash
# Run unit tests
go test -v ./...

# Build
go build

# Install locally
make install

# Format code
make fmt

# Run linter
make lint
```

## 📚 Next Steps

1. ✅ Provider is created with your function
2. ⬜ Install Go and build the provider
3. ⬜ Test locally with the examples
4. ⬜ Push to GitHub
5. ⬜ Set up GPG signing for releases
6. ⬜ Create your first release (v0.1.0)
7. ⬜ Update your existing Terraform configs to use the new provider
8. ⬜ (Optional) Publish to Terraform Registry

## ❓ Need Help?

- See `README.md` for full documentation
- See `SETUP.md` for detailed setup instructions
- Check `examples/` for usage examples
- Review `.github/workflows/` for CI/CD configuration

## 🎉 What Makes This Special

This provider is ready for:
- ✅ GitHub distribution
- ✅ Automated releases with GoReleaser
- ✅ GPG signing for security
- ✅ CI/CD with GitHub Actions
- ✅ Multi-platform support (Linux, macOS, Windows, FreeBSD)
- ✅ Multiple architectures (amd64, arm64, arm, 386)
- ✅ Terraform Registry publishing (when you're ready)

