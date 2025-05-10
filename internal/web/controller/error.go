package controller

import (
	"encoding/json"
	"net/http"

	spec_v1 "github.com/opencontainers/distribution-spec/specs-go/v1"
)

const (
	errBlobUnknown         = "BLOB_UNKNOWN"          // blob unknown to registry
	errBlobUploadInvalid   = "BLOB_UPLOAD_INVALID"   // blob upload invalid
	errBlobUploadUnknown   = "BLOB_UPLOAD_UNKNOWN"   // blob upload unknown to registry
	errDigestInvalid       = "DIGEST_INVALID"        // provided digest did not match uploaded content
	errManifestBlobUnknown = "MANIFEST_BLOB_UNKNOWN" // manifest references a manifest or blob unknown to registry
	errManifestInvalid     = "MANIFEST_INVALID"      // manifest invalid
	errManifestUnknown     = "MANIFEST_UNKNOWN"      // manifest unknown to registry
	errNameInvalid         = "NAME_INVALID"          // invalid repository name
	errNameUnknown         = "NAME_UNKNOWN"          // repository not known to registry
	errSizeInvalid         = "SIZE_INVALID"          // provided length did not match content length
	errUnauthorized        = "UNAUTHORIZED"          // authentication required
	errDenied              = "DENIED"                // requested access to the resource is denied
	errUnsupported         = "UNSUPPORTED"           // the operation is unsupported
	errTooManyRequests     = "TOOMANYREQUESTS"       // too many requests
	errRangeInvalid        = "RANGE_INVALID"         // requested range not satisfiable
)

func sendError(w http.ResponseWriter, status int, code string, message string) {
	response := spec_v1.ErrorResponse{
		Errors: []spec_v1.ErrorInfo{
			{
				Code:    code,
				Message: message,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}
