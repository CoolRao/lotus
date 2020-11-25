package db

func CreateTable() error {

	sql := ` 
 	DROP TABLE IF EXISTS sectors;
	create table sectors (id integer not null primary key, sectorId text,p2Host text);
	delete from sectors;
	`
	_, err := getSqlDb().Exec(sql)
	if err != nil {
		return err
	}
	return nil
}
