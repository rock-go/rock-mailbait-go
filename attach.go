package mailbait

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GenerateAttach 根据文件名和链接生成docx附件蜜饵，返回其路径
func GenerateAttach(filename string, url string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}

	fs, err := f.Stat()
	if err != nil {
		return "", err
	}

	if fs.IsDir() {
		return "", errors.New("the filename gave is dir")
	}

	baseDir := filepath.Dir(filename)
	tmpDir := baseDir + "/tempZip/"
	err = os.Mkdir(tmpDir, 0666)
	if err != nil && !os.IsExist(err) {
		return "", err
	}

	// 解压
	_, err = Unzip(filename, tmpDir)
	if err != nil {
		return "", err
	}
	// 修改
	err = ModifyDocx(tmpDir, url)
	if err != nil {
		return "", err
	}
	// 压缩
	now := time.Now().Format("20060102150405-")
	dst := baseDir + "/" + now + fs.Name()
	err = Zip(dst, tmpDir, 1)
	_ = os.RemoveAll(tmpDir)
	return dst, nil
}

func ModifyDocx(fileDir, url string) error {
	contentSettingsXmlRels := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId333" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/attachedTemplate" Target="target_url" TargetMode="External"/>
</Relationships>`
	contentSettingsXmlRels = strings.ReplaceAll(contentSettingsXmlRels, "target_url", url)
	filePathRels := fileDir + "/word/_rels/settings.xml.rels"
	f, err := os.OpenFile(filePathRels, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}

	_, err = f.WriteString(contentSettingsXmlRels)
	if err != nil {
		return err
	}
	f.Close()

	filePathXml := fileDir + "/word/settings.xml"
	f, err = os.OpenFile(filePathXml, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadFile(filePathXml)
	if err != nil {
		return err
	}

	dataStr := string(data)
	content := strings.Split(dataStr, "<w:bordersDoNotSurroundFooter/>")
	if len(content) < 2 {
		return errors.New(filePathXml + " content parse error")
	}
	dataStr = content[0] + (`<w:bordersDoNotSurroundFooter/><w:attachedTemplate r:id="rId333"/>`) + content[1]

	_, err = f.WriteString(dataStr)
	f.Close()
	return err
}
