package tgscrapper

import (
	"fmt"
	"math"
	"net/url"
	"path"
	"sort"
	"strconv"
	"sync"
	"time"
)

type TGMessage struct {
	DateTime time.Time

	Title  string
	Body   string
	Link   string
	Author string

	ID int
}

func (tgm *TGMessage) PopulateMessageID() error {
	u, err := url.Parse(tgm.Link)
	if err != nil {
		return ErrInvalidData
	}

	id, err := strconv.Atoi(path.Base(u.Path))
	if err != nil {
		return ErrInvalidData
	}

	tgm.ID = id

	return nil
}

type TGMessages struct {
	GenerationTime    time.Time
	OldestMessageDate time.Time

	ChannelName        string
	ChannelTitle       string
	ChannelDescription string
	ChannelLink        string

	Items []*TGMessage

	OldestMessageID int

	mu sync.Mutex
}

// NewMessages creates new messages storage object.
// Note: channel name needs to be validated according
// to https://core.telegram.org/method/account.checkUsername
func NewMessages(channelName string) *TGMessages {
	items := make([]*TGMessage, 0)

	currentTime := time.Now().UTC().Round(time.Second)

	return &TGMessages{
		Items:             items,
		ChannelName:       channelName,
		OldestMessageID:   math.MaxInt64,
		OldestMessageDate: currentTime,
		GenerationTime:    currentTime,
	}
}

func (tgms *TGMessages) Store(m *TGMessage) {
	tgms.mu.Lock()

	tgms.Items = append(tgms.Items, m)

	if tgms.OldestMessageID > m.ID {
		tgms.OldestMessageID = m.ID
	}

	if tgms.OldestMessageDate.After(m.DateTime) {
		tgms.OldestMessageDate = m.DateTime
	}

	tgms.mu.Unlock()
}

func (tgms *TGMessages) Sort() {
	tgms.mu.Lock()

	sort.Slice(tgms.Items, func(i, j int) bool {
		return tgms.Items[i].DateTime.UTC().Unix() < tgms.Items[j].DateTime.UTC().Unix()
	})

	tgms.mu.Unlock()
}

func (tgms *TGMessages) MustPaginationURL() string {
	u := &url.URL{
		Scheme: "https",
		Host:   "t.me",
		Path:   path.Join("s", tgms.ChannelName),
	}

	if tgms.OldestMessageID == math.MaxInt64 {
		return u.String()
	}

	val, err := url.ParseQuery(fmt.Sprintf("before=%d", tgms.OldestMessageID))
	if err != nil {
		panic(err)
	}

	u.RawQuery = val.Encode()

	return u.String()
}
