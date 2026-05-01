package core

type Task struct {
	Id          string `json:"id"`
	Payload     string `json:"payload"`
	Max_retries int    `json:"max_retries"`
	Status      string `json:"-"`
}
