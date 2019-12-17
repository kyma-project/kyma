package v1alpha1

// HasCondition returns a boolean values if the subscription satisfies the condition.
func (in *Subscription) HasCondition(checked SubscriptionCondition) bool {
	if len(in.Status.Conditions) == 0 {
		return false
	}

	for _, cond := range in.Status.Conditions {
		if checked.Type == cond.Type && checked.Status == cond.Status {
			return true
		}
	}

	return false
}
