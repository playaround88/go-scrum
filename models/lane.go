package models

import (
	"fmt"
	"strconv"
	"errors"
)

//Lane 泳道
//暂时不实现泳道的功能，简化前端开发
//redis存储结构
//lane:[id] -hash 对象存储
//lane:seq - string Id序列值
//board-lane:[board_id] - set 看板泳道列表（有序）
type Lane struct {
	Id int64
	BoardId int64
	Tasks []Task
}


const (
	LANE_PREFIX = "lane:"
	LANE_SEQ = LANE_PREFIX+"seq"
	BOARD_LANE = "board_lane:"
)

func (l *Lane) SaveOrUpdate() error {
	if l.Id <= 0 {
		ic := client.Incr(LANE_SEQ)
		if ic.Err() != nil {
			return ic.Err()
		}
		l.Id = ic.Val()
	}

	pipeline := client.Pipeline()

	pipeline.HSet(LANE_PREFIX+fmt.Sprintf("%d",l.Id),"BoardId", fmt.Sprintf("%d", l.BoardId))
	//关联关系
	pipeline.SAdd(BOARD_LANE+fmt.Sprintf("%d", l.BoardId),fmt.Sprintf("%d",l.Id))

	_,err := pipeline.Exec()

	return err
}

func (l *Lane) Del() error {
	if l.Id <= 0 || l.BoardId <= 0 {
		return errors.New("数据异常，未进行操作，check Id or BoardId field")
	}

	pipeline := client.TxPipeline()

	pipeline.Del(LANE_PREFIX+fmt.Sprintf("%d",l.Id))

	pipeline.SRem(BOARD_LANE+fmt.Sprintf("%d", l.BoardId),fmt.Sprintf("%d",l.Id))

	_,err := pipeline.Exec()

	return err
}

func (l *Lane) Load() error{
	if l.Id <= 0 {
		return errors.New("数据异常，未进行操作，check Id field")
	}

	ssc := client.HGetAll(LANE_PREFIX+fmt.Sprintf("%d",l.Id))
	if ssc.Err() != nil {
		return ssc.Err()
	}
	m := ssc.Val()

	return mapLane(l, m)
}

func mapLane(l *Lane, m map[string]string) error {
	bid,err := strconv.ParseInt(m["BoardId"],10, 64)
	if err != nil {
		return err
	}
	l.BoardId = bid

	return nil
}
