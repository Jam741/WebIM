package controllers

import (
	"strings"
	"github.com/astaxie/beego"
	"github.com/beego/i18n"
)

/*供用户选择技术和用户名的欢迎页面*/

var langTypes []string // Language that are supported 语言支持

func init() {
	// Initialize language type list 初始化语言列表
	langTypes = strings.Split(beego.AppConfig.String("lang_types"), "|")

	//Load locale files according to language types 根据语言类型加载语言环境文件。
	for _, lang := range langTypes {
		beego.Trace("Loading language: " + lang)
		if err := i18n.SetMessage(lang, "conf/"+"locale_"+lang+".ini"); err != nil {
			beego.Error("Fail to set message file:", err)
			return
		}
	}
}

//baseController represents base router for all other app routers. baseController 表示所有其他应用程序路由器的基路由器。
//It implemented some methods for the same implementation. 它为相同的实现实现了一些方法。
//thus, it will be embedded into other routers. 因此，它将被嵌入到其他路由器中。
type baseController struct {
	beego.Controller //Embed struct that has stub implementation of the interface. 嵌入struct，该结构有接口的存根实现。
	i18n.Locale      //For i18n usage when process data and render template. 在处理数据和呈现模板时使用i18n。
}

//用户扩展函数，会在其他函数执行之前先执行
func (this *baseController) Prepare() {
	//reset language option.
	this.Lang = "" //This field is from i18n.Locale.

	//1. Get language information form 'Accept-Language'
	al := this.Ctx.Request.Header.Get("Accept-Language")
	if len(al) > 4 {
		al = al[:5] // Only compare first 5 letters. 只比较前5个字母。
		if i18n.IsExist(al) {
			this.Lang = al
		}
	}

	//2. Default language is English.
	if len(this.Lang) == 0 {
		this.Lang = "en-US"
	}

	// Set template level language option .
	this.Data["Lang"] = this.Lang

}

// AppController handles the welcome screen that allows user to pick a technology and username. AppController处理允许用户选择技术和用户名的欢迎屏幕。
type AppController struct {
	baseController // Embed to use methods that are implemented in baseController. 嵌入在baseController中实现的方法。
}

// Get implemented Get() method for AppController
func (this *AppController) Get() {
	this.TplName = "welcome.html"
}

func (this *AppController) Join() {
	//Get from value . 获取表单信息
	uname := this.GetString("uname")
	tech := this.GetString("tech")

	switch tech {
	case "longpolling":
		this.Redirect("/lp?uname="+uname, 302)
	case "websocket":
		this.Redirect("/ws?uname="+uname, 302)
	default:
		this.Redirect("/", 302)
	}

	// Usually put return after redirect. 通常在重定向后返回。
	return
}
