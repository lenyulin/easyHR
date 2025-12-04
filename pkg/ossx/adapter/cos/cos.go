package cos

import (
	"context"
	"crypto/sha256"
	oss "easyHR/pkg/ossx"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/tencentyun/cos-go-sdk-v5"
)

var (
	ErrUploadToOSSFailed = errors.New("upload to cos failed")
)

type OssHandler struct {
	oss *cos.Client
}

func NewCOSHandler(oss *cos.Client) oss.OSSHandler {
	return &OssHandler{oss: oss}
}

const salt = "xiFgge1O4DqWs5og"

func (hdl *OssHandler) filenameToUniqueWithSalt(filename string) string {
	// 结合文件名和盐值生成哈希，降低碰撞概率
	data := []byte(filename + "|" + salt)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (hdl *OssHandler) Upload(ctx context.Context, fileDir string) (string, string, error) {
	splitDir := strings.Split(fileDir, "\\")
	rawName := splitDir[len(splitDir)-1]
	filename := hdl.filenameToUniqueWithSalt(splitDir[len(splitDir)-1])
	_, _, err := hdl.oss.Object.Upload(context.Background(), filename+".mp4", fileDir, nil)
	if err != nil {
		return "", "", err
	}
	return filename, rawName, nil
}

func (hdl *OssHandler) Find(ctx context.Context, uid int64) error {
	return nil
}
func (hdl *OssHandler) Delete(ctx context.Context, uid int64) error {
	return nil
}
