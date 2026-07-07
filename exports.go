package veloz

import "veloz/internal/runtime"

type Value = runtime.Value

type Callable = runtime.Callable

type SafeString = runtime.SafeString

type Attributer = runtime.Attributer

var (
	Nil     = runtime.Nil
	Bool    = runtime.Bool
	Int     = runtime.Int
	Float   = runtime.Float
	Str     = runtime.Str
	Object  = runtime.Object
	FromAny = runtime.FromAny
)
