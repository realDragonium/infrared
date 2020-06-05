package api

type errorResp struct {
	Message string `json:"error"`
}

func (err errorResp) Error() string {
	return err.Message
}

var (
	errNotAnAdmin         = errorResp{Message: "you are not an admin"}
	errUsernameEmpty      = errorResp{Message: "username cannot be empty"}
	errTokenSigning       = errorResp{Message: "failed to sign the token"}
	errTokenOutdated      = errorResp{Message: "time constraints of token exceeded"}
	errTokenInvalid       = errorResp{Message: "invalid token"}
	errInvalidCredentials = errorResp{Message: "invalid username or password"}
	errPasswordHashing    = errorResp{Message: "failed to hash the password"}
	errUsernameTaken      = errorResp{Message: "username already taken"}
)
