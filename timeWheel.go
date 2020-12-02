package akTimeWheel

import (
	"container/list"
	"sync"
	"time"

	"github.com/Peakchen/xgameCommon/utils"

	"github.com/Peakchen/xgameCommon/akLog"
)

type TW int

const (
	TW_NO      TW = 0
	TW_SmallMs TW = 1 //5ms  	(5-100]ms
	TW_BigMs   TW = 2 //50ms 	(100-1000]ms
	TW_Sec     TW = 3 //1s   	(1-60]s
	TW_Min     TW = 4 //60s  	(1-60]min
	TW_Day     TW = 5 //30min 	(1-24]h
	TW_Month   TW = 6 //12h    (1-30]d
	TW_Year    TW = 7 //15d    (1-12]month

	TW_Max TW = TW_Year
)

type TWRange struct {
	min int64
	max int64
}

var (
	twTickers = map[TW]time.Duration{
		TW_SmallMs: 5 * time.Millisecond,
		TW_BigMs:   50 * time.Millisecond,
		TW_Sec:     1 * time.Second,
		TW_Min:     60 * time.Second,
		TW_Day:     30 * time.Minute,
		TW_Month:   12 * time.Hour,
		TW_Year:    15 * 24 * time.Hour,
	}

	twScale = map[TW]int64{
		TW_SmallMs: 5,
		TW_BigMs:   50,
		TW_Sec:     utils.Sec2Ms_Int64(1),
		TW_Min:     utils.Sec2Ms_Int64(60),
		TW_Day:     utils.Min2Ms_Int64(30),
		TW_Month:   utils.Hour2Ms_Int64(12),
		TW_Year:    utils.Day2Ms_Int64(15),
	}

	twRange = map[TW]*TWRange{
		TW_SmallMs: &TWRange{min: 5, max: 100},
		TW_BigMs:   &TWRange{min: 100, max: 1000},
		TW_Sec:     &TWRange{min: utils.Sec2Ms_Int64(1), max: utils.Sec2Ms_Int64(60)},
		TW_Min:     &TWRange{min: utils.Min2Ms_Int64(1), max: utils.Min2Ms_Int64(60)},
		TW_Day:     &TWRange{min: utils.Hour2Ms_Int64(1), max: utils.Hour2Ms_Int64(24)},
		TW_Month:   &TWRange{min: utils.Day2Ms_Int64(1), max: utils.Day2Ms_Int64(30)},
		TW_Year:    &TWRange{min: utils.Month2Ms_Int64(1), max: utils.Month2Ms_Int64(12)},
	}
)

const (
	TaskElementMax = 1000
)

type akTimerMgr struct {
	tws  map[TW]*TimeWheel
	ext  chan *TaskElement
	stop chan bool
}

type TaskElement struct {
	ty   TW
	task *list.Element
}

var (
	_akTimerMgr *akTimerMgr
)

func Start() {
	_akTimerMgr = &akTimerMgr{
		tws:  make(map[TW]*TimeWheel, TW_Max),
		ext:  make(chan *TaskElement, TaskElementMax),
		stop: make(chan bool, 1),
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go _akTimerMgr.wheelLoop(&wg)

	for wi := TW_SmallMs; wi <= TW_Max; wi++ {
		wg.Add(1)
		_akTimerMgr.tws[wi] = &TimeWheel{
			ty:    wi,
			tasks: list.New(),
		}
		go _akTimerMgr.tws[wi].taskLoop(_akTimerMgr.ext, &wg)
	}
}

func AddTimer(ty TW, interval int64, runs int64, fn func()) {
	_akTimerMgr.addTask(ty, interval, runs, fn)
}

/*
	ty :  time wheel type
	interval : task time interval
	runs : time wheel run times
	fn : function call back
*/
func (this *akTimerMgr) addTask(ty TW, interval int64, runs int64, fn func()) {
	tRange, ok := twRange[ty]
	if !ok {
		akLog.Error("can not find tw range data, ty: ", ty)
		return
	}
	if tRange.min > interval || tRange.max < interval {
		akLog.Error("tw interval over range, ty: ", ty, interval)
		return
	}
	this.tws[ty].add(interval, runs, fn)
}

func (this *akTimerMgr) wheelLoop(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case ele := <-this.ext:
			this.tws[ele.ty].pushBack(ele.task.Value.(*WheelTask))
		}
	}
}

func getTimewheelScale(src TW, interval int64) (dst TW) {
	for i := src - 1; i >= TW_SmallMs; i-- {
		if (interval > 0 && twRange[i].min >= interval) ||
			(twRange[i].min < interval && twRange[i].max >= interval) {
			return i
		}
	}
	return TW_NO
}

type TimeWheel struct {
	ty    TW
	tasks *list.List
}

type WheelTask struct {
	create   int64
	start    int64
	interval int64
	runs     int64
	fn       func()
	scales   int64
	mark     int64
}

func (this *WheelTask) reset() {
	this.mark = 0
}

func (this *TimeWheel) add(interval int64, runs int64, fn func()) {
	this.tasks.PushBack(&WheelTask{
		create:   time.Now().Unix(),
		interval: interval,
		runs:     runs,
		fn:       fn,
	})
}

func (this *TimeWheel) pushBack(e interface{}) {
	this.tasks.PushBack(e)
}

func (this *TimeWheel) pop(e interface{}) {
	this.tasks.Remove(e.(*list.Element))
}

func (this *TimeWheel) taskLoop(ext chan *TaskElement, wg *sync.WaitGroup) {
	ticker := time.NewTicker(twTickers[this.ty])
	defer func() {
		ticker.Stop()
		wg.Done()
	}()

	for {
		select {
		case <-ticker.C:
			if this.tasks.Len() == 0 {
				continue
			}

			akLog.FmtPrintf("cur ty: %v.", this.ty)
			e := this.tasks.Front()
			if e != nil {

				task := e.Value.(*WheelTask)
				if task.mark == 0 && task.start == 0 {
					task.start = time.Now().Unix()
				}

				task.scales += int64(twScale[this.ty])
				subscale := task.interval - task.scales
				if subscale <= 0 {
					task.scales = 0
					task.mark++
					task.fn()
				} else if subscale < twScale[this.ty] && subscale > 0 {
					dst := getTimewheelScale(this.ty, subscale)
					if dst != TW_NO {
						akLog.FmtPrintf("get next ty: %v, src: %v, subscale: %v.", dst, this.ty, subscale)
						ext <- &TaskElement{
							ty: dst,
							task: &list.Element{
								Value: &WheelTask{
									create:   time.Now().Unix(),
									interval: subscale,
									runs:     1,
									fn:       task.fn,
								},
							},
						}
						task.scales = 0
						task.mark++
					}
				}
				if task.mark == task.runs {
					this.pop(e)
				}
			}
		}
	}
}
