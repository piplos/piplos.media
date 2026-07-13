// Package server wires HTTP routes.
package server

import (
	"github.com/gofiber/fiber/v3"

	"github.com/piplos/site/internal/handlers"
	"github.com/piplos/site/internal/middleware"
	"github.com/piplos/site/internal/models"
)

// Handlers groups all route handlers.
type Handlers struct {
	Auth     *handlers.AuthHandler
	Users    *handlers.UsersHandler
	Content  *handlers.ContentHandler
	Leads    *handlers.LeadsHandler
	Settings *handlers.SettingsHandler
	Public   *handlers.PublicHandler
	Uploads  *handlers.UploadsHandler
	AIModels *handlers.AIModelsHandler
}

// Register mounts all API routes under /api/v1.
//
// Роли: admin — полный доступ; manager — контент и заявки,
// но не пользователи и не настройки (включая языки).
func Register(app *fiber.App, h *Handlers, auth *middleware.Auth) {
	api := app.Group("/api/v1")

	// Public (site) endpoints.
	api.Post("/leads", h.Leads.Create)
	pub := api.Group("/public")
	pub.Get("/projects", h.Public.Projects)
	pub.Get("/services", h.Public.Services)
	pub.Get("/stack", h.Public.Stack)
	pub.Get("/seo", h.Public.SEO)
	pub.Get("/legal", h.Public.Legal)
	pub.Get("/languages", h.Public.Languages)

	// Auth.
	api.Post("/auth/login", h.Auth.Login)
	api.Post("/auth/refresh", h.Auth.Refresh)
	api.Get("/auth/me", auth.RequireAuth(), h.Auth.Me)

	// Content + leads: admin and manager.
	staff := api.Group("", auth.RequireAuth(), auth.RequireRole(models.RoleAdmin, models.RoleManager))

	staff.Get("/projects", h.Content.ListProjects)
	staff.Post("/projects/reorder", h.Content.ReorderProjects)
	staff.Put("/projects/reorder", h.Content.ReorderProjects)
	staff.Post("/projects", h.Content.CreateProject)
	staff.Get("/projects/:slug", h.Content.GetProject)
	staff.Put("/projects/:slug", h.Content.UpdateProject)
	staff.Delete("/projects/:slug", h.Content.DeleteProject)

	staff.Get("/services", h.Content.ListServices)
	staff.Post("/services/reorder", h.Content.ReorderServices)
	staff.Put("/services/reorder", h.Content.ReorderServices)
	staff.Post("/services", h.Content.CreateService)
	staff.Put("/services/:id", h.Content.UpdateService)
	staff.Delete("/services/:id", h.Content.DeleteService)

	staff.Get("/stack", h.Content.ListStack)
	staff.Post("/stack/reorder", h.Content.ReorderStack)
	staff.Put("/stack/reorder", h.Content.ReorderStack)
	staff.Post("/stack", h.Content.CreateStackItem)
	staff.Put("/stack/:id", h.Content.UpdateStackItem)
	staff.Delete("/stack/:id", h.Content.DeleteStackItem)

	staff.Get("/seo", h.Content.ListSEO)
	staff.Post("/seo", h.Content.CreateSEO)
	staff.Put("/seo/:id", h.Content.UpdateSEO)
	staff.Delete("/seo/:id", h.Content.DeleteSEO)

	staff.Get("/legal", h.Content.ListLegal)
	staff.Get("/legal/:id", h.Content.GetLegal)
	staff.Put("/legal/:id", h.Content.UpdateLegal)

	staff.Post("/uploads", h.Uploads.Upload)

	staff.Get("/leads", h.Leads.List)
	staff.Get("/leads/:id", h.Leads.Get)
	staff.Patch("/leads/:id/status", h.Leads.UpdateStatus)
	staff.Delete("/leads/:id", h.Leads.Delete)

	// Языки нужны контент-редакторам для чтения; AI-перевод — тоже.
	staff.Get("/languages", h.Settings.ListLanguages)
	staff.Post("/translate", h.Settings.Translate)

	// Admin-only: users, settings, languages management.
	adm := api.Group("", auth.RequireAuth(), auth.RequireRole(models.RoleAdmin))
	adm.Get("/users", h.Users.List)
	adm.Post("/users", h.Users.Create)
	adm.Put("/users/:id", h.Users.Update)
	adm.Delete("/users/:id", h.Users.Delete)

	adm.Get("/settings", h.Settings.ListSettings)
	adm.Get("/settings/:key", h.Settings.GetSetting)
	adm.Post("/settings/test", h.Settings.TestSetting)
	adm.Post("/settings/test/translation", h.Settings.TestTranslation)
	adm.Put("/settings/:key", h.Settings.UpdateSetting)
	adm.Post("/languages", h.Settings.UpsertLanguage)
	adm.Delete("/languages/:code", h.Settings.DeleteLanguage)

	adm.Get("/ai-models", h.AIModels.ListAIModels)
	adm.Post("/ai-models", h.AIModels.CreateAIModel)
	adm.Put("/ai-models/:id", h.AIModels.UpdateAIModel)
	adm.Delete("/ai-models/:id", h.AIModels.DeleteAIModel)
}
