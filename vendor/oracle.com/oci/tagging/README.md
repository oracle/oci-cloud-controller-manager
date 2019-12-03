# tagging

Package `tagging` provides a utilities to generate a protobuf-serialized tag slug to send to the OCI Identity service for tagging authorization.

NOTE: The `vendor/` is added to the repo to ensure there are no conflicts with other versions of `protobuf` in clients (e.g. CoreDNS).

## Installation (Dyn)
### Glide
Use glide to alias the repository to `oracle.com` package host
```yaml
- package: oracle.com/oci/tagging
  repo: ssh://git@github.corp.dyndns.com:oci/tagging.git
  vcs: git
```

### Manual
Clone the repo:
```bash
git clone git@github.corp.dyndns.com:oci/tagging.git $GOPATH/src/oracle.com/oci/tagging
```

Install:
```bash
cd $GOPATH/src/oracle.com/oci/tagging && go install
```

## Usage

### FreeformTagSet

`FreeformTagSet` is a map of arbitrary keys to values.

### DefinedTagSet

`DefinedTagSet` is a nested map of tag namespaces to a map of defined key / value pairs.

### `NewTagSlug(freeformTags FreeformTagSet, definedTags DefinedTagSet) (*TagSlug, error)`

`NewTagSlug` creates a protobuf-serialized `[]byte` from a set of freeform and defined tags. This is used by `ociauthz` to perform tagging authorization.
