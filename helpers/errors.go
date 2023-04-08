package helpers

import "errors"

var ErrApiTokenChallengeInProgress error = errors.New("there is already a token challenge in progress. need to wait for timeout")
var ErrApiDisabled error = errors.New("api feature is not enabled for user")
var ErrApiInvalidTokenVerification error = errors.New("invalid token verification")