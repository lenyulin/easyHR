package tos

import (
	"context"
	"crypto/sha256"
	oss "easyHR/pkg/ossx"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

var (
	ErrUploadToOSSFailed = errors.New("upload to cos failed")
)

type TosHandler struct {
	tos *tos.ClientV2
}

// 填写 BucketName
//bucketName = "*** Provide your bucket name ***"

// 将文件上传到 example_dir 目录下的 example.txt 文件
// objectKey = "example_dir/example.txt"
const (
	Bucket = "2hbni3ppllryx90c"
)

func (hdl *TosHandler) Upload(ctx context.Context, fileDir string) (string, string, error) {
	bucket := Bucket
	splitDir := strings.Split(fileDir, "\\")
	rawName := splitDir[len(splitDir)-1]
	filename := hdl.filenameToUniqueWithSalt(splitDir[len(splitDir)-1])
	_, err := hdl.tos.PutObjectFromFile(ctx, &tos.PutObjectFromFileInput{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket: bucket,
			Key:    filename + ".mp4",
		},
		FilePath: fileDir,
	})
	if err != nil {
		return "", "", err
	}
	return filename + ".mp4", rawName, nil
}

func (hdl *TosHandler) Find(ctx context.Context, uid int64) error {
	return nil
}

func (hdl *TosHandler) Delete(ctx context.Context, uid int64) error {
	return nil
}

func NewTOSHandler(tos *tos.ClientV2) oss.OSSHandler {
	return &TosHandler{tos: tos}
}

const salt = "xiFgge1O4DqWs5og"

func (hdl *TosHandler) filenameToUniqueWithSalt(filename string) string {
	// 结合文件名和盐值生成哈希，降低碰撞概率
	data := []byte(filename + "|" + salt)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
