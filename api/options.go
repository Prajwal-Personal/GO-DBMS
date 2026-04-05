package api

// Option allows configuring the DB instance.
type Option func(*DB)

// WithMaxPool returns an Option to configure the connection pool size.
func WithMaxPool(size int) Option {
	return func(db *DB) {
		db.config.PoolSize = size
	}
}
