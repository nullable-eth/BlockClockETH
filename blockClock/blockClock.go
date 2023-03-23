package blockClock

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)


// Client struct
type Client struct {
	httpClient *http.Client
	baseURL string
}

// NewClient create new client object
func NewClient(httpClient *http.Client, baseUrl string, pwd string) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
		httpClient.Timeout = 30 * time.Second
	}
	return &Client{httpClient: httpClient, baseURL: baseUrl}
}

// helper
// doReq HTTP client
func doReq(req *http.Request, client *http.Client) ([]byte, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if 200 != resp.StatusCode {
		return nil, fmt.Errorf("%s", body)
	}
	return body, nil
}

// MakeReq HTTP request helper
func (c *Client) MakeReq(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}
	resp, err := doReq(req, c.httpClient)
	if err != nil {
		return nil, err
	}
	return resp, err
}

func (c *Client) LightsOff() error {
	url := fmt.Sprintf("%s/api/lights/off", c.baseURL)
	_, err := c.MakeReq(url)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) LightsOn(color string) error {
	url := fmt.Sprintf("%s/api/lights/%s", c.baseURL, color)
	_, err := c.MakeReq(url)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) LightsFlash() error {
	url := fmt.Sprintf("%s/api/lights/flash", c.baseURL)
	_, err := c.MakeReq(url)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) DisplayOUText(position int, over string, under string) error {
	url := fmt.Sprintf("%s/api/ou_text/%d/%s/%s", c.baseURL, position, over, under)
	_, err := c.MakeReq(url)
	if err != nil {
		return err
	}


	return nil
}

func (c *Client) DisplayImage(position int, name string) error {
	url := fmt.Sprintf("%s/api/image/%d/%s", c.baseURL, position, name)
	_, err := c.MakeReq(url)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) DisplayLargeText(text string, showCurrency bool) error {
	currency := ""
	if showCurrency {
		currency = "?sym=$"
	}
	url := fmt.Sprintf("%s/api/show/text/%s%s", c.baseURL, text, currency)
	_, err := c.MakeReq(url)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) PauseBuiltInFunctions() error {
	url := fmt.Sprintf("%s/api/action/pause", c.baseURL)
	_, err := c.MakeReq(url)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) ResumeBuiltInFunctions() error {
	url := fmt.Sprintf("%s/api/action/update?rate=5", c.baseURL)
	_, err := c.MakeReq(url)
	if err != nil {
		return err
	}

	return nil
}