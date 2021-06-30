# Azure E2E

E2E tests for Azure are needed to mitigate introduction of new bugs in dependencies like libgit2. The point
is also to verify other Azure service integrations are actually working as expected.

## Tests

* Flux can be successfully installed on AKS using the CLI e.g.:
* flux install --components-extra=image-reflector-controller,image-automation-controller --network-policy=false
* source-controller can clone Azure DevOps repositories (https+ssh)
* source-controller can pull charts from Azure Container Registry Helm repositories
* image-reflector-controller can list tags from Azure Container Registry image repositories
* image-automation-controller can create branches and push to Azure DevOps repositories (https+ssh)
* kustomize-controller can decrypt secrets using SOPS and Azure Key Vault
* notification-controller can send commit status to Azure DevOps
* notification-controller can forward events to Azure Event Hub

## Architecture

The tests are designed so that as little computation has to be done when running the tests. The majority of Git repositories and Helm Charts should already be created.



## Running
