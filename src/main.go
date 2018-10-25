package main

import (
	"fmt"
	"os"
	"syscall"
)

var gGeneratorGo sGenerateGoFile
var gGeneratorCs sGenerateCsFile
var gGeneratorLaya sGenerateLayaFile
var gGeneratorCocos sGenerateCocosFile

var gKernel32 *syscall.LazyDLL
var gProc *syscall.LazyProc

func initConsoleColor() {
	gKernel32 = syscall.NewLazyDLL("kernel32.dll")
	gProc = gKernel32.NewProc("SetConsoleTextAttribute")
}

func clearConsoleColor() {
	gProc.Call(uintptr(syscall.Stdout), 7)
}

func main() {
	//初始化控制台输出颜色
	initConsoleColor()
	defer clearConsoleColor()

	//解析命令
	if analysisArgs(os.Args[1:]) == false {
		return
	}

	//初始化生成器
	if gGeneratorGo.init() == false {
		return
	}
	if gGeneratorCs.init() == false {
		return
	}
	if gGeneratorLaya.init() == false {
		return
	}
	if gGeneratorCocos.init() == false {
		return
	}

	//判断资源路径是否存在
	_, statErr := os.Stat(gCommand.InputPath)
	if os.IsNotExist(statErr) == true {
		logErr("can not find xlsx path.")
		return
	}

	//得到配置文件列表
	if pushFileInfo(gCommand.InputPath) == false {
		return
	}

	//生成rof二进制文件
	for i := 0; i < len(gFileList); i++ {
		if generateRof(gFileList[i]) == false {
			return
		}
	}

	//释放生成器
	gGeneratorGo.clear()
	gGeneratorCs.clear()
	gGeneratorLaya.clear()
	gGeneratorCocos.clear()

	log("[SUCCESS] Generate completely!")
}

func log(aContent string) {
	//green 10
	gProc.Call(uintptr(syscall.Stdout), 10)
	fmt.Println(aContent)
}

func logErr(aContent string) {
	//red 12
	gProc.Call(uintptr(syscall.Stdout), 12)
	fmt.Println("[ERROR] " + aContent)
}
