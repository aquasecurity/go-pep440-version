package version

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConstraints(t *testing.T) {
	tests := []struct {
		constraint string
		wantErr    bool
	}{
		// https://github.com/pypa/packaging/blob/28d2fa0742747cda4bc4530b2a5bc919b7382039/tests/test_specifiers.py#L31-L42
		{"~=2.0", false},
		{"==2.1.*", false},
		{"==2.1.0.3", false},
		{"!=2.2.*", false},
		{"!=2.2.0.5", false},
		{"<=5", false},
		{">=7.9a1", false},
		{"<1.0.dev1", false},
		{">2.0.post1", false},

		// TODO
		// {"===lolwat", false},

		// https://github.com/pypa/packaging/blob/28d2fa0742747cda4bc4530b2a5bc919b7382039/tests/test_specifiers.py#L50-L86
		// Operator-less specifier
		{"2.0", false}, // go-pep-440-version permits this case

		// Invalid operator
		{"=>2.0", true},

		//Version-less specifier
		{"==", true},

		// Local segment on operators which don't support them
		{"~=1.0+5", true},
		{">=1.0+deadbeef", true},
		{"<=1.0+abc123", true},
		{">1.0+watwat", true},
		{"<1.0+1.0", true},

		// Prefix matching on operators which don't support them
		{"~=1.0.*", true},
		{">=1.0.*", true},
		{"<=1.0.*", true},
		{">1.0.*", true},
		{"<1.0.*", true},

		// Combination of local and prefix matching on operators which do
		// support one or the other
		{"==1.0.*+5", true},
		{"!=1.0.*+deadbeef", true},

		// Prefix matching cannot be used inside of a local version
		{"==1.0+5.*", true},
		{"!=1.0+deadbeef.*", true},

		// Prefix matching must appear at the end
		{"==1.0.*.5", true},

		// Compatible operator requires 2 digits in the release operator
		{"~=1", true},

		// Cannot use a prefix matching after a .devN version
		{"==1.0.dev1.*", true},
		{"!=1.0.dev1.*", true},
	}
	for _, tt := range tests {
		t.Run(tt.constraint, func(t *testing.T) {
			_, err := NewSpecifiers(tt.constraint)
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVersion_Check(t *testing.T) {
	tests := []struct {
		version string
		spec    string
		want    bool
	}{
		// Test the equality operation
		{"2.0", "==2", true},
		{"2.0", "==2.0", true},
		{"2.0", "==2.0.0", true},
		{"2.0+deadbeef", "==2", true},
		{"2.0+deadbeef", "==2.0", true},
		{"2.0+deadbeef", "==2.0.0", true},
		{"2.0+deadbeef", "==2+deadbeef", true},
		{"2.0+deadbeef", "==2.0+deadbeef", true},
		{"2.0+deadbeef", "==2.0.0+deadbeef", true},
		{"2.0+deadbeef.0", "==2.0.0+deadbeef.00", true},

		// Test the equality operation with a prefix
		{"2.dev1", "==2.*", true},
		{"2a1", "==2.*", true},
		{"2a1.post1", "==2.*", true},
		{"2b1", "==2.*", true},
		{"2b1.dev1", "==2.*", true},
		{"2c1", "==2.*", true},
		{"2c1.post1.dev1", "==2.*", true},
		{"2rc1", "==2.*", true},
		{"2", "==2.*", true},
		{"2.0", "==2.*", true},
		{"2.0.0", "==2.*", true},
		{"2.0.post1", "==2.0.post1.*", true},
		{"2.0.post1.dev1", "==2.0.post1.*", true},
		{"2.1+local.version", "==2.1.*", true},

		// Test the in-equality operation
		{"2.1", "!=2", true},
		{"2.1", "!=2.0", true},
		{"2.0.1", "!=2", true},
		{"2.0.1", "!=2.0", true},
		{"2.0.1", "!=2.0.0", true},
		{"2.0", "!=2.0+deadbeef", true},

		// Test the in-equality operation with a prefix
		{"2.0", "!=3.*", true},
		{"2.1", "!=2.0.*", true},

		// Test the greater than equal operation
		{"2.0", ">=2", true},
		{"2.0", ">=2.0", true},
		{"2.0", ">=2.0.0", true},
		{"2.0.post1", ">=2", true},
		{"2.0.post1.dev1", ">=2", true},
		{"3", ">=2", true},

		// Test the less than equal operation
		{"2.0", "<=2", true},
		{"2.0", "<=2.0", true},
		{"2.0", "<=2.0.0", true},
		{"2.0.dev1", "<=2", true},
		{"2.0a1", "<=2", true},
		{"2.0a1.dev1", "<=2", true},
		{"2.0b1", "<=2", true},
		{"2.0b1.post1", "<=2", true},
		{"2.0c1", "<=2", true},
		{"2.0c1.post1.dev1", "<=2", true},
		{"2.0rc1", "<=2", true},
		{"1", "<=2", true},

		// Test the greater than operation
		{"3", ">2", true},
		{"2.1", ">2.0", true},
		{"2.0.1", ">2", true},
		{"2.1.post1", ">2", true},
		{"2.1+local.version", ">2", true},

		// Test the less than operation
		{"1", "<2", true},
		{"2.0", "<2.1", true},
		{"2.0.dev0", "<2.1", true},

		// Test the compatibility operation
		{"1", "~=1.0", true},
		{"1.0.1", "~=1.0", true},
		{"1.1", "~=1.0", true},
		{"1.9999999", "~=1.0", true},

		// Test that epochs are handled sanely
		{"2!1.0", "~=2!1.0", true},
		{"2!1.0", "==2!1.*", true},
		{"2!1.0", "==2!1.0", true},
		{"2!1.0", "!=1.0", true},
		{"1.0", "!=2!1.0", true},
		{"1.0", "<=2!0.1", true},
		{"2!1.0", ">=2.0", true},
		{"1.0", "<2!0.1", true},
		{"2!1.0", ">2.0", true},

		// Test some normalization rules
		{"2.0.5", ">2.0dev", true},

		// Test the equality operation
		{"2.1", "==2", false},
		{"2.1", "==2.0", false},
		{"2.1", "==2.0.0", false},
		{"2.0", "==2.0+deadbeef", false},

		//Test the equality operation with a prefix
		{"2.0", "==3.*", false},
		{"2.1", "==2.0.*", false},

		// Test the in-equality operation
		{"2.0", "!=2", false},
		{"2.0", "!=2.0", false},
		{"2.0", "!=2.0.0", false},
		{"2.0+deadbeef", "!=2", false},
		{"2.0+deadbeef", "!=2.0", false},
		{"2.0+deadbeef", "!=2.0.0", false},
		{"2.0+deadbeef", "!=2+deadbeef", false},
		{"2.0+deadbeef", "!=2.0+deadbeef", false},
		{"2.0+deadbeef", "!=2.0.0+deadbeef", false},
		{"2.0+deadbeef.0", "!=2.0.0+deadbeef.00", false},

		// Test the in-equality operation with a prefix
		{"2.dev1", "!=2.*", false},
		{"2a1", "!=2.*", false},
		{"2a1.post1", "!=2.*", false},
		{"2b1", "!=2.*", false},
		{"2b1.dev1", "!=2.*", false},
		{"2c1", "!=2.*", false},
		{"2c1.post1.dev1", "!=2.*", false},
		{"2rc1", "!=2.*", false},
		{"2", "!=2.*", false},
		{"2.0", "!=2.*", false},
		{"2.0.0", "!=2.*", false},
		{"2.0.post1", "!=2.0.post1.*", false},
		{"2.0.post1.dev1", "!=2.0.post1.*", false},

		//Test the greater than equal operation
		{"2.0.dev1", ">=2", false},
		{"2.0a1", ">=2", false},
		{"2.0a1.dev1", ">=2", false},
		{"2.0b1", ">=2", false},
		{"2.0b1.post1", ">=2", false},
		{"2.0c1", ">=2", false},
		{"2.0c1.post1.dev1", ">=2", false},
		{"2.0rc1", ">=2", false},
		{"1", ">=2", false},

		// Test the less than equal operation
		{"2.0.post1", "<=2", false},
		{"2.0.post1.dev1", "<=2", false},
		{"3", "<=2", false},

		// Test the greater than operation
		{"1", ">2", false},
		{"2.0.dev1", ">2", false},
		{"2.0a1", ">2", false},
		{"2.0a1.post1", ">2", false},
		{"2.0b1", ">2", false},
		{"2.0b1.dev1", ">2", false},
		{"2.0c1", ">2", false},
		{"2.0c1.post1.dev1", ">2", false},
		{"2.0rc1", ">2", false},
		{"2.0", ">2", false},
		{"2.0.post1", ">2", false},
		{"2.0.post1.dev1", ">2", false},
		{"2.0+local.version", ">2", false},

		// Test the less than operation
		{"2.0.dev1", "<2", false},
		{"2.0a1", "<2", false},
		{"2.0a1.post1", "<2", false},
		{"2.0b1", "<2", false},
		{"2.0b2.dev1", "<2", false},
		{"2.0c1", "<2", false},
		{"2.0c1.post1.dev1", "<2", false},
		{"2.0rc1", "<2", false},
		{"2.0", "<2", false},
		{"2.post1", "<2", false},
		{"2.post1.dev1", "<2", false},
		{"3", "<2", false},

		// Test the compatibility operation
		{"2.0", "~=1.0", false},
		{"1.1.0", "~=1.0.0", false},
		{"1.1.post1", "~=1.0.0", false},

		// Test that epochs are handled sanely
		{"1.0", "~=2!1.0", false},
		{"2!1.0", "~=1.0", false},
		{"2!1.0", "==1.0", false},
		{"1.0", "==2!1.0", false},
		{"2!1.0", "==1.*", false},
		{"1.0", "==2!1.*", false},
		{"2!1.0", "!=2!1.0", false},

		// local versions
		{"1.0.0+local", "==1.0.0", true},
		{"1.0.0+local", "!=1.0.0", false},
		{"1.0.0+local", "<=1.0.0", true},
		{"1.0.0+local", ">=1.0.0", true},
		{"1.0.0+local", "<1.0.0", false},
		{"1.0.0+local", ">1.0.0", false},

		// and operators
		{"1.0", ">= 1.0, != 1.3.4.*, < 2.0", true},
		{"1.0", "~= 0.9, >= 1.0, != 1.3.4.*, < 2.0", false},
		{"0.9", "~= 0.9, != 1.3.4.*, < 2.0", true},
		{"2.0", ">= 1.0, != 1.3.4.*, < 2.0", false},
		{"1.3.4", ">= 1.0, != 1.3.4.*, < 2.0", false},

		// or operators
		{"1.0", "~= 0.9, >= 1.0, != 1.3.4.*, < 2.0 || ==1.0", true},
		{"1.0", "~= 0.9, >= 1.0, != 1.3.4.*, < 2.0 || !=1.0", false},

		// Test the equality operation not defined in PEP 440
		{"2.0", "2", true},
		{"2.0", "2.0", true},
		{"2.0", "2.0.0", true},
		{"2.1", "2", false},
		{"2.1", "2.0", false},
		{"2.1", "2.0.0", false},
		{"2.0", "2.0+deadbeef", false},
		{"2.0", "=2", true},
		{"2.0", "=2.0", true},
		{"2.0", "=2.0.0", true},
		{"2.1", "=2", false},
		{"2.1", "=2.0", false},
		{"2.1", "=2.0.0", false},
		{"2.0", "=2.0+deadbeef", false},
		{"2.0", "*", true},

		// space separated
		{"1.0", ">= 1.0 != 1.3.4.* < 2.0", true},
		{"1.0", "~= 0.9 >= 1.0 != 1.3.4.* < 2.0", false},
		{"0.9", "~= 0.9 != 1.3.4.* < 2.0", true},
		{"2.0", ">= 1.0 != 1.3.4.* < 2.0", false},
		{"1.3.4", ">= 1.0 != 1.3.4.* < 2.0", false},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s %s", tt.version, tt.spec), func(t *testing.T) {
			c, err := NewSpecifiers(tt.spec)
			require.NoError(t, err)

			v, err := Parse(tt.version)
			require.NoError(t, err)

			assert.Equal(t, tt.want, c.Check(v))
		})
	}
}

func TestVersion_CheckWithPreRelease(t *testing.T) {
	tests := []struct {
		version string
		spec    string
		want    bool
	}{
		{"1.3.4", "< 2.0", true},
		{"2.0a1", "<2", true},
		{"2.1a1", "<2", false},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s %s", tt.version, tt.spec), func(t *testing.T) {
			c, err := NewSpecifiers(tt.spec, WithPreRelease(true))
			require.NoError(t, err)

			v, err := Parse(tt.version)
			require.NoError(t, err)

			assert.Equal(t, tt.want, c.Check(v))
		})
	}
}
