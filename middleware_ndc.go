package main

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"fmt"
 "bytes"
)

// NDCMiddleware is a caching middleware that will pull data from Redis instead of the upstream proxy
type NDCMiddleware struct {
	*TykMiddleware
	CacheStore StorageHandler
	sh         SuccessHandler
}

type NDCMiddlewareConfig struct {
}

// New lets you do any initialisations for the object can be done here
func (m *NDCMiddleware) New() {
        fmt.Println( "INITIALIZE! ")
	m.sh = SuccessHandler{m.TykMiddleware}
}

// GetConfig retrieves the configuration from the API config - we user mapstructure for this for simplicity
func (m *NDCMiddleware) GetConfig() (interface{}, error) {
	var thisModuleConfig NDCMiddlewareConfig
	return thisModuleConfig, nil
}

func (m NDCMiddleware) CreateCheckSum(req *http.Request, keyName string) string {
	h := md5.New()
	toEncode := strings.Join([]string{req.Method, req.URL.String()}, "-")
	log.Debug("Cache encoding: ", toEncode)
	io.WriteString(h, toEncode)
	reqChecksum := hex.EncodeToString(h.Sum(nil))

	cacheKey := m.Spec.APIDefinition.APIID + keyName + reqChecksum

	return cacheKey
}
// ProcessRequest will run any checks on the request on the way through the system, return an error to have the chain fail
func (m *NDCMiddleware) ProcessRequest(w http.ResponseWriter, r *http.Request, configuration interface{}) (error, int) {
	fmt.Println( "*** NDCMiddleware:ProcessRequest")

	var copiedRequest *http.Request = CopyHttpRequest( r )
	// copiedRequest = CopyHttpRequest(r)

	buf := new(bytes.Buffer)
	buf.ReadFrom( copiedRequest.Body )

	fmt.Println( buf.String() )

	return nil, 200
}
