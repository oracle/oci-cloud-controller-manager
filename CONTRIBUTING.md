# Contributing to the OCI Cloud Controller Manager

*Copyright (c) 2017 Oracle and/or its affiliates. All rights reserved.*

## Guidelines to raise a PR

### Contributor Agreement

Pull requests can be made under
[The Oracle Contributor Agreement](https://www.oracle.com/technetwork/community/oca-486395.html)
(OCA).
For pull requests to be accepted, the bottom of
your commit message must have the following line using your name and
e-mail address as it appears in the OCA Signatories list.
```
Signed-off-by: Your Name <you@example.org>
```
This can be automatically added to pull requests by committing with:
```
git commit --signoff
```
**Only pull requests from committers that can be verified as having
signed the OCA can be accepted.**

### Commit Message
* The commits message should prefix "External-ccm:" 
* All commits should be squashed to a single commit before merging

### Best Practices
* Follow the development guidelines [here](docs/development.md)
* govet, golint, gofmt should pass on the PR
* make targets "build" and "test" should be successful on the PR
* E2E should be run on a self managed test cluster, you will have to create a test cluster with the image generated from your changes. Please follow E2E guide [here](test/e2e/cloud-provider-oci/README.md)
* E2E tests should pass on 3 versions of kubernetes currently supported by the repo
