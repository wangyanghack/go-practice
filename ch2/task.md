# ppt:33 页的三条红线位置，请在 runtime 的代码中找到(可以使用课上介绍的编译、反编译工具查找)
1. 关闭nil channel发生panic：
```go
package main

func main() {
	var ch chan int
	close(ch)
	ch <- 1
}
```
先通过go tool compile编译，close(ch)会调用runtime.closechan()，ch<-1会调用runtime.closesend1()
```shell
➜  ch2 go tool compile -S send_to_nil.go
"".main STEXT size=58 args=0x0 locals=0x18 funcid=0x0
        0x0000 00000 (send_to_nil.go:3) TEXT    "".main(SB), ABIInternal, $24-0
        0x0000 00000 (send_to_nil.go:3) CMPQ    SP, 16(R14)
        0x0004 00004 (send_to_nil.go:3) PCDATA  $0, $-2
        0x0004 00004 (send_to_nil.go:3) JLS     51
        0x0006 00006 (send_to_nil.go:3) PCDATA  $0, $-1
        0x0006 00006 (send_to_nil.go:3) SUBQ    $24, SP
        0x000a 00010 (send_to_nil.go:3) MOVQ    BP, 16(SP)
        0x000f 00015 (send_to_nil.go:3) LEAQ    16(SP), BP
        0x0014 00020 (send_to_nil.go:3) FUNCDATA        $0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
        0x0014 00020 (send_to_nil.go:3) FUNCDATA        $1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
        0x0014 00020 (send_to_nil.go:5) XORL    AX, AX
        0x0016 00022 (send_to_nil.go:5) PCDATA  $1, $0
        0x0016 00022 (send_to_nil.go:5) CALL    runtime.closechan(SB)
        0x001b 00027 (send_to_nil.go:6) XORL    AX, AX
        0x001d 00029 (send_to_nil.go:6) LEAQ    ""..stmp_0(SB), BX
        0x0024 00036 (send_to_nil.go:6) CALL    runtime.chansend1(SB)
        0x0029 00041 (send_to_nil.go:7) MOVQ    16(SP), BP
        0x002e 00046 (send_to_nil.go:7) ADDQ    $24, SP
        0x0032 00050 (send_to_nil.go:7) RET
        0x0033 00051 (send_to_nil.go:7) NOP
        0x0033 00051 (send_to_nil.go:3) PCDATA  $1, $-1
        0x0033 00051 (send_to_nil.go:3) PCDATA  $0, $-2
        0x0033 00051 (send_to_nil.go:3) CALL    runtime.morestack_noctxt(SB)
        0x0038 00056 (send_to_nil.go:3) PCDATA  $0, $-1
        0x0038 00056 (send_to_nil.go:3) JMP     0
```
dlv查看closechan()会在关闭channel的时候检查c是否为nil，c为nil时会发生panic: close of nil channel
```shell
> runtime.closechan() /usr/local/go/src/runtime/chan.go:357 (PC: 0x1004425)
Warning: debugging optimized function
   352:		memmove(dst, src, t.size)
   353:	}
   354:
   355:	func closechan(c *hchan) {
   356:		if c == nil {
=> 357:			panic(plainError("close of nil channel"))
   358:		}
   359:
   360:		lock(&c.lock)
   361:		if c.closed != 0 {
   362:			unlock(&c.lock)
(dlv)
```
2. 关闭已经关闭的channel发生panic：
```go
package main

func main() {
	ch := make(chan int)
	close(ch)
	close(ch)
	ch <- 1
}
```
```shell
➜  ch2 go tool compile -S send_to_nil.go
"".main STEXT size=91 args=0x0 locals=0x20 funcid=0x0
        0x0000 00000 (send_to_nil.go:3) TEXT    "".main(SB), ABIInternal, $32-0
        0x0000 00000 (send_to_nil.go:3) CMPQ    SP, 16(R14)
        0x0004 00004 (send_to_nil.go:3) PCDATA  $0, $-2
        0x0004 00004 (send_to_nil.go:3) JLS     84
        0x0006 00006 (send_to_nil.go:3) PCDATA  $0, $-1
        0x0006 00006 (send_to_nil.go:3) SUBQ    $32, SP
        0x000a 00010 (send_to_nil.go:3) MOVQ    BP, 24(SP)
        0x000f 00015 (send_to_nil.go:3) LEAQ    24(SP), BP
        0x0014 00020 (send_to_nil.go:3) FUNCDATA        $0, gclocals·69c1753bd5f81501d95132d08af04464(SB)
        0x0014 00020 (send_to_nil.go:3) FUNCDATA        $1, gclocals·9fb7f0986f647f17cb53dda1484e0f7a(SB)
        0x0014 00020 (send_to_nil.go:4) LEAQ    type.chan int(SB), AX
        0x001b 00027 (send_to_nil.go:4) XORL    BX, BX
        0x001d 00029 (send_to_nil.go:4) PCDATA  $1, $0
        0x001d 00029 (send_to_nil.go:4) NOP
        0x0020 00032 (send_to_nil.go:4) CALL    runtime.makechan(SB)
        0x0025 00037 (send_to_nil.go:4) MOVQ    AX, "".ch+16(SP)
        0x002a 00042 (send_to_nil.go:5) PCDATA  $1, $1
        0x002a 00042 (send_to_nil.go:5) CALL    runtime.closechan(SB)
        0x002f 00047 (send_to_nil.go:6) MOVQ    "".ch+16(SP), AX
        0x0034 00052 (send_to_nil.go:6) CALL    runtime.closechan(SB)
        0x0039 00057 (send_to_nil.go:7) MOVQ    "".ch+16(SP), AX
        0x003e 00062 (send_to_nil.go:7) LEAQ    ""..stmp_0(SB), BX
        0x0045 00069 (send_to_nil.go:7) PCDATA  $1, $0
        0x0045 00069 (send_to_nil.go:7) CALL    runtime.chansend1(SB)
        0x004a 00074 (send_to_nil.go:8) MOVQ    24(SP), BP
        0x004f 00079 (send_to_nil.go:8) ADDQ    $32, SP
        0x0053 00083 (send_to_nil.go:8) RET
        0x0054 00084 (send_to_nil.go:8) NOP
        0x0054 00084 (send_to_nil.go:3) PCDATA  $1, $-1
        0x0054 00084 (send_to_nil.go:3) PCDATA  $0, $-2
        0x0054 00084 (send_to_nil.go:3) CALL    runtime.morestack_noctxt(SB)
        0x0059 00089 (send_to_nil.go:3) PCDATA  $0, $-1
        0x0059 00089 (send_to_nil.go:3) JMP     0
```
closechan()之后会检查c.closed是否为0，如果检查c.closed为0，说明还没有被关闭，执行c.closed=1，关闭channel
```shell
> runtime.closechan() /usr/local/go/src/runtime/chan.go:372 (PC: 0x100424e)
Warning: debugging optimized function
   367:			callerpc := getcallerpc()
   368:			racewritepc(c.raceaddr(), callerpc, funcPC(closechan))
   369:			racerelease(c.raceaddr())
   370:		}
   371:
=> 372:		c.closed = 1
   373:
   374:		var glist gList
   375:
   376:		// release all readers
   377:		for {
(dlv)
```
如果不为0，说明已经被关闭，发生panic： close of closed channel
```shell
> runtime.closechan() /usr/local/go/src/runtime/chan.go:361 (PC: 0x100423f)
Warning: debugging optimized function
Values returned:

   356:		if c == nil {
   357:			panic(plainError("close of nil channel"))
   358:		}
   359:
   360:		lock(&c.lock)
=> 361:		if c.closed != 0 {
   362:			unlock(&c.lock)
   363:			panic(plainError("close of closed channel"))
   364:		}
   365:
   366:		if raceenabled {
(dlv)
```
3. 往关闭的channel写数据发生panic：
```go
package main

func main() {
	ch := make(chan int)
	close(ch)
	ch <- 1
}
```
先通过go tool compile编译，可以看到make(chan int)会调用runtime.makechan()，close(ch)会调用runtime.closechan()，ch<-1会调用runtime.closesend1()
```shell
➜  ch2 go tool compile -S send_to_nil.go
"".main STEXT size=86 args=0x0 locals=0x20 funcid=0x0
        0x0000 00000 (send_to_nil.go:3) TEXT    "".main(SB), ABIInternal, $32-0
        0x0000 00000 (send_to_nil.go:3) CMPQ    SP, 16(R14)
        0x0004 00004 (send_to_nil.go:3) PCDATA  $0, $-2
        0x0004 00004 (send_to_nil.go:3) JLS     79
        0x0006 00006 (send_to_nil.go:3) PCDATA  $0, $-1
        0x0006 00006 (send_to_nil.go:3) SUBQ    $32, SP
        0x000a 00010 (send_to_nil.go:3) MOVQ    BP, 24(SP)
        0x000f 00015 (send_to_nil.go:3) LEAQ    24(SP), BP
        0x0014 00020 (send_to_nil.go:3) FUNCDATA        $0, gclocals·69c1753bd5f81501d95132d08af04464(SB)
        0x0014 00020 (send_to_nil.go:3) FUNCDATA        $1, gclocals·9fb7f0986f647f17cb53dda1484e0f7a(SB)
        0x0014 00020 (send_to_nil.go:4) LEAQ    type.chan int(SB), AX
        0x001b 00027 (send_to_nil.go:4) XORL    BX, BX
        0x001d 00029 (send_to_nil.go:4) PCDATA  $1, $0
        0x001d 00029 (send_to_nil.go:4) NOP
        0x0020 00032 (send_to_nil.go:4) CALL    runtime.makechan(SB)
        0x0025 00037 (send_to_nil.go:4) MOVQ    AX, "".ch+16(SP)
        0x002a 00042 (send_to_nil.go:5) PCDATA  $1, $1
        0x002a 00042 (send_to_nil.go:5) CALL    runtime.closechan(SB)
        0x002f 00047 (send_to_nil.go:6) MOVQ    "".ch+16(SP), AX
        0x0034 00052 (send_to_nil.go:6) LEAQ    ""..stmp_0(SB), BX
        0x003b 00059 (send_to_nil.go:6) PCDATA  $1, $0
        0x003b 00059 (send_to_nil.go:6) NOP
        0x0040 00064 (send_to_nil.go:6) CALL    runtime.chansend1(SB)
        0x0045 00069 (send_to_nil.go:7) MOVQ    24(SP), BP
        0x004a 00074 (send_to_nil.go:7) ADDQ    $32, SP
        0x004e 00078 (send_to_nil.go:7) RET
        0x004f 00079 (send_to_nil.go:7) NOP
        0x004f 00079 (send_to_nil.go:3) PCDATA  $1, $-1
        0x004f 00079 (send_to_nil.go:3) PCDATA  $0, $-2
        0x004f 00079 (send_to_nil.go:3) CALL    runtime.morestack_noctxt(SB)
        0x0054 00084 (send_to_nil.go:3) PCDATA  $0, $-1
        0x0054 00084 (send_to_nil.go:3) JMP     0
```
dlv查看chansend()在向channel写数据之前会先通过closed是否为0检查channel是否已经关闭，如果不为0，就会发生panic: send on closed channel
```shell
> runtime.chansend() /usr/local/go/src/runtime/chan.go:204 (PC: 0x1003f33)
Warning: debugging optimized function
Values returned:

   199:
   200:		lock(&c.lock)
   201:
   202:		if c.closed != 0 {
   203:			unlock(&c.lock)
=> 204:			panic(plainError("send on closed channel"))
   205:		}
   206:
   207:		if sg := c.recvq.dequeue(); sg != nil {
   208:			// Found a waiting receiver. We pass the value we want to send
   209:			// directly to the receiver, bypassing the channel buffer (if any).
```
# 修正压缩包中的 ast_map_expr 文件夹中的 test，使 test 可以通过
```go
package mapexpr

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Eval : 判断 map 是否符合 bool 表达式
//	expr = `a > 1 && b < 0`
func Eval(m map[string]string, expr string) (bool, error) {
	fset := token.NewFileSet()
	exprAst, err := parser.ParseExpr(expr)
	if err != nil {
		return false, err
	}

	ast.Print(fset, exprAst)
	return judge(exprAst, m), nil
}

func isLeaf(bop ast.Node) bool {
	expr, ok := bop.(*ast.BinaryExpr)
	if !ok {
		return false
	}
	_, okL := expr.X.(*ast.Ident)
	_, okR := expr.Y.(*ast.BasicLit)
	if okL && okR {
		return true
	}
	return false
}

// dfs
func judge(bop ast.Node, m map[string]string) bool {
	if isLeaf(bop) {
		// do the leaf logic
		expr := bop.(*ast.BinaryExpr)
		x := expr.X.(*ast.Ident)
		y := expr.Y.(*ast.BasicLit)

		switch expr.Op {
		case token.GTR:
			return m[x.Name]>y.Value
		case token.LSS:
			return m[x.Name]<y.Value
		case token.LEQ:
			return m[x.Name]<=y.Value
		case token.GEQ:
			return m[x.Name]>=y.Value
		}

		// FIXME，修正这里的逻辑，使 test 能够正确通过
		//return m[x.Name] == y.Value
	}

	// not leaf
	// 那么一定是 binary expression
	expr, ok := bop.(*ast.BinaryExpr)
	if !ok {
		println("this cannot be true")
		return false
	}

	switch expr.Op {
	case token.LAND:
		return judge(expr.X, m) && judge(expr.Y, m)
	case token.LOR:
		return judge(expr.X, m) || judge(expr.Y, m)
	}

	println("unsupported operator")
	return false
}

type testCase struct {
	m      map[string]string
	expr   string
	result bool
}

func TestMapExpr(t *testing.T) {
	cases := []testCase{
		testCase{
			m: map[string]string{
				"invest": "20000", "posts": "144",
			},
			expr:   `invest > 10000 && posts > 100`,
			result: true,
		},
	}

	for _, cas := range cases {
		res, err := Eval(cas.m, cas.expr)
		assert.Nil(t, err)
		assert.Equal(t, res, cas.result)
	}
}
```