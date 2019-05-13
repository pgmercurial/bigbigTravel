package helper

import (
	"net"
	"net/http"
	"io/ioutil"
	"strings"
)

func GetLocalIP() string {
	addrs,_ := net.InterfaceAddrs()
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func GetExternalIP() string{
	resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil{
		return ""
	}
	return string(bytes)
}

func ParseHost(host string) (string, string) {
	strs := strings.Split(host, ":")
	if  len(strs) == 0{
		return "",""
	}else if len(strs) == 1 {
		return strs[0], ""
	}else {
		return strs[0], strs[1]
	}
}


