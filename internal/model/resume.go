package model

type Resume struct {
	Id int `json:"id"`
	UserId int `json:"user_id"`
	CategoryId int `json:"category_id"`
	Description string `json:"description"`
	SalaryFrom int `json:"salary_from"`
	SalaryTo int `json:"salary_to"`
	Experiences []*ResumeExperience `json:"experiences"`
	Skills []*ResumeSkill `json:"skills"`
	SalaryPeriod string `json:"salary_period"`
	CreatedAt string `json:"created_at"`
}

type ResumeSkill struct {
	Id int `json:"id"`
	ResumeId int `json:"resume_id"`
	Skill string `json:"skill"`
}

type ResumeExperience struct {
	Id int `json:"id"`
	ResumeId int `json:"resume_id"`
	OrganizationName string `json:"organization_name"`
	CategoryId int `json:"category_id"`
	Description string `json:"description"`
	StartMonth string `json:"start_month"`
	StartYear string `json:"start_year"`
	EndMonth string `json:"end_month"`
	EndYear string `json:"end_year"`
}
