package snowflake

import (
	"sync"
	"time"
)

const (
	//开始时间戳2021-01-01 00:00:00 +0000
	DefaultEpoch = 1609430400000

	//机器id位数
	DefaultWorkerIDBits = 10
	//12位序列号
	DefaultSequenceBits = 12

	TimestampBits = 41
	TotalBits     = 64
)

type SnowFlakeWorkerIdFunc func() int64

type SnowFlake struct {
	epoch        int64
	workerID     int64
	sequenceBits int64
	workerIDBits int64

	mu sync.Mutex
	//开始时间戳
	lastTimestamp int64
	sequence      int64
}

type OptionFunc func(o *SnowFlake)

func WithEpoch(epoch int64) OptionFunc {
	return func(sf *SnowFlake) {
		sf.epoch = epoch
	}
}

func WithSequenceBits(bits int64) OptionFunc {
	return func(o *SnowFlake) {
		o.sequenceBits = bits
	}
}

func WithWorkerIDBits(bits int64) OptionFunc {
	return func(o *SnowFlake) {
		o.workerIDBits = bits
	}
}

func (sf *SnowFlake) Next() int64 {
	sf.mu.Lock()
	defer sf.mu.Unlock()

	now := time.Now().UnixNano()/1000000 - sf.epoch
	if now == sf.lastTimestamp {
		sf.sequence = (sf.sequence + 1) & sf.sequenceMask()
		if sf.sequence == 0 {
			// 阻塞到下一毫秒，获得新的时间戳
			for now <= sf.lastTimestamp {
				now = time.Now().UnixNano()/1000000 - sf.epoch
			}
		}
	} else {
		//时间戳改变，毫秒内序列重置
		sf.sequence = 0
	}

	sf.lastTimestamp = now

	return now<<sf.timestampShift() | sf.workerID<<sf.nodeIdShift() | sf.sequence
}

func (sf *SnowFlake) maxWorkerID() int64 {
	return -1 ^ (-1 << sf.workerIDBits)
}

func (sf *SnowFlake) sequenceMask() int64 {
	return -1 ^ (-1 << sf.sequenceBits)
}

func (sf *SnowFlake) nodeIdShift() int64 {
	return sf.sequenceBits
}

func (sf *SnowFlake) timestampShift() int64 {
	return sf.workerIDBits + sf.sequenceBits
}

func NewSnowFlake(wIDFunc SnowFlakeWorkerIdFunc, options ...OptionFunc) (*SnowFlake, error) {
	sf := &SnowFlake{
		epoch:        DefaultEpoch,
		sequenceBits: DefaultSequenceBits,
		workerIDBits: DefaultWorkerIDBits,
	}

	for _, o := range options {
		o(sf)
	}

	if TotalBits != (sf.sequenceBits + sf.workerIDBits + TimestampBits + 1) {
		return nil, ErrBits
	}

	workerID := wIDFunc()
	if workerID > sf.maxWorkerID() {
		return nil, ErrWorkerID
	}
	sf.workerID = workerID

	return sf, nil
}
