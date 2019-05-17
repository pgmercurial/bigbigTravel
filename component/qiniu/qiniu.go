package qiniu

import (
	"bytes"
	"context"
	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
	"github.com/qiniu/x/errors.v7"
	"io/ioutil"
	"net/http"
	"strconv"
)

type QiniuConfig struct {
	Bucket    string
	Host      string
	AccessKey string
	SecretKey string
}

var envcfg *QiniuConfig
var qncfg *storage.Config

func InitConfig(config *QiniuConfig) {
	envcfg = config
	qncfg = &storage.Config{
		Zone:          &storage.ZoneHuabei,
		UseHTTPS:      false,
		UseCdnDomains: false,
	}
}

func GetUploadToken() string{
	putPolicy := storage.PutPolicy{
		Scope: envcfg.Bucket,
	}
	mac := qbox.NewMac(envcfg.AccessKey, envcfg.SecretKey)
	upToken := putPolicy.UploadToken(mac)
	return upToken
}

func UploadFile(fileBytes []byte, key string) (*storage.PutRet, error) {
	//putPolicy := storage.PutPolicy{
	//	Scope: envcfg.Bucket,
	//}
	//mac := qbox.NewMac(envcfg.AccessKey, envcfg.SecretKey)
	//upToken := putPolicy.UploadToken(mac)
	upToken := GetUploadToken()

	formUploader := storage.NewFormUploader(qncfg)
	ret := storage.PutRet{}
	err := formUploader.Put(context.Background(), &ret, upToken, key, bytes.NewReader(fileBytes), int64(len(fileBytes)), &storage.PutExtra{})
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func DownloadFile(publicURL string) ([]byte, error) {
	////私有云download URL
	//srcUri, _ := url.Parse(publicURL)
	//urlToSign := srcUri.String()
	//deadline := time.Now().Add(time.Second * 3600).Unix()
	//if strings.Contains(urlToSign, "?") {
	//	urlToSign = fmt.Sprintf("%s&e=%d", urlToSign, deadline)
	//} else {
	//	urlToSign = fmt.Sprintf("%s?e=%d", urlToSign, deadline)
	//}
	//
	//mac := qbox.NewMac(envcfg.AccessKey, envcfg.SecretKey)
	//token := mac.Sign([]byte(urlToSign))
	//URL := fmt.Sprintf("%s&token=%s", urlToSign, token)

	//公有云download URL
	URL := publicURL

	resp, respErr := http.Get(URL)
	if respErr != nil {
		return nil, respErr
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("http resp status code : " + strconv.Itoa(resp.StatusCode))
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
