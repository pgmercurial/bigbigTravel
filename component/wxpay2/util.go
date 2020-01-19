package wxpay2

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/pem"
	"encoding/xml"
	"fmt"
	"golang.org/x/crypto/pkcs12"
	"log"
	"strconv"
	"strings"
	"time"
)

func XmlToMap(xmlStr string) Params {

	params := make(Params)
	decoder := xml.NewDecoder(strings.NewReader(xmlStr))

	var (
		key   string
		value string
	)

	for t, err := decoder.Token(); err == nil; t, err = decoder.Token() {
		switch token := t.(type) {
		case xml.StartElement: // 开始标签
			key = token.Name.Local
		case xml.CharData: // 标签内容
			content := string([]byte(token))
			value = content
		}
		if key != "xml" {
			if value != "\n" {
				params.SetString(key, value)
			}
		}
	}

	return params
}

func MapToXml(params Params) string {
	var buf bytes.Buffer
	buf.WriteString(`<xml>`)
	for k, v := range params {
		buf.WriteString(`<`)
		buf.WriteString(k)
		buf.WriteString(`><![CDATA[`)
		buf.WriteString(v)
		buf.WriteString(`]]></`)
		buf.WriteString(k)
		buf.WriteString(`>`)
	}
	buf.WriteString(`</xml>`)

	return buf.String()
}

// 用时间戳生成随机字符串
func nonceStr() string {
	return strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
}

func NonceStr() string {
	return nonceStr()
}

// 将Pkcs12转成Pem
func pkcs12ToPem(p12 []byte, password string) tls.Certificate {

	blocks, err := pkcs12.ToPEM(p12, password)

	// 从恐慌恢复
	defer func() {
		if x := recover(); x != nil {
			log.Print(x)
		}
	}()

	if err != nil {
		panic(err)
	}

	var pemData []byte
	for _, b := range blocks {
		pemData = append(pemData, pem.EncodeToMemory(b)...)
	}

	cert, err := tls.X509KeyPair(pemData, pemData)
	if err != nil {
		panic(err)
	}
	return cert
}

//前端再次签名所需要的数据
func Resign(mp map[string]string, apiKey string) string {
	stringA := fmt.Sprintf("appId=%s&nonceStr=%s&package=%s&signType=%s&timeStamp=%s",
		mp["appId"], mp["nonceStr"], mp["package"], mp["signType"], mp["timeStamp"])
	stringSignTemp := stringA + fmt.Sprintf("&key=%s", apiKey)

	md5Bytes := md5.Sum([]byte(stringSignTemp))
	hexStr := hex.EncodeToString(md5Bytes[:])
	return strings.ToUpper(hexStr)
}