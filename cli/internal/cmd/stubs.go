package cmd

import "errors"

// errNotImplemented is returned by verb command stubs that exist for help
// screen display but lack an API client implementation.
var errNotImplemented = errors.New("not implemented — this command requires an API client that has not been built yet")
