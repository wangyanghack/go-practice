# 1.部署好本机的 docker 环境，使用 ppt 中的 dockerfile build 自己的环境
```docker
FROM centos
RUN yum install dlv -y \
  && yum install binutils -y \
  && yum install vim -y \
  && yum install gdb -y
RUN yum install -y wget
RUN wget https://golang.org/dl/go1.17.linux-amd64.tar.gz \
  && rm -rf /usr/local/go && tar -C /usr/local -xzf go1.17.linux-amd64.tar.gz
```

# 2.使用 readelf 工具，查看编译后的进程入口地址

```shell
[root@9e1026f4fb4d home]# readelf -h ./hello
ELF Header:
  Magic:   7f 45 4c 46 02 01 01 00 00 00 00 00 00 00 00 00
  Class:                             ELF64
  Data:                              2's complement, little endian
  Version:                           1 (current)
  OS/ABI:                            UNIX - System V
  ABI Version:                       0
  Type:                              EXEC (Executable file)
  Machine:                           Advanced Micro Devices X86-64
  Version:                           0x1
  Entry point address:               0x45c1a0
  Start of program headers:          64 (bytes into file)
  Start of section headers:          456 (bytes into file)
  Flags:                             0x0
  Size of this header:               64 (bytes)
  Size of program headers:           56 (bytes)
  Number of program headers:         7
  Size of section headers:           64 (bytes)
  Number of section headers:         23
  Section header string table index: 3
```

# 3.在 dlv 调试工具中，使用断点功能找到代码位置
```shell
[root@9e1026f4fb4d home]# dlv exec ./hello --check-go-version=false
Type 'help' for list of commands.
(dlv) b *0x45c1a0
Breakpoint 1 set at 0x45c1a0 for _rt0_amd64_linux() /usr/local/go/src/runtime/rt0_linux_amd64.s:8
(dlv) c
> _rt0_amd64_linux() /usr/local/go/src/runtime/rt0_linux_amd64.s:8 (hits total:1) (PC: 0x45c1a0)
Warning: debugging optimized function
     3:	// license that can be found in the LICENSE file.
     4:
     5:	#include "textflag.h"
     6:
     7:	TEXT _rt0_amd64_linux(SB),NOSPLIT,$-8
=>   8:		JMP	_rt0_amd64(SB)
     9:
    10:	TEXT _rt0_amd64_linux_lib(SB),NOSPLIT,$0
    11:		JMP	_rt0_amd64_lib(SB)
(dlv)
```


# 4.使用断点调试功能，查看 Go 的 runtime 的下列函数执行流程，使用 IDE 查看函数的调用方：
## runqget：如果runnext非空，获取runnext上的g，为空就从本地队列runq的头获取g
```go
// If inheritTime is true, gp should inherit the remaining time in the
// current time slice. Otherwise, it should start a new time slice.
// Executed only by the owner P.
func runqget(_p_ *p) (gp *g, inheritTime bool) {
	// If there's a runnext, it's the next G to run.
	for {
		next := _p_.runnext
		if next == 0 {
			break
		}
		if _p_.runnext.cas(next, 0) {
			return next.ptr(), true
		}
	}

	for {
		h := atomic.LoadAcq(&_p_.runqhead) // load-acquire, synchronize with other consumers
		t := _p_.runqtail
		if t == h {
			return nil, false
		}
		gp := _p_.runq[h%uint32(len(_p_.runq))].ptr()
		if atomic.CasRel(&_p_.runqhead, h, h+1) { // cas-release, commits consume
			return gp, false
		}
	}
}
```
调用方：schedule,findrunnable,stealWork
## globrunqget：从全局队列里取出一部分放到本地队列
```go
// Try get a batch of G's from the global runnable queue.
// sched.lock must be held.
func globrunqget(_p_ *p, max int32) *g {
	assertLockHeld(&sched.lock)

	if sched.runqsize == 0 {
		return nil
	}

	n := sched.runqsize/gomaxprocs + 1
	if n > sched.runqsize {
		n = sched.runqsize
	}
	if max > 0 && n > max {
		n = max
	}
	if n > int32(len(_p_.runq))/2 {
		n = int32(len(_p_.runq)) / 2
	}

	sched.runqsize -= n

	gp := sched.runq.pop()
	n--
	for ; n > 0; n-- {
		gp1 := sched.runq.pop()
		runqput(_p_, gp1, false)
	}
	return gp
}
```
调用方：schedule:每61次调用一次全局队列，findrunnable
## runqput：将新g放到runnext，如果runnext有旧g，旧g移到本地队列
```go
// runqput tries to put g on the local runnable queue.
// If next is false, runqput adds g to the tail of the runnable queue.
// If next is true, runqput puts g in the _p_.runnext slot.
// If the run queue is full, runnext puts g on the global queue.
// Executed only by the owner P.
func runqput(_p_ *p, gp *g, next bool) {
	if randomizeScheduler && next && fastrand()%2 == 0 {
		next = false
	}

	if next {
	retryNext:
		oldnext := _p_.runnext
		if !_p_.runnext.cas(oldnext, guintptr(unsafe.Pointer(gp))) {
			goto retryNext
		}
		if oldnext == 0 {
			return
		}
		// Kick the old runnext out to the regular run queue.
		gp = oldnext.ptr()
	}

retry:
	h := atomic.LoadAcq(&_p_.runqhead) // load-acquire, synchronize with consumers
	t := _p_.runqtail
	if t-h < uint32(len(_p_.runq)) {
		_p_.runq[t%uint32(len(_p_.runq))].set(gp)
		atomic.StoreRel(&_p_.runqtail, t+1) // store-release, makes the item available for consumption
		return
	}
	if runqputslow(_p_, gp, h, t) {
		return
	}
	// the queue is not full, now the put above must succeed
	goto retry
}
```
调用方：globrunqget，newproc，goyield_m，ready
## globrunqput：把g放到全局队列里
```go
// Put gp on the global runnable queue.
// sched.lock must be held.
// May run during STW, so write barriers are not allowed.
//go:nowritebarrierrec
func globrunqput(gp *g) {
	assertLockHeld(&sched.lock)

	sched.runq.pushBack(gp)
	sched.runqsize++
}
```
调用方：exitsyscall0，goschedImpl，debugCallWrap1
## schedule
exitsyscall0  
goexit0  
goschedImpl  
goyield_m  
mstart1  
park_m  
preemptPark
## findrunnable
schedule
## sysmon
newm
