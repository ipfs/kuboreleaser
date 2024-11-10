package matrix

import (
	"fmt"
	"net/url"

	"github.com/ipfs/kuboreleaser/util"
	"github.com/matrix-org/gomatrix"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	matrix *gomatrix.Client
}

func NewClient() (*Client, error) {
	url := util.GetenvPrompt("MATRIX_URL")
	user := util.GetenvPrompt("MATRIX_USER")
	token := util.GetenvPromptSecret("MATRIX_TOKEN", "If you don't have a token, you can leave it blank and use a password instead. Please enter the token:")
	password := util.GetenvPromptSecret("MATRIX_PASSWORD", "If you don't have a password, you can leave it blank and use a token instead. Please enter the password:")
	if token == "" && password == "" {
		return nil, fmt.Errorf("⚠️ MATRIX_TOKEN nor MATRIX_PASSWORD are set")
	}

	matrix, err := gomatrix.NewClient(url, user, token)
	if err != nil {
		return nil, err
	}

	if password != "" {
		response, err := matrix.Login(&gomatrix.ReqLogin{
			Type:     "m.login.password",
			User:     user,
			Password: password,
		})
		if err != nil {
			return nil, err
		}
		matrix.AccessToken = response.AccessToken
	}

	return &Client{
		matrix: matrix,
	}, nil
}

type RespRoomID struct {
	RoomID  string   `json:"room_id"`
	Servers []string `json:"servers"`
}

func (c *Client) GetRoomID(roomAlias string) (*RespRoomID, error) {
	log.WithFields(log.Fields{
		"roomAlias": roomAlias,
	}).Debug("Getting Room ID...")

	u := c.matrix.BuildBaseURL("_matrix/client/v3/directory/room/", roomAlias)

	var resp *RespRoomID
	err := c.matrix.MakeRequest("GET", u, nil, &resp)
	if err != nil {
		return nil, err
	}
	log.WithFields(log.Fields{
		"roomID": resp.RoomID,
	}).Debug("Got Room ID")
	return resp, nil
}

func (c *Client) GetMessages(roomID, from, to string, dir rune, limit int, filter string) (*gomatrix.RespMessages, error) {
	log.WithFields(log.Fields{
		"roomID": roomID,
		"from":   from,
		"to":     to,
		"dir":    dir,
		"limit":  limit,
		"filter": filter,
	}).Debug("Getting Messages...")

	u, err := url.Parse(c.matrix.BuildBaseURL("_matrix/client/v3/rooms/", roomID, "/messages"))
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("from", from)
	q.Set("dir", string(dir))
	if to != "" {
		q.Set("to", to)
	}
	if limit != 0 {
		q.Set("limit", fmt.Sprintf("%d", limit))
	}
	if filter != "" {
		q.Set("filter", filter)
	}
	u.RawQuery = q.Encode()

	var resp *gomatrix.RespMessages
	err = c.matrix.MakeRequest("GET", u.String(), nil, &resp)
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"chunk": resp.Chunk,
	}).Debug("Got Messages")

	return resp, nil
}

func (c *Client) GetLatestMessagesBy(roomAlias, author string, limit int) ([]gomatrix.Event, error) {
	roomID, err := c.GetRoomID(roomAlias)
	if err != nil {
		return nil, err
	}

	filter := fmt.Sprintf(`{"types":["m.room.message"],"senders":["%s"]}`, author)

	messages, err := c.GetMessages(roomID.RoomID, "", "", 'b', limit, filter)
	if err != nil {
		return nil, err
	}

	return messages.Chunk, nil
}
