var d = db.adminCommand( { listDatabases: 1, nameOnly: true } )


for (var i in d.databases ) {

	nm = d.databases[i].name
	if (nm != "admin" && nm != "local" && nm != "mtx_dev") {
		db = db.getSiblingDB(nm)
		db.dropDatabase()
	}
}
