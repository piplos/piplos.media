package auth_test

import (
	"strings"
	"testing"

	authperms "github.com/piplos/piplos.media/internal/auth"
)

// routeProbe mirrors internal/server/routes_auth_test.go route lists for drift detection.
type routeProbe struct {
	method string
	path   string
	group  string
}

var expectedRouteGroups = []routeProbe{
	// Public auth
	{"POST", "/v1/auth/login", authperms.GroupPublic},
	{"POST", "/v1/auth/refresh", authperms.GroupPublic},
	// Authenticated auth
	{"GET", "/v1/auth/me", authperms.GroupAuthenticated},
	{"POST", "/v1/auth/logout", authperms.GroupAuthenticated},
	{"GET", "/v1/auth/permissions", authperms.GroupAuthenticated},
	// Staff samples
	{"GET", "/v1/projects", authperms.GroupStaff},
	{"POST", "/v1/projects", authperms.GroupStaff},
	{"GET", "/v1/leads", authperms.GroupStaff},
	{"GET", "/v1/languages", authperms.GroupStaff},
	{"POST", "/v1/translate", authperms.GroupStaff},
	// Admin samples
	{"GET", "/v1/users", authperms.GroupAdmin},
	{"POST", "/v1/users", authperms.GroupAdmin},
	{"GET", "/v1/settings", authperms.GroupAdmin},
	{"PUT", "/v1/settings/SMTP", authperms.GroupAdmin},
	{"POST", "/v1/languages", authperms.GroupAdmin},
	{"GET", "/v1/ai-models", authperms.GroupAdmin},
	{"POST", "/v1/ai-models", authperms.GroupAdmin},
	{"GET", "/v1/backups", authperms.GroupAdmin},
}

func matrixMatchesRoute(r authperms.RoutePermission, method, path string) bool {
	if r.Path != path {
		return false
	}
	if r.Method == "*" || r.Method == method {
		return true
	}
	return false
}

func findRouteGroup(method, path string) (string, bool) {
	for _, r := range authperms.APIRoutes {
		if matrixMatchesRoute(r, method, path) {
			return r.Group, true
		}
	}
	// Wildcard prefix match for paths like /v1/public/*
	for _, r := range authperms.APIRoutes {
		prefix := strings.TrimSuffix(r.Path, "/*")
		if strings.HasSuffix(r.Path, "/*") && strings.HasPrefix(path, prefix) {
			if r.Method == "*" || r.Method == method {
				return r.Group, true
			}
		}
	}
	// Sub-resource match: /v1/settings covers /v1/settings/:key
	for _, r := range authperms.APIRoutes {
		if r.Method == "*" || r.Method == method {
			if path == r.Path || strings.HasPrefix(path, r.Path+"/") {
				return r.Group, true
			}
		}
	}
	return "", false
}

func TestAPIRoutesCoverRegisteredEndpoints(t *testing.T) {
	for _, probe := range expectedRouteGroups {
		group, ok := findRouteGroup(probe.method, probe.path)
		if !ok {
			t.Errorf("route %s %s not found in permissions matrix", probe.method, probe.path)
			continue
		}
		if group != probe.group {
			t.Errorf("route %s %s: matrix group %q, want %q", probe.method, probe.path, group, probe.group)
		}
	}
}

func TestStaffAndAdminRolesNonEmpty(t *testing.T) {
	if len(authperms.StaffRoles) == 0 || len(authperms.AdminRoles) == 0 {
		t.Fatal("role lists must not be empty")
	}
	for _, r := range authperms.AdminRoles {
		found := false
		for _, s := range authperms.StaffRoles {
			if r == s {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("admin role %q must be included in staff roles", r)
		}
	}
}
