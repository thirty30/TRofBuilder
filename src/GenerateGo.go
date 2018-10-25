package main

import (
	"fmt"
	"os"
)

type sGenerateGoFile struct {
	mFile            *os.File
	mMapTypeRelation map[int]string
}

func (pOwn *sGenerateGoFile) init() bool {
	if len(gCommand.GoFileName) <= 0 {
		return true
	}
	var err error
	pOwn.mFile, err = os.Create(gCommand.GoFileName)
	if err != nil {
		logErr("can not create go file")
		return false
	}
	pOwn.mMapTypeRelation = make(map[int]string)
	pOwn.mMapTypeRelation[1] = "int32"
	pOwn.mMapTypeRelation[2] = "int64"
	pOwn.mMapTypeRelation[3] = "float32"
	pOwn.mMapTypeRelation[4] = "float64"
	pOwn.mMapTypeRelation[5] = "string"

	pOwn.mFile.WriteString("package rof\n")
	pOwn.mFile.WriteString("import (\n\"encoding/binary\"\n\"math\"\n)\n")
	return true
}

func (pOwn *sGenerateGoFile) generate(aTableName string, aList []*sColInfo) {
	if len(gCommand.GoFileName) <= 0 {
		return
	}
	strRowClassName := fmt.Sprintf("sRof%sRow", aTableName)
	strTableClassName := fmt.Sprintf("sRof%sTable", aTableName)
	//row 结构体
	strContent := fmt.Sprintf("type %s struct {\n", strRowClassName)
	for i := 0; i < len(aList); i++ {
		cell := aList[i]
		strType := pOwn.mMapTypeRelation[cell.Type]
		strContent += fmt.Sprintf("m%s %s\n", cell.Name, strType)
	}
	strContent += "}\n"

	//ReadBody
	strContent += fmt.Sprintf("func (pOwn *%s) readBody(aBuffer []byte) int32 {\nvar nOffset int32\n", strRowClassName)
	for i := 0; i < len(aList); i++ {
		cell := aList[i]
		switch cell.Type {
		case 1: //int32
			{
				strContent += fmt.Sprintf("pOwn.m%s = int32(binary.BigEndian.Uint32(aBuffer[nOffset:]))\nnOffset+=4\n", cell.Name)
			}
			break
		case 2: //int64
			{
				strContent += fmt.Sprintf("pOwn.m%s = int64(binary.BigEndian.Uint64(aBuffer[nOffset:]))\nnOffset+=8\n", cell.Name)
			}
			break
		case 3: //float32
			{
				strContent += fmt.Sprintf("pOwn.m%s = math.Float32frombits(binary.BigEndian.Uint32(aBuffer[nOffset:]))\nnOffset+=4\n", cell.Name)
			}
			break
		case 4: //float64
			{
				strContent += fmt.Sprintf("pOwn.m%s = math.Float64frombits(binary.BigEndian.Uint64(aBuffer[nOffset:]))\nnOffset+=8\n", cell.Name)
			}
			break
		case 5: //string
			{
				strContent += fmt.Sprintf("n%sLen := int32(binary.BigEndian.Uint32(aBuffer[nOffset:]))\nnOffset+=4\n", cell.Name)
				strContent += fmt.Sprintf("pOwn.m%s = string(aBuffer[nOffset:nOffset+n%sLen])\nnOffset+=n%sLen\n", cell.Name, cell.Name, cell.Name)
			}
			break
		}
	}
	strContent += "return nOffset\n}\n"

	//函数
	for i := 0; i < len(aList); i++ {
		cell := aList[i]
		strType := pOwn.mMapTypeRelation[cell.Type]
		strContent += fmt.Sprintf("func (pOwn *%s) Get%s() %s { return pOwn.m%s } \n", strRowClassName, cell.Name, strType, cell.Name)
	}

	//table 结构体
	strContent += fmt.Sprintf("type %s struct { \nmRowNum int32\nmColNum int32\nmIDMap  map[int32]*%s\nmRowMap map[int32]int32\n}\n", strTableClassName, strRowClassName)
	strContent += fmt.Sprintf("func (pOwn *%s) newTypeObj() iRofRow {return new(%s)}\n", strTableClassName, strRowClassName)
	strContent += fmt.Sprintf("func (pOwn *%s) setRowNum(aNum int32) {pOwn.mRowNum = aNum}\n", strTableClassName)
	strContent += fmt.Sprintf("func (pOwn *%s) setColNum(aNum int32) {pOwn.mColNum = aNum}\n", strTableClassName)
	strContent += fmt.Sprintf("func (pOwn *%s) setIDMap(aKey int32, aValue iRofRow) {pOwn.mIDMap[aKey] = aValue.(*%s)}\n", strTableClassName, strRowClassName)
	strContent += fmt.Sprintf("func (pOwn *%s) setRowMap(aKey int32, aValue int32) {pOwn.mRowMap[aKey] = aValue}\n", strTableClassName)
	strContent += fmt.Sprintf("func (pOwn *%s) init(aPath string) bool {\n", strTableClassName)
	strContent += fmt.Sprintf("pOwn.mIDMap = make(map[int32]*%s)\n", strRowClassName)
	strContent += fmt.Sprintf("pOwn.mRowMap = make(map[int32]int32)\nreturn analysisRof(aPath, pOwn)\n}\n")
	strContent += fmt.Sprintf("func (pOwn *%s) GetDataByID(aID int32) *%s {return pOwn.mIDMap[aID]}\n", strTableClassName, strRowClassName)
	strContent += fmt.Sprintf("func (pOwn *%s) GetDataByRow(aIndex int32) *%s {\n", strTableClassName, strRowClassName)
	strContent += fmt.Sprintf("nID, ok := pOwn.mRowMap[aIndex]\nif ok == false {return nil}\nreturn pOwn.mIDMap[nID]\n}\n")
	strContent += fmt.Sprintf("func (pOwn *%s) GetRows() int32 {return pOwn.mRowNum}\n", strTableClassName)
	strContent += fmt.Sprintf("func (pOwn *%s) GetCols() int32 {return pOwn.mColNum}\n", strTableClassName)
	pOwn.mFile.WriteString(strContent)
}

func (pOwn *sGenerateGoFile) clear() {
	if len(gCommand.GoFileName) <= 0 {
		return
	}
	pOwn.mFile.Close()
}
