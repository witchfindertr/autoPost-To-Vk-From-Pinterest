package models

import "fmt"

// Response is the base struct for all responses
// that come back from the Pinterest API.
type Empty struct{}

//type Response struct {
//	Data    interface{} `json:"data"`
//	Message string      `json:"message"`
//	Type    string      `json:"type"`
//	Page    Page        `json:"page"`
//}
////type MyDat struct {
//	Data []Pin `json:"data"`
//	Page Page  `json:"page"`
//}
//type Page struct {
//	Cursor string `json:"cursor"`
//	Next   string `json:"next"`
//}

type UploadPhotoWallResponse struct {
	Server int    `json:"server"`
	Photo  string `json:"photo"`
	Hash   string `json:"hash"`
}
type UploadPhotoResponse struct {
	Server    int    `json:"server"`
	AlbumId   int    `json:"aid"`
	Hash      string `json:"hash"`
	PhotoList string `json:"photos_list"`
}
type Error struct {
	Code          int    `json:"error_code"`
	Message       string `json:"error_msg"`
	RequestParams []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"request_params"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("code %d: %s", e.Code, e.Message)
}
