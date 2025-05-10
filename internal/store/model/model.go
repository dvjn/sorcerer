package model

type UploadInfo struct {
	Name      string
	ID        string
	Path      string
	Size      int64
	Offset    int64
	Completed bool
}
