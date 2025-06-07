package searchreplace

var searchReplaceCases = []test{
	{
		name: "search by value",
		config: `
data:
  by-value: '3'
`,
		input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  foo: 3
---
apiVersion: apps/v1
kind: Service
metadata:
  name: nginx-service
 `,
		out: `${filePath}
fieldPath: spec.replicas
value: 3

${filePath}
fieldPath: spec.foo
value: 3

Matched 2 field(s)
`,
		expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  foo: 3
---
apiVersion: apps/v1
kind: Service
metadata:
  name: nginx-service
`,
	},
	{
		name: "search replace by value",
		config: `
data:
  by-value: '3'
  put-value: '4'
`,
		input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
---
apiVersion: apps/v1
kind: Service
metadata:
  name: nginx-service
foo:
  bar: 3
 `,
		out: `${filePath}
fieldPath: spec.replicas
value: 4

${filePath}
fieldPath: foo.bar
value: 4

Mutated 2 field(s)
`,
		expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 4
---
apiVersion: apps/v1
kind: Service
metadata:
  name: nginx-service
foo:
  bar: 4
 `,
	},
	{
		name: "search replace by value to different type 1",
		config: `
data:
  by-value: '3'
  put-value: four
`,
		input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
---
apiVersion: apps/v1
kind: Service
metadata:
  name: nginx-service
foo:
  bar: 3
 `,
		out: `${filePath}
fieldPath: spec.replicas
value: four

${filePath}
fieldPath: foo.bar
value: four

Mutated 2 field(s)
`,
		expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: four
---
apiVersion: apps/v1
kind: Service
metadata:
  name: nginx-service
foo:
  bar: four
 `,
	},
	{
		name: "search replace by value to different type 2",
		config: `
data:
  by-value: four
  put-value: '4'
`,
		input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  foo: four
 `,
		out: `${filePath}
fieldPath: spec.foo
value: 4

Mutated 1 field(s)
`,
		expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  foo: 4
 `,
	},
	{
		name: "search replace multiple deployments",
		config: `
data:
  by-value: '3'
  put-value: '4'
`,
		input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mysql-deployment
spec:
  replicas: 3
 `,
		out: `${filePath}
fieldPath: spec.replicas
value: 4

${filePath}
fieldPath: spec.replicas
value: 4

Mutated 2 field(s)
`,
		expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 4
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mysql-deployment
spec:
  replicas: 4
 `,
	},
	{
		name: "search replace multiple deployments different value",
		config: `
data:
  by-value: '3'
  put-value: '4'
`,
		input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mysql-deployment
spec:
  replicas: 5
 `,
		out: `${filePath}
fieldPath: spec.replicas
value: 4

Mutated 1 field(s)
`,
		expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 4
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mysql-deployment
spec:
  replicas: 5
 `,
	},
	{
		name: "search by regex",
		config: `
data:
  by-value-regex: nginx-(.*)
`,
		input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
 `,
		out: `${filePath}
fieldPath: metadata.name
value: nginx-deployment

Matched 1 field(s)
`,
		expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
 `,
	},
	{
		name: "search replace by regex",
		config: `
data:
  by-value-regex: nginx-(.*)
  put-value: ubuntu-deployment
`,
		input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
 `,
		out: `${filePath}
fieldPath: metadata.name
value: ubuntu-deployment

Mutated 1 field(s)
`,
		expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ubuntu-deployment
spec:
  replicas: 3
 `,
	},
	{
		name: "search replace by regex helm template and empty values",
		config: `
data:
  by-value-regex: nginx-(.*)
  put-value: ubuntu-deployment
`,
		input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: {? {replicas: ''} : ''}
  foo:
 `,
		out: `${filePath}
fieldPath: metadata.name
value: ubuntu-deployment

Mutated 1 field(s)
`,
		expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ubuntu-deployment
spec:
  replicas: {? {replicas: ''} : ''}
  foo:
 `,
	},
	{
		name: "search by path",
		config: `
data:
  by-path: spec.replicas
`,
		input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  foo: 3
---
apiVersion: apps/v1
kind: Service
metadata:
  name: nginx-service
 `,
		out: `${filePath}
fieldPath: spec.replicas
value: 3

Matched 1 field(s)
`,
		expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  foo: 3
---
apiVersion: apps/v1
kind: Service
metadata:
  name: nginx-service
 `,
	},
	{
		name: "search by array path",
		config: `
data:
  by-path: spec.foo[1]
`,
		input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  foo:
    - a
    - b
---
apiVersion: apps/v1
kind: Service
metadata:
  name: nginx-service
 `,
		out: `${filePath}
fieldPath: spec.foo[1]
value: b

Matched 1 field(s)
`,
		expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  foo:
  - a
  - b
---
apiVersion: apps/v1
kind: Service
metadata:
  name: nginx-service
 `,
	},
	{
		name: "search by array path all elements",
		config: `
data:
  by-path: spec.foo[*]
`,
		input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  foo:
  - a
  - b
---
apiVersion: apps/v1
kind: Service
metadata:
  name: nginx-service
 `,
		out: `${filePath}
fieldPath: spec.foo[0]
value: a

${filePath}
fieldPath: spec.foo[1]
value: b

Matched 2 field(s)
`,
		expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  foo:
  - a
  - b
---
apiVersion: apps/v1
kind: Service
metadata:
  name: nginx-service
 `,
	},
	{
		name: "search replace by array path regex",
		config: `
data:
  by-path: spec.foo[1]
  put-value: c
`,
		input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  foo:
  - a
  - b
 `,
		out: `${filePath}
fieldPath: spec.foo[1]
value: c

Mutated 1 field(s)
`,
		expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  foo:
  - a
  - c
 `,
	},
	{
		name: "search replace by array path out of bounds",
		config: `
data:
  by-path: spec.foo[2]
  put-value: c
`,
		input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  foo:
  - a
  - b
 `,
		out: `Mutated 0 field(s)
`,
		expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  foo:
  - a
  - b
 `,
	},
	{
		name: "search replace by array objects path",
		config: `
data:
  by-path: spec.foo[1].c
  put-value: thing-new
`,
		input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  foo:
  - c: thing0
  - c: thing1
  - c: thing2
 `,
		out: `${filePath}
fieldPath: spec.foo[1].c
value: thing-new

Mutated 1 field(s)
`,
		expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  foo:
  - c: thing0
  - c: thing-new
  - c: thing2
 `,
	},
	{
		name: "replace by path and value",
		config: `
data:
  by-path: spec.replicas
  by-value: '3'
  put-value: '4'
`,
		input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  foo: 3
---
apiVersion: apps/v1
kind: Service
metadata:
  name: nginx-service
 `,
		out: `${filePath}
fieldPath: spec.replicas
value: 4

Mutated 1 field(s)
`,
		expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 4
  foo: 3
---
apiVersion: apps/v1
kind: Service
metadata:
  name: nginx-service
 `,
	},
	{
		name: "add non-existing field",
		config: `
data:
  by-path: metadata.namespace
  put-value: myspace
`,
		input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
 `,
		out: `${filePath}
fieldPath: metadata.namespace
value: myspace

Mutated 1 field(s)
`,
		expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  namespace: myspace
spec:
  replicas: 3
 `,
	},
	{
		name: "put value by regex capture groups",
		config: `
data:
  by-value-regex: (\w+)-dev-(\w+)-us-east-1-(\w+)
  put-value: ${1}-prod-${2}-us-central-1-${3}
`,
		input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo1-dev-bar1-us-east-1-baz1
  namespace: foo2-dev-bar2-us-east-1-baz2
 `,
		out: `${filePath}
fieldPath: metadata.name
value: foo1-prod-bar1-us-central-1-baz1

${filePath}
fieldPath: metadata.namespace
value: foo2-prod-bar2-us-central-1-baz2

Mutated 2 field(s)
`,
		expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo1-prod-bar1-us-central-1-baz1
  namespace: foo2-prod-bar2-us-central-1-baz2
 `,
	},
	{
		name: "error when both by-value and by-regex provided",
		config: `
data:
  by-value: nginx-deployment
  by-value-regex: nginx-(.*)
`,
		input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
 `,
		expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
 `,
		errMsg: `only one of ["by-value", "by-value-regex"] can be provided`,
	},
	{
		name: "do not error when none of the search matchers are provided",
		config: `
data: ~
`,
		input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
 `,
		expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
 `,
		out: `Matched 0 field(s)
`,
	},
	{
		name: "do not error when none of the required search matchers are provided",
		config: `
data:
  put-value: foo
`,
		input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
 `,
		expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
 `,
		out: `Mutated 0 field(s)
`,
	},
}
