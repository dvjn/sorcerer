package service

import (
	"encoding/json"
	"net/http"

	"github.com/dvjn/sorcerer/pkg/models"
)

type RegistryError struct {
	Code   string
	Status int
}

var (
	// BLOB_UNKNOWN - blob unknown to registry
	errBlobUnknown = &RegistryError{
		Code:   "BLOB_UNKNOWN",
		Status: http.StatusNotFound,
	}
	// BLOB_UPLOAD_INVALID - blob upload invalid
	errBlobUploadInvalid = &RegistryError{
		Code:   "BLOB_UPLOAD_INVALID",
		Status: http.StatusBadRequest,
	}
	// BLOB_UPLOAD_UNKNOWN - blob upload unknown to registry
	errBlobUploadUnknown = &RegistryError{
		Code:   "BLOB_UPLOAD_UNKNOWN",
		Status: http.StatusNotFound,
	}
	// DIGEST_INVALID - provided digest did not match uploaded content
	errDigestInvalid = &RegistryError{
		Code:   "DIGEST_INVALID",
		Status: http.StatusBadRequest,
	}
	// MANIFEST_BLOB_UNKNOWN - manifest references a manifest or blob unknown to registry
	errManifestBlobUnknown = &RegistryError{
		Code:   "MANIFEST_BLOB_UNKNOWN",
		Status: http.StatusBadRequest,
	}
	// MANIFEST_INVALID - manifest invalid
	errManifestInvalid = &RegistryError{
		Code:   "MANIFEST_INVALID",
		Status: http.StatusBadRequest,
	}
	// MANIFEST_UNKNOWN - manifest unknown to registry
	errManifestUnknown = &RegistryError{
		Code:   "MANIFEST_UNKNOWN",
		Status: http.StatusNotFound,
	}
	// REPOSITORY_INVALID - invalid repository
	errRepositoryInvalid = &RegistryError{
		Code:   "REPOSITORY_INVALID",
		Status: http.StatusBadRequest,
	}
	// REPOSITORY_UNKNOWN - repository not known to registry
	errRepositoryUnknown = &RegistryError{
		Code:   "REPOSITORY_UNKNOWN",
		Status: http.StatusNotFound,
	}
	// OWNER_INVALID - invalid owner name
	errOwnerInvalid = &RegistryError{
		Code:   "OWNER_INVALID",
		Status: http.StatusBadRequest,
	}
	// OWNER_UNKNOWN - owner not known to registry
	errOwnerUnknown = &RegistryError{
		Code:   "OWNER_UNKNOWN",
		Status: http.StatusNotFound,
	}
	// SIZE_INVALID - provided length did not match content length
	errSizeInvalid = &RegistryError{
		Code:   "SIZE_INVALID",
		Status: http.StatusBadRequest,
	}
	// UNAUTHORIZED - authentication required
	errUnauthorized = &RegistryError{
		Code:   "UNAUTHORIZED",
		Status: http.StatusUnauthorized,
	}
	// DENIED - requested access to the resource is denied
	errDenied = &RegistryError{
		Code:   "DENIED",
		Status: http.StatusForbidden,
	}
	// UNSUPPORTED - the operation is unsupported
	errUnsupported = &RegistryError{
		Code:   "UNSUPPORTED",
		Status: http.StatusMethodNotAllowed,
	}
	// TOOMANYREQUESTS - too many requests
	errTooManyRequests = &RegistryError{
		Code:   "TOOMANYREQUESTS",
		Status: http.StatusTooManyRequests,
	}
	// RANGE_INVALID - requested range not satisfiable
	errRangeInvalid = &RegistryError{
		Code:   "BLOB_UPLOAD_INVALID",
		Status: http.StatusRequestedRangeNotSatisfiable,
	}
)

func (e *RegistryError) WithStatus(status int) *RegistryError {
	return &RegistryError{
		Code:   e.Code,
		Status: status,
	}
}

func sendError(w http.ResponseWriter, err *RegistryError, message string) {
	errorResponse := models.ErrorResponse{
		Errors: []models.Error{
			{
				Code:    err.Code,
				Message: message,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Status)
	json.NewEncoder(w).Encode(errorResponse)
}
