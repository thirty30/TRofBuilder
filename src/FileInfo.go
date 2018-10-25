package main

import (
	"os"
	"strings"
)

var gFileList = make([]*sFileInfo, 0)       //文件列表
var gCheckRepetition = make(map[string]int) //判断文件名是否重复

type sFileInfo struct {
	mPath string
	mName string
}

func (pOwn *sFileInfo) getXlsxPath() string {
	return pOwn.mPath
}

func (pOwn *sFileInfo) getRofPath() string {
	return strings.Replace(pOwn.mPath, gCommand.InputPath, gCommand.OutputPath, 1)
}

func (pOwn *sFileInfo) getXlsxFileName() string {
	return pOwn.mPath + pOwn.mName
}

func pushFileInfo(aDir string) bool {
	fileDir, _ := os.Open(aDir)
	fileList, _ := fileDir.Readdir(0)
	for _, v := range fileList {
		fileName := v.Name()
		if v.IsDir() == true {
			if pushFileInfo(aDir+fileName+"/") == false {
				return false
			}
		}
		if len(fileName) < 6 || fileName[len(fileName)-5:] != ".xlsx" {
			continue
		}
		if fileName[:2] == "~$" {
			continue
		}

		_, isExist := gCheckRepetition[fileName]
		if isExist == true {
			logErr("repetitive file: " + aDir + fileName)
			return false
		}
		gCheckRepetition[fileName] = 1
		pInfo := new(sFileInfo)
		pInfo.mPath = aDir
		pInfo.mName = fileName
		gFileList = append(gFileList, pInfo)
	}
	fileDir.Close()
	return true
}
