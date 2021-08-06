package mailbait

import (
	"encoding/json"
	"github.com/rock-go/rock/lua"
	"github.com/rock-go/rock/xcall"
)

func (m *MailBait) Index(L *lua.LState, key string) lua.LValue {
	if key == "start" {
		return lua.NewFunction(m.start)
	}
	if key == "close" {
		return lua.NewFunction(m.close)
	}
	if key == "new_attach" {
		return lua.NewFunction(newBaitAttach)
	}
	if key == "send" {
		return lua.NewFunction(m.sendBait)
	}
	if key == "progress" {
		return lua.NewFunction(m.checkProgress)
	}

	return lua.LNil
}

func (m *MailBait) NewIndex(L *lua.LState, key string, val lua.LValue) {
	switch key {
	case "name":
		m.cfg.name = lua.CheckString(L, val)
	case "email":
		m.cfg.email = CheckEmail(L, val)
	case "to":
		m.cfg.to = lua.CheckString(L, val)
	case "subject":
		m.cfg.subject = lua.CheckString(L, val)
	case "mail_link":
		m.cfg.mailLink = lua.CheckString(L, val)
	case "click_link":
		m.cfg.clickLink = lua.CheckString(L, val)
	case "attach_link":
		m.cfg.attachLink = lua.CheckString(L, val)
	case "content":
		m.cfg.content = lua.CheckString(L, val)
	case "attachments":
		m.cfg.attachments = lua.CheckString(L, val)
	case "heartbeat":
		m.cfg.heartbeat = lua.CheckInt(L, val)
	}
}

func (m *MailBait) start(L *lua.LState) int {
	_ = m.Start()
	return 0
}

func (m *MailBait) close(L *lua.LState) int {
	_ = m.Close()
	return 0
}

func newBaitAttach(L *lua.LState) int {
	n := L.GetTop()
	if n < 2 {
		L.RaiseError("need 2 args, got %d", n)
		return 0
	}

	old := L.CheckString(1)
	url := L.CheckString(2)
	newPath, err := GenerateAttach(old, url)
	if err != nil {
		L.RaiseError("generate new bait attach error: %v", err)
		return 0
	}

	L.Push(lua.LString(newPath))
	return 1
}

func (m *MailBait) sendBait(L *lua.LState) int {
	if m.status == BUSY {
		L.RaiseError("a task is in progress, wait or close it")
		return 0
	}

	go m.SendBait()
	return 0
}

func (m *MailBait) checkProgress(L *lua.LState) int {
	stat, err := json.Marshal(m.stat)
	if err != nil {
		L.RaiseError("get mail bait send progress error: %v", err)
		return 0
	}

	L.Push(lua.LString(lua.B2S(stat)))
	return 1
}

func newLuaMailBait(L *lua.LState) int {
	cfg := newConfig(L)
	proc := L.NewProc(cfg.name, baitTypeOF)
	if proc.IsNil() {
		proc.Set(newMailBait(cfg))
		goto done
	}
	proc.Value.(*MailBait).cfg = cfg

done:
	L.Push(proc)
	return 1
}

func LuaInjectApi(env xcall.Env) {
	env.Set("mail_bait", lua.NewFunction(newLuaMailBait))
}
