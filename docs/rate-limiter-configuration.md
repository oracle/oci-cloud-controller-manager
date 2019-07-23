# Rate Limiting Configuration

This file defines a list of Rate Limiting configurations properties,
supported by the `oci-cloud-controller-manager`.

These properties exist in the `yaml` configuration under `rateLimiter`
in root. For example:

```yaml
auth:
  ...
loadBalancer:
  ...
rateLimiter:
  rateLimitQPSRead: ...
  rateLimitBucketRead: ...
  rateLimitQPSWrite: ...
  rateLimitBucketWrite: ...
```

## Request read/write rate limiting properties

| Name | Description | Default |
| ---- | ----------- | ------- |
| `rateLimitQPSRead` | The maximum queries allowed per second for read requests. | 20.0 |
| `rateLimitBucketRead` | The maximum token bucket burst size for read requests. | 5.0 |
| `rateLimitQPSWrite` | The maximum queries allwoed per second for write requests. | 20.0 |
| `rateLimitBucketWrite` | The maximim token bucket burst size for write requests. | 5.0 |

## Disable Rate Limiting
The rate limiting can be completely disabled by adding a property `disableRateLimiter: true`.
By default the property is `false`

```yaml
auth:
  ...
loadBalancer:
  ...
rateLimiter:
  disableRateLimiter: true
```
