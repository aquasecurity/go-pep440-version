package version

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/xerrors"
)

var (
	specifierOperators = map[string]operatorFunc{
		"":    specifierEqual, // not defined in PEP 440
		"=":   specifierEqual, // not defined in PEP 440
		"==":  specifierEqual,
		"!=":  specifierNotEqual,
		">":   specifierGreaterThan,
		"<":   specifierLessThan,
		">=":  specifierGreaterThanEqual,
		"<=":  specifierLessThanEqual,
		"~=":  specifierCompatible,
		"===": specifierArbitrary,
	}

	specifierRegexp       *regexp.Regexp
	validConstraintRegexp *regexp.Regexp
	prefixRegexp          *regexp.Regexp
)

func init() {
	ops := make([]string, 0, len(specifierOperators))
	for k := range specifierOperators {
		ops = append(ops, regexp.QuoteMeta(k))
	}

	specifierRegexp = regexp.MustCompile(fmt.Sprintf(
		`(?i)(?P<operator>(%s))\s*(?P<version>%s(\.\*)?)`,
		strings.Join(ops, "|"), regex))

	validConstraintRegexp = regexp.MustCompile(fmt.Sprintf(
		`^\s*(\s*(%s)\s*(%s(\.\*)?)\s*\,?)*\s*$`,
		strings.Join(ops, "|"), regex))

	prefixRegexp = regexp.MustCompile(`^([0-9]+)((?:a|b|c|rc)[0-9]+)$`)
}

type operatorFunc func(v Version, c string, allowLocalVersion bool) bool

type Specifiers struct {
	specifiers [][]specifier
	conf       conf
}

type specifier struct {
	version           string
	operator          operatorFunc
	original          string
	allowLocalVersion bool
}

// NewSpecifiers parses a given specifier and returns a new instance of Specifiers
func NewSpecifiers(v string, opts ...SpecifierOption) (Specifiers, error) {
	c := new(conf)

	// Apply options
	for _, o := range opts {
		o.apply(c)
	}

	var sss [][]specifier
	for _, vv := range strings.Split(v, "||") {
		if strings.TrimSpace(vv) == "*" {
			vv = ">=0.0.0"
		}

		// Validate the segment
		if !validConstraintRegexp.MatchString(vv) {
			return Specifiers{}, xerrors.Errorf("improper constraint: %s", vv)
		}

		ss := specifierRegexp.FindAllString(vv, -1)
		if ss == nil {
			ss = append(ss, strings.TrimSpace(vv))
		}

		var specs []specifier
		for _, single := range ss {
			s, err := newSpecifier(single, c)
			if err != nil {
				return Specifiers{}, err
			}
			specs = append(specs, s)
		}
		sss = append(sss, specs)
	}

	return Specifiers{
		specifiers: sss,
		conf:       *c,
	}, nil

}

func newSpecifier(s string, c *conf) (specifier, error) {
	m := specifierRegexp.FindStringSubmatch(s)
	if m == nil {
		return specifier{}, xerrors.Errorf("improper specifier: %s", s)
	}

	operator := m[specifierRegexp.SubexpIndex("operator")]
	version := m[specifierRegexp.SubexpIndex("version")]

	if operator != "===" {
		if err := validate(operator, version, c); err != nil {
			return specifier{}, err
		}
	}

	return specifier{
		version:           version,
		operator:          specifierOperators[operator],
		original:          s,
		allowLocalVersion: c != nil && c.allowLocalVersion,
	}, nil
}

func validate(operator, version string, c *conf) error {
	hasWildcard := false
	if strings.HasSuffix(version, ".*") {
		hasWildcard = true
		version = strings.TrimSuffix(version, ".*")
	}
	v, err := Parse(version)
	if err != nil {
		return xerrors.Errorf("version parse error (%s): %w", v, err)
	}

	allowLocal := c != nil && c.allowLocalVersion

	switch operator {
	case "", "=", "==", "!=":
		if hasWildcard && (!v.dev.isNull() || v.local != "") {
			return xerrors.New("the (non)equality operators don't allow to use a wild card and a dev" +
				" or local version together")
		}
	case "~=":
		if hasWildcard {
			return xerrors.New("a wild card is not allowed")
		} else if len(v.release) < 2 {
			return xerrors.New("the compatible operator requires at least two digits in the release segment")
		} else if v.local != "" && !allowLocal {
			return xerrors.New("local versions cannot be specified")
		}
	default:
		if hasWildcard {
			return xerrors.New("a wild card is not allowed")
		} else if v.local != "" && !allowLocal {
			return xerrors.New("local versions cannot be specified")
		}
	}
	return nil
}

// Check tests if a version satisfies all the specifiers.
func (ss Specifiers) Check(v Version) bool {
	if ss.conf.includePreRelease {
		v.preReleaseIncluded = true
	}

	for _, s := range ss.specifiers {
		if andCheck(v, s) {
			return true
		}
	}

	return false
}

func (s specifier) check(v Version) bool {
	return s.operator(v, s.version, s.allowLocalVersion)
}

func (s specifier) String() string {
	return s.original
}

// String returns the string format of the specifiers
func (ss Specifiers) String() string {
	var ssStr []string
	for _, orS := range ss.specifiers {
		var sstr []string
		for _, andS := range orS {
			sstr = append(sstr, andS.String())
		}
		ssStr = append(ssStr, strings.Join(sstr, ","))
	}

	return strings.Join(ssStr, "||")
}

func andCheck(v Version, specifiers []specifier) bool {
	for _, c := range specifiers {
		if !c.check(v) {
			return false
		}
	}
	return true
}

func versionSplit(version string) []string {
	var result []string
	for _, v := range strings.Split(version, ".") {
		m := prefixRegexp.FindStringSubmatch(v)
		if m != nil {
			result = append(result, m[1:]...)
		} else {
			result = append(result, v)
		}
	}
	return result
}

func isDigist(s string) bool {
	if _, err := strconv.Atoi(s); err == nil {
		return true
	}
	return false
}

func padVersion(left, right []string) ([]string, []string) {
	var leftRelease, rightRelease []string
	for _, l := range left {
		if isDigist(l) {
			leftRelease = append(leftRelease, l)
		}
	}

	for _, r := range right {
		if isDigist(r) {
			rightRelease = append(rightRelease, r)
		}
	}

	// Get the rest of our versions
	leftRest := left[len(leftRelease):]
	rightRest := left[len(rightRelease):]

	for i := 0; i < len(leftRelease)-len(rightRelease); i++ {
		rightRelease = append(rightRelease, "0")
	}
	for i := 0; i < len(rightRelease)-len(leftRelease); i++ {
		leftRelease = append(leftRelease, "0")
	}

	return append(leftRelease, leftRest...), append(rightRelease, rightRest...)
}

//-------------------------------------------------------------------
// Specifier functions
//-------------------------------------------------------------------

func specifierCompatible(prospective Version, spec string, allowLocalVersion bool) bool {
	// Compatible releases have an equivalent combination of >= and ==. That is that ~=2.2 is equivalent to >=2.2,==2.*.
	// This allows us to implement this in terms of the other specifiers instead of implementing it ourselves.
	// The only thing we need to do is construct the other specifiers.

	// For prefix matching, we need to use the public version (without local segment).
	// The local segment is only relevant for the >= comparison.
	specVersion := MustParse(spec)
	publicSpec := specVersion.Public()

	var prefixElements []string
	for _, s := range versionSplit(publicSpec) {
		if strings.HasPrefix(s, "post") || strings.HasPrefix(s, "dev") {
			break
		}
		prefixElements = append(prefixElements, s)
	}

	// We want everything but the last item in the version, but we want to ignore post and dev releases and
	// we want to treat the pre-release as it's own separate segment.
	prefix := strings.Join(prefixElements[:len(prefixElements)-1], ".")

	// Add the prefix notation to the end of our string
	prefix += ".*"

	return specifierGreaterThanEqual(prospective, spec, allowLocalVersion) && specifierEqual(prospective, prefix, allowLocalVersion)
}

func specifierEqual(prospective Version, spec string, allowLocalVersion bool) bool {
	// https://github.com/pypa/packaging/blob/a6407e3a7e19bd979e93f58cfc7f6641a7378c46/packaging/specifiers.py#L476
	// We need special logic to handle prefix matching
	if strings.HasSuffix(spec, ".*") {
		// In the case of prefix matching we want to ignore local segment.
		prospective = MustParse(prospective.Public())

		// Split the spec out by dots, and pretend that there is an implicit
		// dot in between a release segment and a pre-release segment.
		splitSpec := versionSplit(strings.TrimSuffix(spec, ".*"))

		// Split the prospective version out by dots, and pretend that there is an implicit dot
		//  in between a release segment and a pre-release segment.
		splitProspective := versionSplit(prospective.String())

		// Shorten the prospective version to be the same length as the spec
		// so that we can determine if the specifier is a prefix of the
		// prospective version or not.
		if len(splitProspective) > len(splitSpec) {
			splitProspective = splitProspective[:len(splitSpec)]
		}

		paddedSpec, paddedProspective := padVersion(splitSpec, splitProspective)
		return reflect.DeepEqual(paddedSpec, paddedProspective)
	}

	specVersion := MustParse(spec)
	// When allowLocalVersion is disabled and spec has no local, strip local from prospective (PEP 440 default).
	// When allowLocalVersion is enabled, keep local for strict comparison.
	if !allowLocalVersion && specVersion.local == "" {
		prospective = MustParse(prospective.Public())
	}

	return specVersion.Equal(prospective)
}

func specifierNotEqual(prospective Version, spec string, allowLocalVersion bool) bool {
	return !specifierEqual(prospective, spec, allowLocalVersion)
}

func specifierLessThan(prospective Version, spec string, allowLocalVersion bool) bool {
	// Convert our spec to a Version instance, since we'll want to work with it as a version.
	s := MustParse(spec)

	// When allowLocalVersion is disabled, strip local from prospective for comparison (like packaging does).
	if !allowLocalVersion && prospective.local != "" {
		prospective = MustParse(prospective.Public())
	}

	// Check to see if the prospective version is less than the spec version.
	// If it's not we can short circuit and just return False now instead of doing extra unneeded work.
	if !prospective.LessThan(s) {
		return false
	}

	// This special case is here so that, unless the specifier itself includes is a pre-release version,
	// that we do not accept pre-release versions for the version mentioned in the specifier
	// (e.g. <3.1 should not match 3.1.dev0, but should match 3.0.dev0).
	if !s.IsPreRelease() && prospective.IsPreRelease() {
		if MustParse(prospective.BaseVersion()).Equal(MustParse(s.BaseVersion())) {
			return false
		}
	}
	return true
}

func specifierGreaterThan(prospective Version, spec string, allowLocalVersion bool) bool {
	// Convert our spec to a Version instance, since we'll want to work with it as a version.
	s := MustParse(spec)

	// Check to see if the prospective version is greater than the spec version.
	// If it's not we can short circuit and just return False now instead of doing extra unneeded work.
	if !prospective.GreaterThan(s) {
		return false
	}

	// This special case is here so that, unless the specifier itself includes is a post-release version,
	// that we do not accept post-release versions for the version mentioned in the specifier
	// (e.g. >3.1 should not match 3.0.post0, but should match 3.2.post0).
	if !s.IsPostRelease() && prospective.IsPostRelease() {
		if MustParse(prospective.BaseVersion()).Equal(MustParse(s.BaseVersion())) {
			return false
		}
	}

	// Ensure that we do not allow a local version of the version mentioned
	// in the specifier, which is technically greater than, to match.
	// Skip this check when allowLocalVersion is enabled since we want to compare local versions.
	if !allowLocalVersion && prospective.local != "" {
		if MustParse(prospective.BaseVersion()).Equal(MustParse(s.BaseVersion())) {
			return false
		}
	}
	return true
}

func specifierArbitrary(prospective Version, spec string, _ bool) bool {
	return strings.EqualFold(prospective.String(), spec)
}

func specifierLessThanEqual(prospective Version, spec string, allowLocalVersion bool) bool {
	p := prospective
	if !allowLocalVersion {
		p = MustParse(prospective.Public())
	}
	s := MustParse(spec)
	return p.LessThanOrEqual(s)
}

func specifierGreaterThanEqual(prospective Version, spec string, allowLocalVersion bool) bool {
	p := prospective
	if !allowLocalVersion {
		p = MustParse(prospective.Public())
	}
	s := MustParse(spec)
	return p.GreaterThanOrEqual(s)
}
