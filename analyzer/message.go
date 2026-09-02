package analyzer

import (
	"fmt"
	"math"
	"strconv"

	policydomain "github.com/gostafa/distance/internal/features/policy/domain"
)

// formatViolation renders one policy violation as a diagnostic message.
func formatViolation(v policydomain.Violation) string {
	return fmt.Sprintf("%s (package): %s %s exceeds max %s",
		v.Package, v.Key, formatNumber(v.Value), formatNumber(v.Threshold))
}

func formatNumber(value float64) string {
	if value == math.Trunc(value) && !math.IsInf(value, 0) {
		return strconv.FormatFloat(value, 'f', -1, 64)
	}

	return strconv.FormatFloat(value, 'f', 2, 64)
}
