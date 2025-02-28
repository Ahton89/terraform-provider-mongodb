package types

var (
	DefaultDatabases = []string{"admin", "config", "local"} // Default databases to exclude from listing
	DefaultDatabase  = "admin"                              // Default database for making queries
)

type Databases struct {
	Databases []Database `tfsdk:"databases"`
}

type Database struct {
	Name string `tfsdk:"name"`
}
