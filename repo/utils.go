package repo

import (
	"fmt"
	"strings"

	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/aarondl/sqlboiler/v4/types"
)

// ConvertInt64ArrayToIntSlice converts a types.Int64Array to a slice of int.
func ConvertInt64ArrayToIntSlice(arr types.Int64Array) []int {
	intSlice := make([]int, len(arr))
	for i, v := range arr {
		intSlice[i] = int(v)
	}
	return intSlice
}

func JoinOn(rightTable, leftName, rightColumn string, join func(string, ...interface{}) qm.QueryMod) qm.QueryMod {
	return join(fmt.Sprintf("%s ON %s = %s", rightTable, leftName, rightColumn))
}

func AndIn(column string, values []string) string {
	return fmt.Sprintf("AND %s IN ('%s')", column, strings.Join(values, "','"))
}
