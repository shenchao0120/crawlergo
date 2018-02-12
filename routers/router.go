package routers

import (
	"chaoshen.com/crawlergo/controllers"
	"github.com/astaxie/beego"
)

func init() {
    beego.Router("/", &controllers.MainController{})
}
