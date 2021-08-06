package mailbait

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rock-go/rock/logger"
	"github.com/rock-go/rock/lua"
	"io"
	"os"
	"reflect"
	"strings"
	"time"
)

const (
	IDLE = iota
	BUSY
)

type MailBait struct {
	lua.Super
	cfg *config

	receiver []string

	status int
	stat   statistic

	ctx    context.Context
	cancel context.CancelFunc
}

type statistic struct {
	Total   int `json:"total"`
	Success int `json:"success"`
	Fail    int `json:"fail"`
}

var baitTypeOF = reflect.TypeOf((*MailBait)(nil)).String()

func newMailBait(cfg *config) *MailBait {
	bait := &MailBait{cfg: cfg}
	bait.S = lua.INIT
	bait.T = baitTypeOF
	return bait
}

func (m *MailBait) Start() error {
	m.ctx, m.cancel = context.WithCancel(context.Background())
	m.S = lua.RUNNING
	m.U = time.Now()
	m.status = IDLE
	logger.Infof("%s mail bait start successfully", m.cfg.name)
	return nil
}

func (m *MailBait) Close() error {
	if m.cancel != nil {
		m.cancel()
	}

	logger.Infof("%s mail bait close successfully", m.cfg.name)
	m.S = lua.CLOSE
	return nil
}

func (m *MailBait) Type() string {
	return baitTypeOF
}

func (m *MailBait) Name() string {
	return m.cfg.name
}

// SendBait 发送蜜饵邮件
func (m *MailBait) SendBait() {
	receivers := verifyReceiver(m.cfg.to)
	attaches, err := CheckAttachments(m.cfg.attachments)
	if err != nil {
		logger.Errorf("send mail bait error: %v", err)
		return
	}

	m.stat = statistic{Total: len(receivers)}
	tk := time.Now().Unix()
	baitAttach := make([]string, 0)
	m.status = BUSY
	for _, receiver := range receivers {
		// 打印发送进度
		now := time.Now().Unix()
		if now-tk >= int64(m.cfg.heartbeat) {
			logger.Infof("%s mail bait sending, total %d, send %d, success %d, fail %d",
				m.cfg.name, m.stat.Total, m.stat.Success+m.stat.Fail, m.stat.Success, m.stat.Fail)
			tk = now
		}

		select {
		case <-m.ctx.Done():
			m.status = IDLE
			logger.Infof("%s mail bait exit, total %d, send %d, success %d, fail %d",
				m.cfg.name, m.stat.Total, m.stat.Success+m.stat.Fail, m.stat.Success, m.stat.Fail)
			return
		default:
			baitAttach = m.newBaitAttach(attaches, receiver)
			obj := m.formatContent(receiver, baitAttach)
			err := m.cfg.email.SendMail(obj)
			if err != nil {
				m.stat.Fail++
				logger.Errorf("%s mail bait send error: %v", m.cfg.name, err)
				goto REMOVE
			}
			m.stat.Success++
		REMOVE:
			time.Sleep(1 * time.Second)
			removeAttach(baitAttach)
		}
	}
	// 发送完成
	m.status = IDLE
	logger.Infof("%s mail bait complete, total %d, send %d, success %d, fail %d",
		m.cfg.name, m.stat.Total, m.stat.Success+m.stat.Fail, m.stat.Success, m.stat.Fail)
}

// 校验邮件接收者
func verifyReceiver(receiver string) []string {
	// 文件
	f, err := os.OpenFile(receiver, os.O_RDONLY, 0666)
	if err == nil {
		return readLines(f)
	}

	// 非文件
	receiver = strings.Trim(receiver, " ")
	return strings.Split(receiver, ",")
}

// read lines
func readLines(f *os.File) []string {
	if f == nil {
		return nil
	}

	lines := make([]string, 0)
	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		line = strings.Trim(line, "\r\n")
		line = strings.Trim(line, "\n")
		lines = append(lines, line)
		if err == io.EOF {
			return lines
		}
	}
}

func CheckAttachments(val string) ([]string, error) {
	val = strings.Trim(val, " ")
	if val == "" {
		return nil, nil
	}

	attaches := strings.Split(val, ",")
	for _, attach := range attaches {
		info, err := os.Stat(attach)
		if err != nil || info.IsDir() {
			logger.Errorf("attachments list error, check the stat of file %s", attach)
			return nil, errors.New("check attaches error")
		}
	}

	return attaches, nil
}

func (m *MailBait) newBaitAttach(old []string, mail string) []string {
	url := fmt.Sprintf("%s%s", m.cfg.attachLink, mail)
	var newFile = make([]string, 0)
	for _, o := range old {
		n, err := GenerateAttach(o, url)
		if err != nil {
			continue
		}
		newFile = append(newFile, n)
	}
	return newFile
}

func removeAttach(attaches []string) {
	for _, a := range attaches {
		e := os.Remove(a)
		if e != nil {
			logger.Errorf("bait remove attach %s error: %v", a, e)
		}
	}
}

// 根据不同的接收人生成对应的发送对象
func (m *MailBait) formatContent(receiver string, attaches []string) []byte {
	content := fmt.Sprintf(`<body background="%s%s">%s</body>`, m.cfg.mailLink, receiver, m.cfg.content)
	content = strings.ReplaceAll(content, "{CLICK_LINK}", m.cfg.clickLink+receiver)
	obj := Obj{
		To:          receiver,
		Subject:     m.cfg.subject,
		Typ:         "html",
		Content:     []byte(content),
		Attachments: attaches,
	}

	data, err := json.Marshal(obj)
	if err != nil {
		return nil
	}

	return data
}
