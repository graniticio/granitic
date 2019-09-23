// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
Package querymanager provides the QueryManager facility which allows database queries to be stored away from code and looked up by ID.

The facility adds a component of type dsquery.QueryManager to the IoC container that other components may use to
lookup templated queries by an ID and have those queries populated with supplied parameters.

A full description of this facility and how to configure it can be found at https://granitic.io/ref/query-management-facility . Also refer to the GoDoc for the
GoDoc for the dsquery package.

Template locations and template formats

The QueryManager manager facility is configured with the QueryManager configuration element. For most applications, only the
TemplateLocation might need changing from it's default. The default setting is:

	{
	  "QueryManager":{
		"TemplateLocation": "resource/queries"
	  }
	}

On startup, any files in the the TemplateLocation will be treated as containing query templates. The name of each file
is not significant. A typical file might look like:

	ID:ARTIST_ID_SELECT

	SELECT
		id
	FROM
		artist
	WHERE
		name = '${artistName}'

	ID:ARTIST_INSERT

	INSERT INTO artist (
		name
	) VALUES (
		'${artistName}'
	)


	ID:RECORD_INSERT

	INSERT INTO record (
		cat_ref,
		name,
		artist_id
	) VALUES (
		'${catRef}',
		'${recordName}',
		${artistID}
	)

This file defines three query templates. A new template is signified by a line starting

	ID:QUERY_ID

In this example, an application would use the query ID ARTIST_ID_SELECT to recover the first query.

Parameters

A query template may optionally include parameters. Any string inside a ${} structure is considered a parameter name. In the example
above, query ID RECORD_INSERT defines three parameters catRef, recordName, artistID. The query manager can be supplied with a map
containing keys that match those parameter names and will populate the template with the values associated with those keys.

Required parameters

If you put a ! character before a parameter name in your template (e.g. ${!artistID}), an error will be returned if that parameter is
not available when a query is built.

Parameter Values

Parameter values are injected into the query using a component called a ParamValueProcessor. Granitic includes two
built-in implementations - ConfigurableProcessor and SQLProcessor. These components a) decide how to handle missing parameter
values and b) perform any escaping/substitution/conversion of values before they are injected into the query.

To enable one of the default processors, set QueryManager.ProcessorName to Configurable or SQL (the default is Configurable). If
you want to implement your own processor, set QueryManager.CreateDefaultValueProcessor to false and define a component that
implements ParamValueProcessor
*/
package querymanager

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/v2/config"
	"github.com/graniticio/granitic/v2/dsquery"
	"github.com/graniticio/granitic/v2/instance"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
)

// QueryManagerComponentName is the name of the query manager in the IoC container.
const QueryManagerComponentName = instance.FrameworkPrefix + "QueryManager"

// QueryManagerFacilityName is the name of the facility
const QueryManagerFacilityName = "QueryManager"

const processorDecorator = instance.FrameworkPrefix + "ParamValueProcessorDecorator"

const confValueProcess = "Configurable"
const sqlValueProcess = "SQL"

// FacilityBuilder creates an instance of dsquery.QueryManager and stores it in the IoC container.
type FacilityBuilder struct {
}

// BuildAndRegister implements FacilityBuilder.BuildAndRegister
func (qmfb *FacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.Accessor, cn *ioc.ComponentContainer) error {

	queryManager := new(dsquery.TemplatedQueryManager)
	ca.Populate("QueryManager", queryManager)

	cn.WrapAndAddProto(QueryManagerComponentName, queryManager)

	if build, _ := ca.BoolVal("QueryManager.CreateDefaultValueProcessor"); build == false {
		//Construction of stock value processor has been disabled

		vpd := new(valueProcessorDecorator)
		vpd.QueryManager = queryManager
		cn.WrapAndAddProto(processorDecorator, vpd)

		return nil
	}

	vpName, err := ca.StringVal("QueryManager.ProcessorName")

	if err != nil || (vpName != confValueProcess && vpName != sqlValueProcess) {
		return fmt.Errorf("QueryManager.ProcessorName must be set to '%s' or '%s' if you want to use a stock ValueProcessor", confValueProcess, sqlValueProcess)
	}

	vpConfig := "QueryManager.ValueProcessors." + vpName

	if !ca.PathExists(vpConfig) {
		return errors.New("Missing configuration path for ValueProcessor: " + vpConfig)
	}

	var vp dsquery.ParamValueProcessor

	if vpName == confValueProcess {
		vp = new(dsquery.ConfigurableProcessor)
	} else if vpName == sqlValueProcess {

		vp = new(dsquery.SQLProcessor)

	}

	ca.Populate(vpConfig, vp)

	queryManager.ValueProcessor = vp

	return nil
}

// FacilityName implements FacilityBuilder.FacilityName
func (qmfb *FacilityBuilder) FacilityName() string {
	return QueryManagerFacilityName
}

// DependsOnFacilities implements FacilityBuilder.DependsOnFacilities
func (qmfb *FacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}

// Finds ValueProccessor implementations and injects them into the QueryManager
type valueProcessorDecorator struct {
	QueryManager *dsquery.TemplatedQueryManager
}

func (vpd *valueProcessorDecorator) OfInterest(component *ioc.Component) bool {

	ci := component.Instance

	_, found := ci.(dsquery.ParamValueProcessor)

	return found

}

func (vpd *valueProcessorDecorator) DecorateComponent(component *ioc.Component, container *ioc.ComponentContainer) {
	vpd.QueryManager.ValueProcessor, _ = component.Instance.(dsquery.ParamValueProcessor)
}
