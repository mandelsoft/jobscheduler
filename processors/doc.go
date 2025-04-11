/*
Package processors provides support for processing with limited parallelism.

Special synchronization objects support this functionality.
Everything works on context.Context objects. They have to feature
the pool attribute, which can be set with WithPool(ctx)`
Additionally, the operations react on cancelled contexts.

Basically, the pool should be assiged to a Go routine
to offer a really save operation, but this functionality
is explicitly forbidden in Go.
*/
package processors

import "context"

var _ context.Context
