package translator

import (
	"strconv"

	"github.com/ebukreev/go-z3/z3"
)

type stringTable struct {
	ids    map[string]z3.Int
	values map[string]string
	nextID int64
}

func newStringTable(ctx *z3.Context) stringTable {
	return stringTable{
		ids:    map[string]z3.Int{"": ctx.FromInt(0, ctx.IntSort()).(z3.Int)},
		values: map[string]string{"0": ""},
		nextID: 1,
	}
}

func (table *stringTable) intern(ctx *z3.Context, value string) z3.Int {
	if id, ok := table.ids[value]; ok {
		return id
	}
	for {
		next := table.nextID
		table.nextID++
		key := strconv.FormatInt(next, 10)
		if _, used := table.values[key]; used {
			continue
		}
		id := ctx.FromInt(next, ctx.IntSort()).(z3.Int)
		table.ids[value], table.values[key] = id, value
		return id
	}
}

func (table *stringTable) resolve(id z3.Int) (string, bool) {
	number, ok := id.AsBigInt()
	if !ok {
		return "", false
	}
	key := number.String()
	if value, ok := table.values[key]; ok {
		return value, true
	}

	value := "$string." + key
	for {
		if _, used := table.ids[value]; !used {
			table.ids[value], table.values[key] = id, value
			return value, true
		}
		value += "_"
	}
}
