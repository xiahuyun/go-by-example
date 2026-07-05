# informer 410 Gone real cluster demo

This demo uses a real local Kubernetes cluster to show what happens when a
watch starts from an expired `resourceVersion`.

It uses ConfigMaps in a dedicated namespace so the demo is cheap to run and
does not create Pods.

## Run

```bash
go run . --mode=manual
```

The manual mode:

1. Lists demo ConfigMaps and captures an old `resourceVersion`.
2. Creates, updates, and deletes many ConfigMaps to advance cluster revisions.
3. Starts a real watch from the old `resourceVersion`.
4. Detects `410 Gone` / `Expired`.
5. Recovers by relisting and starting a new watch from the fresh
   `resourceVersion`.

If your local cluster still accepts the old version, force an older one:

```bash
go run . --mode=manual --resource-version=1
```

Or increase churn:

```bash
go run . --mode=manual --churn=20000
```

## Informer mode

```bash
go run . --mode=informer --resource-version=1
```

Informer mode builds a real `cache.NewSharedIndexInformer` with a custom
`ListWatch`. The first watch deliberately uses the stale resourceVersion so the
apiserver can return `410 Gone`. The Reflector inside client-go then relists
and resumes watching.

This is the important production lesson:

- If you use informer, client-go handles expired watch history in the Reflector.
- Your event handlers normally see only add, update, and delete events.
- If you write a manual watch loop, you must handle `apierrors.IsGone(err)` or
  a `watch.Error` event with status code `410` / reason `Expired`, then relist.

## Useful flags

```text
--kubeconfig          path to kubeconfig, defaults to ~/.kube/config
--namespace           demo namespace, defaults to informer-410-demo
--mode                manual or informer
--resource-version    stale resourceVersion to use directly
--churn               number of ConfigMap operations, defaults to 3000
--cleanup             delete demo ConfigMaps on exit, defaults to true
--timeout             whole demo timeout, defaults to 3m
--watch-timeout       single watch timeout, defaults to 20s
--informer-run-for    informer mode runtime, defaults to 45s
```
