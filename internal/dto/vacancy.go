package dto

type CreateVacancyRequest struct {
	Vacancy Vacancy `json:"vacancy"`
}

type CreateVacancyResponse struct {
	Vacancy Vacancy `json:"vacancy"`
}

type UpdateVacancyRequest struct {
	Vacancy Vacancy `json:"vacancy"`
}

type UpdateVacancyResponse struct {
	Vacancy Vacancy `json:"vacancy"`
}

type ListVacancyRequest struct {
	CategoryID int     `json:"category_id"`
	SalaryFrom float64 `json:"salary_from"`
	SalaryTo   float64 `json:"salary_to"`
	Search     string  `json:"search"`
	Limit      int     `json:"limit"`
	Offset     int     `json:"offset"`
}

type ListVacancyResponse struct {
	Vacancie []Vacancy `json:"vacancies"`
	Total    int       `json:"total"`
}

type Vacancy struct {
	ID             int64                   `json:"id"`
	Title          string                  `json:"title"`
	Description    string                  `json:"description"`
	SalaryFrom     *float64                `json:"salary_from,omitempty"`
	SalaryTo       *float64                `json:"salary_to,omitempty"`
	SalaryExact    *float64                `json:"salary_exact,omitempty"`
	SalaryType     string                  `json:"salary_type"`
	SalaryCurrency string                  `json:"salary_currency"`
	OrganizationID int64                   `json:"organization_id"`
	Organization   Organization            `json:"organization"`
	CategoryID     int64                   `json:"category_id"`
	Category       CategoryResponse        `json:"category"`
	Details        []VacancyDetailResponse `json:"details"`
	CreatedAt      string                  `json:"created_at"`
	UpdatedAt      string                  `json:"updated_at"`
}

type VacancyDetailResponse struct {
	ID        int64  `json:"id"`
	GroupName string `json:"group_name"`
	Name      string `json:"name"`
	Value     string `json:"value"`
	IconURL   string `json:"icon_url,omitempty"`
	VacancyID int64  `json:"vacancy_id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
