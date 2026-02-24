package version_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aquasecurity/go-pep440-version"
)

var (
	versions = []string{
		// Implicit epoch of 0
		"1.0.dev456",
		"1.0a1",
		"1.0a2.dev456",
		"1.0a12.dev456",
		"1.0a12",
		"1.0b1.dev456",
		"1.0b2",
		"1.0b2.post345.dev456",
		"1.0b2.post345",
		"1.0b2-346",
		"1.0c1.dev456",
		"1.0c1",
		"1.0rc2",
		"1.0c3",
		"1.0",
		"1.0.post456.dev34",
		"1.0.post456",
		"1.1.dev1",
		"1.2+123abc",
		"1.2+123abc456",
		"1.2+abc",
		"1.2+abc123",
		"1.2+abc123def",
		"1.2+1234.abc",
		"1.2+123456",
		"1.2.r32+123456",
		"1.2.rev33+123456",
		// Explicit epoch of 1
		"1!1.0.dev456",
		"1!1.0a1",
		"1!1.0a2.dev456",
		"1!1.0a12.dev456",
		"1!1.0a12",
		"1!1.0b1.dev456",
		"1!1.0b2",
		"1!1.0b2.post345.dev456",
		"1!1.0b2.post345",
		"1!1.0b2-346",
		"1!1.0c1.dev456",
		"1!1.0c1",
		"1!1.0rc2",
		"1!1.0c3",
		"1!1.0",
		"1!1.0.post456.dev34",
		"1!1.0.post456",
		"1!1.1.dev1",
		"1!1.2+123abc",
		"1!1.2+123abc456",
		"1!1.2+abc",
		"1!1.2+abc123",
		"1!1.2+abc123def",
		"1!1.2+1234.abc",
		"1!1.2+123456",
		"1!1.2.r32+123456",
		"1!1.2.rev33+123456",
	}
)

// https://github.com/pypa/packaging/blob/a6407e3a7e19bd979e93f58cfc7f6641a7378c46/tests/test_version.py#L85-L87
func TestParseValidVersion(t *testing.T) {
	for _, v := range versions {
		t.Run(v, func(t *testing.T) {
			_, err := version.Parse(v)
			assert.NoError(t, err)
		})
	}
}

// https://github.com/pypa/packaging/blob/a6407e3a7e19bd979e93f58cfc7f6641a7378c46/tests/test_version.py#L102-L104
func TestParseInvalidVersion(t *testing.T) {
	versions := []string{
		// Non sensical versions should be invalid
		"french toast",
		// Versions with invalid local versions
		"1.0+a+",
		"1.0++",
		"1.0+_foobar",
		"1.0+foo&asd",
		"1.0+1+1",
	}
	for _, v := range versions {
		t.Run(v, func(t *testing.T) {
			_, err := version.Parse(v)
			assert.Error(t, err)
		})
	}
}

func TestVersion_String(t *testing.T) {
	tests := []struct {
		version string
		want    string
	}{
		// Various development release incarnations
		{"1.0dev", "1.0.dev0"},
		{"1.0.dev", "1.0.dev0"},
		{"1.0dev1", "1.0.dev1"},
		{"1.0-dev", "1.0.dev0"},
		{"1.0-dev1", "1.0.dev1"},
		{"1.0DEV", "1.0.dev0"},
		{"1.0.DEV", "1.0.dev0"},
		{"1.0DEV1", "1.0.dev1"},
		{"1.0DEV", "1.0.dev0"},
		{"1.0.DEV1", "1.0.dev1"},
		{"1.0-DEV", "1.0.dev0"},
		{"1.0-DEV1", "1.0.dev1"},
		// Various alpha incarnations
		{"1.0a", "1.0a0"},
		{"1.0.a", "1.0a0"},
		{"1.0.a1", "1.0a1"},
		{"1.0-a", "1.0a0"},
		{"1.0-a1", "1.0a1"},
		{"1.0alpha", "1.0a0"},
		{"1.0.alpha", "1.0a0"},
		{"1.0.alpha1", "1.0a1"},
		{"1.0-alpha", "1.0a0"},
		{"1.0-alpha1", "1.0a1"},
		{"1.0A", "1.0a0"},
		{"1.0.A", "1.0a0"},
		{"1.0.A1", "1.0a1"},
		{"1.0-A", "1.0a0"},
		{"1.0-A1", "1.0a1"},
		{"1.0ALPHA", "1.0a0"},
		{"1.0.ALPHA", "1.0a0"},
		{"1.0.ALPHA1", "1.0a1"},
		{"1.0-ALPHA", "1.0a0"},
		{"1.0-ALPHA1", "1.0a1"},
		// Various beta incarnations
		{"1.0b", "1.0b0"},
		{"1.0.b", "1.0b0"},
		{"1.0.b1", "1.0b1"},
		{"1.0-b", "1.0b0"},
		{"1.0-b1", "1.0b1"},
		{"1.0beta", "1.0b0"},
		{"1.0.beta", "1.0b0"},
		{"1.0.beta1", "1.0b1"},
		{"1.0-beta", "1.0b0"},
		{"1.0-beta1", "1.0b1"},
		{"1.0B", "1.0b0"},
		{"1.0.B", "1.0b0"},
		{"1.0.B1", "1.0b1"},
		{"1.0-B", "1.0b0"},
		{"1.0-B1", "1.0b1"},
		{"1.0BETA", "1.0b0"},
		{"1.0.BETA", "1.0b0"},
		{"1.0.BETA1", "1.0b1"},
		{"1.0-BETA", "1.0b0"},
		{"1.0-BETA1", "1.0b1"},
		// Various release candidate incarnations
		{"1.0c", "1.0rc0"},
		{"1.0.c", "1.0rc0"},
		{"1.0.c1", "1.0rc1"},
		{"1.0-c", "1.0rc0"},
		{"1.0-c1", "1.0rc1"},
		{"1.0rc", "1.0rc0"},
		{"1.0.rc", "1.0rc0"},
		{"1.0.rc1", "1.0rc1"},
		{"1.0-rc", "1.0rc0"},
		{"1.0-rc1", "1.0rc1"},
		{"1.0C", "1.0rc0"},
		{"1.0.C", "1.0rc0"},
		{"1.0.C1", "1.0rc1"},
		{"1.0-C", "1.0rc0"},
		{"1.0-C1", "1.0rc1"},
		{"1.0RC", "1.0rc0"},
		{"1.0.RC", "1.0rc0"},
		{"1.0.RC1", "1.0rc1"},
		{"1.0-RC", "1.0rc0"},
		{"1.0-RC1", "1.0rc1"},
		// Various post release incarnations
		{"1.0post", "1.0.post0"},
		{"1.0.post", "1.0.post0"},
		{"1.0post1", "1.0.post1"},
		{"1.0post", "1.0.post0"},
		{"1.0-post", "1.0.post0"},
		{"1.0-post1", "1.0.post1"},
		{"1.0POST", "1.0.post0"},
		{"1.0.POST", "1.0.post0"},
		{"1.0POST1", "1.0.post1"},
		{"1.0POST", "1.0.post0"},
		{"1.0r", "1.0.post0"},
		{"1.0rev", "1.0.post0"},
		{"1.0.POST1", "1.0.post1"},
		{"1.0.r1", "1.0.post1"},
		{"1.0.rev1", "1.0.post1"},
		{"1.0-POST", "1.0.post0"},
		{"1.0-POST1", "1.0.post1"},
		{"1.0-5", "1.0.post5"},
		{"1.0-r5", "1.0.post5"},
		{"1.0-rev5", "1.0.post5"},
		// Local version case insensitivity
		{"1.0+AbC", "1.0+abc"},
		// Integer Normalization
		{"1.01", "1.1"},
		{"1.0a05", "1.0a5"},
		{"1.0b07", "1.0b7"},
		{"1.0c056", "1.0rc56"},
		{"1.0rc09", "1.0rc9"},
		{"1.0.post000", "1.0.post0"},
		{"1.1.dev09000", "1.1.dev9000"},
		{"00!1.2", "1.2"},
		{"0100!0.0", "100!0.0"},
		// Various other normalizations
		{"v1.0", "1.0"},
		{"   v1.0\t\n", "1.0"},
	}
	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			v, err := version.Parse(tt.version)
			require.NoError(t, err)

			assert.Equal(t, tt.want, v.String())
		})
	}
	t.Run("Zero Value", func(t *testing.T) {
		v := version.Version{}
		assert.Equal(t, "", v.String())
	})
}

func TestVersion_LessThan_LessThanOrEqual(t *testing.T) {
	var tests [][2]string
	for i, v1 := range versions {
		for _, v2 := range versions[i+1:] {
			tests = append(tests, [2]string{v1, v2})
		}
	}
	for _, tt := range tests {
		t.Run(tt[0]+" < "+tt[1], func(t *testing.T) {
			v1, v2 := parseVersions(t, tt[0], tt[1])
			assert.True(t, v1.LessThan(v2))
		})
		t.Run(tt[0]+" >= "+tt[1], func(t *testing.T) {
			v1, v2 := parseVersions(t, tt[0], tt[1])
			assert.False(t, v1.GreaterThanOrEqual(v2))
		})
	}
}

func TestVersion_LessThanOrEqual(t *testing.T) {
	var tests [][2]string
	for i, v1 := range versions {
		for _, v2 := range versions[i:] {
			tests = append(tests, [2]string{v1, v2})
		}
	}
	for _, tt := range tests {
		t.Run(tt[0]+" <= "+tt[1], func(t *testing.T) {
			v1, v2 := parseVersions(t, tt[0], tt[1])
			assert.True(t, v1.LessThanOrEqual(v2))
		})
		t.Run(tt[0]+" > "+tt[1], func(t *testing.T) {
			v1, v2 := parseVersions(t, tt[0], tt[1])
			assert.False(t, v1.GreaterThan(v2))
		})
	}
}

func TestVersion_Equal(t *testing.T) {
	var tests [][2]string
	for _, v := range versions {
		tests = append(tests, [2]string{v, v})
	}
	for _, tt := range tests {
		t.Run(tt[0]+" = "+tt[1], func(t *testing.T) {
			v1, v2 := parseVersions(t, tt[0], tt[1])
			assert.True(t, v1.Equal(v2))
		})
	}
}

func TestVersion_GreaterThan(t *testing.T) {
	var tests [][2]string
	for i, v1 := range versions {
		for _, v2 := range versions[:i] {
			tests = append(tests, [2]string{v1, v2})
		}
	}
	for _, tt := range tests {
		t.Run(tt[0]+" > "+tt[1], func(t *testing.T) {
			v1, v2 := parseVersions(t, tt[0], tt[1])
			assert.True(t, v1.GreaterThan(v2))
		})
		t.Run(tt[0]+" <= "+tt[1], func(t *testing.T) {
			v1, v2 := parseVersions(t, tt[0], tt[1])
			assert.False(t, v1.LessThanOrEqual(v2))
		})
	}
}

func TestVersion_GreaterThanOrEqual(t *testing.T) {
	var tests [][2]string
	for i, v1 := range versions {
		for _, v2 := range versions[:i+1] {
			tests = append(tests, [2]string{v1, v2})
		}
	}
	for _, tt := range tests {
		t.Run(tt[0]+" >= "+tt[1], func(t *testing.T) {
			v1, v2 := parseVersions(t, tt[0], tt[1])
			assert.True(t, v1.GreaterThanOrEqual(v2))
		})
		t.Run(tt[0]+" < "+tt[1], func(t *testing.T) {
			v1, v2 := parseVersions(t, tt[0], tt[1])
			assert.False(t, v1.LessThan(v2))
		})
	}
}

func parseVersions(t *testing.T, s1, s2 string) (version.Version, version.Version) {
	t.Helper()

	v1, err := version.Parse(s1)
	require.NoError(t, err)

	v2, err := version.Parse(s2)
	require.NoError(t, err)

	return v1, v2
}

func TestVersion_JSON(t *testing.T) {
	versionString := "1.0.post456.dev34"
	v := version.MustParse(versionString)

	jsonString := fmt.Sprintf(`{"version": "%s"}`, versionString)
	type vs struct {
		Version version.Version `json:"version"`
	}

	t.Run("Unmarshal", func(t *testing.T) {
		fromJson := vs{}
		if err := json.Unmarshal([]byte(jsonString), &fromJson); assert.NoError(t, err) {
			assert.Equal(t, vs{Version: v}, fromJson)
		}
	})

	t.Run("Marshal", func(t *testing.T) {
		if toJson, err := json.Marshal(vs{Version: v}); assert.NoError(t, err) {
			assert.JSONEq(t, jsonString, string(toJson))
		}
	})
}
