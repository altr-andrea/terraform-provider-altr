// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package service

import "strings"

const (
	UUIDv4Regex                    = `^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`
	AlphanumericAndUnderscoreRegex = `^[a-zA-Z0-9_]+$`
	HostnameRegexStringRFC1123     = `^([a-zA-Z0-9]{1}[a-zA-Z0-9-]{0,62}){1}(\.[a-zA-Z0-9]{1}[a-zA-Z0-9-]{0,62})*?$`
)

// OltpDatabaseTypes are the database types accepted by the sc-control API
// for repos (altr_repo.type) and sidecar listeners
// (altr_sidecar_listener.database_type).
var OltpDatabaseTypes = []string{
	"Oracle",
	"MSSQL",
	"MySQL",
	"Postgres",
	"MongoDB",
}

// MongoDBVersionNote is appended to the type descriptions of the resources
// that can create a MongoDB object.
const MongoDBVersionNote = " (MongoDB requires sidecar >= 1.59.0)"

// OltpDatabaseTypesList formats the accepted types for schema descriptions.
func OltpDatabaseTypesList() string {
	return strings.Join(OltpDatabaseTypes, ", ")
}
