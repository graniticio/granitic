// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package rdbms

// A function able execute an insert statement and return an RDBMS generated ID as an int64.
// If your implementation requires access to the context, it is available on the *Client
type InsertWithReturnedId func(string, *Client, *int64) error

// An implementation of InsertWithReturnedId that will work with any Go database driver that implements LastInsertId
func DefaultInsertWithReturnedId(query string, client *Client, target *int64) error {

	if r, err := client.Exec(query); err != nil {
		return err
	} else {
		if id, err := r.LastInsertId(); err != nil {
			return err
		} else {

			*target = id

			return nil
		}
	}

}
