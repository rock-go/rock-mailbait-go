package mailbait

import (
	"github.com/rock-go/rock/lua"
)

type config struct {
	name        string
	email       Mail   // email 接口
	to          string // 接收者的邮件列表
	subject     string // 主题
	mailLink    string // 邮件隐藏链接
	clickLink   string // 正文部分点击链接
	attachLink  string // 附件隐藏链接
	content     string // 内容
	attachments string // 附件列表
	heartbeat   int
}

type Mail interface {
	lua.LFace
	SendMail(interface{}) error
}

type Obj struct {
	To          string   `json:"to"`
	Subject     string   `json:"subject"`
	Typ         string   `json:"typ"` // text, html
	Content     []byte   `json:"content"`
	Attachments []string `json:"attachments"` // 附件链接
}

func newConfig(L *lua.LState) *config {
	tb := L.CheckTable(1)
	cfg := &config{}
	cfg.name = tb.CheckString("name", "rock-go-mailbait")
	cfg.heartbeat = tb.CheckInt("heartbeat", 10)
	cfg.email = CheckEmail(L, tb.RawGetString("email"))
	cfg.to = tb.CheckString("to", "")
	cfg.subject = tb.CheckString("subject", "")
	cfg.mailLink = tb.CheckString("mail_link", "")
	cfg.clickLink = tb.CheckString("click_link", "")
	cfg.attachLink = tb.CheckString("attach_link", "")
	cfg.content = tb.CheckString("content", "")
	cfg.attachments = tb.CheckString("attachment", "")

	return cfg
}

func CheckEmail(L *lua.LState, val lua.LValue) Mail {
	ud := lua.CheckLightUserData(L, val)
	mail, ok := ud.Value.(Mail)
	if !ok {
		L.RaiseError("got email userdata error")
		return nil
	}

	return mail
}
