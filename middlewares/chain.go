package middlewares

import "net/http"

func ChainMiddlewares(h http.Handler, mws ...func(http.Handler) http.Handler)http.Handler {
	for i := len(mws) - 1; i>=0; i-- {
		h = mws[i](h)
	}

	return h
}