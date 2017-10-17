// Copyright 2016 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
	Package querymanager provides the QueryManager facility which allows database queries to be stored away from code and looked up by Id.

	The facility adds a component of type dsquery.QueryManager to the IoC container that other components may use to
	lookup templated queries by an Id and have those queries populated with supplied parameters.

	A full description of this facility and how to configure it can be found at http://granitic.io/1.0/ref/query-manager . Also refer to the GoDoc for the
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

		Id:ARTIST_ID_SELECT

		SELECT
			id
		FROM
			artist
		WHERE
			name = '${artistName}'

		Id:ARTIST_INSERT

		INSERT INTO artist (
			name
		) VALUES (
			'${artistName}'
		)


		Id:RECORD_INSERT

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

		Id:QUERY_ID

	In this example, an application would use the query Id ARTIST_ID_SELECT to recover the first query.

	Parameters

	A query template may optionally include parameters. Any string inside a ${} structure is considered a parameter name. In the example
	above, query Id RECORD_INSERT defines three parameters catRef, recordName, artistID. The query manager can be supplied with a map
	containing keys that match those parameter names and will populate the template with the values associated with those keys.
*/
package querymanager

import (
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/dsquery"
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
)

// The name of the query manager in the IoC container.
const QueryManagerComponentName = instance.FrameworkPrefix + "QueryManager"

// The name of the facility
const QueryManagerFacilityName = "QueryManager"

// Creates an instance of dsquery.QueryManager and stores it in the IoC container.
type QueryManagerFacilityBuilder struct {
}

// See FacilityBuilder.BuildAndRegister
func (qmfb *QueryManagerFacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.ConfigAccessor, cn *ioc.ComponentContainer) error {

	queryManager := new(dsquery.TemplatedQueryManager)
	ca.Populate("QueryManager", queryManager)

	cn.WrapAndAddProto(QueryManagerComponentName, queryManager)

	return nil
}

// See FacilityBuilder.FacilityName
func (qmfb *QueryManagerFacilityBuilder) FacilityName() string {
	return QueryManagerFacilityName
}

// See FacilityBuilder.DependsOnFacilities
func (qmfb *QueryManagerFacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}
