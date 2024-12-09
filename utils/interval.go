package utils

import (
	"log"
	"time"
)

func IntervalAsync(procName string,secondS int64,proc func()error){
	go func() {
		IntervalSync(procName ,secondS,proc)
	}()
}
func IntervalSync(procName string,secondS int64,proc func()error){
	MIntervalSync(procName ,secondS*1000,proc)
}
func MIntervalSync(procName string,millsecondS int64,proc func()error){
	var timer = time.NewTicker(time.Duration(millsecondS) * time.Millisecond)
	defer timer.Stop()
	log.Println(procName,"timer thread begin with interval millsecondS:",millsecondS)
	for {
		select {
		case <-timer.C:
			func() {
				defer func() {
					//必须要先声明defer，否则不能捕获到panic异常
					if err := recover(); err != nil {
						log.Println(err)    //这里的err其实就是panic传入的内容，55
						log.Println("timer thread Sleep 10秒重试")
						time.Sleep(time.Second * 10)
					}
				}()
				log.Println(procName,"begin")
				err:=proc()
				if err!=nil{
					log.Println(procName,err)
				}
			}()
		}
		//等触发时的信号
		//log.Println("timer end!")
	}
	log.Println("timer thread exit!")
}