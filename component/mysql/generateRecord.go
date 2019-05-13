package mysql

import (
	"bigbigTravel/component/helper"
	"strings"
	"fmt"
	"os"
)

func init() {
	RegisterRecord("mysql_table_scheme", new(TableScheme))
}

type TableScheme struct {
	Field 		string  `column:"field"`
	Type 		string  `column:"type"`
	Null		string	`column:"null"`
	Key 		string	`column:"key"`
	Default		interface{}	`column:"default"`
	Extra		string	`column:"extra"`
}

func (t *TableScheme) Name() string {
	return "mysql_table_scheme"
}

type CodeText struct {
	Dir			string
	FileName	string
	PackageName	string
	Lines		[]string
}

type fieldLine struct {
	k			string
	t			string
	tags		map[string]string
}

func NewCodeText(dir , filename string) *CodeText {
	ct := &CodeText{
		Dir: dir,
		FileName: filename,
		Lines:make([]string,0, 50),
	}
	strs := strings.Split(dir, "/")
	ct.PackageName = strs[len(strs)-1]
	return ct
}

func (ct *CodeText)AddLine(string string)  {
	ct.Lines = append(ct.Lines, string)
}

func generateTableSchemeCode(dir, table string) *CodeText {
	db := GetInstance(false)
	rl := db.Query("mysql_table_scheme", "desc `"+table+"`")
	structName := helper.StrUnderline2CamelCase(table, true)
	ct := NewCodeText(dir, table+".go")
	ct.AddLine(
		`package `+ct.PackageName +`
import (
	"github.com/golib/mysql"
)

const RecordName`+structName+` = "`+table+`"

func init()  {
	var r = &`+structName+`{}
	mysql.RegisterRecord(r.Name(), r)
}

type `+structName+` struct{`)
	for i := 0; i<rl.Len(); i++ {
		field := rl.GetAt(i).(*TableScheme)
		fline := &fieldLine{
			k:helper.StrUnderline2CamelCase(field.Field, true),
			t:mysqlTypeToCodeType(field.Type),
			tags:make(map[string]string),
		}

		fline.tags["json"] = helper.StrUnderline2CamelCase(field.Field, false)
		fline.tags["form"] = helper.StrUnderline2CamelCase(field.Field, false)
		fline.tags["column"] = field.Field
		if field.Key == "PRI" {
			fline.tags["primary"] = "true"
			fline.tags["modify"] = "false"
		}else {
			switch field.Field {
			case "create_time", "update_time":
				fline.tags["modify"] = "false"
			default:
				fline.tags["modify"] = "true"
			}
		}

		line := `	`+fline.k+`	`+fline.t
		tagStrSlice := make([]string, 0, 3)
		for k,v := range fline.tags {
			tagStrSlice = append(tagStrSlice, k+":\""+v+"\"")
		}
		line += "	`"+strings.Join(tagStrSlice, " ")+"`"
		ct.AddLine(line)
	}
	ct.AddLine("}")
	ct.AddLine(
		`
func (r *`+helper.StrUnderline2CamelCase(table, true)+`) Name() string {
	return RecordName`+structName+`
}
`)
	return ct
}

func GenerateSignalRecord(dir string, name string)  {
	fmt.Println("Generating Record:",name)
	codeText := generateTableSchemeCode(dir, name)
	_,err := os.Stat(dir)
	if err != nil {
		err = os.Mkdir(dir, 0755)
		if err != nil{
			fmt.Println(err.Error())
			os.Exit(0)
		}
	}
	filePath := strings.TrimRight(dir, "/")+"/"+codeText.FileName
	f,err := os.Create(filePath)
	checkErr(err)
	_,err = f.WriteString(strings.Join(codeText.Lines,"\n"))
	checkErr(err)
	fmt.Println("Success:", filePath)
}

func GenerateAllRecord(dir string)  {
	db := GetInstance(false)
	rows, err := db.Db().Query("show tables")
	checkErr(err)
	var tableName string
	for rows.Next() {
		rows.Scan(&tableName)
		GenerateSignalRecord(dir, tableName)
	}
}

func mysqlTypeToCodeType(mysqlType string) string {
	switch  {
	case strings.Contains(mysqlType,"int"):
		return "int"
	case strings.Contains(mysqlType,"char") ,
		strings.Contains(mysqlType,"text"),
		strings.Contains(mysqlType,"timestamp"):
		return "string"
	case strings.Contains(mysqlType,"float") :
		return "float"
	default:
		return "interface{}"
	}
}

func checkErr(err error)  {
	if err != nil {
		fmt.Println("Failed:", err.Error())
		os.Exit(0)
	}
}