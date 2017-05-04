package main

type ServiceStats struct {
	SubscribersCount int `json:"subscribers_count"`
}

type ActiveSessions struct {
	SubscribersSessionsCount map[string]int `json:"subscribers_sessions"`
}
