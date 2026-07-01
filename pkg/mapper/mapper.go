package mapper

import (
	"fmt"

	"github.com/jinzhu/copier"
)

func MapOne[S any, D any](src S) (*D, error) {
	var dest D

	if err := copier.Copy(&dest, src); err != nil {
		return nil, fmt.Errorf("mapper.MapOne: %w", err)
	}

	return &dest, nil
}

func MapList[S any, D any](src []S) ([]D, error) {
	if len(src) == 0 {
		return []D{}, nil
	}

	dest := make([]D, len(src))

	if err := copier.Copy(&dest, src); err != nil {
		return nil, fmt.Errorf("mapper.MapList: %w", err)
	}

	return dest, nil
}

func Copy(dest, src any, ignoreEmpty bool) error {
	return copier.CopyWithOption(dest, src, copier.Option{IgnoreEmpty: ignoreEmpty})
}
