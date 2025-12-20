package constants

// Subscription tier constants
const (
	// TierFree - Free tier with limited features
	// Limits: 10 patients, 1 clinician
	TierFree = "free"

	// TierPaid - Paid tier with unlimited features
	// Limits: Unlimited patients, unlimited clinicians
	TierPaid = "paid"
)

// TierLimits defines the limits for each subscription tier
type TierLimits struct {
	MaxPatients   int
	MaxClinicians int
}

// GetTierLimits returns the limits for a given subscription tier
func GetTierLimits(tier string) TierLimits {
	switch tier {
	case TierFree:
		return TierLimits{
			MaxPatients:   10,
			MaxClinicians: 1,
		}
	case TierPaid:
		return TierLimits{
			MaxPatients:   -1, // -1 means unlimited
			MaxClinicians: -1, // -1 means unlimited
		}
	default:
		// Default to free tier limits if tier is unknown
		return TierLimits{
			MaxPatients:   10,
			MaxClinicians: 1,
		}
	}
}

// IsValidTier checks if a tier string is valid
func IsValidTier(tier string) bool {
	return tier == TierFree || tier == TierPaid
}

// AllTiers returns a slice of all valid tiers
func AllTiers() []string {
	return []string{
		TierFree,
		TierPaid,
	}
}

// IsUnlimited checks if a limit value represents unlimited (-1)
func IsUnlimited(limit int) bool {
	return limit == -1
}

