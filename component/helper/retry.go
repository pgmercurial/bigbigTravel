package helper

import "time"

type Retry interface {
	Get(get func() interface{})
	Do(do func(interface{}))
}

type retry struct {
	channel  chan interface{}
	num      int
	duration time.Duration
	timeout  time.Duration
	cLen     int
}

func GetRetry(cLen int, duration time.Duration, timeout time.Duration) Retry {
	r := &retry{
		duration: duration,
		timeout:  timeout,
		cLen:     cLen,
	}
	r.Reset()
	return r
}

func (r *retry) Get(get func() interface{}) {
	RetryGet(r.channel, r.duration, r.timeout, get)
	r.num++
}

func (r *retry) Do(do func(interface{})) {
	RetryDo(r.channel, r.num, r.duration, r.timeout, do)
}

func (r *retry) Reset() {
	c := make(chan interface{}, r.cLen)
	r.channel = c
	r.num = 0
}

func Delay(fn func(), delay time.Duration) {
	go func() {
		time.Sleep(delay)
		fn()
	}()
}

func RetryGet(channel chan interface{}, duration time.Duration, timeout time.Duration, get func() interface{}) {
	var res interface{}
	stop := false
	dealFn := func() bool {
		res = get()
		if stop || res != nil {
			channel <- res
			return true
		}
		return false
	}
	if dealFn() {
		return
	}
	if duration <= 0 {
		duration = 2 * time.Second
	}
	go func() {
		if timeout > 0 {
			Delay(func() {
				stop = true
			}, timeout)
		}
		ticker := time.NewTicker(duration)
		for range ticker.C {
			if dealFn() {
				ticker.Stop()
				break
			}
		}
	}()
}

func RetryDo(channel chan interface{}, n int, duration time.Duration, timeout time.Duration, do func(interface{})) {
	if duration <= 0 {
		duration = 2 * time.Second
	}

	go func() {
		doneN := 0
		running := true
		if timeout > 0 {
			Delay(func() {
				running = false
			}, timeout)
		}

		var res interface{}
		for running {
			select {
			case res = <-channel:
				do(res)
				doneN++
				if n <= doneN {
					running = false
				}
			default:
				time.Sleep(duration)
			}
		}
	}()
}
