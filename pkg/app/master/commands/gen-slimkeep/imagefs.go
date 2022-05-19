package genslimkeep

import (
	"fmt"
	"io/fs"
)

func newImageFS(imageName string) (fs.FS, error) {
	return nil, fmt.Errorf("image fs walker not implemented yet")
}
