package domain

// IntervalOffset returns the (years, months, days) to add to a base time (via time.Time.AddDate) to
// advance it by n intervals of the given PlanIntervalType.
//
//	YEAR  -> (n, 0, 0)
//	MONTH -> (0, n, 0)
//	WEEK  -> (0, 0, 7*n)
//	DAY   -> (0, 0, n)
//	(any other / empty) -> (0, 0, 0)
func IntervalOffset(it PlanIntervalType, n int32) (years, months, days int) {
	switch it {
	case PlanIntervalYear:
		return int(n), 0, 0
	case PlanIntervalMonth:
		return 0, int(n), 0
	case PlanIntervalWeek:
		return 0, 0, int(n) * 7
	case PlanIntervalDay:
		return 0, 0, int(n)
	default:
		return 0, 0, 0
	}
}
