package api

import (
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func (a AppState) RegisterCORSMiddleware(ctx huma.Context, next func(huma.Context)) {
	ctx.SetHeader("Access-Control-Allow-Origin", a.CORS)
	ctx.SetHeader("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	ctx.SetHeader("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

	fmt.Println("received request with method: ", ctx.Method())

	if ctx.Method() == http.MethodOptions {

		ctx.SetStatus(204)
		ctx.BodyWriter().Write([]byte{})
		return
	}

	next(ctx)
}
