package models

//User 系统用户
//redis存储结构
//user:seq - string Id序列值
//user:[Id] - Hash 对象存储
//contactor:[User_Id] - list 联系人
//登录映射
//un2id - hash 用户名到id的映射
//email2id - hash Email到id的映射
type User struct {
	Id int64
	Username string
	Password string
	Email    string

	Company string

	Contactors []User
}