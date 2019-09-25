package initial

import (
	"fmt"
	"github.com/go-ini/ini"
	"os"
)

type InitS struct {
	File  string
	Sect  string
}

func (i *InitS) InitF(s interface{}) error {
	// загрузить файл настроек -------------------------------------------
	cfg, err := ini.Load(i.File)
	if err != nil {
		os.Exit(1)
		return fmt.Errorf("fail to read file: %v", err)
	}

	// считать переменные из раздела section -------------------------------
	if err = cfg.Section(i.Sect).MapTo(s); err != nil {
		return fmt.Errorf("failed to maping ini: %v", err)
	}

	return nil
}
