package busworker

import (
	"encoding/json"
	"github.com/huajiao-tv/pepperbus/utility/msgRedis"
)

type Client struct {
	addr string
	*msgRedis.MultiPool
}

// redis address string format
// ip:port:password
func NewClient(addr string) *Client {
	return &Client{
		addr: addr,
		MultiPool: msgRedis.NewMultiPool(
			[]string{addr},
			msgRedis.DefaultMaxConnNumber,
			msgRedis.DefaultMaxIdleNumber,
			msgRedis.DefaultMaxIdleSeconds),
	}
}

// add jobs to specified queue
func (c *Client) AddJobs(queue string, password string, contents ...interface{}) error {
	var (
		key  = queue + "/default"
		data = make([]string, 0, len(contents))
	)

	for _, content := range contents {
		if sv, ok := content.(string); ok {
			data = append(data, sv)
		} else if bv, ok := content.([]byte); ok {
			data = append(data, string(bv))
		} else if v, err := json.Marshal(content); err == nil {
			data = append(data, string(v))
		} else {
			// cannot convert to json string
			return err
		}
	}
	if err := c.Auth(queue, password); err != nil {
		return err
	}
	_, err := c.Call(c.addr).LPUSH(key, data)
	return err
}
func (c *Client) Auth(queue string, password string) error {
	args := queue + ":" + password
	_, err := c.Call(c.addr).AUTH(args)
	return err
}
