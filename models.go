package main

import "time"

type ApiResponse struct {
	Ok bool`json:"Ok"`
	Token string `json:"Token"`
	UserId string `json:"UserId"`
	Email string`json:"Email"`
	UserName string`json:"UserName"`
	Msg string `json:"Msg"`
}



type Note struct {
	NoteId     string
	NotebookId string
	UserId     string
	Title      string
	Tags       []string
	Content    string
	IsMarkdown bool
	IsBlog     bool
	IsTrash bool
	Files []NoteFile // 图片, 附件
	CreatedTime time.Time
	UpdatedTime time.Time
	PublicTime time.Time

	// 更新序号
	Usn int
}

type Notebook struct {
	NotebookId	string
	UserId	string
	ParentNotebookId string // 上级
	Seq              int // 排序
	Title            string
	IsBlog           bool
	IsDeleted	 bool
	CreatedTime      time.Time
	UpdatedTime      time.Time

	// 更新序号
	Usn int  // UpdateSequenceNum

}

type NoteFile struct {
	FileId string // 服务器端Id
	LocalFileId string // 客户端Id
	Type string // images/png, doc, xls, 根据fileName确定
	Title string
	HasBody bool // 传过来的值是否要更新内容, 如果有true, 则必须传文件
	IsAttach bool // 是否是附件, 不是附件就是图片
}




