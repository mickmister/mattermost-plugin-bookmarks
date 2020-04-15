package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestExecuteCommandAdd(t *testing.T) {
	tests := map[string]struct {
		commandArgs       *model.CommandArgs
		bookmarks         *Bookmarks
		expectedMsgPrefix string
		expectedContains  []string
	}{
		"User doesn't provide an ID": {
			commandArgs:       &model.CommandArgs{Command: "/bookmarks add"},
			bookmarks:         nil,
			expectedMsgPrefix: strings.TrimSpace("Missing "),
			expectedContains:  []string{"Missing sub-command", "bookmarks add"},
		},
		"PostID doesn't exist": {
			commandArgs:       &model.CommandArgs{Command: fmt.Sprintf("/bookmarks add %v", PostIDDoesNotExist)},
			bookmarks:         nil,
			expectedMsgPrefix: strings.TrimSpace(fmt.Sprintf("PostID `%v` is not a valid postID", PostIDDoesNotExist)),
			expectedContains:  nil,
		},
		"Bookmark added  no title provided": {
			commandArgs:       &model.CommandArgs{Command: fmt.Sprintf("/bookmarks add %v", b1ID)},
			bookmarks:         getExecuteCommandTestBookmarks(),
			expectedMsgPrefix: strings.TrimSpace(fmt.Sprintf("Added bookmark: [:link:](https://myhost.com//pl/ID1) `label1` `label2` `TitleFromPost` this is the post.Message")),
			expectedContains:  nil,
		},
		"Bookmark added  title provided": {
			commandArgs:       &model.CommandArgs{Command: fmt.Sprintf("/bookmarks add %v %v", PostIDExists, "TitleProvidedByUser")},
			bookmarks:         getExecuteCommandTestBookmarks(),
			expectedMsgPrefix: strings.TrimSpace(fmt.Sprintf("Added bookmark: [:link:](https://myhost.com//pl/ID2) `label1` `label2` TitleProvidedByUser")),
			expectedContains:  nil,
		},
		"Bookmark added  title provided with labels": {
			commandArgs:       &model.CommandArgs{Command: fmt.Sprintf("/bookmarks add %v %v --labels label1,label2", PostIDExists, "TitleProvidedByUser")},
			bookmarks:         getExecuteCommandTestBookmarks(),
			expectedMsgPrefix: strings.TrimSpace(fmt.Sprintf("Added bookmark: [:link:](https://myhost.com//pl/ID2) `label1` `label2` TitleProvidedByUser")),
			expectedContains:  nil,
		},
		"Bookmark unknown flag provided": {
			commandArgs:       &model.CommandArgs{Command: fmt.Sprintf("/bookmarks add %v --unknownflag", b1ID)},
			bookmarks:         getExecuteCommandTestBookmarks(),
			expectedMsgPrefix: strings.TrimSpace(fmt.Sprintf("Unable to parse options, unknown flag: --unknownflag")),
			expectedContains:  nil,
		},
		"Bookmark --labels provided without options": {
			commandArgs:       &model.CommandArgs{Command: fmt.Sprintf("/bookmarks add %v --labels", b1ID)},
			bookmarks:         getExecuteCommandTestBookmarks(),
			expectedMsgPrefix: strings.TrimSpace(fmt.Sprintf("Unable to parse options, flag needs an argument: --labels")),
			expectedContains:  nil,
		},
		"Bookmark --labels provided with one label": {
			commandArgs:       &model.CommandArgs{Command: fmt.Sprintf("/bookmarks add %v --labels label1", b1ID)},
			bookmarks:         getExecuteCommandTestBookmarks(),
			expectedMsgPrefix: strings.TrimSpace(fmt.Sprintf("Added bookmark: [:link:](https://myhost.com//pl/ID1)")),
			expectedContains:  nil,
		},
		"Bookmark --labels provided with two labels": {
			commandArgs:       &model.CommandArgs{Command: fmt.Sprintf("/bookmarks add %v --labels label1,label2", b1ID)},
			bookmarks:         getExecuteCommandTestBookmarks(),
			expectedMsgPrefix: strings.TrimSpace(fmt.Sprintf("Added bookmark: [:link:](https://myhost.com//pl/ID1)")),
			expectedContains:  nil,
		},
	}
	for name, tt := range tests {
		api := makeAPIMock()
		tt.commandArgs.UserId = UserID
		siteURL := "https://myhost.com"
		api.On("GetPost", PostIDDoesNotExist).Return(nil, &model.AppError{Message: "An Error Occurred"})
		api.On("GetPost", b1ID).Return(&model.Post{Message: "this is the post.Message"}, nil)
		api.On("GetPost", b2ID).Return(&model.Post{Message: "this is the post.Message"}, nil)
		api.On("GetPost", b3ID).Return(&model.Post{Message: "this is the post.Message"}, nil)
		api.On("GetPost", b4ID).Return(&model.Post{Message: "this is the post.message"}, nil)
		api.On("addBookmark", UserID, tt.bookmarks).Return(mock.Anything)
		api.On("GetTeam", mock.Anything).Return(&model.Team{Id: teamID1}, nil)
		api.On("GetConfig", mock.Anything).Return(&model.Config{ServiceSettings: model.ServiceSettings{SiteURL: &siteURL}})
		api.On("exists", mock.Anything).Return(true)
		// api.On("ByID", mock.Anything).Return(true)

		jsonBmarks, err := json.Marshal(tt.bookmarks)
		api.On("KVGet", getBookmarksKey(tt.commandArgs.UserId)).Return(jsonBmarks, nil)
		api.On("KVSet", mock.Anything, mock.Anything).Return(nil)

		t.Run(name, func(t *testing.T) {
			assert.Nil(t, err)
			// isSendEphemeralPostCalled := false
			api.On("SendEphemeralPost", mock.AnythingOfType("string"), mock.AnythingOfType("*model.Post")).Run(func(args mock.Arguments) {
				// isSendEphemeralPostCalled = true

				post := args.Get(1).(*model.Post)
				actual := strings.TrimSpace(post.Message)
				assert.True(t, strings.HasPrefix(actual, tt.expectedMsgPrefix), "Expected returned message to start with: \n%s\nActual:\n%s", tt.expectedMsgPrefix, actual)
				if tt.expectedContains != nil {
					for i := range tt.expectedContains {
						assert.Contains(t, actual, tt.expectedContains[i])
					}
				}
				// assert.Contains(t, actual, tt.expectedMsgPrefix)
			}).Once().Return(&model.Post{})
			// assert.Equal(t, true, isSendEphemeralPostCalled)

			p := makePlugin(api)
			cmdResponse, appError := p.ExecuteCommand(&plugin.Context{}, tt.commandArgs)
			require.Nil(t, appError)
			require.NotNil(t, cmdResponse)
			// assert.True(t, isSendEphemeralPostCalled)
		})
	}
}