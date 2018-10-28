package db

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"eonbot/pkg"
	"eonbot/pkg/asset"
	"eonbot/pkg/exchange"
	"eonbot/pkg/file"
	"errors"
	"path"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
)

const (
	dbFile         = "data.db"
	maxSavedCycles = 10 // max amount of cycles of specific pair to save
)

var (
	telegramSubsBucket = []byte("telegram-subs")
	cyclesBucket       = []byte("cycles")
	ordersBucket       = []byte("orders")
)

var (
	// ErrDataNotFound is returned if bucket, pair,
	// interval, etc is not found in the db.
	ErrDataNotFound = errors.New("data not found")
)

// PersistentStorer defines methods
// used to store and retrieve
// specific data from persistent
// database.
type PersistentStorer interface {
	// close closes open database's connection.
	close()

	// SaveTelegramSubscriber saves provided chat
	// id to the db, so that when bot is restarted
	// chat's wouldn't need to be re-subscribed.
	SaveTelegramSubscriber(id int64) error

	// GetTelegramSubscribers retrieves currently
	// saved telegram chat ids from the db.
	GetTelegramSubscribers() ([]int64, error)

	// DeleteTelegramSubscriber removes specific id
	// from the db.
	DeleteTelegramSubscriber(id int64) error

	// SavePairCycle saves specific pair's cycle to the db.
	SavePairCycle(pair asset.Pair, cyc *pkg.StreamCycle) error

	// GetPairCycle retrieves specific pair's cycle by its id from the
	// db.
	GetPairCycle(pair asset.Pair, id int64) (*pkg.StreamCycle, error)

	// GetPairCyclesIDs retrieves all saved orders ids from the db.
	GetPairCyclesIDs(pair asset.Pair) ([]uint64, error)

	// SavePairOrder saves specific pair's order to the db.
	SavePairOrder(pair asset.Pair, ord exchange.Order, strat string) error

	// GetPairOrders retrieves all specific pair's orders in the provided time
	// interval from the db.
	GetPairOrders(pair asset.Pair, start, end time.Time) (map[string][]exchange.BotOrder, error)

	// GetOrders retrieves all pairs all orders in the provided time
	// interval from the db.
	GetOrders(start, end time.Time) (map[string][]exchange.BotOrder, error)

	// GetPairActivityOnHours retrieves specific pair's orders in the provided
	// time interval activity on specific hours in a day (0-23) from the db.
	// Returned slice will contain 24 elements, each representing an hour
	// in a day.
	GetPairActivityOnHours(pair asset.Pair, end time.Time, days int) ([]int, error)

	// GetActivityOnHours retrieves all pairs orders in the provided
	// time interval activity on specific hours in a day (0-23) from the db.
	// Returned slice will contain 24 elements, each representing an hour
	// in a day.
	GetActivityOnHours(end time.Time, days int) ([]int, error)

	// GetPairOrdersCount retrieves pair's total orders
	// count from the db.
	GetPairOrdersCount(pair asset.Pair) (int, error)

	// GetOrdersCount retrieves all pairs total orders
	// count from the db.
	GetOrdersCount() (int, error)
}

// persistentStore contains persistent
// db connection.
type persistentStore struct {
	db *bolt.DB
}

// newPersistentStore creates new persistentStore.
func newPersistentStore() (*persistentStore, error) {
	dir, err := file.ExecDir()
	if err != nil {
		return nil, err
	}

	// open a connection to the db file.
	db, err := bolt.Open(path.Join(dir, dbFile), 0600, &bolt.Options{Timeout: time.Second * 20})
	if err != nil {
		if err == bolt.ErrTimeout {
			return nil, errors.New("another bot instance is running on the same machine")
		}
		return nil, err
	}

	return &persistentStore{db: db}, nil
}

func (p *persistentStore) close() {
	p.db.Close()
}

/*
   Telegram subscribers
*/

func (p *persistentStore) SaveTelegramSubscriber(id int64) error {
	return p.db.Update(func(tx *bolt.Tx) error {
		// find or create telegram subscribers bucket.
		b, err := tx.CreateBucketIfNotExists(telegramSubsBucket)
		if err != nil {
			return err
		}

		// save or update data.
		return b.Put([]byte(strconv.FormatInt(id, 10)), []byte{})
	})
}

func (p *persistentStore) GetTelegramSubscribers() ([]int64, error) {
	subs := make([]int64, 0)
	err := p.db.View(func(tx *bolt.Tx) error {
		// retrieve subscribers bucket.
		b := tx.Bucket(telegramSubsBucket)
		if b == nil {
			return nil // no need to error if subs don't exist
		}

		// loop over all chat ids and add them to
		// the slice.
		return b.ForEach(func(k []byte, v []byte) error {
			// convert to required format.
			id, err := strconv.ParseInt(string(k), 10, 64)
			if err != nil {
				return err
			}

			subs = append(subs, id)
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return subs, nil
}

func (p *persistentStore) DeleteTelegramSubscriber(id int64) error {
	return p.db.Update(func(tx *bolt.Tx) error {
		// retrieve subscribers bucket.
		b := tx.Bucket(telegramSubsBucket)
		if b == nil {
			return nil // no need to error if subs don't exist
		}

		// remove data.
		return b.Delete([]byte(strconv.FormatInt(id, 10)))
	})
}

/*
   Pair cycles snapshots
*/

func (p *persistentStore) SavePairCycle(pair asset.Pair, cyc *pkg.StreamCycle) error {
	return p.db.Update(func(tx *bolt.Tx) error {
		// find or create cycles bucket.
		b, err := tx.CreateBucketIfNotExists(cyclesBucket)
		if err != nil {
			return err
		}

		// find or create pairs bucket.
		pb, err := b.CreateBucketIfNotExists([]byte(pair.String()))
		if err != nil {
			return err
		}

		// convert cycle data to json.
		bCyc, err := json.Marshal(cyc)
		if err != nil {
			return err
		}

		// generate cycle db id.
		id, err := pb.NextSequence()
		if err != nil {
			return err
		}

		// save or update data.
		if err := pb.Put(itob(id), bCyc); err != nil {
			return err
		}

		// clean up old cycles (if more than allowed exist).
		if pb.Stats().KeyN > maxSavedCycles {
			// calculate how many to remove.
			count := pb.Stats().KeyN - maxSavedCycles
			c := pb.Cursor()

			// loop over all saved pair's cycles.
			for k, _ := c.First(); k != nil; k, _ = c.Next() {
				// remove cycle.
				if err := pb.Delete(k); err != nil {
					return err
				}

				// decrement cycles-to-remove count.
				count--

				// stop the loop when count is not valid anymore.
				if count <= 0 {
					break
				}
			}
		}

		return nil
	})
}

func (p *persistentStore) GetPairCycle(pair asset.Pair, id int64) (*pkg.StreamCycle, error) {
	// pair must be valid to continue.
	if err := pair.RequireValid(); err != nil {
		return nil, err
	}

	cyc := new(pkg.StreamCycle)
	err := p.db.View(func(tx *bolt.Tx) error {
		// find cycles bucket.
		b := tx.Bucket(cyclesBucket)
		if b == nil {
			return ErrDataNotFound
		}

		// find pair bucket.
		pb := b.Bucket([]byte(pair.String()))
		if pb == nil {
			return ErrDataNotFound
		}

		var bCyc []byte
		if id > 0 {
			// retrieve cycle by provided id.
			bCyc = pb.Get(itob(uint64(id)))
		} else { // if number is -1 or less, return last element
			c := pb.Cursor()

			// retrieve last cycle in db.
			_, bCyc = c.Last()
		}

		if bCyc == nil {
			return ErrDataNotFound
		}

		// convert json data.
		if err := json.Unmarshal(bCyc, cyc); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return cyc, nil
}

func (p *persistentStore) GetPairCyclesIDs(pair asset.Pair) ([]uint64, error) {
	// pair must be valid to continue.
	if err := pair.RequireValid(); err != nil {
		return nil, err
	}

	ids := make([]uint64, 0)
	err := p.db.View(func(tx *bolt.Tx) error {
		// find cycles bucket.
		b := tx.Bucket(cyclesBucket)
		if b == nil {
			return ErrDataNotFound
		}

		// find pair bucket.
		pb := b.Bucket([]byte(pair.String()))
		if pb == nil {
			return ErrDataNotFound
		}

		// loop over all pair's cycles and collect
		// their ids.
		return pb.ForEach(func(k, v []byte) error {
			// convert and append id.
			ids = append(ids, boti(k))
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return ids, nil
}

/*
   Pair orders made by bot
*/

func (p *persistentStore) SavePairOrder(pair asset.Pair, ord exchange.Order, strat string) error {
	return p.db.Update(func(tx *bolt.Tx) error {
		// find or create orders bucket.
		b, err := tx.CreateBucketIfNotExists(ordersBucket)
		if err != nil {
			return err
		}

		// find or create pair bucket.
		pb, err := b.CreateBucketIfNotExists([]byte(pair.String()))
		if err != nil {
			return err
		}

		date := ord.Timestamp
		if date.IsZero() {
			date = time.Now()
		}

		// find or create current timestamp bucket.
		tb, err := pb.CreateBucketIfNotExists([]byte(date.UTC().Format(time.RFC3339)))
		if err != nil {
			return err
		}

		// convert to json.
		bOrd, err := json.Marshal(exchange.NewBotOrder(ord, strat))
		if err != nil {
			return err
		}

		// generate id.
		id, err := tb.NextSequence()
		if err != nil {
			return err
		}

		// save or update data.
		return tb.Put(itob(id), bOrd)
	})
}

func (p *persistentStore) GetPairOrders(pair asset.Pair, start, end time.Time) (map[string][]exchange.BotOrder, error) {
	// pair must be valid to continue.
	if err := pair.RequireValid(); err != nil {
		return nil, err
	}

	orders := make(map[string][]exchange.BotOrder)
	err := p.db.View(func(tx *bolt.Tx) error {
		// find orders bucket.
		b := tx.Bucket(ordersBucket)
		if b == nil {
			return ErrDataNotFound
		}

		// find pair bucket.
		pb := b.Bucket([]byte(pair.String()))
		if pb == nil {
			return ErrDataNotFound
		}

		orders[pair.String()] = make([]exchange.BotOrder, 0)

		c := pb.Cursor()

		min := start.UTC().Format(time.RFC3339)
		max := end.UTC().Format(time.RFC3339)

		// loop over orders that are in specific time interval.
		for k, v := c.Seek([]byte(min)); k != nil && bytes.Compare(k, []byte(max)) <= 0; k, v = c.Next() {
			if v == nil { // bucket
				// find timestamp bucket.
				tb := pb.Bucket(k)

				// loop over orders in timestamp.
				err := tb.ForEach(func(tk []byte, tv []byte) error {
					var order exchange.BotOrder

					// convert from json.
					if err := json.Unmarshal(tv, &order); err != nil {
						return err
					}
					orders[pair.String()] = append(orders[pair.String()], order)
					return nil
				})

				if err != nil {
					return err
				}
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (p *persistentStore) GetOrders(start, end time.Time) (map[string][]exchange.BotOrder, error) {
	orders := make(map[string][]exchange.BotOrder)
	err := p.db.View(func(tx *bolt.Tx) error {
		// find orders bucket.
		b := tx.Bucket(ordersBucket)
		if b == nil {
			return ErrDataNotFound
		}

		// loop over pairs.
		return b.ForEach(func(k []byte, v []byte) error {
			if v == nil { // bucket
				orders[string(k)] = make([]exchange.BotOrder, 0)

				// find pairs bucket.
				pb := b.Bucket(k)
				c := pb.Cursor()

				min := start.UTC().Format(time.RFC3339)
				max := end.UTC().Format(time.RFC3339)

				// loop over orders that are in specific time interval.
				for pk, pv := c.Seek([]byte(min)); pk != nil && bytes.Compare(pk, []byte(max)) <= 0; pk, pv = c.Next() {
					if pv == nil { // bucket
						// find timestamp bucket.
						tb := pb.Bucket(pk)

						// loop over orders in timestamp.
						err := tb.ForEach(func(tk []byte, tv []byte) error {
							var order exchange.BotOrder

							// convert from json.
							if err := json.Unmarshal(tv, &order); err != nil {
								return err
							}
							orders[string(k)] = append(orders[string(pk)], order)
							return nil
						})

						if err != nil {
							return err
						}
					}
				}
			}
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (p *persistentStore) GetPairActivityOnHours(pair asset.Pair, end time.Time, days int) ([]int, error) {
	if err := pair.RequireValid(); err != nil {
		return nil, err
	}

	var start time.Time

	if days == 0 {
		return nil, errors.New("days count cannot be 0")
	} else if days < 0 {
		start = time.Unix(0, 0)
	} else {
		start = end.Add(-time.Hour * 24 * time.Duration(days))
	}

	hours := make([]int, 24) // 24 hours
	err := p.db.View(func(tx *bolt.Tx) error {
		// find orders bucket.
		b := tx.Bucket(ordersBucket)
		if b == nil {
			return ErrDataNotFound
		}

		// find pair bucket.
		pb := b.Bucket([]byte(pair.String()))
		if pb == nil {
			return ErrDataNotFound
		}

		c := pb.Cursor()

		min := start.UTC().Format(time.RFC3339)
		max := end.UTC().Format(time.RFC3339)

		// loop over orders that are in specific time interval.
		for k, v := c.Seek([]byte(min)); k != nil && bytes.Compare(k, []byte(max)) <= 0; k, v = c.Next() {
			if v == nil { // bucket
				// find timestamp bucket.
				tb := pb.Bucket(k)

				// convert timestamp.
				t, err := time.Parse(time.RFC3339, string(k))
				if err != nil {
					return err
				}

				// find hour.
				for i := 0; i < 24; i++ {
					if t.Hour() == i {
						hours[i] = hours[i] + tb.Stats().KeyN
						break
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return hours, nil
}

func (p *persistentStore) GetActivityOnHours(end time.Time, days int) ([]int, error) {
	if days == 0 {
		return nil, errors.New("days count cannot be 0")
	}

	var start time.Time

	if days == 0 {
		return nil, errors.New("days count cannot be 0")
	} else if days < 0 {
		start = time.Unix(0, 0)
	} else {
		start = end.Add(-time.Hour * 24 * time.Duration(days))
	}

	hours := make([]int, 24) // 24 hours
	err := p.db.View(func(tx *bolt.Tx) error {
		// find orders bucket.
		b := tx.Bucket(ordersBucket)
		if b == nil {
			return ErrDataNotFound
		}

		// loop over pairs.
		return b.ForEach(func(k, v []byte) error {
			if v == nil { // bucket

				// find pair bucket.
				pb := b.Bucket(k)
				c := pb.Cursor()

				min := start.UTC().Format(time.RFC3339)
				max := end.UTC().Format(time.RFC3339)

				// loop over orders that are in specific time interval.
				for pk, pv := c.Seek([]byte(min)); pk != nil && bytes.Compare(pk, []byte(max)) <= 0; pk, pv = c.Next() {
					if pv == nil { // bucket
						// find timestamp bucket.
						tb := pb.Bucket(pk)

						// convert timestamp.
						t, err := time.Parse(time.RFC3339, string(pk))
						if err != nil {
							return err
						}

						// find hour.
						for i := 0; i < 24; i++ {
							if t.Hour() == i {
								hours[i] = hours[i] + tb.Stats().KeyN
								break
							}
						}
					}
				}
			}
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return hours, nil
}

func (p *persistentStore) GetPairOrdersCount(pair asset.Pair) (int, error) {
	if err := pair.RequireValid(); err != nil {
		return 0, err
	}

	var res int
	err := p.db.View(func(tx *bolt.Tx) error {
		// find orders bucket.
		b := tx.Bucket(ordersBucket)
		if b == nil {
			return ErrDataNotFound
		}

		// find pair bucket.
		pb := b.Bucket([]byte(pair.String()))
		if pb == nil {
			return ErrDataNotFound
		}

		// loop over timestamp buckets and count
		// how many entries they have.
		return pb.ForEach(func(pk, pv []byte) error {
			if pv == nil { // bucket
				tb := pb.Bucket(pk)
				res += tb.Stats().KeyN
			}
			return nil
		})
	})

	if err != nil {
		return 0, err
	}

	return res, nil
}

func (p *persistentStore) GetOrdersCount() (int, error) {
	var res int
	err := p.db.View(func(tx *bolt.Tx) error {
		// find orders bucket.
		b := tx.Bucket(ordersBucket)
		if b == nil {
			return ErrDataNotFound
		}

		// loop over pairs.
		return b.ForEach(func(k, v []byte) error {
			if v == nil { // bucket
				// find pair bucket.
				pb := b.Bucket(k)

				// loop over timestamp buckets and count
				// how many entries they have.
				err := pb.ForEach(func(pk, pv []byte) error {
					if pv == nil { // bucket
						tb := pb.Bucket(pk)
						res += tb.Stats().KeyN
					}
					return nil
				})
				if err != nil {
					return err
				}
			}
			return nil
		})
	})

	if err != nil {
		return 0, err
	}

	return res, nil
}

// itob returns an 8-byte big endian representation of v.
// From: https://github.com/boltdb/bolt#autoincrementing-integer-for-the-bucket
func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

// boti returns uint64 version of v.
// Oposite to itob.
func boti(v []byte) uint64 {
	return binary.BigEndian.Uint64(v)
}
