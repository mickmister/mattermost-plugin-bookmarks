package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/mattermost/mattermost-server/v5/plugin"
)

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.URL.Path {
	case "/add":
		p.handleAdd(w, r)
	case "/get":
		p.handleView(w, r)
	// case "/delete":
	// 	p.handleDelete(w, r)
	case "/labels/get":
		p.handleLabelsGet(w, r)
	case "/labels/add":
		p.handleLabelsAdd(w, r)
	// case "/delete":
	default:
		http.NotFound(w, r)
	}
}

func (p *Plugin) handleAdd(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-ID")
	if userID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var bmark *Bookmark
	if err = json.Unmarshal(body, &bmark); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b := NewBookmarksWithUser(p.API, userID)
	bmarks, err := b.getBookmarks()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bmarkD, _ := json.MarshalIndent(bmark, "", "    ")
	fmt.Printf("bmark = %+v\n", string(bmarkD))

	// check if labelIDs exist.  If not, this is a label name and needs to be
	// converted to label struct with UUID value
	l := NewLabelsWithUser(p.API, userID)
	l, err = l.getLabels()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	labelIDs := bmark.getLabelIDs()

	var labelIDsForBookmark []string
	for _, labelID := range labelIDs {
		label, err := l.get(labelID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		fmt.Printf("label = %+v\n", label)

		// if doesn't exist, this is a name and needs to be added to the labels
		// store.  also save the id to the bookmark, not the name
		if label == nil {
			labelNew, err := l.addLabel(labelID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			fmt.Printf("label is nil = %+v\n", label)
			labelIDsForBookmark = append(labelIDsForBookmark, labelNew.ID)
			continue
		}
		labelIDsForBookmark = append(labelIDsForBookmark, labelID)
	}

	bmark.LabelIDs = labelIDsForBookmark
	fmt.Printf("labelIDsForBookmark = %+v\n", labelIDsForBookmark)
	err = bmarks.addBookmark(bmark)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// text, err := p.getBmarkTextOneLine(bmark, bmark.LabelIDs, nil)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	// post := &model.Post{
	// 	UserId: p.getBotID(),
	// ChannelId: args.ChannelId,
	// Message: text,
	// }
	// _ = p.API.SendEphemeralPost(args.UserId, post)
}

// func (p *Plugin) handleDelete(w http.ResponseWriter, r *http.Request) {
// 	return
// }

func (p *Plugin) handleView(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-ID")
	if userID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query()
	postID := query["postID"][0]
	fmt.Printf("postID = %+v\n", postID)

	b := NewBookmarksWithUser(p.API, userID)
	bmarks, err := b.getBookmarks()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	bmark, err := bmarks.getBookmark(postID)
	fmt.Printf("bmark = %+v\n", bmark)
	if err != nil {
		p.handleErrorWithCode(w, http.StatusBadRequest, "Unable to get bookmark", err)
	}

	resp, err := json.Marshal(bmark)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(resp)

}

func (p *Plugin) handleLabelsGet(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-ID")
	if userID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	l := NewLabelsWithUser(p.API, userID)
	labels, err := l.getLabels()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	resp, err := json.Marshal(labels)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(resp)
}

func (p *Plugin) handleLabelsAdd(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-ID")
	if userID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query()
	labelName := query["labelName"][0]
	fmt.Println("1. IN HERE!")
	fmt.Printf("labelName = %+v\n", labelName)
	// body, err := ioutil.ReadAll(r.Body)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusBadRequest)
	// 	return
	// }
	//
	// fmt.Printf("body = %+v\n", body)
	// fmt.Println("2. IN HERE!")
	// var label *Label
	// if err = json.Unmarshal(body, &labelName); err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	fmt.Println("3. IN HERE!")
	l := NewLabelsWithUser(p.API, userID)
	labels, err := l.getLabels()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("4. IN HERE!")
	label, err := labels.addLabel(labelName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("5. IN HERE!")
	resp, err := json.Marshal(label)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(resp)
}

func (p *Plugin) handleErrorWithCode(w http.ResponseWriter, code int, errTitle string, err error) {
	w.WriteHeader(code)
	b, _ := json.Marshal(struct {
		Error   string `json:"error"`
		Details string `json:"details"`
	}{
		Error:   errTitle,
		Details: err.Error(),
	})
	_, _ = w.Write(b)
}
