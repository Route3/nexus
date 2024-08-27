package engine

import (
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	client *http.Client
	url    *url.URL
	token  string
}

func NewClient(rawUrl string, token string) (*Client, error) {
	url, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	return &Client{
		client,
		url,
		token,
	}, nil
}

func (c *Client) GetPayloadV1() (*GetPayloadV1Response, error) {
	return nil, nil
}

func (c *Client) NewPayloadV1() (*GetPayloadV1Response, error) {
	return nil, nil
}

func (c *Client) ForkchoiceUpdatedV1() (*ForkchoiceUpdatedV1Response, error) {
	return nil, nil
}

func (c *Client) ExchangeCapabilities() (*ExchangeCapabilitiesResponse, error) {
	return nil, nil
}
