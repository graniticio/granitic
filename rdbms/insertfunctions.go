// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package rdbms

import "database/sql"

// InsertWithReturnedID is a function able execute an insert statement and return an RDBMS generated ID as an int64.
// If your implementation requires access to the context, it is available on the *ManagedClient
type InsertWithReturnedID func(string, Client, *int64) error

// DefaultInsertWithReturnedID is an implementation of InsertWithReturnedID that will work with any Go database driver that implements LastInsertId
func DefaultInsertWithReturnedID(query string, client Client, target *int64) error {
	var r sql.Result
	var err error
	var id int64

	if r, err = client.Exec(query); err != nil {
		return err
	}

	if id, err = r.LastInsertId(); err != nil {
		return err
	}

	*target = id

	return nil
}
