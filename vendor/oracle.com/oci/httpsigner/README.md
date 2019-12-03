# httpsigner

The `httpsigner` library provides a Go implementation of [cavage-http-signatures][sig-spec]. It provides the underlying
signature functionality along with a few useful abstractions to simplify integration.  In addition, the package provides
a few extension points to allow adopters to add custom behavior.

Currently Supported Algorithms:
- rsa-sha256
- rsa-pss-sha256 (w/ 256-bit salt)


[sig-spec]: https://tools.ietf.org/html/draft-cavage-http-signatures-08

# Documentation

- [Usage](docs/usage.md)
    - Sign
        - [Sign a Request with `httpsigner.SignRequest`](docs/usage.md#sign-a-request-with-httpsignersignrequest)
        - [Sign a Request with `httpsigner.RequestSigner`](docs/usage.md#sign-a-request-with-httpsignerrequestsigner)
        - [Use `httpsigner.Client` to automatically sign outgoing requests](docs/usage.md#use-httpsignerclient-to-automatically-sign-outgoing-requests)
    - Verify
        - [Verify a request signature with `httpsigner.VerifyRequest`](docs/usage.md#verify-a-request-signature-with-httpsignerverifyrequest)
        - [Verify a request signature using an `httpsigner.RequestVerifier`](docs/usage.md#verify-a-request-signature-using-an-httpsignerrequestverifier)
- [Extension Points](docs/extension.md)
    - [`httpsigner.Algorithm`](docs/extension.md#algorithm)
    - [`httpsigner.AlgorithmSupplier` & `httpsigner.Algorithms`](docs/extension.md#algorithmsupplier-and-algorithms)
    - [`httpsigner.KeySupplier`](docs/extension.md#keysupplier)
    - [`httpsigner.RequestSigner`](docs/extension.md#requestsigner)
    - [`httpsigner.Client` & `httpsigner.SigningClient` Interfaces](docs/extension.md#client-and-signingclient-interfaces)
    - [`httpsigner.DefaultSigningClient`](docs/extension.md#defaultsigningclient)
    - [Key Rotation](docs/extension.md#key-rotation)
    - [`httpsigner.RequestVerifier`](docs/extension.md#requestverifier)

# Installation (Dyn)
### Glide
Use glide to alias the repository to `oracle.com` package host
```yaml
- package: oracle.com/oci/httpiam
  repo: ssh://git@github.corp.dyndns.com:oci/httpsigner.git
  vcs: git
```

### Manual
Clone the repo:
```bash
git clone git@github.corp.dyndns.com:oci/httpsigner.git $GOPATH/src/oracle.com/oci/httpsigner
```

Install:
```bash
cd $GOPATH/src/oracle.com/oci/httpsigner && go install
```

# Installation (OCI)
Use the same commands above but substitute git@github.corp.dyndns.com:oci for ssh://git@bitbucket.oci.oraclecorp.com:7999/goiam

### Glide
Use glide to alias the repository to `oracle.com` package host
```yaml
- package: oracle.com/oci/httpiam
  repo: ssh://git@bitbucket.oci.oraclecorp.com:7999/goiam/httpsigner.git
  vcs: git
```

### Manual
Clone the repo:
```bash
git clone ssh://git@bitbucket.oci.oraclecorp.com:7999/goiam/httpsigner.git $GOPATH/src/oracle.com/oci/httpsigner
```

Install:
```bash
cd $GOPATH/src/oracle.com/oci/httpsigner && go install
```

