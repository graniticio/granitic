package rdbms

type InsertWithReturnedID func(string, *RDBMSClient, *int64) error

func DefaultInsertWithReturnedID(query string, client *RDBMSClient, target *int64) error {

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
