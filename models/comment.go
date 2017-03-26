package models

import (
	"time"
	"fmt"
	"github.com/pkg/errors"
	"strconv"
)

//Comment 评论
//redis存储结构
//comment:[id] - hash
//comment:seq - string Id序列
//task_comment:[task_id] - list 任务评论列表
type Comment struct {
	Id int64
	UserId int64
	UserName string
	TaskId int64
	Content string
	CreateTime time.Time
}

const (
	CM_PREFIX = "comment:"
	CM_SEQ = CM_PREFIX+"seq"
	TASK_COMMENT = "task_comment:"
)

func (cm *Comment) SaveOrUpdate() error {
	if cm.Id <= 0 {
		ic := client.Incr(CM_SEQ)
		if ic.Err() != nil {
			return ic.Err()
		}
		cm.Id = ic.Val()
	}

	pipeline := client.TxPipeline()

	pipeline.HSet(CM_PREFIX+fmt.Sprintf("%d",cm.Id),"UserId", fmt.Sprintf("%d",cm.UserId))
	pipeline.HSet(CM_PREFIX+fmt.Sprintf("%d",cm.Id),"UserName", cm.UserName)
	pipeline.HSet(CM_PREFIX+fmt.Sprintf("%d",cm.Id),"TaskId", fmt.Sprintf("%d",cm.TaskId))
	pipeline.HSet(CM_PREFIX+fmt.Sprintf("%d",cm.Id),"Content", cm.Content)
	pipeline.HSet(CM_PREFIX+fmt.Sprintf("%d",cm.Id),"CreateTime", cm.CreateTime.Unix())
	//
	pipeline.LPush(TASK_COMMENT+fmt.Sprintf("%d",cm.TaskId),fmt.Sprintf("%d",cm.Id))

	_,err := pipeline.Exec()

	return err
}

func (cm *Comment) Del() error{
	if cm.Id <= 0 || cm.TaskId <= 0 {
		return errors.New("数据异常，未进行操作，check Id or TaskId field")
	}

	pipeline := client.TxPipeline()

	pipeline.Del(CM_PREFIX+fmt.Sprintf("%d",cm.Id))
	pipeline.LRem(TASK_COMMENT+fmt.Sprintf("%d",cm.TaskId),1,fmt.Sprintf("%d",cm.Id))

	_,err := pipeline.Exec()

	return err
}

func (cm *Comment) Load() error {
	if cm.Id <= 0 || cm.TaskId <= 0 {
		return errors.New("数据异常，未进行操作，check Id field")
	}

	ssc := client.HGetAll(CM_PREFIX+fmt.Sprintf("%d",cm.Id))
	if ssc.Err() != nil {
		return ssc.Err()
	}
	m := ssc.Val()

	return mapComment(cm, m)
}

func mapComment(cm *Comment, m map[string]string) error {

	uid,err := strconv.ParseInt(m["UserId"],10, 64)
	if err != nil {
		return err
	}
	cm.UserId = uid

	cm.UserName = m["UserName"]

	tid,err := strconv.ParseInt(m["TaskId"],10,64)
	if err != nil {
		return err
	}
	cm.TaskId=tid

	cm.Content = m["Content"]

	secs,err := strconv.ParseInt(m["CreateTime"],10,64)
	if err != nil {
		return err
	}
	cm.CreateTime = time.Unix(secs,0)

	return err
}

func QueryCommentByTask(tid int64, start, end int64) ([]Comment,error) {
	var cms []Comment
	ssc := client.LRange(TASK_COMMENT+fmt.Sprintf("%d",tid),start, end)
	if ssc.Err() != nil {
		return cms,ssc.Err()
	}

	m := ssc.Val()

	for _,commentId := range m {
		cid,err := strconv.ParseInt(commentId,10,64)
		if err != nil {
			return cms, err
		}

		comment := Comment{Id:cid}
		comment.Load()

		cms = append(cms, comment)
	}

	return cms, nil
}