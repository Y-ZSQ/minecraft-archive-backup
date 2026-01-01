package pool

import (
	"sync"
	"time"
)

// Pool 通用对象池
type Pool struct {
	pool       sync.Pool
	entries    []interface{}
	mu         sync.Mutex
	cond       *sync.Cond
	closed     bool
	maxRetain  int           // 最大保留数量
	minRetain  int           // 最小保留数量
	cleanupInt time.Duration // 清理间隔
	lastClean  time.Time     // 上次清理时间
}

// NewPool 创建新的对象池
func NewPool(newFunc func() interface{}, maxRetain, minRetain int, cleanupInt time.Duration) *Pool {
	p := &Pool{
		pool: sync.Pool{
			New: newFunc,
		},
		entries:    make([]interface{}, 0, maxRetain),
		maxRetain:  maxRetain,
		minRetain:  minRetain,
		cleanupInt: cleanupInt,
		lastClean:  time.Now(),
		closed:     false,
	}
	p.cond = sync.NewCond(&p.mu)
	go p.cleanupWorker()
	return p
}

// cleanupWorker 清理工作协程
func (p *Pool) cleanupWorker() {
	ticker := time.NewTicker(p.cleanupInt)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.mu.Lock()
			if p.closed {
				p.mu.Unlock()
				return
			}
			p.lastClean = time.Now()
			p.cleanup()
			p.mu.Unlock()
		}
	}
}

// cleanup 清理过期的对象
func (p *Pool) cleanup() {
	if len(p.entries) > p.minRetain {
		// 移除超出最小保留数量的对象
		removeCount := len(p.entries) - p.minRetain
		for i := 0; i < removeCount; i++ {
			// 从后面开始移除（先进后出）
			lastIndex := len(p.entries) - 1
			p.entries[lastIndex] = nil
			p.entries = p.entries[:lastIndex]
		}
	}
}

// Get 从池中获取对象
func (p *Pool) Get() interface{} {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 首先尝试从内部列表获取
	if len(p.entries) > 0 {
		lastIndex := len(p.entries) - 1
		item := p.entries[lastIndex]
		p.entries = p.entries[:lastIndex]
		return item
	}

	// 内部列表为空，从sync.Pool获取
	return p.pool.Get()
}

// Put 将对象放回池中
func (p *Pool) Put(x interface{}) {
	if x == nil {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// 添加到内部列表
	p.entries = append(p.entries, x)

	// 如果内部列表太长，调用sync.Pool的Put
	if len(p.entries) > p.maxRetain {
		lastIndex := len(p.entries) - 1
		p.pool.Put(p.entries[lastIndex])
		p.entries = p.entries[:lastIndex]
	}
}

// Close 关闭对象池
func (p *Pool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.closed = true
	p.cond.Broadcast()

	// 清理所有引用
	for i := range p.entries {
		p.entries[i] = nil
	}
	p.entries = nil
}

// Size 返回当前池中对象数量
func (p *Pool) Size() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.entries)
}
