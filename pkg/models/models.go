package models

type ErrorResponse struct {
	Errors []Error `json:"errors"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
	Detail  any    `json:"detail,omitempty"`
}

type UploadInfo struct {
	Name      string
	ID        string
	Path      string
	Size      int64
	Offset    int64
	Completed bool
}

type TagsResponse struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}
