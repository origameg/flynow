package schedule

// Gets a list of scheduled destination airports, using the given API client.
func GetDestinations(origin string, client ScheduleClient) (destinations []string, err error) {

	return client.GetScheduledDestinations(origin)
}
