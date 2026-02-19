package version

type conf struct {
	includePreRelease   bool
	allowLocalSpecifier bool
}

type SpecifierOption interface {
	apply(*conf)
}

type WithPreRelease bool

func (o WithPreRelease) apply(c *conf) {
	c.includePreRelease = bool(o)
}

// AllowLocalSpecifier is an option that allows local version labels in specifiers.
// By default (PEP 440), local versions are not permitted in specifiers and local version
// labels are ignored when checking if candidate versions match a given specifier.
// When enabled, specifiers like ">= 1.0+local.1" become valid and local version
// segments are compared strictly.
type AllowLocalSpecifier bool

func (o AllowLocalSpecifier) apply(c *conf) {
	c.allowLocalSpecifier = bool(o)
}
