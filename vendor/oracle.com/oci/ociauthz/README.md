# ociauthz
Package `ociauthz` provides a collection of utilities to perform authentication and authorization against the OCI Identity Data plane.

**macOS Developers, see [Special Note for Developers running docker-for-mac on macOS](#macos-dev-note) below.**

## Documentation
- [KeySupplier](docs/key-supplier.md)
  - [X509CertificateSupplier](docs/key-supplier.md#x509certificatesupplier)
  - [STSKeySupplier](docs/key-supplier.md#stskeysupplier)
- [SigningClient](docs/client.md)
- [AuthorizationClient](docs/authz.md#ociauthzauthorizationclient)

## Installation (Dyn)
### Glide
Use glide to alias the repository to `oracle.com` package host
```yaml
- package: oracle.com/oci/ociauthz
  repo: ssh://git@github.corp.dyndns.com:oci/ociauthz.git
  vcs: git
```

### Manual
Clone the repo:
```bash
git clone git@github.corp.dyndns.com:oci/ociauthz.git $GOPATH/src/oracle.com/oci/ociauthz
```

Install:
```bash
cd $GOPATH/src/oracle.com/oci/ociauthz && go install
```

## Installation (OCI)
Use the same commands above but substitute `git@github.corp.dyndns.com:oci` for `ssh://git@bitbucket.oci.oraclecorp.com:7999/goiam`

### Glide
Use glide to alias the repository to `oracle.com` package host
```yaml
- package: oracle.com/oci/ociauthz
  repo: ssh://git@bitbucket.oci.oraclecorp.com:7999/goiam/ociauthz.git
  vcs: git
```

### Manual
Clone the repo:
```bash
git clone ssh://git@bitbucket.oci.oraclecorp.com:7999/goiam/ociauthz.git $GOPATH/src/oracle.com/oci/ociauthz
```

Install:
```bash
cd $GOPATH/src/oracle.com/oci/ociauthz && go install
```

## Build Notes
To build and run unit tests run make with the default target.
```
make
```

### DNS Resolution for Internal Dependencies
For the most part, developers should be able to build without issue.  In some environments (notably some Linux
environments) the docker container gets the default DNS server configuration which cannot resolve GitHub Enterprise or
BitBucket.  In this case, the environment variable `DOCKER_DNS_OVERRIDE` can be set to a DNS resolver which can provide
the address for GitHub Enterprise or BitBucket.

```
DOCKER_DNS_OVERRIDE=«dns resolver IP» make
```

Or include it in your local dev environment

```
export DOCKER_DNS_OVERRIDE=«dns resolver IP»
make
```

<a id="macos-dev-note"></a>

### Special Note for Developers running docker-for-mac on macOS
Unfortunately at present there is a bug related to how docker-for-mac handles forwarding of sockets.
[https://github.com/docker/for-mac/issues/410].  This causes problems when running the `vendor` make target.  There are
several workarounds to this involving various bits of chicanery with the ssh-agent, socket forwarding, etc.  The reader
is welcome to pursue getting those to work.  The most straight forward way to work around this is to run glide locally
and then run the `unittest` and `lint` targets manually.  Something to the effect of:

```
glide install
make unittest lint
```

Developers are advised to run at least `unittest` and `lint` before submitting as both will be run during the automated
pull request builder in TeamCity.  TeamCity build agents are Linux instances and are not affected by the socket
forwarding issue.
