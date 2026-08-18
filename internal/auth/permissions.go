// Package auth defines the role/permission matrix shared by routes and admin UI.
package auth

import "github.com/piplos/piplos.media/internal/models"

// Permission groups (single source of truth for route access).
const (
	GroupPublic        = "public"
	GroupAuthenticated = "authenticated"
	GroupStaff         = "staff"
	GroupAdmin         = "admin"
)

// StaffRoles may access content, uploads, files, leads, read languages, translate.
var StaffRoles = []models.UserRole{models.RoleAdmin, models.RoleManager}

// AdminRoles may access users, settings, language CRUD, AI models, backups.
var AdminRoles = []models.UserRole{models.RoleAdmin}

// AuthenticatedRoles may call /auth/me and /auth/logout.
var AuthenticatedRoles = StaffRoles

// RoutePermission documents an API endpoint and its required group (for codegen/CI).
type RoutePermission struct {
	Method string
	Path   string
	Group  string
}

// APIRoutes is the canonical permission matrix for /v1 endpoints.
var APIRoutes = []RoutePermission{
	// Public
	{Method: "POST", Path: "/v1/leads", Group: GroupPublic},
	{Method: "GET", Path: "/v1/public/*", Group: GroupPublic},
	{Method: "POST", Path: "/v1/auth/login", Group: GroupPublic},
	{Method: "POST", Path: "/v1/auth/refresh", Group: GroupPublic},
	// Authenticated
	{Method: "GET", Path: "/v1/auth/me", Group: GroupAuthenticated},
	{Method: "POST", Path: "/v1/auth/logout", Group: GroupAuthenticated},
	{Method: "GET", Path: "/v1/auth/permissions", Group: GroupAuthenticated},
	// Staff (content, uploads, files, leads, languages read, translate)
	{Method: "*", Path: "/v1/projects", Group: GroupStaff},
	{Method: "*", Path: "/v1/services", Group: GroupStaff},
	{Method: "*", Path: "/v1/stack", Group: GroupStaff},
	{Method: "*", Path: "/v1/seo", Group: GroupStaff},
	{Method: "*", Path: "/v1/pages", Group: GroupStaff},
	{Method: "*", Path: "/v1/legal", Group: GroupStaff},
	{Method: "*", Path: "/v1/uploads", Group: GroupStaff},
	{Method: "*", Path: "/v1/files", Group: GroupStaff},
	{Method: "*", Path: "/v1/leads", Group: GroupStaff},
	{Method: "GET", Path: "/v1/languages", Group: GroupStaff},
	{Method: "POST", Path: "/v1/translate", Group: GroupStaff},
	// Admin only
	{Method: "*", Path: "/v1/users", Group: GroupAdmin},
	{Method: "*", Path: "/v1/settings", Group: GroupAdmin},
	{Method: "POST", Path: "/v1/languages", Group: GroupAdmin},
	{Method: "DELETE", Path: "/v1/languages/*", Group: GroupAdmin},
	{Method: "*", Path: "/v1/ai-models", Group: GroupAdmin},
	{Method: "*", Path: "/v1/backups", Group: GroupAdmin},
}
