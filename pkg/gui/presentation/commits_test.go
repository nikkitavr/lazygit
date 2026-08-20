package presentation

import (
	"strings"
	"testing"
	"time"

	"github.com/gookit/color"
	"github.com/jesseduffield/generics/set"
	"github.com/jesseduffield/lazygit/pkg/commands/git_commands"
	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/jesseduffield/lazygit/pkg/common"
	"github.com/jesseduffield/lazygit/pkg/config"
	"github.com/jesseduffield/lazygit/pkg/utils"
	"github.com/samber/lo"
	"github.com/stefanhaller/git-todo-parser/todo"
	"github.com/stretchr/testify/assert"
	"github.com/xo/terminfo"
)

func formatExpected(expected string) string {
	return strings.TrimSpace(strings.ReplaceAll(expected, "\t", ""))
}

func TestGetCommitListDisplayStrings(t *testing.T) {
	scenarios := []struct {
		testName                  string
		commitOpts                []models.NewCommitOpts
		branches                  []*models.Branch
		currentBranchName         string
		hasUpdateRefConfig        bool
		fullDescription           bool
		commitColumnOrder         config.CommitColumnOrder
		cherryPickedCommitHashSet *set.Set[string]
		markedBaseCommit          string
		diffName                  string
		timeFormat                string
		shortTimeFormat           string
		now                       time.Time
		parseEmoji                bool
		selectedCommitHashPtr     *string
		startIdx                  int
		endIdx                    int
		showGraph                 bool
		bisectInfo                *git_commands.BisectInfo
		expected                  string
		focus                     bool
		setupConfig               func(*common.Common)
	}{
		{
			testName:                  "no commits",
			commitOpts:                []models.NewCommitOpts{},
			startIdx:                  0,
			endIdx:                    1,
			showGraph:                 false,
			bisectInfo:                git_commands.NewNullBisectInfo(),
			cherryPickedCommitHashSet: set.New[string](),
			now:                       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			expected:                  "",
		},
		{
			testName: "some commits",
			commitOpts: []models.NewCommitOpts{
				{Name: "commit1", Hash: "hash1"},
				{Name: "commit2", Hash: "hash2"},
			},
			startIdx:                  0,
			endIdx:                    2,
			showGraph:                 false,
			bisectInfo:                git_commands.NewNullBisectInfo(),
			cherryPickedCommitHashSet: set.New[string](),
			now:                       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: formatExpected(`
		hash1 commit1
		hash2 commit2
						`),
		},
		{
			testName: "commit with tags",
			commitOpts: []models.NewCommitOpts{
				{Name: "commit1", Hash: "hash1", Tags: []string{"tag1", "tag2"}},
				{Name: "commit2", Hash: "hash2"},
			},
			startIdx:                  0,
			endIdx:                    2,
			showGraph:                 false,
			bisectInfo:                git_commands.NewNullBisectInfo(),
			cherryPickedCommitHashSet: set.New[string](),
			now:                       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: formatExpected(`
		hash1 tag1 tag2 commit1
		hash2 commit2
						`),
		},
		{
			testName: "show local branch head, except the current branch, main branches, or merged branches",
			commitOpts: []models.NewCommitOpts{
				{Name: "commit1", Hash: "hash1"},
				{Name: "commit2", Hash: "hash2"},
				{Name: "commit3", Hash: "hash3"},
				{Name: "commit4", Hash: "hash4", Status: models.StatusMerged},
			},
			branches: []*models.Branch{
				{Name: "current-branch", CommitHash: "hash1", Head: true},
				{Name: "other-branch", CommitHash: "hash2", Head: false},
				{Name: "master", CommitHash: "hash3", Head: false},
				{Name: "old-branch", CommitHash: "hash4", Head: false},
			},
			currentBranchName:         "current-branch",
			hasUpdateRefConfig:        true,
			startIdx:                  0,
			endIdx:                    4,
			showGraph:                 false,
			bisectInfo:                git_commands.NewNullBisectInfo(),
			cherryPickedCommitHashSet: set.New[string](),
			now:                       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: formatExpected(`
		hash1 commit1
		hash2 * commit2
		hash3 commit3
		hash4 commit4
						`),
		},
		{
			testName: "show local branch head for head commit if updateRefs is on",
			commitOpts: []models.NewCommitOpts{
				{Name: "commit1", Hash: "hash1"},
				{Name: "commit2", Hash: "hash2"},
			},
			branches: []*models.Branch{
				{Name: "current-branch", CommitHash: "hash1", Head: true},
				{Name: "other-branch", CommitHash: "hash1", Head: false},
			},
			currentBranchName:         "current-branch",
			hasUpdateRefConfig:        true,
			startIdx:                  0,
			endIdx:                    2,
			showGraph:                 false,
			bisectInfo:                git_commands.NewNullBisectInfo(),
			cherryPickedCommitHashSet: set.New[string](),
			now:                       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: formatExpected(`
		hash1 * commit1
		hash2 commit2
						`),
		},
		{
			testName: "don't show local branch head for head commit if updateRefs is off",
			commitOpts: []models.NewCommitOpts{
				{Name: "commit1", Hash: "hash1"},
				{Name: "commit2", Hash: "hash2"},
			},
			branches: []*models.Branch{
				{Name: "current-branch", CommitHash: "hash1", Head: true},
				{Name: "other-branch", CommitHash: "hash1", Head: false},
			},
			currentBranchName:         "current-branch",
			hasUpdateRefConfig:        false,
			startIdx:                  0,
			endIdx:                    2,
			showGraph:                 false,
			bisectInfo:                git_commands.NewNullBisectInfo(),
			cherryPickedCommitHashSet: set.New[string](),
			now:                       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: formatExpected(`
		hash1 commit1
		hash2 commit2
						`),
		},
		{
			testName: "show local branch head and tag if both exist",
			commitOpts: []models.NewCommitOpts{
				{Name: "commit1", Hash: "hash1"},
				{Name: "commit2", Hash: "hash2", Tags: []string{"some-tag"}},
				{Name: "commit3", Hash: "hash3"},
			},
			branches: []*models.Branch{
				{Name: "some-branch", CommitHash: "hash2"},
			},
			startIdx:                  0,
			endIdx:                    3,
			showGraph:                 false,
			bisectInfo:                git_commands.NewNullBisectInfo(),
			cherryPickedCommitHashSet: set.New[string](),
			now:                       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: formatExpected(`
		hash1 commit1
		hash2 * some-tag commit2
		hash3 commit3
						`),
		},
		{
			testName: "showing graph",
			commitOpts: []models.NewCommitOpts{
				{Name: "commit1", Hash: "hash1", Parents: []string{"hash2", "hash3"}},
				{Name: "commit2", Hash: "hash2", Parents: []string{"hash3"}},
				{Name: "commit3", Hash: "hash3", Parents: []string{"hash4"}},
				{Name: "commit4", Hash: "hash4", Parents: []string{"hash5"}},
				{Name: "commit5", Hash: "hash5", Parents: []string{"hash7"}},
			},
			startIdx:                  0,
			endIdx:                    5,
			showGraph:                 true,
			bisectInfo:                git_commands.NewNullBisectInfo(),
			cherryPickedCommitHashSet: set.New[string](),
			now:                       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: formatExpected(`
		hash1 ◎─╮ commit1
		hash2 ○ │ commit2
		hash3 ○─╯ commit3
		hash4 ○ commit4
		hash5 ○ commit5
						`),
		},
		{
			testName: "showing graph, including rebase commits",
			commitOpts: []models.NewCommitOpts{
				{Name: "commit1", Hash: "hash1", Parents: []string{"hash2", "hash3"}, Action: todo.Pick},
				{Name: "commit2", Hash: "hash2", Parents: []string{"hash3"}, Action: todo.Pick},
				{Name: "commit3", Hash: "hash3", Parents: []string{"hash4"}},
				{Name: "commit4", Hash: "hash4", Parents: []string{"hash5"}},
				{Name: "commit5", Hash: "hash5", Parents: []string{"hash7"}},
			},
			startIdx:                  0,
			endIdx:                    5,
			showGraph:                 true,
			bisectInfo:                git_commands.NewNullBisectInfo(),
			cherryPickedCommitHashSet: set.New[string](),
			now:                       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: formatExpected(`
		hash1 pick commit1
		hash2 pick commit2
		hash3      ○ commit3
		hash4      ○ commit4
		hash5      ○ commit5
				`),
		},
		{
			testName: "showing graph, including rebase commits, with offset",
			commitOpts: []models.NewCommitOpts{
				{Name: "commit1", Hash: "hash1", Parents: []string{"hash2", "hash3"}, Action: todo.Pick},
				{Name: "commit2", Hash: "hash2", Parents: []string{"hash3"}, Action: todo.Pick},
				{Name: "commit3", Hash: "hash3", Parents: []string{"hash4"}},
				{Name: "commit4", Hash: "hash4", Parents: []string{"hash5"}},
				{Name: "commit5", Hash: "hash5", Parents: []string{"hash7"}},
			},
			startIdx:                  1,
			endIdx:                    5,
			showGraph:                 true,
			bisectInfo:                git_commands.NewNullBisectInfo(),
			cherryPickedCommitHashSet: set.New[string](),
			now:                       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: formatExpected(`
		hash2 pick commit2
		hash3      ○ commit3
		hash4      ○ commit4
		hash5      ○ commit5
				`),
		},
		{
			testName: "startIdx is past TODO commits",
			commitOpts: []models.NewCommitOpts{
				{Name: "commit1", Hash: "hash1", Parents: []string{"hash2", "hash3"}, Action: todo.Pick},
				{Name: "commit2", Hash: "hash2", Parents: []string{"hash3"}, Action: todo.Pick},
				{Name: "commit3", Hash: "hash3", Parents: []string{"hash4"}},
				{Name: "commit4", Hash: "hash4", Parents: []string{"hash5"}},
				{Name: "commit5", Hash: "hash5", Parents: []string{"hash7"}},
			},
			startIdx:                  3,
			endIdx:                    5,
			showGraph:                 true,
			bisectInfo:                git_commands.NewNullBisectInfo(),
			cherryPickedCommitHashSet: set.New[string](),
			now:                       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: formatExpected(`
		hash4 ○ commit4
		hash5 ○ commit5
				`),
		},
		{
			testName: "only showing TODO commits",
			commitOpts: []models.NewCommitOpts{
				{Name: "commit1", Hash: "hash1", Parents: []string{"hash2", "hash3"}, Action: todo.Pick},
				{Name: "commit2", Hash: "hash2", Parents: []string{"hash3"}, Action: todo.Pick},
				{Name: "commit3", Hash: "hash3", Parents: []string{"hash4"}},
				{Name: "commit4", Hash: "hash4", Parents: []string{"hash5"}},
				{Name: "commit5", Hash: "hash5", Parents: []string{"hash7"}},
			},
			startIdx:                  0,
			endIdx:                    2,
			showGraph:                 true,
			bisectInfo:                git_commands.NewNullBisectInfo(),
			cherryPickedCommitHashSet: set.New[string](),
			now:                       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: formatExpected(`
		hash1 pick commit1
		hash2 pick commit2
				`),
		},
		{
			testName: "no TODO commits, towards bottom",
			commitOpts: []models.NewCommitOpts{
				{Name: "commit1", Hash: "hash1", Parents: []string{"hash2", "hash3"}},
				{Name: "commit2", Hash: "hash2", Parents: []string{"hash3"}},
				{Name: "commit3", Hash: "hash3", Parents: []string{"hash4"}},
				{Name: "commit4", Hash: "hash4", Parents: []string{"hash5"}},
				{Name: "commit5", Hash: "hash5", Parents: []string{"hash7"}},
			},
			startIdx:                  4,
			endIdx:                    5,
			showGraph:                 true,
			bisectInfo:                git_commands.NewNullBisectInfo(),
			cherryPickedCommitHashSet: set.New[string](),
			now:                       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: formatExpected(`
			hash5 ○ commit5
				`),
		},
		{
			testName: "only TODO commits except last",
			commitOpts: []models.NewCommitOpts{
				{Name: "commit1", Hash: "hash1", Parents: []string{"hash2", "hash3"}, Action: todo.Pick},
				{Name: "commit2", Hash: "hash2", Parents: []string{"hash3"}, Action: todo.Pick},
				{Name: "commit3", Hash: "hash3", Parents: []string{"hash4"}, Action: todo.Pick},
				{Name: "commit4", Hash: "hash4", Parents: []string{"hash5"}, Action: todo.Pick},
				{Name: "commit5", Hash: "hash5", Parents: []string{"hash7"}},
			},
			startIdx:                  0,
			endIdx:                    2,
			showGraph:                 true,
			bisectInfo:                git_commands.NewNullBisectInfo(),
			cherryPickedCommitHashSet: set.New[string](),
			now:                       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: formatExpected(`
			hash1 pick commit1
			hash2 pick commit2
				`),
		},
		{
			testName: "graph in divergence view - all commits visible",
			commitOpts: []models.NewCommitOpts{
				{Name: "commit1", Hash: "hash1r", Parents: []string{"hash2r"}, Divergence: models.DivergenceRight},
				{Name: "commit2", Hash: "hash2r", Parents: []string{"hash3r", "hash5r"}, Divergence: models.DivergenceRight},
				{Name: "commit3", Hash: "hash3r", Parents: []string{"hash4r"}, Divergence: models.DivergenceRight},
				{Name: "commit1", Hash: "hash1l", Parents: []string{"hash2l"}, Divergence: models.DivergenceLeft},
				{Name: "commit2", Hash: "hash2l", Parents: []string{"hash3l", "hash4l"}, Divergence: models.DivergenceLeft},
				{Name: "commit3", Hash: "hash3l", Parents: []string{"hash4l"}, Divergence: models.DivergenceLeft},
				{Name: "commit4", Hash: "hash4l", Parents: []string{"hash5l"}, Divergence: models.DivergenceLeft},
				{Name: "commit5", Hash: "hash5l", Parents: []string{"hash6l"}, Divergence: models.DivergenceLeft},
			},
			startIdx:                  0,
			endIdx:                    8,
			showGraph:                 true,
			bisectInfo:                git_commands.NewNullBisectInfo(),
			cherryPickedCommitHashSet: set.New[string](),
			now:                       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: formatExpected(`
		↓ hash1r ○ commit1
		↓ hash2r ◎─╮ commit2
		↓ hash3r ○ │ commit3
		↑ hash1l ○ commit1
		↑ hash2l ◎─╮ commit2
		↑ hash3l ○ │ commit3
		↑ hash4l ○─╯ commit4
		↑ hash5l ○ commit5
				`),
		},
		{
			testName: "graph in divergence view - not all remote commits visible",
			commitOpts: []models.NewCommitOpts{
				{Name: "commit1", Hash: "hash1r", Parents: []string{"hash2r"}, Divergence: models.DivergenceRight},
				{Name: "commit2", Hash: "hash2r", Parents: []string{"hash3r", "hash5r"}, Divergence: models.DivergenceRight},
				{Name: "commit3", Hash: "hash3r", Parents: []string{"hash4r"}, Divergence: models.DivergenceRight},
				{Name: "commit1", Hash: "hash1l", Parents: []string{"hash2l"}, Divergence: models.DivergenceLeft},
				{Name: "commit2", Hash: "hash2l", Parents: []string{"hash3l", "hash4l"}, Divergence: models.DivergenceLeft},
				{Name: "commit3", Hash: "hash3l", Parents: []string{"hash4l"}, Divergence: models.DivergenceLeft},
				{Name: "commit4", Hash: "hash4l", Parents: []string{"hash5l"}, Divergence: models.DivergenceLeft},
				{Name: "commit5", Hash: "hash5l", Parents: []string{"hash6l"}, Divergence: models.DivergenceLeft},
			},
			startIdx:                  2,
			endIdx:                    8,
			showGraph:                 true,
			bisectInfo:                git_commands.NewNullBisectInfo(),
			cherryPickedCommitHashSet: set.New[string](),
			now:                       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: formatExpected(`
		↓ hash3r ○ │ commit3
		↑ hash1l ○ commit1
		↑ hash2l ◎─╮ commit2
		↑ hash3l ○ │ commit3
		↑ hash4l ○─╯ commit4
		↑ hash5l ○ commit5
				`),
		},
		{
			testName: "graph in divergence view - not all local commits",
			commitOpts: []models.NewCommitOpts{
				{Name: "commit1", Hash: "hash1r", Parents: []string{"hash2r"}, Divergence: models.DivergenceRight},
				{Name: "commit2", Hash: "hash2r", Parents: []string{"hash3r", "hash5r"}, Divergence: models.DivergenceRight},
				{Name: "commit3", Hash: "hash3r", Parents: []string{"hash4r"}, Divergence: models.DivergenceRight},
				{Name: "commit1", Hash: "hash1l", Parents: []string{"hash2l"}, Divergence: models.DivergenceLeft},
				{Name: "commit2", Hash: "hash2l", Parents: []string{"hash3l", "hash4l"}, Divergence: models.DivergenceLeft},
				{Name: "commit3", Hash: "hash3l", Parents: []string{"hash4l"}, Divergence: models.DivergenceLeft},
				{Name: "commit4", Hash: "hash4l", Parents: []string{"hash5l"}, Divergence: models.DivergenceLeft},
				{Name: "commit5", Hash: "hash5l", Parents: []string{"hash6l"}, Divergence: models.DivergenceLeft},
			},
			startIdx:                  0,
			endIdx:                    5,
			showGraph:                 true,
			bisectInfo:                git_commands.NewNullBisectInfo(),
			cherryPickedCommitHashSet: set.New[string](),
			now:                       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: formatExpected(`
		↓ hash1r ○ commit1
		↓ hash2r ◎─╮ commit2
		↓ hash3r ○ │ commit3
		↑ hash1l ○ commit1
		↑ hash2l ◎─╮ commit2
				`),
		},
		{
			testName: "graph in divergence view - no remote commits visible",
			commitOpts: []models.NewCommitOpts{
				{Name: "commit1", Hash: "hash1r", Parents: []string{"hash2r"}, Divergence: models.DivergenceRight},
				{Name: "commit2", Hash: "hash2r", Parents: []string{"hash3r", "hash5r"}, Divergence: models.DivergenceRight},
				{Name: "commit3", Hash: "hash3r", Parents: []string{"hash4r"}, Divergence: models.DivergenceRight},
				{Name: "commit1", Hash: "hash1l", Parents: []string{"hash2l"}, Divergence: models.DivergenceLeft},
				{Name: "commit2", Hash: "hash2l", Parents: []string{"hash3l", "hash4l"}, Divergence: models.DivergenceLeft},
				{Name: "commit3", Hash: "hash3l", Parents: []string{"hash4l"}, Divergence: models.DivergenceLeft},
				{Name: "commit4", Hash: "hash4l", Parents: []string{"hash5l"}, Divergence: models.DivergenceLeft},
				{Name: "commit5", Hash: "hash5l", Parents: []string{"hash6l"}, Divergence: models.DivergenceLeft},
			},
			startIdx:                  4,
			endIdx:                    8,
			showGraph:                 true,
			bisectInfo:                git_commands.NewNullBisectInfo(),
			cherryPickedCommitHashSet: set.New[string](),
			now:                       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: formatExpected(`
		↑ hash2l ◎─╮ commit2
		↑ hash3l ○ │ commit3
		↑ hash4l ○─╯ commit4
		↑ hash5l ○ commit5
				`),
		},
		{
			testName: "graph in divergence view - no local commits visible",
			commitOpts: []models.NewCommitOpts{
				{Name: "commit1", Hash: "hash1r", Parents: []string{"hash2r"}, Divergence: models.DivergenceRight},
				{Name: "commit2", Hash: "hash2r", Parents: []string{"hash3r", "hash5r"}, Divergence: models.DivergenceRight},
				{Name: "commit3", Hash: "hash3r", Parents: []string{"hash4r"}, Divergence: models.DivergenceRight},
				{Name: "commit1", Hash: "hash1l", Parents: []string{"hash2l"}, Divergence: models.DivergenceLeft},
				{Name: "commit2", Hash: "hash2l", Parents: []string{"hash3l", "hash4l"}, Divergence: models.DivergenceLeft},
				{Name: "commit3", Hash: "hash3l", Parents: []string{"hash4l"}, Divergence: models.DivergenceLeft},
				{Name: "commit4", Hash: "hash4l", Parents: []string{"hash5l"}, Divergence: models.DivergenceLeft},
				{Name: "commit5", Hash: "hash5l", Parents: []string{"hash6l"}, Divergence: models.DivergenceLeft},
			},
			startIdx:                  0,
			endIdx:                    2,
			showGraph:                 true,
			bisectInfo:                git_commands.NewNullBisectInfo(),
			cherryPickedCommitHashSet: set.New[string](),
			now:                       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: formatExpected(`
		↓ hash1r ○ commit1
		↓ hash2r ◎─╮ commit2
				`),
		},
		{
			testName: "graph in divergence view - no remote commits present",
			commitOpts: []models.NewCommitOpts{
				{Name: "commit1", Hash: "hash1l", Parents: []string{"hash2l"}, Divergence: models.DivergenceLeft},
				{Name: "commit2", Hash: "hash2l", Parents: []string{"hash3l", "hash4l"}, Divergence: models.DivergenceLeft},
				{Name: "commit3", Hash: "hash3l", Parents: []string{"hash4l"}, Divergence: models.DivergenceLeft},
				{Name: "commit4", Hash: "hash4l", Parents: []string{"hash5l"}, Divergence: models.DivergenceLeft},
				{Name: "commit5", Hash: "hash5l", Parents: []string{"hash6l"}, Divergence: models.DivergenceLeft},
			},
			startIdx:                  0,
			endIdx:                    5,
			showGraph:                 true,
			bisectInfo:                git_commands.NewNullBisectInfo(),
			cherryPickedCommitHashSet: set.New[string](),
			now:                       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: formatExpected(`
		↑ hash1l ○ commit1
		↑ hash2l ◎─╮ commit2
		↑ hash3l ○ │ commit3
		↑ hash4l ○─╯ commit4
		↑ hash5l ○ commit5
				`),
		},
		{
			testName: "graph in divergence view - no local commits present",
			commitOpts: []models.NewCommitOpts{
				{Name: "commit1", Hash: "hash1r", Parents: []string{"hash2r"}, Divergence: models.DivergenceRight},
				{Name: "commit2", Hash: "hash2r", Parents: []string{"hash3r", "hash5r"}, Divergence: models.DivergenceRight},
				{Name: "commit3", Hash: "hash3r", Parents: []string{"hash4r"}, Divergence: models.DivergenceRight},
			},
			startIdx:                  0,
			endIdx:                    3,
			showGraph:                 true,
			bisectInfo:                git_commands.NewNullBisectInfo(),
			cherryPickedCommitHashSet: set.New[string](),
			now:                       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: formatExpected(`
		↓ hash1r ○ commit1
		↓ hash2r ◎─╮ commit2
		↓ hash3r ○ │ commit3
				`),
		},
		{
			testName: "empty half-screen column order preserves legacy layout",
			commitOpts: []models.NewCommitOpts{
				{Name: "commit1", Hash: "hash1"},
			},
			fullDescription:           true,
			commitColumnOrder:         config.CommitColumnOrder{},
			startIdx:                  0,
			endIdx:                    1,
			showGraph:                 false,
			bisectInfo:                git_commands.NewNullBisectInfo(),
			cherryPickedCommitHashSet: set.New[string](),
			now:                       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			setupConfig: func(c *common.Common) {
				c.UserConfig().Gui.CommitAuthorLongLength = 0
			},
			expected: "hash1 commit1",
		},
		{
			testName: "custom half-screen column order",
			commitOpts: []models.NewCommitOpts{
				{Name: "commit1", Hash: "hash1", UnixTimestamp: 1577844184, AuthorName: "Jesse Duffield"},
				{Name: "commit2", Hash: "hash2", UnixTimestamp: 1576844184, AuthorName: "Jesse Duffield"},
			},
			fullDescription:           true,
			commitColumnOrder:         config.CommitColumnOrder{config.CommitColumnMessage, config.CommitColumnAuthor, config.CommitColumnTime, config.CommitColumnHash},
			timeFormat:                "2006-01-02",
			shortTimeFormat:           "3:04PM",
			startIdx:                  0,
			endIdx:                    2,
			showGraph:                 false,
			bisectInfo:                git_commands.NewNullBisectInfo(),
			cherryPickedCommitHashSet: set.New[string](),
			now:                       time.Date(2020, 1, 1, 5, 3, 4, 0, time.UTC),
			expected: formatExpected(`
		commit1 Jesse Duffield    2:03AM     hash1
		commit2 Jesse Duffield    2019-12-20 hash2
						`),
		},
		{
			testName: "custom half-screen column subset hides omitted columns",
			commitOpts: []models.NewCommitOpts{
				{Name: "commit1", Hash: "hash1", UnixTimestamp: 1577844184, AuthorName: "Jesse Duffield"},
				{Name: "commit2", Hash: "hash2", UnixTimestamp: 1576844184, AuthorName: "Jesse Duffield"},
			},
			fullDescription:           true,
			commitColumnOrder:         config.CommitColumnOrder{config.CommitColumnMessage, config.CommitColumnTime},
			timeFormat:                "2006-01-02",
			shortTimeFormat:           "3:04PM",
			startIdx:                  0,
			endIdx:                    2,
			showGraph:                 false,
			bisectInfo:                git_commands.NewNullBisectInfo(),
			cherryPickedCommitHashSet: set.New[string](),
			now:                       time.Date(2020, 1, 1, 5, 3, 4, 0, time.UTC),
			expected: formatExpected(`
		commit1 2:03AM
		commit2 2019-12-20
						`),
		},
		{
			testName: "custom half-screen columns override legacy hiding",
			commitOpts: []models.NewCommitOpts{
				{Name: "commit1", Hash: "hash1", AuthorName: "Jesse Duffield"},
			},
			fullDescription:           true,
			commitColumnOrder:         config.CommitColumnOrder{config.CommitColumnMessage, config.CommitColumnAuthor, config.CommitColumnHash},
			startIdx:                  0,
			endIdx:                    1,
			showGraph:                 false,
			bisectInfo:                git_commands.NewNullBisectInfo(),
			cherryPickedCommitHashSet: set.New[string](),
			now:                       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			setupConfig: func(c *common.Common) {
				c.UserConfig().Gui.CommitAuthorLongLength = 0
				c.UserConfig().Gui.CommitHashLength = 0
			},
			expected: "commit1 JD hash1",
		},
		{
			testName: "custom half-screen columns keep structural indicators and hide message decorations",
			commitOpts: []models.NewCommitOpts{
				{Name: "commit1", Hash: "hash1", Parents: []string{"hash2"}, AuthorName: "Jesse Duffield", Status: models.StatusConflicted, Action: todo.Pick, Divergence: models.DivergenceRight},
			},
			fullDescription:           true,
			commitColumnOrder:         config.CommitColumnOrder{config.CommitColumnAuthor},
			startIdx:                  0,
			endIdx:                    1,
			showGraph:                 true,
			bisectInfo:                git_commands.NewNullBisectInfo(),
			cherryPickedCommitHashSet: set.New[string](),
			now:                       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			setupConfig: func(c *common.Common) {
				c.UserConfig().Gui.CommitAuthorLongLength = 2
			},
			expected: "↓ pick JD",
		},
		{
			testName: "custom half-screen columns preserve an otherwise empty row",
			commitOpts: []models.NewCommitOpts{
				{Hash: "hash1"},
			},
			fullDescription:           true,
			commitColumnOrder:         config.CommitColumnOrder{config.CommitColumnMessage},
			startIdx:                  0,
			endIdx:                    1,
			showGraph:                 false,
			bisectInfo:                git_commands.NewNullBisectInfo(),
			cherryPickedCommitHashSet: set.New[string](),
			now:                       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			expected:                  " ",
		},
		{
			testName: "custom time format",
			commitOpts: []models.NewCommitOpts{
				{Name: "commit1", Hash: "hash1", UnixTimestamp: 1577844184, AuthorName: "Jesse Duffield"},
				{Name: "commit2", Hash: "hash2", UnixTimestamp: 1576844184, AuthorName: "Jesse Duffield"},
			},
			fullDescription:           true,
			timeFormat:                "2006-01-02",
			shortTimeFormat:           "3:04PM",
			startIdx:                  0,
			endIdx:                    2,
			showGraph:                 false,
			bisectInfo:                git_commands.NewNullBisectInfo(),
			cherryPickedCommitHashSet: set.New[string](),
			now:                       time.Date(2020, 1, 1, 5, 3, 4, 0, time.UTC),
			expected: formatExpected(`
		hash1 2:03AM     Jesse Duffield    commit1
		hash2 2019-12-20 Jesse Duffield    commit2
						`),
		},
	}

	oldColorLevel := color.ForceSetColorLevel(terminfo.ColorLevelNone)
	defer color.ForceSetColorLevel(oldColorLevel)

	focusing := false
	for _, scenario := range scenarios {
		if scenario.focus {
			focusing = true
		}
	}

	for _, s := range scenarios {
		if !focusing || s.focus {
			t.Run(s.testName, func(t *testing.T) {
				common := common.NewDummyCommon()
				if s.setupConfig != nil {
					s.setupConfig(common)
				}

				hashPool := &utils.StringPool{}

				commits := lo.Map(s.commitOpts,
					func(opts models.NewCommitOpts, _ int) *models.Commit { return models.NewCommit(hashPool, opts) })

				result := GetCommitListDisplayStrings(
					common,
					commits,
					s.branches,
					s.currentBranchName,
					s.hasUpdateRefConfig,
					s.fullDescription,
					s.commitColumnOrder,
					s.cherryPickedCommitHashSet,
					s.diffName,
					s.markedBaseCommit,
					s.timeFormat,
					s.shortTimeFormat,
					s.now,
					s.parseEmoji,
					s.selectedCommitHashPtr,
					s.startIdx,
					s.endIdx,
					s.showGraph,
					s.bisectInfo,
				)

				renderedLines, _ := utils.RenderDisplayStrings(result, nil)
				renderedResult := strings.Join(renderedLines, "\n")
				t.Logf("\n%s", renderedResult)

				assert.EqualValues(t, s.expected, renderedResult)
			})
		}
	}
}

func TestCommitColumnIndex(t *testing.T) {
	t.Run("legacy layout", func(t *testing.T) {
		assert.Equal(t, 1, CommitColumnIndex(nil, config.CommitColumnHash))
		assert.Equal(t, 3, CommitColumnIndex(nil, config.CommitColumnTime))
		assert.Equal(t, 5, CommitColumnIndex(nil, config.CommitColumnAuthor))
		assert.Equal(t, 6, CommitColumnIndex(nil, config.CommitColumnMessage))
	})

	t.Run("custom layout", func(t *testing.T) {
		order := config.CommitColumnOrder{
			config.CommitColumnMessage,
			config.CommitColumnAuthor,
			config.CommitColumnTime,
		}
		assert.Equal(t, -1, CommitColumnIndex(order, config.CommitColumnHash))
		assert.Equal(t, 5, CommitColumnIndex(order, config.CommitColumnTime))
		assert.Equal(t, 4, CommitColumnIndex(order, config.CommitColumnAuthor))
		assert.Equal(t, 3, CommitColumnIndex(order, config.CommitColumnMessage))
	})
}
