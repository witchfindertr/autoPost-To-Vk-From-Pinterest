package models

type IniMySQL struct {
	Database string
	User     string
	Pass     string
	Host     string
	Port     string
}
type Pages struct {
	Exists bool
	Cursor string
	Writes int8
}
type DbVideoInfo struct {
	Id         int
	Originlink string
	Note       string
}