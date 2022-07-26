# Kf Operator

The Kf Operator, is an implementation of the [Operator Pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/) for Kf.

## Development

Development is done in a similar way to Kf codebase. Use `ko apply -f config/` with a kubeconfig that points to a GKE on GCP cluster.

## Testing

Testing is run automatically via prow during a code review. You can manually run the ci tests with:

```sh
./scripts/ci-test.sh
```
