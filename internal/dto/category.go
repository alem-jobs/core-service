package dto

type CreateCategory struct {
	Name     string `json:"name"`
	ParentID *int   `json:"parent_id"`
}

type UpdateCategory struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	ParentID *int    `json:"parent_id"`
}

type CategoryResponse struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	ParentID *int    `json:"parent_id"`
	Left     int     `json:"lft"`
	Right    int     `json:"rgt"`
	Depth    int     `json:"depth"`
}
