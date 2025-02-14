package pikpak

import (
	"github.com/MunifTanjim/stremthru/core"
	"github.com/MunifTanjim/stremthru/store"
)

var ErrByCode = map[int]string{
	3:    "invalid_argument",
	5:    "not_found", // 'file_not_found'
	8:    "file_space_not_enough",
	9:    "captcha_invalid", // 'file_in_recycle_bin'
	16:   "unauthenticated",
	4002: "captcha_invalid",
	4022: "invalid_account_or_password",
	4126: "invalid_grant",
}

var errorCodeByErr = map[string]core.ErrorCode{
	"invalid_argument":            core.ErrorCodeBadRequest,
	"not_found":                   core.ErrorCodeNotFound,
	"file_space_not_enough":       core.ErrorCodeStoreLimitExceeded,
	"captcha_invalid":             core.ErrorCodeForbidden,
	"unauthenticated":             core.ErrorCodeUnauthorized,
	"invalid_account_or_password": core.ErrorCodeUnauthorized,
	"invalid_grant":               core.ErrorCodeUnauthorized,
}

func TranslateErr(err string) core.ErrorCode {
	if code, ok := errorCodeByErr[err]; ok {
		return code
	}
	return core.ErrorCodeUnknown
}

func UpstreamErrorWithCause(cause error) *core.UpstreamError {
	err := core.NewUpstreamError("")
	err.StoreName = string(store.StoreNamePikPak)

	if rerr, ok := cause.(*ResponseContainer); ok {
		err.Msg = rerr.Err
		if rerr.ErrDesc != "" {
			err.Msg += ": " + rerr.ErrDesc
		}
		err.Code = TranslateErr(rerr.Err)
		err.UpstreamCause = rerr
	} else {
		err.Cause = cause
	}

	return err
}
