package apispec

import (
	"reflect"

	"github.com/tks/backend/internal/asset"
	"github.com/tks/backend/internal/auth"
	"github.com/tks/backend/internal/carousel"
	"github.com/tks/backend/internal/design"
	"github.com/tks/backend/internal/export"
	"github.com/tks/backend/internal/image"
	"github.com/tks/backend/internal/job"
	"github.com/tks/backend/internal/language"
	"github.com/tks/backend/internal/project"
	"github.com/tks/backend/internal/schema"
)

// Operation is one endpoint in the contract.
type Operation struct {
	Method      string
	Path        string
	OperationID string
	Summary     string
	// Body and Response are Go values whose types define the payloads.
	Body     any
	Response any
	// Query parameters, name -> description.
	Query map[string]string
}

// Envelope shapes that handlers wrap responses in.
type ProjectsResponse struct {
	Projects []project.Project `json:"projects"`
}
type CarouselsResponse struct {
	Carousels []carousel.Carousel `json:"carousels"`
}
type SchemaResponse struct {
	Version int         `json:"version"`
	Form    schema.Form `json:"form"`
}
type CarouselResponse struct {
	Carousel carousel.Carousel `json:"carousel"`
	Job      *job.Job          `json:"job,omitempty"`
}
type ImageCandidatesResponse struct {
	Query      string                  `json:"query"`
	Page       int                     `json:"page,omitempty"`
	Candidates []image.PublicCandidate `json:"candidates"`
}
type SelectImageResponse struct {
	Version int         `json:"version"`
	Asset   asset.Asset `json:"asset"`
}
type SourcesResponse struct {
	Sources []project.Source `json:"sources"`
}
type RegistryResponse struct {
	Fonts     []registryFont      `json:"fonts"`
	Formulas  []registryFormula   `json:"formulas"`
	Presets   []design.Preset     `json:"presets"`
	Languages []language.Language `json:"languages"`
}
type registryFont struct {
	ID         string `json:"id"`
	Family     string `json:"family"`
	CSSStack   string `json:"css_stack"`
	Weights    []int  `json:"weights"`
	Vietnamese bool   `json:"vietnamese"`
}
type registryFormula struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Structure      []string `json:"structure"`
	RecommendedFor string   `json:"recommended_for"`
}
type AuthConfigResponse struct {
	GoogleEnabled   bool `json:"google_enabled"`
	DevLoginEnabled bool `json:"dev_login_enabled"`
}
type DevLoginRequest struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}
type CreateProjectRequest struct {
	Niche    string `json:"niche"`
	Language string `json:"language"`
}
type UpdateProjectRequest struct {
	Name        *string          `json:"name,omitempty"`
	Niche       *string          `json:"niche,omitempty"`
	Description *string          `json:"description,omitempty"`
	Language    *string          `json:"language,omitempty"`
	Brand       *project.Brand   `json:"brand,omitempty"`
	Context     *project.Context `json:"context,omitempty"`
}
type GenerateCarouselRequest struct {
	Inputs map[string]any `json:"inputs"`
	Ratio  string         `json:"ratio"`
}
type GenerateCarouselResponse struct {
	CarouselID string            `json:"carousel_id"`
	Carousel   carousel.Carousel `json:"carousel"`
	JobID      string            `json:"job_id"`
	Status     string            `json:"status"`
}
type UpdateCarouselRequest struct {
	Title  *string `json:"title,omitempty"`
	Status *string `json:"status,omitempty"`
}
type SaveDesignRequest struct {
	Version int           `json:"version"`
	Design  design.Design `json:"design"`
}
type SearchImagesRequest struct {
	Query string `json:"query,omitempty"`
	Page  int    `json:"page,omitempty"`
}
type SelectImageRequest struct {
	CandidateID string `json:"candidate_id"`
	Version     int    `json:"version"`
}
type UpdateCaptionRequest struct {
	Caption  string   `json:"caption"`
	Hashtags []string `json:"hashtags"`
}
type JobResponse struct {
	ID           string           `json:"id"`
	CarouselID   string           `json:"carousel_id"`
	Status       string           `json:"status"`
	Progress     int              `json:"progress"`
	CurrentStep  string           `json:"current_step"`
	ErrorMessage string           `json:"error_message,omitempty"`
	Steps        []map[string]any `json:"steps"`
}
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// Operations is the contract. Adding an endpoint here is what publishes it to
// every generated client.
func Operations() []Operation {
	return []Operation{
		{"GET", "/auth/config", "getAuthConfig", "Which sign-in methods are enabled", nil, AuthConfigResponse{}, nil},
		{"POST", "/auth/dev-login", "devLogin", "Sign in with an email address (development only)", DevLoginRequest{}, auth.User{}, nil},
		{"POST", "/auth/logout", "logout", "End the session", nil, nil, nil},
		{"GET", "/auth/me", "getCurrentUser", "The signed-in user", nil, auth.User{}, nil},

		{"GET", "/projects", "listProjects", "Every project the user owns", nil, ProjectsResponse{}, nil},
		{"POST", "/projects", "createProject", "Create a project from a niche description", CreateProjectRequest{}, project.Project{}, nil},
		{"GET", "/projects/{projectId}", "getProject", "One project", nil, project.Project{}, nil},
		{"PATCH", "/projects/{projectId}", "updateProject", "Update identity, brand kit or AI strategy", UpdateProjectRequest{}, project.Project{}, nil},
		{"DELETE", "/projects/{projectId}", "deleteProject", "Soft-delete a project", nil, nil, nil},
		{"GET", "/projects/{projectId}/schema", "getCarouselForm", "The AI-generated input form", nil, SchemaResponse{}, nil},
		{"POST", "/projects/{projectId}/generate-schema", "regenerateCarouselForm", "Rebuild the input form", nil, SchemaResponse{}, nil},

		{"GET", "/projects/{projectId}/carousels", "listCarousels", "Carousels in a project", nil, CarouselsResponse{},
			map[string]string{"status": "draft | ready | archived | all", "q": "title search"}},
		{"POST", "/projects/{projectId}/carousels/generate", "generateCarousel", "Queue a carousel generation", GenerateCarouselRequest{}, GenerateCarouselResponse{}, nil},

		{"GET", "/carousels/{carouselId}", "getCarousel", "One carousel and its latest job", nil, CarouselResponse{}, nil},
		{"PATCH", "/carousels/{carouselId}", "updateCarousel", "Rename or change status", UpdateCarouselRequest{}, carousel.Carousel{}, nil},
		{"DELETE", "/carousels/{carouselId}", "deleteCarousel", "Soft-delete a carousel", nil, nil, nil},
		{"POST", "/carousels/{carouselId}/duplicate", "duplicateCarousel", "Copy a carousel", nil, carousel.Carousel{}, nil},
		{"POST", "/carousels/{carouselId}/archive", "archiveCarousel", "Archive a carousel", nil, carousel.Carousel{}, nil},
		{"POST", "/carousels/{carouselId}/regenerate", "regenerateCarousel", "Run generation again with the stored inputs", nil, GenerateCarouselResponse{}, nil},
		{"POST", "/carousels/{carouselId}/apply-brand", "applyBrand", "Re-stamp the project brand kit", nil, carousel.DesignView{}, nil},
		{"GET", "/carousels/{carouselId}/content", "getCarouselContent", "The generated copy", nil, ContentDocument{}, nil},
		{"GET", "/carousels/{carouselId}/sources", "getResearchSources", "Sources cited during research", nil, SourcesResponse{}, nil},

		{"GET", "/carousels/{carouselId}/design", "getDesign", "The current Design JSON", nil, carousel.DesignView{}, nil},
		{"PATCH", "/carousels/{carouselId}/design", "saveDesign", "Autosave with optimistic concurrency", SaveDesignRequest{}, carousel.DesignView{}, nil},

		{"GET", "/carousels/{carouselId}/caption", "getCaption", "Post caption and hashtags", nil, carousel.Caption{}, nil},
		{"PATCH", "/carousels/{carouselId}/caption", "updateCaption", "Edit the caption", UpdateCaptionRequest{}, carousel.Caption{}, nil},
		{"POST", "/carousels/{carouselId}/caption/generate", "regenerateCaption", "Rewrite the caption with AI", nil, carousel.Caption{}, nil},

		{"GET", "/carousels/{carouselId}/slides/{slideId}/images", "getSlideImages", "Image candidates for a slide", nil, ImageCandidatesResponse{}, nil},
		{"POST", "/carousels/{carouselId}/slides/{slideId}/images/search", "searchSlideImages", "Search again, optionally with a new keyword", SearchImagesRequest{}, ImageCandidatesResponse{}, nil},
		{"POST", "/carousels/{carouselId}/slides/{slideId}/images/select", "selectSlideImage", "Use a candidate as the slide background", SelectImageRequest{}, SelectImageResponse{}, nil},

		{"POST", "/carousels/{carouselId}/export", "createExport", "Queue a PNG/ZIP export", nil, export.Job{}, nil},
		{"GET", "/exports/{exportId}", "getExport", "Export progress and download URL", nil, export.Job{}, nil},

		{"GET", "/assets/{assetId}", "getAsset", "One asset", nil, asset.Asset{}, nil},
		{"DELETE", "/assets/{assetId}", "deleteAsset", "Delete an asset", nil, nil, nil},

		{"GET", "/jobs/{jobId}", "getJob", "Generation job progress", nil, JobResponse{}, nil},
		{"GET", "/registry", "getRegistry", "Fonts, formulas, canvas presets and content languages", nil, RegistryResponse{}, nil},
	}
}

// ContentDocument mirrors what GET /content returns (content.Content).
type ContentDocument struct {
	SchemaVersion string `json:"schema_version"`
	Title         string `json:"title"`
	FormulaID     string `json:"formula_id"`
	FormulaReason string `json:"formula_reason"`
	Slides        []struct {
		Index       int    `json:"index"`
		Role        string `json:"role"`
		Headline    string `json:"headline"`
		Body        string `json:"body,omitempty"`
		ImageIntent string `json:"image_intent,omitempty"`
		ImageQuery  string `json:"image_query,omitempty"`
	} `json:"slides"`
}

// Build reflects over every operation and returns the populated registry.
func Build() (*Registry, []Operation) {
	reg := NewRegistry()
	ops := Operations()
	for _, op := range ops {
		if op.Body != nil {
			reg.Register(reflect.TypeOf(op.Body))
		}
		if op.Response != nil {
			reg.Register(reflect.TypeOf(op.Response))
		}
	}
	return reg, ops
}
