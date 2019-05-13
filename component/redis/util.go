package redis

import "fmt"

func GenKey(appname string, isCache bool, typ , module, model string, identify ...interface{}) string {
	typeDesc := "store"
	if isCache {
		typeDesc = "cache"
	}
	key :=  appname+":"+typeDesc+":"+typ+":"+module+":"+model
	if len(identify) != 0 {
		for _,v := range identify {
			key += ":"+fmt.Sprint(v)
		}
	}
	return key
}
