package model

// const (
// 	// UserMaxStoreSizeNormal 普通用户最大存储空间
// 	UserMaxStoreSizeNormal = 1073741824
// 	// UserMaxStoreSizeVip vip用户最大存储空间
// 	UserMaxStoreSizeVip = UserMaxStoreSizeNormal * 5
// 	// InitializeStoreSize 初始化大小
// 	InitializeStoreSize = 0
// )

// User ...
type User struct {
	UUID     string `json:"uuid"`
	Name     string `json:"name"`
	Pwd      string `json:"pwd"`
	Mailbox  string `json:"mailbox"`
	RegTime  string `json:"regTime"`
	LastTime string `json:"lastTime"`
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
	Status   string `json:"status"`
}

// Tag ...
type Tag struct {
	ID      int    `json:"id"`
	TagName string `json:"tagname"`
	Sum     int    `json:"sum"`
}

// ToDo ...
type ToDo struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Text        string `json:"text"`
	UserID      string `json:"userID"`
	UserMailbox string `json:"userMailbox"`
	RegTime     string `json:"regTime"`
	RemindTime  string `json:"remindTime"`
}
