package main

import (
	"net/http"
 	"bytes"
 	"encoding/xml"
	// "github.com/influxdb/influxdb/client/v2"
)

// NDCMiddleware is a middleware to perform analytics based on the request body / message

const (
	// TODO: support a range of content types
	acceptedContentType string = "application/xml"
)

type NDCMiddleware struct {
	*TykMiddleware
	CacheStore StorageHandler
	sh         SuccessHandler
}

type NDCMiddlewareConfig struct {
}

// Custom structs (these should be in a separate place later):

type AirShoppingRQType struct {
	PointOfSale pointOfSaleType
}

type pointOfSaleType struct {
	Location locationType
}

type locationType struct {
	CountryCode string
	CityCode    string
}

// New lets you do any initialisations for the object can be done here

func (m *NDCMiddleware) New() {
	log.Info( "NDCMiddleware init")
	m.sh = SuccessHandler{m.TykMiddleware}
}


// Sample RecordHit()

func RecordHit( r *AirShoppingRQType ) {
	log.Debug( "RecordHit" )
}
// GetConfig retrieves the configuration from the API config - we user mapstructure for this for simplicity

func (m *NDCMiddleware) GetConfig() (interface{}, error) {
	var thisModuleConfig NDCMiddlewareConfig
	return thisModuleConfig, nil
}

// ProcessRequest will run any checks on the request on the way through the system, return an error to have the chain fail

func (m *NDCMiddleware) ProcessRequest(w http.ResponseWriter, r *http.Request, configuration interface{}) (error, int) {
	log.Debug( "NDCMiddleware ProcessRequest")
	var copiedRequest *http.Request = CopyHttpRequest( r )

	buf := new(bytes.Buffer)
	buf.ReadFrom( copiedRequest.Body )

	// _, err = buf.WriteTo(w)

	// fmt.Println( buf.String() )

	// TODO: handle nil contentType
	var contentType string = r.Header[ "Content-Type" ][0]

	if( acceptedContentType != contentType ) {
		log.Debug( "Not tracking this request" )
		return nil, 200
	}

	log.Debug( "Tracking this request")

	var AirShoppingRQ AirShoppingRQType
	xml.Unmarshal( buf.Bytes(), &AirShoppingRQ )

	go RecordHit( &AirShoppingRQ )

	return nil, 200


}
