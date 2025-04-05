package router

import "net/http"

type Service interface {
	ApiVersionCheck(w http.ResponseWriter, r *http.Request)

	CheckBlobExists(w http.ResponseWriter, r *http.Request)
	GetBlob(w http.ResponseWriter, r *http.Request)
	DeleteBlob(w http.ResponseWriter, r *http.Request)

	InitiateUpload(w http.ResponseWriter, r *http.Request)
	UploadBlobChunk(w http.ResponseWriter, r *http.Request)
	CompleteUpload(w http.ResponseWriter, r *http.Request)
	GetBlobUploadStatus(w http.ResponseWriter, r *http.Request)

	CheckManifestExists(w http.ResponseWriter, r *http.Request)
	GetManifest(w http.ResponseWriter, r *http.Request)
	PutManifest(w http.ResponseWriter, r *http.Request)
	DeleteManifest(w http.ResponseWriter, r *http.Request)

	ListTags(w http.ResponseWriter, r *http.Request)

	ListReferrers(w http.ResponseWriter, r *http.Request)
}
