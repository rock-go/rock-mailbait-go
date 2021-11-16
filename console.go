package mailbait

import "github.com/rock-go/rock/lua"

func (m *MailBait) Header(out lua.Printer) {
	out.Printf("type: %s", m.Type())
	out.Printf("uptime: %s", m.Uptime)
	out.Printf("status: %s", m.Status)
	out.Printf("version: v1.0.0")
	out.Println("")
}

func (m *MailBait) Show(out lua.Printer) {
	m.Header(out)
	out.Printf("name: %s", m.cfg.name)
	out.Printf("email: %s", m.cfg.email.Name())
	out.Printf("to: %s", m.cfg.to)
	out.Printf("subject: %s", m.cfg.subject)
	out.Printf("mail_link: %s", m.cfg.mailLink)
	out.Printf("click_link: %s", m.cfg.clickLink)
	out.Printf("attach_link: %s", m.cfg.attachLink)
	out.Printf("content: %s", m.cfg.content)
	out.Printf("attachment: %s", m.cfg.attachments)
	out.Printf("heartbeat: %d", m.cfg.heartbeat)
	out.Println("")
}

func (m *MailBait) Help(out lua.Printer) {
	m.Header(out)

	out.Printf(".start() 启动")
	out.Printf(".close() 关闭")
	out.Println("")
}
