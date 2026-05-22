package detector

import (
	"context"
	"regexp"

	"github.com/raccoonrat/control-sci/tianmu/core"
)

type RegexPIIDetector struct {
	phoneRegex  *regexp.Regexp
	idCardRegex *regexp.Regexp
}

func NewRegexPIIDetector() *RegexPIIDetector {
	return &RegexPIIDetector{
		phoneRegex:  regexp.MustCompile(`1[3-9]\d{9}`),
		idCardRegex: regexp.MustCompile(`[1-6]\d{5}(18|19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]`),
	}
}

func (d *RegexPIIDetector) ID() string {
	return "cn-regex-pii-v1"
}

func (d *RegexPIIDetector) Category() string {
	return core.ChinesePIICategory
}

func (d *RegexPIIDetector) Version() string {
	return "1.0.0"
}

func (d *RegexPIIDetector) Detect(ctx context.Context, normalizedPrompt string) (core.DetectorSignal, error) {
	if err := ctx.Err(); err != nil {
		return core.DetectorSignal{}, err
	}

	if d.phoneRegex.MatchString(normalizedPrompt) || d.idCardRegex.MatchString(normalizedPrompt) {
		return core.DetectorSignal{
			DetectorID: d.ID(),
			Category:   d.Category(),
			Version:    d.Version(),
			Confidence: 0.95,
			Triggered:  true,
		}, nil
	}

	return core.DetectorSignal{
		DetectorID: d.ID(),
		Category:   d.Category(),
		Version:    d.Version(),
		Confidence: 0,
		Triggered:  false,
	}, nil
}
