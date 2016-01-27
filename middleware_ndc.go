package main

import (
	"net/http"
 	"bytes"
 	"encoding/xml"
	"time"
	"errors"
	"github.com/influxdb/influxdb/client/v2"
)

// NDCMiddleware is a middleware to perform analytics based on the request body / message

const (
	// TODO: support a range of content types
	acceptedContentType string = "application/xml"
)

var NDCSupportedMethods = map[string] struct{} {
	"AirShoppingRQ": {},
}

type NDCMiddleware struct {
	*TykMiddleware
	CacheStore StorageHandler
	sh         SuccessHandler
	db         client.Client
}

type NDCMiddlewareConfig struct {
}

type NDCMiddlewareRecord struct {
	method	string
	remoteAddress	string
	elapsedTime	float64
}

// Custom structs (these should be in a separate place later):

type NDCGenericMessage struct {
  XMLName xml.Name
	_method string
}

type NDCMessage interface {}

func ParseNDCMessage( RequestBody *[]byte ) ( NDCGenericMessage, NDCMessage, error ) {

	log.Info( "ParseNDCMessage")

	var genericMessage NDCGenericMessage

	var message NDCMessage

	xml.Unmarshal( *RequestBody, &genericMessage )

	var method = genericMessage.XMLName.Local

	_, supported := NDCSupportedMethods[ method ]

	if supported  {
		switch method {
			case "AirShoppingRQ":
				var currentMessage AirShoppingRQType
				xml.Unmarshal( *RequestBody, &currentMessage)

				message = currentMessage
		}

		genericMessage._method = method

		return genericMessage, message, nil
	}

	log.Info( "NDCMessage not supported")
	return genericMessage, nil, errors.New( "NDCMessage not suported")
}

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

func (m *NDCMiddleware) RecordHit( Record *NDCMiddlewareRecord ) {

	log.Info( Record )

	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database: config.NDCMiddlewareConfig.InfluxDbName,
		Precision: "s",
	})

	tags := map[string]string{"ndc_method": Record.method}

  fields := map[string]interface{}{
	    "remoteAddress": Record.remoteAddress,
			"elapsedTime":	Record.elapsedTime,
  }

  pt, _ := client.NewPoint( "ndc", tags, fields, time.Now())

  bp.AddPoint(pt)

  m.db.Write(bp)
}
// GetConfig retrieves the configuration from the API config - we user mapstructure for this for simplicity

func (m *NDCMiddleware) GetConfig() (interface{}, error) {
	var thisModuleConfig NDCMiddlewareConfig // config.NDCMiddlewareConfig?
	return thisModuleConfig, nil
}

// Extra functions

func ComputeRequestTime( t1 time.Time, t2 time.Time ) float64 {
	return float64(t2.UnixNano()-t1.UnixNano()) * 0.000001
}

func ValidNDCMessage( RequestBody *[]byte ) ( NDCMessage ) {
	var message NDCMessage
  return message
}

// ProcessRequest will run any checks on the request on the way through the system, return an error to have the chain fail

func (m *NDCMiddleware) ProcessRequest(w http.ResponseWriter, r *http.Request, configuration interface{}) (error, int) {
	log.Debug( "NDCMiddleware ProcessRequest")

	var copiedRequest *http.Request = CopyHttpRequest( r )

	buf := new(bytes.Buffer)
	buf.ReadFrom( copiedRequest.Body )

	if r.Header["Content-Type"] == nil {
		log.Info( "content type nil")
		return nil, 666
	} else {
		var contentType string = r.Header[ "Content-Type" ][0]
		if( acceptedContentType != contentType ) {
			log.Info( "Ignoring, content type mismatch" )
			return nil, 666
		}
	}

	var body = buf.Bytes()
	var message NDCMessage

	messageInfo, message, err := ParseNDCMessage(&body)

	if( err != nil ) {
		log.Info( "Ignoring, invalid message or not supported method?")
		return nil, 666
	}

	log.Info( "Following request, message is: ")
	log.Info( message )

	var startTime = time.Now()

	reqVal := new(http.Response)
	reqVal = m.sh.ServeHTTPWithCache(w, r)

	var wireFormatReq bytes.Buffer
	reqVal.Write(&wireFormatReq)

	var endTime = time.Now()

	var Record NDCMiddlewareRecord
	Record.method = messageInfo._method
	Record.remoteAddress = r.RemoteAddr
	Record.elapsedTime = ComputeRequestTime( startTime, endTime )

	go m.RecordHit( &Record )

	return nil, 666


}
