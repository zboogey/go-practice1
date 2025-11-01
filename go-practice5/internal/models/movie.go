package models

type Movie struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	Year       int    `json:"year"`
	ActorCount int    `json:"actor_count"`
}
