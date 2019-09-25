package base

import (
	"../initial"
	"../models"
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type DB struct {
	Db *gorm.DB
}

type Initer interface {
	conn(string, string) error
}

func (d *DB) conn(iniFile string, section string) error {

	sql := models.IniMySQL{}

	i := initial.InitS{ File:iniFile, Sect:section}

	if err := i.InitF(&sql); err != nil {
		return fmt.Errorf("fail to read ini: %v", err)
	} else {
		// если файл настроек считан - подключить к базе данных --------------
		d.Db, err = gorm.Open("mysql", sql.User+":"+sql.Pass+"@tcp("+sql.Host+":"+sql.Port+")/"+sql.Database)
		if err != nil {
			return fmt.Errorf("failed to connect databaze: %v", err)
		}
	}

	return nil
}

func Connect(iniFile string, section string, in Initer) error {
	err := in.conn(iniFile, section)
	return err
}
