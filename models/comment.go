package models

import "time"

//Comment 评论
//redis存储结构
//comment:[id] - hash
//comment:seq - string Id序列
//task-comment:[task_id] - list 任务评论列表
type Comment struct {
	Id int64
	UserId int64
	UserName string
	Content string
	CreateTime time.Time
}
