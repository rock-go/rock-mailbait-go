package mailbait

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Zip 压缩文件。offset为0表示压缩时将src目录放入压缩包，为1则表示从src目录内开始压缩
// 如：src/1,src/2,src/3，offset为0，则压缩包根目录为src；否则，压缩包为1，2，3
func Zip(dst, src string, offset int) error {
	fw, err := os.Create(dst)
	defer fw.Close()
	if err != nil {
		return err
	}

	zw := zip.NewWriter(fw)
	defer zw.Close()

	err = addFile(src, zw, offset)
	return err
}

func addFile(src string, zw *zip.Writer, offset int) error {
	src = strings.ReplaceAll(src, "\\", "/")
	return filepath.Walk(src, func(path string, fi os.FileInfo, errBack error) (err error) {
		if errBack != nil {
			return errBack
		}

		// 通过文件信息，创建 zip 的文件信息
		fh, err := zip.FileInfoHeader(fi)
		if err != nil {
			return
		}

		// 替换文件信息中的文件名
		fh.Name = strings.ReplaceAll(path, string(filepath.Separator), "/")
		fh.Name = strings.TrimPrefix(fh.Name, "./")
		fh.Name = strings.TrimPrefix(fh.Name, "/")
		if offset == 1 {
			fh.Name = strings.TrimPrefix(fh.Name, src)
		}

		// 这步开始没有加，会发现解压的时候说它不是个目录
		if fi.IsDir() {
			fh.Name += "/"
		}

		// 检测，如果不是标准文件就只写入头信息，不写入文件数据到 w
		// 如目录，也没有数据需要写
		if !fh.Mode().IsRegular() {
			return nil
		}
		// 写入文件信息，并返回一个 Write 结构
		w, err := zw.CreateHeader(fh)
		if err != nil {
			return
		}
		// 打开要压缩的文件
		fr, err := os.Open(path)
		defer fr.Close()
		if err != nil {
			return
		}

		// 将打开的文件 Copy 到 w
		_, err = io.Copy(w, fr)
		if err != nil {
			return
		}
		return nil
	})
}

// Unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
func Unzip(src string, dest string) ([]string, error) {
	var filenames []string
	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()
	for _, f := range r.File {
		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)
		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}
		filenames = append(filenames, fpath)
		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}
		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}
		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}
		_, err = io.Copy(outFile, rc)
		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()
		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}
