package main

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/mattermost/mattermost-server/v5/model"
)

// getCodeBlockedLabels returns a list of individually codeblocked names
func getCodeBlockedLabels(names []string) string {
	labels := ""
	sort.Strings(names)
	for _, name := range names {
		labels += fmt.Sprintf(" `%s`", name)
	}
	return labels
}

// getPermaLink returns a link to a postID
func (p *Plugin) getPermaLink(postID string) string {
	return fmt.Sprintf("%v/_redirect/pl/%v", p.GetSiteURL(), postID)
}

// getPostIDFromLink extracts a PostID from a link
func (p *Plugin) getPostIDFromLink(s string) string {
	r := regexp.MustCompile(`http:.*\/\w+\/\w+\/(\w+)`)
	if len(r.FindStringSubmatch(s)) == 2 {
		return r.FindStringSubmatch(s)[1]
	}
	return s
}

// getIconLink returns a markdown link to a postID including a :link: icon
func (p *Plugin) getIconLink(postID string) string {
	iconLink := fmt.Sprintf("[:link:](%s)", p.getPermaLink(postID))
	return iconLink
}

// getTitleFromPost returns a title generated from a Post.Message
func (p *Plugin) getTitleFromPost(postID string) (string, error) {
	// MaxTitleCharacters is the maximum length of characters displayed in a
	// bookmark title
	// MaxTitleCharacters = 30

	// TODO: set limit to number of character from post.Message
	// numChars := math.Min(float64(len(post.Message)), MaxTitleCharacters)
	// bookmark.Title = post.Message[0:int(numChars)]
	post, appErr := p.API.GetPost(postID)
	if appErr != nil {
		return "", appErr
	}
	title := post.Message
	return title, nil
}

// getBmarksEphemeralText returns a the text for posting all bookmarks in an
// ephemeral message
func (p *Plugin) getBmarksEphemeralText(userID string) (string, error) {
	b, err := NewBookmarksWithUser(p.API, userID).getBookmarks()
	if err != nil {
		return "", err
	}

	bmarksSorted, err := b.ByPostCreateAt()
	if err != nil {
		return "", err
	}

	text := "#### Bookmarks List\n"
	for _, bmark := range bmarksSorted {
		labelNames, err := b.getBmarkLabelNames(bmark)
		if err != nil {
			return "", err
		}
		nextText, err := p.getBmarkTextOneLine(bmark, labelNames)
		if err != nil {
			return "", err
		}
		text += nextText
	}
	return text, nil
}

// getBmarkTextOneLine returns a single line bookmark text used for an ephemeral post
func (p *Plugin) getBmarkTextOneLine(bmark *Bookmark, labelNames []string) (string, error) {
	postMessage, err := p.getTitleFromPost(bmark.PostID)
	if err != nil {
		return "", err
	}

	title := "`TitleFromPost` " + postMessage
	if bmark.hasUserTitle(bmark) {
		title = bmark.getTitle()
	}

	codeBlockedNames := getCodeBlockedLabels(labelNames)
	text := fmt.Sprintf("%s%s %s\n", p.getIconLink(bmark.PostID), codeBlockedNames, title)

	return text, nil
}

// getBmarkTextDetailed returns detailed, multi-line bookmark text used for an ephemeral post
func (p *Plugin) getBmarkTextDetailed(bmark *Bookmark, labelNames []string, args *model.CommandArgs) (string, error) {
	title, err := p.getTitleFromPost(bmark.PostID)
	if err != nil {
		return "", err
	}

	if bmark.hasUserTitle(bmark) {
		title = bmark.Title
	}

	codeBlockedNames := getCodeBlockedLabels(labelNames)
	post, appErr := p.API.GetPost(bmark.PostID)
	if appErr != nil {
		return "", appErr
	}

	iconLink := p.getIconLink(bmark.PostID)

	text := fmt.Sprintf("%s\n#### Bookmark Title %s\n", codeBlockedNames, iconLink)
	text += fmt.Sprintf("**%s**\n", title)
	text += "##### Post Message \n"
	text += post.Message

	return text, nil
}
