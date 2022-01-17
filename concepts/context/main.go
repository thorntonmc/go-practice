package main

import (
	"context"
	"net/http"
)

/*
 *
 * basics
 *
 */

// a context is an instance that meets the Context interface
// context.Context

// it is conventionally the first arg to a function
func takeContext(ctx context.Context) {

}

// when there is no existing context, like at the beginning of
// your program, create one with conext.Background()

func _main() {
	// this returns empty context.Context interface
	ctx := context.Background()
	takeContext(ctx)

	// if you are unsure of the purpose of the context, use
	// context.TODO. This also returns an empty conext.Context
	// interface, but shouldn't ever be in production code.

	toDo := context.TODO()
	takeContext(toDo)
}

/*
 *
 * middleware, an example
 *
 */

func Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		// wrap the context with stuff
		req = req.WithContext(ctx)
		handler.ServeHTTP(rw, req)
	})
}
