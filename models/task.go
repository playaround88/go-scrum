package models

import "time"

//Task 任务条目
//redis存储结构
//task:[id] - hash
//task:seq - string Id序列值
type Task struct {
	Id int64
	Name string
	Desc string
	CreateTime time.Time
	EndTime time.Time
	//附件Ids（comma seperate）
	Attachment string

}
