# render-helm-chart: Kustomize Private Github-hosted Chart

### Overview

This example demonstrates how to declaratively invoke the `render-helm-chart`
function with kustomize with charts hosted in a private Github repo.

### Function invocation

To use the function with kustomize, you can specify the `functionConfig`
in your kustomization's `generators` field and create the Secret holding
the private repo credentials using the `secretGenerator`:

kustomization.yaml:
```yaml
secretGenerator:
  - name: my-secret
    envs:
    - credentials.env
    options:
      disableNameSuffixHash: true
      annotations:
        config.kubernetes.io/local-config: "true"

transformers:
  - |-
    apiVersion: fn.kpt.dev/v1alpha1
    kind: RenderHelmChart
    metadata:
      name: demo
      annotations:
        config.kubernetes.io/function: |
          container:
            network: true
            image: ghcr.io/kptdev/krm-functions-catalog/render-helm-chart:latest
    helmCharts:
      - chartArgs:
          name: mychart
          repo: https://raw.githubusercontent.com/kpt-helmfn-test-bot/private-helm-repo/main # change this to point to your private chart
          auth:
            kind: Secret
            name: my-secret
```

Then, to build the kustomization with kustomize v4:

```shell
kustomize build --enable-alpha-plugins --network .
```

### Expected result

You should see your inflated private chart in the output, demonstrating that the function was able to
use the provided credentials to access your private Github chart.
