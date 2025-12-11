package utils

import (
	"github.com/jinzhu/copier"
)

func Copy(dest interface{}, src interface{}) error {
	//data, _ := json.Marshal(src)
	//_ = json.Unmarshal(data, dest)
	return copier.Copy(dest, src)
}
