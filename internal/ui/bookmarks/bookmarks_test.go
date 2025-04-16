package bookmarks

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/stretchr/testify/assert"
	"slices"
	"testing"
)

func TestDistanceMap(t *testing.T) {
	selectedCommitId := "x"
	changeIds := []string{"a", "x", "b", "c", "d"}
	distanceMap := calcDistanceMap(selectedCommitId, changeIds)
	assert.Equal(t, 0, distanceMap["x"])
	assert.Equal(t, -1, distanceMap["a"])
	assert.Equal(t, 1, distanceMap["b"])
	assert.Equal(t, 2, distanceMap["c"])
	assert.Equal(t, 3, distanceMap["d"])
	assert.Equal(t, 0, distanceMap["nonexistent"])
}

func Test_Sorting_MoveCommands(t *testing.T) {
	items := []list.Item{
		item{name: "move feature", dist: 5, priority: moveCommand},
		item{name: "move main", dist: 1, priority: moveCommand},
		item{name: "move very-old-feature", dist: 15, priority: moveCommand},
		item{name: "move backwards", dist: -2, priority: moveCommand},
	}
	slices.SortFunc(items, itemSorter)
	var sorted []string
	for _, i := range items {
		sorted = append(sorted, i.(item).name)
	}
	assert.Equal(t, []string{"move main", "move feature", "move very-old-feature", "move backwards"}, sorted)
}
