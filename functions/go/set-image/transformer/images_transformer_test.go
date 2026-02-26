// Copyright 2026 The kpt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package transformer

import (
	"testing"

	"github.com/kptdev/krm-functions-sdk/go/fn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/types"
)

func TestUpdateContainerImages(t *testing.T) {
	const podYAML = `
apiVersion: v1
kind: Pod
metadata:
  name: myPod
spec:
  initContainers:
  - name: initBusybox
    image: busybox:1.36.1
  - name: initAlpine
    image: alpine:3.21
  containers:
  - name: liveNginx
    image: nginx:1.28.1
  - name: liveAlpine
    image: alpine:3.22
`
	const deploymentYAML = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
  labels:
    app: myApp
spec:
  selector:
    matchLabels:
      app: myApp
  template:
    spec:
      initContainers:
      - name: initBusybox
        image: busybox:1.36.1
      - name: initAlpine
        image: alpine:3.21
      containers:
      - name: liveNginx
        image: nginx:1.28.1
      - name: liveAlpine
        image: alpine:3.22
`

	t.Run("container images replaced by default", func(t *testing.T) {
		setImage := &SetImage{
			Image: types.Image{Name: "nginx", NewTag: "1.29.0"},
		}

		podKO, err := fn.ParseKubeObject([]byte(podYAML))
		require.NoError(t, err)

		err = setImage.updateContainerImages(podKO)
		require.NoError(t, err)

		assert.Equal(t, "nginx:1.29.0", podKO.GetMap("spec").GetSlice("containers")[0].GetString("image"))
		assert.Equal(t, 1, setImage.resultCount)
	})

	t.Run("initContainer images replaced by default", func(t *testing.T) {
		setImage := &SetImage{
			Image: types.Image{Name: "busybox", NewTag: "1.37.0"},
		}

		podKO, err := fn.ParseKubeObject([]byte(podYAML))
		require.NoError(t, err)

		err = setImage.updateContainerImages(podKO)
		require.NoError(t, err)

		assert.Equal(t, "busybox:1.37.0", podKO.GetMap("spec").GetSlice("initContainers")[0].GetString("image"))
		assert.Equal(t, 1, setImage.resultCount)
	})

	t.Run("just tag is replaced", func(t *testing.T) {
		setImage := &SetImage{
			Image: types.Image{Name: "alpine", NewTag: "3.24"},
		}

		podKO, err := fn.ParseKubeObject([]byte(podYAML))
		require.NoError(t, err)

		err = setImage.updateContainerImages(podKO)
		require.NoError(t, err)

		assert.Equal(t, "alpine:3.24", podKO.GetMap("spec").GetSlice("initContainers")[1].GetString("image"))
		assert.Equal(t, "alpine:3.24", podKO.GetMap("spec").GetSlice("containers")[1].GetString("image"))
		assert.Equal(t, 2, setImage.resultCount)
	})

	t.Run("tag preserved when only changing name", func(t *testing.T) {
		setImage := &SetImage{
			Image: types.Image{Name: "alpine", NewName: "my.docker.mirror.com/alpine"},
		}

		podKO, err := fn.ParseKubeObject([]byte(podYAML))
		require.NoError(t, err)

		err = setImage.updateContainerImages(podKO)
		require.NoError(t, err)

		assert.Equal(t, "my.docker.mirror.com/alpine:3.21", podKO.GetMap("spec").GetSlice("initContainers")[1].GetString("image"))
		assert.Equal(t, "my.docker.mirror.com/alpine:3.22", podKO.GetMap("spec").GetSlice("containers")[1].GetString("image"))
		assert.Equal(t, 2, setImage.resultCount)
	})

	t.Run("name and tag replace", func(t *testing.T) {
		setImage := &SetImage{
			Image: types.Image{Name: "alpine", NewName: "debian", NewTag: "bookworm"},
		}

		podKO, err := fn.ParseKubeObject([]byte(podYAML))
		require.NoError(t, err)

		err = setImage.updateContainerImages(podKO)
		require.NoError(t, err)

		assert.Equal(t, "debian:bookworm", podKO.GetMap("spec").GetSlice("initContainers")[1].GetString("image"))
		assert.Equal(t, "debian:bookworm", podKO.GetMap("spec").GetSlice("containers")[1].GetString("image"))
		assert.Equal(t, 2, setImage.resultCount)
	})

	t.Run("replace specific tag only", func(t *testing.T) {
		setImage := &SetImage{
			Image: types.Image{Name: "alpine:3.22", NewName: "debian", NewTag: "bookworm"},
		}

		podKO, err := fn.ParseKubeObject([]byte(podYAML))
		require.NoError(t, err)

		err = setImage.updateContainerImages(podKO)
		require.NoError(t, err)

		assert.Equal(t, "alpine:3.21", podKO.GetMap("spec").GetSlice("initContainers")[1].GetString("image"))
		assert.Equal(t, "debian:bookworm", podKO.GetMap("spec").GetSlice("containers")[1].GetString("image"))
		assert.Equal(t, 1, setImage.resultCount)
	})

	// This is just for demonstration purposes
	t.Run("no-op update still logs mutation", func(t *testing.T) {
		setImage := &SetImage{
			Image: types.Image{Name: "alpine", NewTag: "3.22"},
		}

		podKO, err := fn.ParseKubeObject([]byte(podYAML))
		require.NoError(t, err)

		err = setImage.updateContainerImages(podKO)
		require.NoError(t, err)

		assert.Equal(t, "alpine:3.22", podKO.GetMap("spec").GetSlice("initContainers")[1].GetString("image"))
		assert.Equal(t, "alpine:3.22", podKO.GetMap("spec").GetSlice("containers")[1].GetString("image"))
		assert.Equal(t, 2, setImage.resultCount)
	})

	t.Run("deployments are substituted", func(t *testing.T) {
		setImage := &SetImage{
			Image: types.Image{Name: "alpine", NewTag: "3.24"},
		}

		deploymentKO, err := fn.ParseKubeObject([]byte(deploymentYAML))
		require.NoError(t, err)

		err = setImage.updateContainerImages(deploymentKO)
		require.NoError(t, err)

		assert.Equal(t, "alpine:3.24", deploymentKO.GetMap("spec").
			GetMap("template").
			GetMap("spec").
			GetSlice("initContainers")[1].
			GetString("image"))

		assert.Equal(t, "alpine:3.24", deploymentKO.GetMap("spec").
			GetMap("template").
			GetMap("spec").
			GetSlice("containers")[1].
			GetString("image"))
	})
}
