package models

import "time"

//Board 看板
//redis存储结构
//board:[Id] - hash 对象存储
//board:seq - string Id序列值
//project_board:[project_id] - list 项目看板列表
//project_a_board:[project_id] - list 项目已归档看板列表
type Board struct {
	Id int64
	Name string
	CreateTime time.Time

	//XXX 是否已归档
	State string
}
