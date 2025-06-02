# some-function-name

Note: Please ensure you follow the [kpt doc style guide].

## Overview

<!--mdtogo:Short-->

Explain what this function does in one or two sentences.

<!--mdtogo-->

Describe why the user should care about this function.

What problem does it solve?

Provide some context (e.g. In the `gatekeeper` function, explain what's
is `Gatekeeper` project)

[//]: <> (Note: The content between `<!--mdtogo:Short-->` and the following
`<!--mdtogo-->` will be used as the short description for the command.)

<!--mdtogo:Long-->

## Usage

How do I use this function?

Explain what does it do in details.

Is this function meant to be used declaratively, imperatively or both?

### FunctionConfig

Omit this section, if the function doesn't support any `functionConfigs`.
Otherwise, explain the function config and behavior for this function in detail.
For each field in the function config, specify:

- An example value
- Whether it is optional, and if so, the default value

If showing the function orchestrator (e.g. kpt) can make it clear about how to
use the function, it's recommended to use it.

[//]: <> (Note: The content between `<!--mdtogo:Long-->` and the following
`<!--mdtogo-->` will be used as the long description for the command.)

<!--mdtogo-->

## Examples

<!--mdtogo:Examples-->

Omit this section if you are providing complete example kpt packages which are
linked from the catalog site.

Otherwise, provide inline examples in this section.

[//]: <> (Note: The content between `<!--mdtogo:Examples-->` and the following
`<!--mdtogo-->` will be used as the examples for the command.)

<!--mdtogo-->

[kpt doc style guide]: https://github.com/kptdev/kpt/blob/main/docs/style-guides/docs.md
