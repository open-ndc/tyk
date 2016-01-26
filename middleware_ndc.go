package main

import (
	"net/http"
 	"bytes"
 	"encoding/xml"
	"time"
	"math/rand"
	"github.com/influxdb/influxdb/client/v2"
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
	db         client.Client
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

	m.db, _ = client.NewHTTPClient(client.HTTPConfig{
			 Addr: "http://localhost:8086",
			 Username: config.NDCMiddlewareConfig.InfluxDbUsername,
			 Password: config.NDCMiddlewareConfig.InfluxDbPassword,
	 })

	m.sh = SuccessHandler{m.TykMiddleware}
}


// Sample RecordHit()

func (m *NDCMiddleware) RecordHit( r *AirShoppingRQType ) {

	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database: config.NDCMiddlewareConfig.InfluxDbName,
		Precision: "s",
	})

	tags := map[string]string{"ndc_method": "AirShoppingRQ"}

  fields := map[string]interface{}{
	    "a":	rand.Float32(),
			"b":	rand.Float32(),
  }

  pt, _ := client.NewPoint( "ndc", tags, fields, time.Now())

  bp.AddPoint(pt)

  // Write the batch
  m.db.Write(bp)

	log.Debug( "RecordHit" )
}
// GetConfig retrieves the configuration from the API config - we user mapstructure for this for simplicity

func (m *NDCMiddleware) GetConfig() (interface{}, error) {
	var thisModuleConfig NDCMiddlewareConfig // config.NDCMiddlewareConfig?
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

	go m.RecordHit( &AirShoppingRQ )

	return nil, 200


}
