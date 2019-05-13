package helper

import  (
	"time"
	"os"
	"sync"
	"fmt"
)

var autoincrement int64
var mutex sync.Mutex

func GenerateUUID() string {
	mutex.Lock()
	t := time.Now().UnixNano() <<8 + int64(os.Getpid()%16 << 4) + autoincrement%16
	autoincrement += 1
	mutex.Unlock()
	return fmt.Sprintf("%x", t)
}
