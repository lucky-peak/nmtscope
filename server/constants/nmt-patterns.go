package constants

import "regexp"

var (
	NMT_TOTAL_HEADER_REGEX    = regexp.MustCompile(`^Total:\s+reserved=(\d+)KB,\s+committed=(\d+)KB`)
	NMT_CATEGORY_HEADER_REGEX = regexp.MustCompile(`^-\s+(.*?)\s+\(reserved=(\d+)KB,\s+committed=(\d+)KB.*\)`)
)
