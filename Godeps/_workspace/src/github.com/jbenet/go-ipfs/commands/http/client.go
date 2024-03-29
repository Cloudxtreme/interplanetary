package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	cmds "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/commands"
	config "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/config"
)

const (
	ApiUrlFormat = "http://%s%s/%s?%s"
	ApiPath      = "/api/v0" // TODO: make configurable
)

// Client is the commands HTTP client interface.
type Client interface {
	Send(req cmds.Request) (cmds.Response, error)
}

type client struct {
	serverAddress string
}

func NewClient(address string) Client {
	return &client{address}
}

func (c *client) Send(req cmds.Request) (cmds.Response, error) {

	// save user-provided encoding
	var previousUserProvidedEncoding string
	if encopt := req.Option(cmds.EncShort); encopt != nil {
		var err error
		previousUserProvidedEncoding, _, err = encopt.String()
		if err != nil {
			return nil, err
		}
	}

	// override with json to send to server
	req.SetOption(cmds.EncShort, cmds.JSON)

	query, err := getQuery(req)
	if err != nil {
		return nil, err
	}

	var fileReader *MultiFileReader
	if req.Files() != nil {
		fileReader = NewMultiFileReader(req.Files(), true)
	}

	path := strings.Join(req.Path(), "/")
	url := fmt.Sprintf(ApiUrlFormat, c.serverAddress, ApiPath, path, query)

	httpReq, err := http.NewRequest("POST", url, fileReader)
	if err != nil {
		return nil, err
	}

	// TODO extract string consts?
	if fileReader != nil {
		httpReq.Header.Set("Content-Type", "multipart/form-data; boundary="+fileReader.Boundary())
		httpReq.Header.Set("Content-Disposition", "form-data: name=\"files\"")
	} else {
		httpReq.Header.Set("Content-Type", "application/octet-stream")
	}
	version := config.CurrentVersionNumber
	httpReq.Header.Set("User-Agent", fmt.Sprintf("/go-ipfs/%s/", version))

	httpRes, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}

	// using the overridden JSON encoding in request
	res, err := getResponse(httpRes, req)
	if err != nil {
		return nil, err
	}

	if len(previousUserProvidedEncoding) > 0 {
		// reset to user provided encoding after sending request
		// NB: if user has provided an encoding but it is the empty string,
		// still leave it as JSON.
		req.SetOption(cmds.EncShort, previousUserProvidedEncoding)
	}

	return res, nil
}

func getQuery(req cmds.Request) (string, error) {
	query := url.Values{}
	for k, v := range req.Options() {
		str := fmt.Sprintf("%v", v)
		query.Set(k, str)
	}

	args := req.Arguments()
	argDefs := req.Command().Arguments
	var argDef cmds.Argument

	for i, arg := range args {
		if i < len(argDefs) {
			argDef = argDefs[i]
		}

		if argDef.Type == cmds.ArgString {
			query.Add("arg", arg)
		}
	}

	return query.Encode(), nil
}

// getResponse decodes a http.Response to create a cmds.Response
func getResponse(httpRes *http.Response, req cmds.Request) (cmds.Response, error) {
	var err error
	res := cmds.NewResponse(req)

	contentType := httpRes.Header["Content-Type"][0]
	contentType = strings.Split(contentType, ";")[0]

	if len(httpRes.Header.Get(streamHeader)) > 0 {
		res.SetOutput(httpRes.Body)
		return res, nil
	}

	dec := json.NewDecoder(httpRes.Body)

	if httpRes.StatusCode >= http.StatusBadRequest {
		e := cmds.Error{}

		if httpRes.StatusCode == http.StatusNotFound {
			// handle 404s
			e.Message = "Command not found."
			e.Code = cmds.ErrClient

		} else if contentType == "text/plain" {
			// handle non-marshalled errors
			buf := bytes.NewBuffer(nil)
			io.Copy(buf, httpRes.Body)
			e.Message = string(buf.Bytes())
			e.Code = cmds.ErrNormal

		} else {
			// handle marshalled errors
			err = dec.Decode(&e)
			if err != nil {
				return nil, err
			}
		}

		res.SetError(e, e.Code)

	} else {
		v := req.Command().Type
		err = dec.Decode(&v)
		if err != nil && err != io.EOF {
			return nil, err
		}

		res.SetOutput(v)
	}

	return res, nil
}
