package response

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type PaginatedResponse struct {
	Data       any   `json:"data"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

func NewPaginated(data any, page, perPage int, total int64) PaginatedResponse {
	totalPages := total / int64(perPage)
	if total%int64(perPage) != 0 {
		totalPages++
	}
	return PaginatedResponse{
		Data:       data,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}
}
