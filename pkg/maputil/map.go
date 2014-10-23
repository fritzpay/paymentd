package maputil

import (
	"fmt"
	"io"
	"sort"
)

func WriteSortedMap(w io.Writer, m map[string]string) error {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	var err error
	for _, k := range keys {
		_, err = w.Write([]byte(k))
		if err != nil {
			return fmt.Errorf("buffer error: %v", err)
		}
		_, err = w.Write([]byte(m[k]))
		if err != nil {
			return fmt.Errorf("buffer error: %v", err)
		}
	}
	return nil
}
