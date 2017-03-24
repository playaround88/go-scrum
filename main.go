package main

import (
	_ "go-scrum/routers"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"go-scrum/models"
)

func main() {
	beego.InsertFilter("/*",beego.BeforeRouter,FilterUser)

	beego.Run()
}

var FilterUser = func(ctx *context.Context) {

	_, ok := ctx.Input.Session("USER").(models.User)

	if !ok && ctx.Request.RequestURI != "/user/login" && ctx.Request.RequestURI != "/user/signup" {
		ctx.Redirect(302, "/user/login")
	}
}

