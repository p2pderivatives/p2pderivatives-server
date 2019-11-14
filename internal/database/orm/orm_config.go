package orm

// Config contains the configuration parameter to set up the orm.
type Config struct {
	EnableLogging bool   `configkey:"database.log"`                          // Whether to enable logging of the database
	DbFilePath    string `configkey:"database.filepath" validate:"required"` // The path to the file where to store the database
}
