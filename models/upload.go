package models

type UploadRequest struct {
	FileData string `json:"file_data"`
}

type Users struct {
	UsersId []string `json:"user_ids"`
}
