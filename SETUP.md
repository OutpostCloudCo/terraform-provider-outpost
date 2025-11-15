# Setup Instructions for Terraform Provider Outpost

## Prerequisites

1. **Install Go** (version 1.22 or higher)
   ```bash
   # On Ubuntu/Debian
   sudo apt update
   sudo apt install -y golang-go
   
   # Or download from https://go.dev/dl/
   ```

2. **Verify Go installation**
   ```bash
   go version
   ```

## Building the Provider

1. **Navigate to the provider directory**
   ```bash
   cd /home/umbrel/outpost/terraform-provider-outpost
   ```

2. **Download dependencies**
   ```bash
   go mod tidy
   ```

3. **Build the provider**
   ```bash
   go build -o terraform-provider-outpost
   ```

4. **Run tests**
   ```bash
   go test -v ./...
   ```

## Local Installation

To use the provider locally with Terraform:

```bash
make install
```

Or manually:

```bash
# Build the provider
go build -o terraform-provider-outpost

# Create the local plugin directory
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/outpost/outpost/0.1.0/$(go env GOOS)_$(go env GOARCH)/

# Copy the binary
cp terraform-provider-outpost ~/.terraform.d/plugins/registry.terraform.io/outpost/outpost/0.1.0/$(go env GOOS)_$(go env GOARCH)/
```

## Testing the Provider

1. **Create a test Terraform configuration**
   ```bash
   cd examples/functions/helm_values_encode
   terraform init
   terraform plan
   ```

2. **View the outputs**
   ```bash
   terraform apply
   ```

## Publishing to GitHub

1. **Initialize a Git repository** (if not already done)
   ```bash
   cd /home/umbrel/outpost/terraform-provider-outpost
   git init
   git add .
   git commit -m "Initial commit: Terraform Provider Outpost with helm_values_encode function"
   ```

2. **Create a GitHub repository**
   - Go to https://github.com/new
   - Create a repository named `terraform-provider-outpost`
   - Do NOT initialize with README, license, or .gitignore

3. **Push to GitHub**
   ```bash
   git remote add origin https://github.com/YOUR_ORG/terraform-provider-outpost.git
   git branch -M main
   git push -u origin main
   ```

4. **Set up GPG signing for releases** (required for Terraform Registry)
   ```bash
   # Generate a GPG key if you don't have one
   gpg --full-generate-key
   
   # List your keys
   gpg --list-secret-keys --keyid-format=long
   
   # Export your public key
   gpg --armor --export YOUR_KEY_ID > public-key.asc
   ```

5. **Add GitHub Secrets** for automated releases:
   - Go to your repository settings → Secrets and variables → Actions
   - Add these secrets:
     - `GPG_PRIVATE_KEY`: Your GPG private key (use `gpg --armor --export-secret-keys YOUR_KEY_ID`)
     - `PASSPHRASE`: Your GPG key passphrase

6. **Create a release**
   ```bash
   git tag -a v0.1.0 -m "Initial release"
   git push origin v0.1.0
   ```
   
   The GitHub Actions workflow will automatically build and publish the release.

## Using the Provider

### From Local Installation

```hcl
terraform {
  required_providers {
    outpost = {
      source  = "registry.terraform.io/outpost/outpost"
      version = "0.1.0"
    }
  }
}

output "test" {
  value = provider::outpost::helm_values_encode({
    name = "test"
    enabled = true
    nothing = null
  })
}
```

### From GitHub (after publishing)

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

### For Terraform Registry

After publishing to the Terraform Registry:

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

## Updating Your Existing Terraform Configuration

To use this provider in your existing configuration at:
`/home/umbrel/outpost/outpost-service-management/terraform/terraform-argocd-outpost-service/terraform.tf`

Replace:
```hcl
helmvaluesencoding = {
  source  = "spacelift.io/outpostcloud25/helmvaluesencoding"
  version = "0.0.0-SNAPSHOT-none"
}
```

With:
```hcl
outpost = {
  source  = "registry.terraform.io/outpost/outpost"
  version = "0.1.0"
}
```

And update any function calls from:
```hcl
provider::helmvaluesencoding::helm_values_encode(...)
```

To:
```hcl
provider::outpost::helm_values_encode(...)
```

## Troubleshooting

### "provider not found" error
- Make sure you've run `make install` or manually installed the provider
- Check that the provider binary exists in `~/.terraform.d/plugins/`
- Run `terraform init` in your Terraform configuration directory

### Build errors
- Ensure Go 1.22 or higher is installed
- Run `go mod tidy` to ensure all dependencies are downloaded
- Check for any syntax errors in the Go files

### Testing errors
- Make sure all dependencies are installed
- Run `go mod download` to fetch missing dependencies
- Check that test files have proper imports

## Next Steps

1. Install Go if not already installed
2. Build and test the provider locally
3. Push to GitHub
4. Set up GPG signing
5. Create a release
6. Optionally: Publish to Terraform Registry (https://www.terraform.io/registry/providers/publishing)

