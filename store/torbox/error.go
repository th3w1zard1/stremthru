package torbox

import (
	"net/http"

	"github.com/MunifTanjim/stremthru/core"
	"github.com/MunifTanjim/stremthru/store"
)

type ErrorCode string

const (
	ErrorCodeDatabaseError           ErrorCode = "DATABASE_ERROR"
	ErrorCodeUnknownError            ErrorCode = "UNKNOWN_ERROR"
	ErrorCodeNoAuth                  ErrorCode = "NO_AUTH"
	ErrorCodeBadToken                ErrorCode = "BAD_TOKEN"
	ErrorCodeAuthError               ErrorCode = "AUTH_ERROR"
	ErrorCodeInvalidOption           ErrorCode = "INVALID_OPTION"
	ErrorCodeRedirectError           ErrorCode = "REDIRECT_ERROR"
	ErrorCodeOAuthVerificationError  ErrorCode = "OAUTH_VERIFICATION_ERROR"
	ErrorCodeEndpointNotFound        ErrorCode = "ENDPOINT_NOT_FOUND"
	ErrorCodeItemNotFound            ErrorCode = "ITEM_NOT_FOUND"
	ErrorCodePlanRestrictedFeature   ErrorCode = "PLAN_RESTRICTED_FEATURE"
	ErrorCodeDuplicateItem           ErrorCode = "DUPLICATE_ITEM"
	ErrorCodeBozoRssFeed             ErrorCode = "BOZO_RSS_FEED"
	ErrorCodeSellixError             ErrorCode = "SELLIX_ERROR"
	ErrorCodeTooMuchData             ErrorCode = "TOO_MUCH_DATA"
	ErrorCodeDownloadTooLarge        ErrorCode = "DOWNLOAD_TOO_LARGE"
	ErrorCodeMissingRequiredOption   ErrorCode = "MISSING_REQUIRED_OPTION"
	ErrorCodeTooManyOptions          ErrorCode = "TOO_MANY_OPTIONS"
	ErrorCodeBozoTorrent             ErrorCode = "BOZO_TORRENT"
	ErrorCodeNoServersAvailableError ErrorCode = "NO_SERVERS_AVAILABLE_ERROR"
	ErrorCodeMonthlyLimit            ErrorCode = "MONTHLY_LIMIT"
	ErrorCodeCooldownLimit           ErrorCode = "COOLDOWN_LIMIT"
	ErrorCodeActiveLimit             ErrorCode = "ACTIVE_LIMIT"
	ErrorCodeDownloadServerError     ErrorCode = "DOWNLOAD_SERVER_ERROR"
	ErrorCodeBozoNzb                 ErrorCode = "BOZO_NZB"
	ErrorCodeSearchError             ErrorCode = "SEARCH_ERROR"
	ErrorCodeInvalidDevice           ErrorCode = "INVALID_DEVICE"
	ErrorCodeDiffIssue               ErrorCode = "DIFF_ISSUE"
	ErrorCodeLinkOffline             ErrorCode = "LINK_OFFLINE"
	ErrorCodeVendorDisabled          ErrorCode = "VENDOR_DISABLED"
)

func UpstreamErrorWithCause(cause error) *core.UpstreamError {
	err := core.NewUpstreamError("")
	err.StoreName = string(store.StoreNameTorBox)

	if rerr, ok := cause.(*ResponseError); ok {
		err.Msg = rerr.Detail
		err.UpstreamCause = rerr
	} else {
		err.Cause = cause
	}

	return err
}

func UpstreamErrorFromRequest(cause error, req *http.Request, res *http.Response) error {
	err := UpstreamErrorWithCause(cause)
	err.InjectReq(req)
	if res != nil {
		err.StatusCode = res.StatusCode
	}
	if err.StatusCode <= http.StatusBadRequest {
		err.StatusCode = http.StatusBadRequest
	}
	return err
}
