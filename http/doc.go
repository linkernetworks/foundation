// Copyright 2018 The LinkerNetworks Authors.
// All Rights Reserved

// This package provides HTTP response helper functions for writing error messages
// To use the helper:
//
//  import "bitbucket.org/linkernetworks/src/net/http/response"
//  response.InternalServerError(req, resp, errors.New("some error message"))
//
// To use the responsetest pacakge to verify the HTTP response:
//
//  responsetest.AssertErrorMessage(t, resp, "some error message")
package http
