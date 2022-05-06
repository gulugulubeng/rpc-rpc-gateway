package pool

import (
	"time"
)

// Worker 真正执行任务的执行体(对应一个协程)
type Worker struct {
	// pool 所在协程池
	pool *Pool

	// args 需要处理的任务
	args chan interface{}

	// recycleTime 记录执行体被放入pool中空闲执行体队列的时间
	recycleTime time.Time
}

// run go关键词开启协程 处理任务
func (w *Worker) run() {
	go func() {
		// 捕捉panic
		defer func() {
			if p := recover(); p != nil {
				w.pool.decRunning()
				if w.pool.PanicHandler != nil {
					w.pool.PanicHandler(p)
				} else {
					panic(p)
				}
			}
		}()
		// 执行体阻塞获取任务
		for arg := range w.args {
			if arg == nil {
				w.pool.decRunning()
				return
			}
			// 执行任务
			w.pool.poolFunc(arg)
			// 任务执行完毕后 将执行体添加进pool空闲执行体队列中
			w.pool.putWorker(w)
		}
	}()
}
