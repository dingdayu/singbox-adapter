// Package entity defines common responses and entities.
package entity

var (
	// SuccessResponse is the default successful response.
	SuccessResponse = Response{Code: 0, Message: "ok"} // success default

	// ErrInternalServer indicates an internal server error.
	ErrInternalServer = Response{Code: 10001, Message: "系统错误"}
	// ErrMissParams indicates missing parameters.
	ErrMissParams = Response{Code: 10002, Message: "缺少参数"}
	// ErrFailParams indicates invalid parameter format.
	ErrFailParams = Response{Code: 10003, Message: "参数格式错误"}
	// ErrNotExist indicates a missing record.
	ErrNotExist = Response{Code: 10004, Message: "数据不存在"}
	// ErrDefault indicates a general failure.
	ErrDefault = Response{Code: 10005, Message: "操作失败"}
	// ErrDataPermission indicates insufficient data permissions.
	ErrDataPermission = Response{Code: 10006, Message: "没有此数据权限"}
	// ErrDuplicatedKey indicates creation of a duplicate record is not allowed.
	ErrDuplicatedKey = Response{Code: 10007, Message: "无法创建重复数据"}

	// ErrAuthForbidden indicates authentication info not found.
	ErrAuthForbidden = Response{Code: 10100, Message: "登录信息不存在"}
	// ErrAuthUserNotFound indicates the user does not exist.
	ErrAuthUserNotFound = Response{Code: 10101, Message: "登陆用户不存在"}
	// ErrAuthTokenInvalid indicates the token is invalid or expired.
	ErrAuthTokenInvalid = Response{Code: 10102, Message: "Token 失效，需要重新登录"}
)

// Response is the common response structure.
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// NewErrResponse builds an error response.
func NewErrResponse(err any) Response {
	s := ErrInternalServer
	s.Error = err
	return s
}

// NewSucResponse builds a success response.
func NewSucResponse(data interface{}) Response {
	s := SuccessResponse
	s.Data = data
	return s
}

// NewErrParamsResponse builds a parameter error response.
func NewErrParamsResponse(message string) Response {
	s := ErrMissParams
	s.Error = message
	return s
}
