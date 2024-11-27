package model

import (
	g "src/global"
)
func SelectConstant(name string) string {
	var value string
	g.Mysql.Table("constant").Where("name = ?", name).Pluck("value", &value)
	return value
}
