package context

import (
	"testing"
	"time"

	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/jesseduffield/lazygit/pkg/config"
	"github.com/jesseduffield/lazygit/pkg/gocui"
	"github.com/jesseduffield/lazygit/pkg/gui/style"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
	"github.com/jesseduffield/lazygit/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddCommitDropIndicator(t *testing.T) {
	pendingHeader := &NonModelItem{Index: 0, Content: "pending"}
	commitsHeader := &NonModelItem{Index: 3, Content: "commits"}
	indicator := &commitDropIndicator{insertionIndex: 3}
	spinnerConfig := config.SpinnerConfig{Frames: []string{"one", "two"}, Rate: 100}

	items := addCommitDropIndicator(
		[]*NonModelItem{pendingHeader}, indicator, "drop here", "moving commits here", spinnerConfig, time.UnixMilli(0),
		6,
	)
	items = append(items, commitsHeader)

	assert.Equal(t, []*NonModelItem{
		pendingHeader,
		{
			Index:   3,
			Content: style.FgCyan.SetBold().Sprint("━━━━━━ drop here ━━━━━━"),
			Column:  6,
		},
		commitsHeader,
	}, items)
	assert.Equal(t, 6, modelIndexToViewIndex(4, items, 3))
	assert.Equal(t, 3, viewIndexToModelIndex(4, items, 4))
}

func TestAddMovingCommitsIndicator(t *testing.T) {
	items := addCommitDropIndicator(
		nil,
		&commitDropIndicator{insertionIndex: 2, moving: true},
		"drop here",
		"moving commits here",
		config.SpinnerConfig{Frames: []string{"one", "two"}, Rate: 100},
		time.UnixMilli(100),
		6,
	)

	assert.Equal(t, []*NonModelItem{
		{
			Index:   2,
			Content: style.FgCyan.SetBold().Sprint("━━━━━━ moving commits here two ━━━━━━"),
			Column:  6,
		},
	}, items)
}

func TestAddCommitDropIndicatorUsesConfiguredColumn(t *testing.T) {
	items := addCommitDropIndicator(
		nil,
		&commitDropIndicator{insertionIndex: 2},
		"drop here",
		"moving commits here",
		config.SpinnerConfig{Frames: []string{"one"}, Rate: 100},
		time.UnixMilli(0),
		3,
	)

	assert.Equal(t, 3, items[0].Column)
}

func TestCommitColumnIndexesForScreenMode(t *testing.T) {
	order := config.CommitColumnOrder{
		config.CommitColumnMessage,
		config.CommitColumnAuthor,
		config.CommitColumnTime,
	}

	assert.Equal(t, 6, commitMessageColumnForScreenMode(types.SCREEN_NORMAL, order))
	assert.Equal(t, 3, commitMessageColumnForScreenMode(types.SCREEN_HALF, order))
	assert.Equal(t, 6, commitMessageColumnForScreenMode(types.SCREEN_FULL, order))
	assert.Equal(t, -1, commitColumnIndexForScreenMode(types.SCREEN_HALF, order, config.CommitColumnHash))
	assert.Equal(t, 0, commitMessageColumnForScreenMode(
		types.SCREEN_HALF,
		config.CommitColumnOrder{config.CommitColumnHash},
	))
}

func TestSearchModelCommitsUsesHashColumnPosition(t *testing.T) {
	hashPool := &utils.StringPool{}
	commits := []*models.Commit{
		models.NewCommit(hashPool, models.NewCommitOpts{Hash: "abcdef123456", Name: "message"}),
	}
	columnPositions := []int{0, 2, 4, 6, 10, 20}
	modelToViewIndex := func(index int) int { return index + 1 }

	t.Run("visible hash column", func(t *testing.T) {
		result := searchModelCommits(false, commits, columnPositions, 3, modelToViewIndex, "abcdef123456")
		assert.Equal(t, []gocui.SearchPosition{{XStart: 6, XEnd: 9, Y: 1}}, result)
	})

	t.Run("hash in last column", func(t *testing.T) {
		result := searchModelCommits(false, commits, columnPositions, 5, modelToViewIndex, "abcdef123456")
		assert.Equal(t, []gocui.SearchPosition{{XStart: 20, XEnd: 32, Y: 1}}, result)
	})

	t.Run("hidden hash column", func(t *testing.T) {
		result := searchModelCommits(false, commits, columnPositions, -1, modelToViewIndex, "abcdef123456")
		assert.Equal(t, []gocui.SearchPosition{{XStart: -1, XEnd: -1, Y: 1}}, result)
	})
}
