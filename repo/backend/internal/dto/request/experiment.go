package request

type CreateExperimentRequest struct {
	Name          string                  `json:"name" binding:"required,min=3,max=128"`
	Slug          string                  `json:"slug" binding:"omitempty,max=128"`
	Description   string                  `json:"description" binding:"omitempty"`
	Variants      []CreateVariantRequest  `json:"variants" binding:"required,min=2"`
	MinSampleSize int                     `json:"min_sample_size" binding:"omitempty,min=50"`
}

type CreateVariantRequest struct {
	Name       string `json:"name" binding:"required,max=64"`
	TrafficPct float64 `json:"traffic_pct" binding:"required,min=0,max=100"`
	Config     []byte  `json:"config" binding:"required"`
}

type UpdateExperimentRequest struct {
	Name        string `json:"name" binding:"omitempty,min=3,max=128"`
	Description string `json:"description" binding:"omitempty"`
}

type UpdateCanaryRequest struct {
	Variants []VariantTrafficUpdate `json:"variants" binding:"required,min=2"`
}

type VariantTrafficUpdate struct {
	Name       string  `json:"name" binding:"required"`
	TrafficPct float64 `json:"traffic_pct" binding:"required,min=0,max=100"`
}
