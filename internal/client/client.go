// client パッケージははてなブログのAPIを利用するためのHTTPクライアントです
// このパッケージで利用しているAPIは予告なく変更される可能性があります
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"sync"
)

type BlogMember struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

// membersCache is a cache for ListMembers
type membersCache struct {
	sync.RWMutex
	Members []*BlogMember
}

type Client struct {
	client *http.Client

	username       string
	owner          string
	blogHost       string
	hatenablogHost string
	insecure       bool
	membersCache   membersCache
}

func NewClient(version, username, apikey, owner, blogHost string) *Client {
	return &Client{
		client: &http.Client{
			Transport: newTransport(username, apikey, version),
		},
		username:       username,
		owner:          owner,
		blogHost:       blogHost,
		hatenablogHost: "blog.hatena.ne.jp",
		insecure:       false,
		membersCache: membersCache{
			Members: nil,
		},
	}
}

func (c *Client) SetHatenablogHost(host string) {
	c.hatenablogHost = host
}

func (c *Client) SetInsecure(insecure bool) {
	c.insecure = insecure
}

func (c *Client) AddMember(username, role string) (*BlogMember, error) {
	data := BlogMember{
		Username: username,
		Role:     role,
	}
	body, _ := json.Marshal(data)

	req, err := http.NewRequest("POST", c.buildURL("members").String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if !(200 <= res.StatusCode && res.StatusCode < 300) {
		respBody, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", res.StatusCode, respBody)
	}

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var resData BlogMember
	if err := json.Unmarshal(buf, &resData); err != nil {
		return nil, err
	}

	// ListMembersのキャッシュを破棄
	c.membersCache.Lock()
	defer c.membersCache.Unlock()
	c.membersCache.Members = nil

	return &resData, nil
}

// ListMembers lists members of the blog.
func (c *Client) ListMembers() ([]*BlogMember, error) {
	// terraform plan時にresourceの数だけこのメソッドが実行される
	// キャッシュがあるときはそれを返すことでリクエストの実行数を減らす
	c.membersCache.RLock()
	if c.membersCache.Members != nil {
		members := append([]*BlogMember{}, c.membersCache.Members...)
		c.membersCache.RUnlock()
		return members, nil
	}
	c.membersCache.RUnlock()

	res, err := c.client.Get(c.buildURL("members").String())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if !(200 <= res.StatusCode && res.StatusCode < 300) {
		respBody, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", res.StatusCode, respBody)
	}

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	resData := new(struct {
		Members []*BlogMember `json:"members"`
	})
	if err := json.Unmarshal(buf, resData); err != nil {
		return nil, err
	}

	// ListMembersのキャッシュを破棄
	c.membersCache.Lock()
	defer c.membersCache.Unlock()
	c.membersCache.Members = append([]*BlogMember{}, resData.Members...)

	return c.membersCache.Members, nil
}

func (c *Client) DeleteMember(username string) error {
	req, err := http.NewRequest("DELETE", c.buildURL("members", username).String(), nil)
	if err != nil {
		return err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if !(200 <= res.StatusCode && res.StatusCode < 300) {
		respBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", res.StatusCode, respBody)
	}

	c.membersCache.Lock()
	defer c.membersCache.Unlock()
	c.membersCache.Members = nil

	return nil
}

// buildURL ははてなブログのAPIのURLを生成するためのヘルパ関数
// 生成するURLは次の形式
// (http|https)://<hatenablogHost>/<owner>/<blogHost>/api/<p...>
func (c *Client) buildURL(p ...string) *url.URL {
	paths := make([]string, 0, len(p)+3)
	paths = append(paths, c.owner, c.blogHost, "api")
	paths = append(paths, p...)

	scheme := "https"
	if c.insecure {
		scheme = "http"
	}
	return &url.URL{
		Scheme: scheme,
		Host:   c.hatenablogHost,
		Path:   path.Join(paths...),
	}
}
