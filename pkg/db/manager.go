package db

// Manager specifies databases manager
// methods.
type Manager interface {
	// InMemory returns in-memoryStore database
	// usage service.
	InMemory() InMomoryStorer

	// Persistent returns persistent database
	// usage service.
	Persistent() PersistentStorer

	// CloseAll closes all connections
	// to the databases.
	CloseAll()
}

// DBManager is an implementation of
// Manager interface.
type DBManager struct {
	mem InMomoryStorer
	per PersistentStorer
}

// New creates new DBManager.
func New() (*DBManager, error) {
	per, err := newPersistentStore()
	if err != nil {
		return nil, err
	}
	return &DBManager{
		mem: newMemory(),
		per: per,
	}, nil
}

func (d *DBManager) InMemory() InMomoryStorer {
	return d.mem
}

func (d *DBManager) Persistent() PersistentStorer {
	return d.per
}

func (d *DBManager) CloseAll() {
	d.per.close()
}
