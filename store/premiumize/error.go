package premiumize

import (
	"net/http"

	"github.com/MunifTanjim/stremthru/core"
	"github.com/MunifTanjim/stremthru/store"
)

func UpstreamErrorWithCause(cause error) *core.UpstreamError {
	err := core.NewUpstreamError("")
	err.StoreName = string(store.StoreNamePremiumize)

	if rerr, ok := cause.(*ResponseContainer); ok {
		err.Msg = rerr.Message
		if err.Msg == "Not logged in." {
			err.Code = core.ErrorCodeUnauthorized
			err.StatusCode = http.StatusUnauthorized
		} else if err.Msg == "Account not premium." {
			err.Code = core.ErrorCodePaymentRequired
			err.StatusCode = http.StatusPaymentRequired
		}
		err.UpstreamCause = rerr
	} else {
		err.Cause = cause
	}

	return err
}
