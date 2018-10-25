package main

import (
	"fmt"
	"os"
)

type sGenerateLayaFile struct {
	mFile            *os.File
	mMapTypeRelation map[int]string
}

func (pOwn *sGenerateLayaFile) init() bool {
	if len(gCommand.LayaFileName) <= 0 {
		return true
	}
	var err error
	pOwn.mFile, err = os.Create(gCommand.LayaFileName)
	if err != nil {
		logErr("can not create ts file")
		return false
	}
	pOwn.mMapTypeRelation = make(map[int]string)
	pOwn.mMapTypeRelation[1] = "number"
	pOwn.mMapTypeRelation[2] = "string"
	pOwn.mMapTypeRelation[3] = "number"
	pOwn.mMapTypeRelation[4] = "number"
	pOwn.mMapTypeRelation[5] = "string"
	return true
}

func (pOwn *sGenerateLayaFile) generate(aTableName string, aList []*sColInfo) {
	if len(gCommand.LayaFileName) <= 0 {
		return
	}
	strRowClassName := fmt.Sprintf("Rof%sRow", aTableName)
	//类
	strContent := fmt.Sprintf("class %s implements IRofBase\n{\n", strRowClassName)
	for i := 0; i < len(aList); i++ {
		cell := aList[i]
		strType := pOwn.mMapTypeRelation[cell.Type]
		strInitValue := ""
		if strType == "number" {
			strInitValue = "0"
		} else {
			strInitValue = "\"\""
		}
		strContent += fmt.Sprintf("private m%s : %s = %s;\n", cell.Name, strType, strInitValue)
	}

	//ReadBody
	strContent += "public ReadBody(rData : Laya.Byte)\n{\n"
	for i := 0; i < len(aList); i++ {
		cell := aList[i]
		switch cell.Type {
		case 1: //int32
			{
				strContent += fmt.Sprintf("this.m%s = rData.getInt32();\n", cell.Name)
			}
			break
		case 2: //int64
			{
				strContent += fmt.Sprintf("for (let i = 0; i < 8; i++) {let temp = rData.getUint8().toString(16);if (temp.length < 2){ temp = \"0\" + temp;}this.m%s += temp;}\n", cell.Name)
			}
			break
		case 3: //float32
			{
				strContent += fmt.Sprintf("this.m%s = rData.getFloat32();\n", cell.Name)
			}
			break
		case 4: //float64
			{
				strContent += fmt.Sprintf("this.m%s = rData.getFloat64();\n", cell.Name)
			}
			break
		case 5: //string
			{
				strContent += fmt.Sprintf("let n%sLen = rData.getInt32();this.m%s= rData.getUTFBytes(n%sLen)\n", cell.Name, cell.Name, cell.Name)
			}
			break
		}
	}
	strContent += "}\n"
	//方法
	for i := 0; i < len(aList); i++ {
		cell := aList[i]
		strType := pOwn.mMapTypeRelation[cell.Type]
		if cell.IsLan == true {
			strContent += fmt.Sprintf("public Get%s() : string { return GetMultiLanguage(this.m%s) }\n", cell.Name, cell.Name)
		} else {
			strContent += fmt.Sprintf("public Get%s() : %s { return this.m%s }\n", cell.Name, strType, cell.Name)
		}

	}
	strContent += "}\n"
	pOwn.mFile.WriteString(strContent)
}

func (pOwn *sGenerateLayaFile) clear() {
	if len(gCommand.LayaFileName) <= 0 {
		return
	}
	pOwn.mFile.Close()
}
