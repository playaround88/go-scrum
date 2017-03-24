package models

//Lane 泳道
//redis存储结构
//lane:seq - string Id序列值
//board-lane:[board_id] - list 看板泳道列表（有序）
//lane:[id] -list 任务集合
type Lane struct {
	Id int64
	Items []Item
}
