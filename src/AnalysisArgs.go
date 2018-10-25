package main

var gCommand sCommand //指令
type sCommand struct {
	InputPath     string
	OutputPath    string
	GoFileName    string
	CsFileName    string
	LayaFileName  string
	CocosFileName string
}

func init() {
	gCommand.InputPath = "./xlsx/"
	gCommand.OutputPath = "./res/"
	gCommand.GoFileName = ""
	gCommand.CsFileName = ""
	gCommand.LayaFileName = ""
	gCommand.CocosFileName = ""
}

//eg. xlsx2rof.exe -i ./xlsx -o ./rof -go ./rofserver -cs ./rofclient -ts ./rofclient
func analysisArgs(args []string) bool {
	var curFunc dealCommand
	for i := 0; i < len(args); i++ {
		parm := args[i]
		if parm[0] == '-' {
			if curFunc != nil {
				if curFunc("") == false {
					return false
				}
			}
			curFunc = checkCommand(parm)
			if curFunc == nil {
				logErr("illegal command:" + parm)
				return false
			}
		} else {
			if curFunc == nil {
				logErr("illegal command:" + parm)
				return false
			}
			if curFunc(parm) == false {
				return false
			}
			curFunc = nil
		}
	}

	if curFunc != nil {
		if curFunc("") == false {
			return false
		}
	}

	return true
}

func checkCommand(aCommand string) dealCommand {
	switch aCommand {
	case "-i":
		return dealCommandI
	case "-o":
		return dealCommandO
	case "-go":
		return dealCommandGO
	case "-cs":
		return dealCommandCS
	case "-laya":
		return dealCommandLaya
	case "-cocos":
		return dealCommandCocos
	case "-h":
		return dealCommandHelp
	}
	return nil
}

type dealCommand func(arg string) bool

func dealCommandI(arg string) bool {
	if len(arg) <= 0 {
		logErr("the command -i lack of arg.")
		return false
	}
	if arg[len(arg)-1] != '/' {
		gCommand.InputPath = arg + "/"
	} else {
		gCommand.InputPath = arg
	}

	return true
}

func dealCommandO(arg string) bool {
	if len(arg) <= 0 {
		logErr("the command -i lack of arg.")
		return false
	}
	if arg[len(arg)-1] != '/' {
		gCommand.OutputPath = arg + "/"
	} else {
		gCommand.OutputPath = arg
	}
	return true
}

func dealCommandGO(arg string) bool {
	if len(arg) <= 0 {
		logErr("the command -go lack of arg.")
		return false
	}
	if len(arg) < 3 || arg[len(arg)-3:] != ".go" {
		logErr("the arg of -go must be a file name. like ./rof/RofElements.go")
		return false
	}
	gCommand.GoFileName = arg
	return true
}

func dealCommandCS(arg string) bool {
	if len(arg) <= 0 {
		logErr("the command -cs lack of arg.")
		return false
	}
	if len(arg) < 3 || arg[len(arg)-3:] != ".cs" {
		logErr("the arg of -cs must be a file name. like ./rof/RofElements.cs")
		return false
	}
	gCommand.CsFileName = arg
	return true
}

func dealCommandLaya(arg string) bool {
	if len(arg) <= 0 {
		logErr("the command -laya lack of arg.")
		return false
	}
	if len(arg) < 3 || arg[len(arg)-3:] != ".ts" {
		logErr("the arg of -laya must be a file name. like ./rof/RofElements.ts")
		return false
	}
	gCommand.LayaFileName = arg
	return true
}

func dealCommandCocos(arg string) bool {
	if len(arg) <= 0 {
		logErr("the command -cocos lack of arg.")
		return false
	}
	if len(arg) < 3 || arg[len(arg)-3:] != ".ts" {
		logErr("the arg of -cocos must be a file name. like ./rof/RofElements.ts")
		return false
	}
	gCommand.CocosFileName = arg
	return true
}

func dealCommandHelp(arg string) bool {
	log("-i : [path] input .xlsx files floder path")
	log("-o : [path] output .bytes files floder path")
	log("-go : optional command. [path] go files floder path and file name. eg. ./rof/RofElements.go")
	log("-ts : optional command. [path] ts files floder path and file name. eg. ./rof/RofElements.ts")
	log("-cs : optional command. [path] cs files floder path and file name. eg. ./rof/RofElements.cs")
	return false
}
