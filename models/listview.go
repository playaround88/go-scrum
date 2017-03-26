package models

import (
	"fmt"
	"github.com/pkg/errors"
	"strconv"
)

//Listview 列表
//redis 存储结构
//listview:[Id] - listview
//listview:seq - string Id序列值
//board-listview:[board_id] -list 看板任务列表（有序）

type Listview struct {
	Id int64
	BoardId int64
	Tasks []Task
}

const (
	LV_PREFIX = "listview:"
	LV_SEQ = LV_PREFIX+"seq"
	BOARD_LV = "board_listview:"
)

func (lv *Listview) SaveOrUpdate() error {
	if lv.Id <= 0 {
		ic := client.Incr(LV_SEQ)
		if ic.Err() != nil {
			return ic.Err()
		}
		lv.Id = ic.Val()
	}

	pipeline := client.Pipeline()

	pipeline.HSet(LV_PREFIX+fmt.Sprintf("%d",lv.Id),"BoardId", fmt.Sprintf("%d", lv.BoardId))
	//关联关系
	pipeline.LPush(BOARD_LV+fmt.Sprintf("%d", lv.BoardId),fmt.Sprintf("%d",lv.Id))

	_,err := pipeline.Exec()

	return err
}

func (lv *Listview) Del() error {
	if lv.Id <= 0 || lv.BoardId <= 0 {
		return errors.New("数据异常，未进行操作，check Id or BoardId field")
	}

	pipeline := client.TxPipeline()

	pipeline.Del(LV_PREFIX+fmt.Sprintf("%d",lv.Id))

	pipeline.LRem(BOARD_LV+fmt.Sprintf("%d", lv.BoardId),1,fmt.Sprintf("%d",lv.Id))

	_,err := pipeline.Exec()

	return err
}

func (lv *Listview) Load() error{
	if lv.Id <= 0 {
		return errors.New("数据异常，未进行操作，check Id field")
	}

	ssc := client.HGetAll(LV_PREFIX+fmt.Sprintf("%d",lv.Id))
	if ssc.Err() != nil {
		 return ssc.Err()
	}
	m := ssc.Val()

	return mapLv(lv, m)
}

func mapLv(lv *Listview, m map[string]string) error {
	bid,err := strconv.ParseInt(m["BoardId"],10, 64)
	if err != nil {
		return err
	}
	lv.BoardId = bid

	return nil
}
