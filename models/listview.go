package models

//Listview 列表
//redis 存储结构
//listview:seq - string Id序列值
//board-listview:[board_id] -list 看板任务列表（有序）
//listview:[Id] - list 任务集合
type Listview struct {
	Id int64
	Items []Item
}
