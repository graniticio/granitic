// Copyright 2016-2020 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package httpserver

import (
	"context"
	"fmt"
	"github.com/graniticio/granitic/v2/config"
	"github.com/graniticio/granitic/v2/instance"
	"github.com/graniticio/granitic/v2/instrument"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
	"github.com/graniticio/granitic/v2/uuid"
	"net/http"
	"strings"
)

// HTTPServerComponentName is the name of the HTTPServer component as stored in the IoC framework.
const HTTPServerComponentName = instance.FrameworkPrefix + "HTTPServer"
const contextIDDecoratorName = instance.FrameworkPrefix + "RequestIDContextDecorator"
const instrumentationDecoratorName = instance.FrameworkPrefix + "RequestInstrumentationDecorator"

const textEntryMode = "TEXT"
const jsonEntryMode = "JSON"

// HTTPServerAbnormalStatusFieldName is the field on the HTTPServer component into which a ws.AbnormalStatusWriter can be injected. Most applications will use either
// the JSONWs or XMLWs facility, in which case a AbnormalStatusWriter that will respond to requests with an abnormal result
// (404, 503 etc) by sending a JSON or XML response respectively.
//
// If this behaviour is undesirable, an alternative AbnormalStatusWriter can set by using the frameworkModifiers mechanism
// (see https://granitic.io/ref/component-definition-files )
const HTTPServerAbnormalStatusFieldName = "AbnormalStatusWriter"
const accessLogWriterName = instance.FrameworkPrefix + "AccessLogWriter"

// FacilityBuilder creates the components that make up the HTTPServer facility (the server and an access log writer).
type FacilityBuilder struct {
}

// BuildAndRegister implements FacilityBuilder.BuildAndRegister
func (hsfb *FacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.Accessor, cn *ioc.ComponentContainer) error {

	log := lm.CreateLogger(instance.FrameworkPrefix + "FacilityBuilder")

	httpServer := new(HTTPServer)
	ca.Populate("HTTPServer", httpServer)

	cn.WrapAndAddProto(HTTPServerComponentName, httpServer)

	if httpServer.AccessLogging {
		if err := hsfb.setupAccessLogging(ca, log, httpServer, cn); err != nil {
			return err
		}
	}

	idbd := new(contextBuilderDecorator)
	idbd.Server = httpServer
	cn.WrapAndAddProto(contextIDDecoratorName, idbd)

	if !httpServer.DisableInstrumentationAutoWire {

		log.LogDebugf("Will attempt to auto-wire an implementation of instrument.RequestInstrumentationManager")

		id := new(instrumentationDecorator)
		id.Server = httpServer
		id.Log = lm.CreateLogger(instrumentationDecoratorName)

		cn.WrapAndAddProto(instrumentationDecoratorName, id)

	} else {
		log.LogDebugf("Auto wiring of instrumentation managers disabled")
	}

	if err := configureRequestIDGeneration(ca, log, httpServer); err != nil {
		return err
	}

	return nil

}

func (hsfb *FacilityBuilder) setupAccessLogging(ca *config.Accessor, log logging.Logger, httpServer *HTTPServer, cn *ioc.ComponentContainer) error {
	accessLogWriter := new(AccessLogWriter)
	ca.Populate("HTTPServer.AccessLog", accessLogWriter)

	var lb LineBuilder
	var mode string
	var err error

	entryPath := "HTTPServer.AccessLog.Entry"

	if mode, err = ca.StringVal(entryPath); err != nil {
		return err
	}

	if mode == textEntryMode {
		ulb := new(UnstructuredLineBuilder)
		ulb.LogLineFormat = accessLogWriter.LogLineFormat
		ulb.LogLinePreset = accessLogWriter.LogLinePreset
		ulb.utcTimes = accessLogWriter.UtcTimes

		lb = ulb
	} else if mode == jsonEntryMode {

		jlb := new(JSONLineBuilder)

		jc := new(AccessLogJSONConfig)
		ca.Populate("HTTPServer.AccessLog.JSON", jc)
		jlb.Config = jc

		jc.UTC = accessLogWriter.UtcTimes

		if err := ValidateJSONFields(jc.Fields); err != nil {
			return err
		}

		if mb, err := CreateMapBuilder(jc); err != nil {
			return err
		} else {
			jlb.MapBuilder = mb
		}

		lb = jlb
	} else {
		return fmt.Errorf("%s is a not a supported value for %s. Should be %s or %s", mode, entryPath, textEntryMode, jsonEntryMode)
	}

	accessLogWriter.builder = lb

	file := accessLogWriter.LogPath

	file = strings.TrimSpace(file)
	file = strings.ToUpper(file)

	if file == stdoutMode {
		log.LogDebugf("Access logs will be written to STDOUT")
		accessLogWriter.LogPath = stdoutMode
	}

	httpServer.AccessLogWriter = accessLogWriter

	cn.WrapAndAddProto(accessLogWriterName, accessLogWriter)

	return nil
}

func configureRequestIDGeneration(ca *config.Accessor, log logging.Logger, s *HTTPServer) error {

	cfg := new(requestIDConfig)
	basePath := "HTTPServer.RequestID"

	if err := ca.Populate(basePath, cfg); err != nil {
		return fmt.Errorf("Unable to read configuration for request ID generation %s", err.Error())
	} else if !cfg.Enabled {
		return nil
	}

	log.LogDebugf("Generation of request IDs enabled")

	if cfg.Format != "UUIDV4" {
		return fmt.Errorf("%s is not a valid configuration value for %s.Format. Must be UUIDV4", cfg.Format, basePath)
	}

	var encodingFunc uuid.EncodeFrom16Byte

	switch cfg.UUID.Encoding {

	case "RFC4122":
		encodingFunc = uuid.StandardEncoder
	case "Base32":
		encodingFunc = uuid.Base32Encoder
	case "Base64":
		encodingFunc = uuid.Base64Encoder
	default:
		return fmt.Errorf("%s is not a valid configuration value for %s.UUID.Encoding. Must be one of RFC4122, Base32, Base64", cfg.UUID.Encoding, basePath)
	}

	rcb := new(requestContextBuilder)
	rcb.encoder = encodingFunc
	rcb.idGen = uuid.GenerateCryptoRand

	s.IDContextBuilder = rcb

	return nil

}

// FacilityName implements FacilityBuilder.FacilityName
func (hsfb *FacilityBuilder) FacilityName() string {
	return "HTTPServer"
}

// DependsOnFacilities implements FacilityBuilder.DependsOnFacilities
func (hsfb *FacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}

// Injects a component whose instance is an implementation of IdentifiedRequestContextBuilder into the HTTP Server
type instrumentationDecorator struct {
	Server *HTTPServer
	Log    logging.Logger
}

// OfInterest returns true if the supplied component is an instance of instrument.RequestInstrumentationManager
func (id *instrumentationDecorator) OfInterest(subject *ioc.Component) bool {
	result := false

	switch subject.Instance.(type) {
	case instrument.RequestInstrumentationManager:
		result = true
	}

	return result
}

// DecorateComponent injects the instrument.RequestInstrumentationManager into the HTTP server
func (id *instrumentationDecorator) DecorateComponent(subject *ioc.Component, cc *ioc.ComponentContainer) {

	im := subject.Instance.(instrument.RequestInstrumentationManager)

	if id.Server.InstrumentationManager != nil {
		id.Log.LogWarnf("Multiple components implementing instrument.RequestInstrumentationManager found. Using %s", subject.Name)

	}

	id.Log.LogDebugf("HTTP server using %s for instrumentation", subject.Name)

	id.Server.InstrumentationManager = im
}

type requestIDConfig struct {
	Enabled bool
	Format  string
	UUID    struct {
		Encoding string
	}
}

type requestContextBuilder struct {
	idGen   uuid.Generate16Byte
	encoder uuid.EncodeFrom16Byte
}

type idKey string

const ridKey idKey = "GRNCREQID"

func (rcb *requestContextBuilder) WithIdentity(ctx context.Context, req *http.Request) (context.Context, error) {

	id := uuid.V4Custom(rcb.idGen, rcb.encoder)

	return context.WithValue(ctx, ridKey, id), nil

}

func (rcb *requestContextBuilder) ID(ctx context.Context) string {
	id := ctx.Value(ridKey)

	if id == nil {
		return ""
	}

	return id.(string)

}
