package model

const (
	// UserMaxStoreSizeNormal 普通用户最大存储空间
	UserMaxStoreSizeNormal = 1073741824
	// UserMaxStoreSizeVip vip用户最大存储空间
	UserMaxStoreSizeVip = UserMaxStoreSizeNormal * 5
	// InitializeStoreSize 初始化大小
	InitializeStoreSize = 0
)

// User ...
type User struct {
	UUID string `json:"uuid"`
	Date string `json:"date"`
	Name string `json:"name"`
	Pwd  string `json:"pwd"`
}

// Record ...
type Record struct {
	ID       string `json:"id"`
	UUID     string `json:"uuid"`
	Title    string `json:"title"`
	Text     string `json:"text"`
	TagName  string `json:"tagname"`
	FilePath string `json:"filepath"`
	FileName string `json:"filename"`
	FileType string `json:"filetype"`
	Date     string `json:"date"`
}

// Tag ...
type Tag struct {
	ID      int    `json:"id"`
	TagName string `json:"tagname"`
	Sum     int    `json:"sum"`
}
