package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"strconv"
	xlsx "xlsx2rof/libxlsx"
)

var gMapTypeRelation = make(map[string]int)     //类型映射
var gTableNameRepetition = make(map[string]int) //判断Rof表名是否重复

func init() {
	gMapTypeRelation["int32"] = 1
	gMapTypeRelation["int64"] = 2
	gMapTypeRelation["float32"] = 3
	gMapTypeRelation["float64"] = 4
	gMapTypeRelation["string"] = 5
}

const (
	cUnexportableColor  = "FF808080"
	cMultiLanguageColor = "FF00B0F0"
	cRofMaxSize         = 1024 * 1024 * 4
)

type sColInfo struct {
	Index int
	Name  string
	Type  int
	IsLan bool
}

func generateRof(aFileInfo *sFileInfo) bool {
	strXlsxFileName := aFileInfo.getXlsxFileName()
	pFile, err := xlsx.OpenFile(strXlsxFileName)
	if err != nil {
		logErr("can not open file: " + strXlsxFileName)
		return false
	}

	strRofPath := aFileInfo.getRofPath()
	os.MkdirAll(strRofPath, os.ModeDir)

	pSheet := pFile.Sheets[0]
	strRofTableName := pSheet.Name

	//判断表名是否重复
	_, exist := gTableNameRepetition[strRofTableName]
	if exist == true {
		logErr("repetitive table name: " + strRofTableName + ", in file: " + strXlsxFileName)
		return false
	}
	gTableNameRepetition[strRofTableName] = 1

	strOutFileName := aFileInfo.getRofPath() + "Rof" + strRofTableName + ".bytes"
	pOutFile, err := os.Create(strOutFileName)
	if err != nil {
		logErr("can not create file: " + strOutFileName)
		return false
	}
	defer pOutFile.Close()

	pFlagRow := pSheet.Rows[0]
	pNameRow := pSheet.Rows[1]
	pTypeRow := pSheet.Rows[2]
	mapRepetitionID := make(map[int32]int8)

	if pSheet.MaxCol != len(pFlagRow.Cells) || pSheet.MaxCol != len(pNameRow.Cells) || pSheet.MaxCol != len(pTypeRow.Cells) {
		logErr("column num is not match max column num  in file: " + strXlsxFileName)
		return false
	}

	//得到真实的列
	colList := make([]*sColInfo, 0)
	for i := 0; i < pSheet.MaxCol; i++ {
		pFlagCell := pFlagRow.Cells[i]
		pNameCell := pNameRow.Cells[i]
		pTypeCell := pTypeRow.Cells[i]
		strFlagColor := pFlagCell.GetStyle().Fill.FgColor
		//判断第一列规范
		if i == 0 {
			//必须可导出
			if strFlagColor == cUnexportableColor {
				logErr("the first column can not be marked as unexportable in file: " + strXlsxFileName)
				return false
			}
			//不能是多语言列
			if strFlagColor == cMultiLanguageColor {
				logErr("the first column can not be marked as  multi-language in file: " + strXlsxFileName)
				return false
			}
			//判断列名是否是ID
			if pNameCell.String() != "ID" {
				logErr("the first column name is not ID in file: " + strXlsxFileName)
				return false
			}
			//判断列类型是否是int32
			if pTypeCell.String() != "int32" {
				logErr("the first column type is not int32 in file: " + strXlsxFileName)
				return false
			}
		}

		//判断列名存在
		if len(pNameCell.String()) <= 0 {
			logErr("the column name is empty in file: " + strXlsxFileName)
			return false
		}
		//判断是不导出列
		if strFlagColor == cUnexportableColor {
			continue
		}
		//判断列名重复
		if checkNameRepetitive(colList, pNameCell.String()) == false {
			logErr("the column name is repetitive in file: " + strXlsxFileName)
			return false
		}
		//判断类型存在
		nType, exist := gMapTypeRelation[pTypeCell.String()]
		if exist == false {
			logErr("the column type is not existed in file: " + strXlsxFileName)
			return false
		}

		pColInfo := new(sColInfo)
		pColInfo.Index = i
		pColInfo.Name = pNameCell.String()
		pColInfo.Type = nType
		pColInfo.IsLan = false
		if strFlagColor == cMultiLanguageColor {
			pColInfo.IsLan = true
		}
		colList = append(colList, pColInfo)
		//判断多语言列的类型必须是int32
		if pColInfo.IsLan == true {
			if pColInfo.Type != 1 {
				logErr("the language column type is not int32 in file: " + strXlsxFileName)
				return false
			}
		}
	}

	//填写头
	nRealRowNum := pSheet.MaxRow - 3
	nRealColNum := len(colList)
	pBuffer := make([]byte, cRofMaxSize)
	nOffset := 64

	binary.BigEndian.PutUint32(pBuffer[nOffset:], uint32(nRealRowNum))
	nOffset += 4
	binary.BigEndian.PutUint32(pBuffer[nOffset:], uint32(nRealColNum))
	nOffset += 4

	//填写列属性
	for i := 0; i < len(colList); i++ {
		pInfo := colList[i]
		if len(pInfo.Name) > 255 {
			logErr("the length of column name is more than 255 in file: " + strXlsxFileName)
			return false
		}
		nNameLen := int8(len(pInfo.Name))
		pBuffer[nOffset] = byte(nNameLen)
		nOffset++
		copy(pBuffer[nOffset:], []byte(pInfo.Name))
		nOffset += len(pInfo.Name)
		binary.BigEndian.PutUint16(pBuffer[nOffset:], uint16(pInfo.Type))
		nOffset += 2
	}

	//填写内容
	for i := 3; i < pSheet.MaxRow; i++ {
		pRow := pSheet.Rows[i]
		for j := 0; j < len(colList); j++ {
			realCell := colList[j]
			cell := pRow.Cells[realCell.Index]
			//判断不能为空
			if len(cell.String()) <= 0 {
				strErr := fmt.Sprintf("cell is empty in file: %s, row num:%d, cell name:%s", strXlsxFileName, i+1, realCell.Name)
				logErr(strErr)
				return false
			}
			//检查ID不能重复
			if realCell.Index == 0 {
				nValue, err := strconv.ParseInt(cell.String(), 10, 32)
				if err != nil {
					strErr := fmt.Sprintf("the value type is not int32 in file: %s, row num:%d, cell name:%s", strXlsxFileName, i+1, realCell.Name)
					logErr(strErr)
					return false
				}
				_, ok := mapRepetitionID[int32(nValue)]
				if ok == true {
					strErr := fmt.Sprintf("ID repetition in file: %s, row num:%d, cell name:%s", strXlsxFileName, i+1, realCell.Name)
					logErr(strErr)
					return false
				}
				mapRepetitionID[int32(nValue)] = 1
			}

			if realCell.Type == 1 { //int32
				nValue, err := strconv.ParseInt(cell.String(), 10, 32)
				if err != nil {
					strErr := fmt.Sprintf("the value type is not int32 in file: %s, row num:%d, cell name:%s", strXlsxFileName, i+1, realCell.Name)
					logErr(strErr)
					return false
				}
				binary.BigEndian.PutUint32(pBuffer[nOffset:], uint32(nValue))
				nOffset += 4
				if nOffset >= cRofMaxSize {
					logErr(strOutFileName + " is more than 4M")
					return false
				}
			}
			if realCell.Type == 2 { //int64
				nValue, err := strconv.ParseInt(cell.String(), 10, 64)
				if err != nil {
					strErr := fmt.Sprintf("the value type is not int64 in file: %s, row num:%d, cell name:%s", strXlsxFileName, i+1, realCell.Name)
					logErr(strErr)
					return false
				}
				binary.BigEndian.PutUint64(pBuffer[nOffset:], uint64(nValue))
				nOffset += 8
				if nOffset >= cRofMaxSize {
					logErr(strOutFileName + " is more than 4M")
					return false
				}
			}
			if realCell.Type == 3 { //float32
				fValue, err := strconv.ParseFloat(cell.String(), 32)
				if err != nil {
					strErr := fmt.Sprintf("the value type is not float32 in file: %s, row num:%d, cell name:%s", strXlsxFileName, i+1, realCell.Name)
					logErr(strErr)
					return false
				}
				bits := math.Float32bits(float32(fValue))
				binary.BigEndian.PutUint32(pBuffer[nOffset:], bits)
				nOffset += 4
				if nOffset >= cRofMaxSize {
					logErr(strOutFileName + " is more than 4M")
					return false
				}
			}
			if realCell.Type == 4 { //float64
				fValue, err := strconv.ParseFloat(cell.String(), 64)
				if err != nil {
					strErr := fmt.Sprintf("the value type is not float64 in file: %s, row num:%d, cell name:%s", strXlsxFileName, i+1, realCell.Name)
					logErr(strErr)
					return false
				}
				bits := math.Float64bits(fValue)
				binary.BigEndian.PutUint64(pBuffer[nOffset:], bits)
				nOffset += 8
				if nOffset >= cRofMaxSize {
					logErr(strOutFileName + " is more than 4M")
					return false
				}
			}
			if realCell.Type == 5 { //string
				strValue := cell.String()
				nLen := len(strValue)
				binary.BigEndian.PutUint32(pBuffer[nOffset:], uint32(nLen))
				nOffset += 4
				if nOffset >= cRofMaxSize {
					logErr(strOutFileName + " is more than 4M")
					return false
				}
				copy(pBuffer[nOffset:], []byte(strValue))
				nOffset += nLen
				if nOffset >= cRofMaxSize {
					logErr(strOutFileName + " is more than 4M")
					return false
				}
			}
		}
	}

	_, err = pOutFile.Write(pBuffer[0:nOffset])
	if err != nil {
		logErr(err.Error())
		return false
	}

	//生成代码文件

	gGeneratorGo.generate(strRofTableName, colList)
	gGeneratorCs.generate(strRofTableName, colList)
	gGeneratorLaya.generate(strRofTableName, colList)
	gGeneratorCocos.generate(strRofTableName, colList)

	return true
}

func checkNameRepetitive(aList []*sColInfo, aName string) bool {
	for i := 0; i < len(aList); i++ {
		if aList[i].Name == aName {
			return false
		}
	}
	return true
}
