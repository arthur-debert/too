package idm

import "testing"

func TestResolvePositionPath(t *testing.T) {
	r := NewRegistry()
	r.scopes["root"] = []string{"uid1", "uid2"}
	r.scopes["uid1"] = []string{"uid1-1", "uid1-2", "uid1-3"}
	r.scopes["uid2"] = []string{"uid2-1"}
	r.scopes["uid1-2"] = []string{"uid1-2-1"}

	testCases := []struct {
		name      string
		path      string
		startScope string
		expectedUID string
		expectErr bool
	}{
		{"Root level", "2", "root", "uid2", false},
		{"Nested level", "1.3", "root", "uid1-3", false},
		{"Deeply nested", "1.2.1", "root", "uid1-2-1", false},
		{"Invalid path (non-numeric)", "1.a", "root", "", true},
		{"Invalid path (zero)", "1.0", "root", "", true},
		{"Not found (bad root)", "3", "root", "", true},
		{"Not found (bad nest)", "1.4", "root", "", true},
		{"Not found (too deep)", "1.2.2", "root", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			uid, err := r.ResolvePositionPath(tc.startScope, tc.path)

			if tc.expectErr {
				if err == nil {
					t.Errorf("Expected an error for path '%s', but got nil", tc.path)
				}
			} else {
				if err != nil {
					t.Errorf("Got unexpected error for path '%s': %v", tc.path, err)
				}
				if uid != tc.expectedUID {
					t.Errorf("Expected UID '%s' for path '%s', but got '%s'", tc.expectedUID, tc.path, uid)
				}
			}
		})
	}
}
