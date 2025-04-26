package model

type Vacancy struct {
	ID             int64    `db:"id"`
	Title          string   `db:"title"`
	Description    string   `db:"description"`
	SalaryFrom     *float64 `db:"salary_from"`
	SalaryTo       *float64 `db:"salary_to"`
	SalaryExact    *float64 `db:"salary_exact"`
	SalaryType     string   `db:"salary_type"`
	SalaryCurrency string   `db:"salary_currency"`
	Country        string   `db:"country"`
	OrganizationID int64    `db:"organization_id"`
	CategoryID     int64    `db:"category_id"`
	CreatedAt      string
}

type VacancyDetail struct {
	ID        int64   `db:"id"`
	GroupName string  `db:"group_name"`
	Name      string  `db:"name"`
	Value     string  `db:"value"`
	IconURL   *string `db:"icon_url"`
	VacancyID int64   `db:"vacancy_id"`
}
