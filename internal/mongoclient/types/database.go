package types

type Databases struct {
	Databases []Database `tfsdk:"databases"`
}

type Database struct {
	Name string `tfsdk:"name"`
}
