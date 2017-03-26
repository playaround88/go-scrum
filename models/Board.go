package models

import (
	"time"
	"fmt"
	"github.com/pkg/errors"
	"strconv"
)

//Board 看板
//redis存储结构
//board:[Id] - hash 对象存储
//board:seq - string Id序列值
//project_board:[project_id] - list 项目看板列表
//project_a_board:[project_id] - list 项目已归档看板列表
type Board struct {
	Id int64
	Name string
	ProjectId int64
	Creator int64
	CreateTime time.Time

	//XXX 是否已归档
	State string
}

const (
	BOARD_PREFIX = "board:"
	BOARD_SEQ = BOARD_PREFIX + "SEQ"
	PROJECT_BOARD = "project_board:"
	PROJECT_A_BOARD = "project_a_board"
)

func (b *Board) SaveOrUpdate() error{
	if b.Id <= 0 {
		ic := client.Incr(BOARD_SEQ)
		b.Id = ic.Val()
	}

	pipeline := client.TxPipeline()
	//对象存储
	pipeline.HSet(BOARD_PREFIX+fmt.Sprintf("%d", b.Id),"Name",b.Name)
	pipeline.HSet(BOARD_PREFIX+fmt.Sprintf("%d", b.Id),"Creator",fmt.Sprintf("%d",b.Creator))
	pipeline.HSet(BOARD_PREFIX+fmt.Sprintf("%d", b.Id),"CreateTime",fmt.Sprintf("%d",b.CreateTime.Unix()))
	pipeline.HSet(BOARD_PREFIX+fmt.Sprintf("%d", b.Id),"State",b.State)

	//关联关系
	if b.State == "archived" {
		pipeline.LPush(PROJECT_A_BOARD, b.Id)
	}else {
		pipeline.LPush(PROJECT_BOARD,b.Id)
	}
	//执行
	_,err := pipeline.Exec()

	return err
}

func (b *Board) Del() error{
	if b.Id <= 0 || b.ProjectId <= 0 {
		return errors.New("看板编码异常，未进行操作，Id or ProjectId")
	}

	pipeline := client.TxPipeline()

	pipeline.Del(BOARD_PREFIX+fmt.Sprintf("%d",b.Id))
	//删除关联
	pipeline.LRem(PROJECT_BOARD+fmt.Sprintf("%d",b.ProjectId), 1, fmt.Sprintf("%d",b.Id))
	pipeline.LRem(PROJECT_A_BOARD+fmt.Sprintf("%d",b.ProjectId), 1, fmt.Sprintf("%d",b.Id))

	_,err := pipeline.Exec()

	return err
}

func (b *Board) Load() error{
	if b.Id <= 0 {
		return errors.New("看板编码异常，未进行操作")
	}

	ssc := client.HGetAll(BOARD_PREFIX+fmt.Sprintf("%d",b.Id))
	if ssc.Err() != nil {
		return ssc.Err()
	}

	m := ssc.Val()

	return mapBoard(b,m)
}

func mapBoard(b *Board, m map[string]string) error{
	b.Name = m["Name"]

	pid,err := strconv.ParseInt(m["ProjectId"],10, 64)
	if err != nil {
		return err
	}
	b.ProjectId = pid

	c,err := strconv.ParseInt(m["Creator"], 10, 64)
	if err != nil {
		return err
	}
	b.Creator = c

	secs,err := strconv.ParseInt(m["CreateTime"],10,64)
	if err != nil {
		return nil
	}
	b.CreateTime = time.Unix(secs,0)

	b.State = m["State"]

	return nil
}



