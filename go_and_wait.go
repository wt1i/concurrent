package concurrent

import (
	"context"
	"fmt"
	"runtime"
	"sync"
)

const (
	DefaultStackSize = 4096
)

// int 为 for range 的下标, 处理 for range 里面创建 Handle 才能更加得心应手
type ContextHandle func(context.Context, int) error

type errList []error

func (e errList) FilterNil() []error {
	errList := make(errList, 0)
	for _, v := range e {
		if v != nil {
			errList = append(errList, v)
		}
	}
	return errList
}

func (e errList) HasError() bool {
	for _, v := range e {
		if v != nil {
			return true
		}
	}
	return false
}

// GoAndWait 封装了 sync.WaitGroup 直接构造好 func 拉函数即可
// 该函数不解决数据竞争问题,需要 func 自行解决
func GoAndWait(ctx context.Context, handlers []ContextHandle) errList {
	wg := &sync.WaitGroup{}
	wg.Add(len(handlers))
	errList := make([]error, len(handlers))
	for i := range handlers {
		go func(ctx context.Context, i int) {
			defer wg.Done()
			defer func() {
				if err := recover(); err != nil {
					errList[i] = fmt.Errorf("[panic] err: %v\nstack: %s", err, getCurrentGoroutineStack())
				}
			}()
			if err := handlers[i](ctx, i); err != nil {
				errList[i] = err
			}
		}(ctx, i)
	}
	wg.Wait()
	return errList
}

// getCurrentGoroutineStack 获取当前Goroutine的调用栈，便于排查panic异常
func getCurrentGoroutineStack() string {
	var buf [DefaultStackSize]byte
	n := runtime.Stack(buf[:], false)
	return string(buf[:n])
}
