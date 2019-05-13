package orm_gen

import "bigbigTravel/component/mysql"

const (
	GenerateSingle = iota
	GenerateGlobal
)

func GenerateRecords(path string, typ int , records ...string){
	if path == "" {
		return
	}
	switch typ {
	case GenerateSingle:
		for _, v := range records {
			mysql.GenerateSignalRecord(path, v)
		}
	case GenerateGlobal:
		mysql.GenerateAllRecord(path)
	}
}
