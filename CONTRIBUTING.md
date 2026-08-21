# Contributing to the Gitea Terraform Provider

Thank you for your interest in contributing to the Gitea Terraform Provider! This document outlines the basic process to contribute effectively.

## Getting Started

1. **Fork the repository** on Gitea.
2. **Clone your fork** locally.
3. **Create a branch** for your changes.


See also [Hashicorp's Terraform Provider tutorial](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider)

## Testing your changes

### Running a local Gitea instance

```bash
docker run -p 3000:3000 gitea/gitea:latest-rootless
```

Visit http://localhost:3000 and go through the setup:
1. Url: http://localhost:3000
2. Username: gitea_admin
3. Password: gitea_admin
4. Email: admin@gitea.local

### Using the local version of the Provider
This involves 2 main steps:
- `go install .` to install the provider binary.
- Use .terraformrc file to point to that binary instead of the terraform registry.

See Terraform docs on [development overrides for provider developers](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers)

Also see [this tutorial section](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider#prepare-terraform-for-local-provider-install) for a bit more context.

Example linux `~/.terraformrc` file:
```
provider_installation {

  dev_overrides {
      "tfmm/gitea" = "/home/<username>/go/bin"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

### Applying the example module

With the local provider and the local Gitea instance set up properly as per above chapters, the example module can be applied:

- navigate into `examples/`
- run terraform commands



