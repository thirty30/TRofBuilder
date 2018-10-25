package main

import (
	"fmt"
	"os"
)

type sGenerateCsFile struct {
	mFile            *os.File
	mMapTypeRelation map[int]string
}

func (pOwn *sGenerateCsFile) init() bool {
	if len(gCommand.CsFileName) <= 0 {
		return true
	}
	var err error
	pOwn.mFile, err = os.Create(gCommand.CsFileName)
	if err != nil {
		logErr("can not create cs file")
		return false
	}
	pOwn.mMapTypeRelation = make(map[int]string)
	pOwn.mMapTypeRelation[1] = "int"
	pOwn.mMapTypeRelation[2] = "long"
	pOwn.mMapTypeRelation[3] = "float"
	pOwn.mMapTypeRelation[4] = "double"
	pOwn.mMapTypeRelation[5] = "string"

	pOwn.mFile.WriteString("using System;\nusing System.Text;\n")
	pOwn.mFile.WriteString("namespace Game\n{\n")
	return true
}

func (pOwn *sGenerateCsFile) generate(aTableName string, aList []*sColInfo) {
	if len(gCommand.CsFileName) <= 0 {
		return
	}
	strRowClassName := fmt.Sprintf("Rof%sRow", aTableName)
	//类
	strContent := fmt.Sprintf("public class %s : IRofBase\n{\n", strRowClassName)
	for i := 0; i < len(aList); i++ {
		cell := aList[i]
		strType := pOwn.mMapTypeRelation[cell.Type]
		if cell.IsLan == true {
			strContent += fmt.Sprintf("public int %sLanID { get; private set; }\n", cell.Name)
			strContent += fmt.Sprintf("public string %s { get { return RofManager.Instance.GetMultiLanguage(this.%sLanID); } }\n", cell.Name, cell.Name)
		} else {
			strContent += fmt.Sprintf("public %s %s { get; private set; }\n", strType, cell.Name)
		}
	}

	//ReadBody
	strContent += "public int ReadBody(byte[] rData, int nOffset)\n{\n"
	for i := 0; i < len(aList); i++ {
		cell := aList[i]
		switch cell.Type {
		case 1: //int32
			{
				strContent += "if (BitConverter.IsLittleEndian){Array.Reverse(rData, nOffset, 4);}\n"
				if cell.IsLan == true {
					strContent += fmt.Sprintf("%sLanID = (int)BitConverter.ToUInt32(rData, nOffset); nOffset += 4;\n", cell.Name)
				} else {
					strContent += fmt.Sprintf("%s = (int)BitConverter.ToUInt32(rData, nOffset); nOffset += 4;\n", cell.Name)
				}
			}
			break
		case 2: //int64
			{
				strContent += "if (BitConverter.IsLittleEndian){Array.Reverse(rData, nOffset, 8);}\n"
				strContent += fmt.Sprintf("%s = (long)BitConverter.ToUInt64(rData, nOffset); nOffset += 8;\n", cell.Name)
			}
			break
		case 3: //float32
			{
				strContent += "if (BitConverter.IsLittleEndian){Array.Reverse(rData, nOffset, 4);}\n"
				strContent += fmt.Sprintf("%s = BitConverter.ToSingle(rData, nOffset); nOffset += 4;\n", cell.Name)
			}
			break
		case 4: //float64
			{
				strContent += "if (BitConverter.IsLittleEndian){Array.Reverse(rData, nOffset, 8);}\n"
				strContent += fmt.Sprintf("%s = BitConverter.ToDouble(rData, nOffset); nOffset += 8;\n", cell.Name)
			}
			break
		case 5: //string
			{
				strContent += "if (BitConverter.IsLittleEndian){Array.Reverse(rData, nOffset, 4);}\n"
				strContent += fmt.Sprintf("int n%sLen = (int)BitConverter.ToUInt32(rData, nOffset); nOffset += 4;\n", cell.Name)
				strContent += fmt.Sprintf("%s = Encoding.UTF8.GetString(rData, nOffset, n%sLen); nOffset += n%sLen;\n", cell.Name, cell.Name, cell.Name)
			}
			break
		}
	}
	strContent += "return nOffset;\n}\n}\n"
	pOwn.mFile.WriteString(strContent)
}

func (pOwn *sGenerateCsFile) clear() {
	if len(gCommand.CsFileName) <= 0 {
		return
	}
	pOwn.mFile.WriteString("}\n")
	pOwn.mFile.Close()
}
