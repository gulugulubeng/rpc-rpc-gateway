package pool

import (
	"sync"
	"sync/atomic"
	"time"
)

type sig struct{}

// Pool 协程池结构体
// 接受外部任务 循环利用Pool中一定数量的协程池
type Pool struct {
	// capacity 容量
	capacity int32

	// running 正在运行的协程数
	running int32

	// workers 可用执行体队列（可以看作是空闲协程集合）
	workers []*Worker

	// poolFunc 执行体操作函数
	poolFunc func(arg interface{})

	// expiryDuration 执行任务超时时长
	expiryDuration time.Duration

	// release 接受关闭协程池信号
	release chan sig

	// lock 互斥锁
	lock sync.Mutex

	// cond 为获取一个空闲执行体(Cond对象不能被拷贝 所以只能使用指针)
	cond *sync.Cond

	// once 用于保证协程池只被关闭一次
	once sync.Once

	// PanicHandler 处理任务执行过程中发生的panic
	// 如果为空 将会抛出panic
	PanicHandler func(interface{})
}

// NewPool 创建任务操作函数为f的协程池
func NewPool(f func(interface{})) (*Pool, error) {
	return NewTimingPool(DefaultPoolSize, DefaultCleanIntervalTime, f)
}

// NewTimingPool 创建容量为size任务超时时间为expiry的协程池
func NewTimingPool(size, expiry int, f func(interface{})) (*Pool, error) {
	// 参数校验
	if size <= 0 {
		return nil, ErrInvalidPoolSize
	}
	if expiry <= 0 {
		return nil, ErrInvalidPoolExpiry
	}
	if f == nil {
		return nil, ErrInvalidPoolFunc
	}
	// 初始化pool
	p := &Pool{
		capacity:       int32(size),
		release:        make(chan sig, 1),
		expiryDuration: time.Duration(expiry) * time.Second,
		poolFunc:       f,
	}
	p.cond = sync.NewCond(&p.lock)
	// 异步协程清除pool中过期的空闲执行体
	go p.periodicallyPurge()
	return p, nil
}

// 清除pool中过期的空闲执行体
func (p *Pool) periodicallyPurge() {
	// 创建定时器
	heartbeat := time.NewTicker(p.expiryDuration)
	defer heartbeat.Stop()

	// 定时循环执行清除任务
	for range heartbeat.C {
		currentTime := time.Now()
		p.lock.Lock()
		idleWorkers := p.workers
		// 当pool中不存在空闲、运行的执行体 且pool已经被关闭时 则直接返回
		if len(idleWorkers) == 0 && p.Running() == 0 && len(p.release) > 0 {
			p.lock.Unlock()
			return
		}
		// 遍历pool中空闲执行体
		n := -1
		for i, w := range idleWorkers {
			// 查询到过期空闲执行体
			if currentTime.Sub(w.recycleTime) <= p.expiryDuration {
				break
			}
			n = i
			w.args <- nil
			idleWorkers[i] = nil
		}
		// 清除过期空闲执行体
		if n > -1 {
			if n >= len(idleWorkers)-1 {
				p.workers = idleWorkers[:0]
			} else {
				p.workers = idleWorkers[n+1:]
			}
		}
		p.lock.Unlock()
	}
}

//-------------------------------------------------------------------------

// Submit 提交执行任务到协程池
func (p *Pool) Submit(arg interface{}) error {
	if len(p.release) > 0 {
		return ErrPoolClosed
	}
	p.getWorker().args <- arg
	return nil
}

// Running 返回当前活跃的协程数量
func (p *Pool) Running() int {
	return int(atomic.LoadInt32(&p.running))
}

// Free 返回当前空闲的协程数量
func (p *Pool) Free() int {
	return int(atomic.LoadInt32(&p.capacity) - atomic.LoadInt32(&p.running))
}

// Cap 返回当前协程池容量
func (p *Pool) Cap() int {
	return int(atomic.LoadInt32(&p.capacity))
}

// ReSize 重置协程池容量
func (p *Pool) ReSize(size int) {
	if size == p.Cap() {
		return
	}
	atomic.StoreInt32(&p.capacity, int32(size))
	diff := p.Running() - size
	for i := 0; i < diff; i++ {
		p.getWorker().args <- nil
	}
}

// Release 关闭协程池
func (p *Pool) Release() error {
	p.once.Do(func() {
		// 向通道中发送pool关闭信号
		p.release <- sig{}
		// 给pool加锁 保证线程安全
		p.lock.Lock()
		// 将pool中执行体全置为空
		idleWorkers := p.workers
		for i, w := range idleWorkers {
			w.args <- nil
			idleWorkers[i] = nil
		}
		p.workers = nil
		// 解锁
		p.lock.Unlock()
	})
	return nil
}

//-------------------------------------------------------------------------

// incRunning 活跃协程数量+1
func (p *Pool) incRunning() {
	atomic.AddInt32(&p.running, 1)
}

// decRunning 活跃协程数量-1
func (p *Pool) decRunning() {
	atomic.AddInt32(&p.running, -1)
}

// getWorker 返回协程池中当前可用的执行体
func (p *Pool) getWorker() *Worker {
	var w *Worker
	waiting := false

	// 将pool加锁
	p.lock.Lock()
	// 获取pool中空间执行体数量n
	idleWorkers := p.workers
	n := len(idleWorkers) - 1
	if n < 0 {
		// pool中运行的执行体数量达到pool容量时 waiting = true
		waiting = p.Running() >= p.Cap()
	} else {
		// 当pool中存在空闲执行体是 取出执行体队列最后的执行体作为返回的执行体
		w = idleWorkers[n]
		p.workers = idleWorkers[:n]
	}
	p.lock.Unlock()

	// waiting = true 则代表pool中运行的执行体已经达到pool容量
	// 这就需要等待获取一个pool中运行的执行体运行完毕
	if waiting {
		for {
			// 阻塞等待获取空闲的执行体
			p.cond.Wait()
			p.lock.Lock()
			l := len(p.workers) - 1
			if l < 0 {
				p.lock.Unlock()
				continue
			}
			w = p.workers[l]
			p.workers[l] = nil
			p.workers = p.workers[:l]
			p.lock.Unlock()
			break
		}
	} else if w == nil {
		// 初始化新的执行体返回
		w = &Worker{
			pool: p,
			args: make(chan interface{}, 1),
		}
		w.run()
		p.incRunning()
	}
	return w
}

// putWorker 将运行完毕的执行体加入pool中空闲执行体队列
func (p *Pool) putWorker(worker *Worker) {
	// 设置空闲执行体过期开始时间
	worker.recycleTime = time.Now()
	p.lock.Lock()
	// 将空闲执行体加入pool
	p.workers = append(p.workers, worker)
	// 唤起一个p.cond.Wait()阻塞的协程 通知有pool中有新的空闲执行体
	p.cond.Signal()
	p.lock.Unlock()
}
