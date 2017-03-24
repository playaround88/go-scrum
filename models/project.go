package models

import "time"

//Project 项目
//redis存储结构
//project:[Id] - hash 对象存储
//project:seq - string Id序列值
//user_projects:[user_id] - list 用户项目列表
//user_a_projects:[user_id] -list 用户已归档项目列表
type Project struct {
	Id int64
	Name string
	Desc string
	CreateTime time.Time

	//状态，是否归档
	State string
}
