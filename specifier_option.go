package version

type conf struct {
	includePreRelease bool
	allowLocalVersion bool
}

type SpecifierOption interface {
	apply(*conf)
}

type WithPreRelease bool

func (o WithPreRelease) apply(c *conf) {
	c.includePreRelease = bool(o)
}

type WithLocalVersion bool

func (o WithLocalVersion) apply(c *conf) {
	c.allowLocalVersion = bool(o)
}
