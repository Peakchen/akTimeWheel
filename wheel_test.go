package akTimeWheel

import (
	"testing"
	"time"

	"github.com/Peakchen/xgameCommon/utils"

	"github.com/Peakchen/xgameCommon/akLog"
)

func init() {
	Start()
}

//TW_SmallMs TW = 1 //5ms  	(5-100]ms
func Test_TW_SmallMs(t *testing.T) {
	var exit = make(chan int, 1)
	var count int64
	var interval int64 = 5
	var runs int64 = 1000
	var now = time.Now()
	AddTimer(TW_SmallMs, interval, runs, func() {
		count++
		akLog.FmtPrintln("TW_SmallMs hello...", count)
		if count >= runs && runs > 0 {
			exit <- 1
		}
	})
	select {
	case <-exit:
		akLog.FmtPrintln("TW_SmallMs exit...", time.Since(now))
	}
}

//TW_BigMs   TW = 2 //50ms 	(100-1000]ms
func Test_TW_BigMs(t *testing.T) {
	var exit = make(chan int, 1)
	var count int64
	var interval int64 = 101
	var runs int64 = 2
	var now = time.Now()
	AddTimer(TW_BigMs, interval, runs, func() {
		count++
		akLog.FmtPrintln("TW_BigMs hello...", count)
		if count >= runs && runs > 0 {
			exit <- 1
		}
	})
	select {
	case <-exit:
		akLog.FmtPrintln("TW_BigMs exit...", time.Since(now))
	}
}

//TW_Sec     TW = 3 //1s   	(1-60]s
func Test_TW_Sec(t *testing.T) {
	var exit = make(chan int, 1)
	var count int64
	var interval int64 = utils.Sec2Ms_Float64(2.101)
	var runs int64 = 2
	var now = time.Now()
	AddTimer(TW_Sec, interval, runs, func() {
		count++
		akLog.FmtPrintln("TW_Sec hello...", count)
		if count >= runs && runs > 0 {
			exit <- 1
		}
	})
	select {
	case <-exit:
		akLog.FmtPrintln("TW_Sec exit...", time.Since(now))
	}
}

//TW_Min     TW = 4 //60s  	(1-60]min
func Test_TW_Min(t *testing.T) {
	var exit = make(chan int, 1)
	var count int64
	var interval int64 = utils.Min2Ms_Float64(1.101)
	var runs int64 = 2
	var now = time.Now()
	AddTimer(TW_Min, interval, runs, func() {
		count++
		akLog.FmtPrintln("TW_Min hello...", count)
		if count >= runs && runs > 0 {
			exit <- 1
		}
	})
	select {
	case <-exit:
		akLog.FmtPrintln("TW_Min exit...", time.Since(now))
	}
}

func TestTimeWheel(t *testing.T) {
	var exit = make(chan int, 1)
	var count int64
	var interval int64 = 2
	var runs int64 = 2

	AddTimer(TW_Sec, interval, runs, func() {
		count++
		akLog.FmtPrintln("hello...", count)
		if count >= runs {
			exit <- 1
		}
	})
	select {
	case <-exit:
		akLog.FmtPrintln("exit...")
	}
}
