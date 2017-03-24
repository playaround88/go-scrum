package routers

import (
	"github.com/astaxie/beego"
	"go-scrum/controllers"
)

func init() {
	beego.Router("/", &controllers.MainController{})

	beego.AutoRouter(&controllers.UserController{})
}
