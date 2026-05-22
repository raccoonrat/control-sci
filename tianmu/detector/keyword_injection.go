package detector

import (
	"context"
	"strings"

	"github.com/raccoonrat/control-sci/tianmu/core"
)

type KeywordInjectionDetector struct {
	keywords []string
}

func NewKeywordInjectionDetector() *KeywordInjectionDetector {
	return &KeywordInjectionDetector{
		keywords: []string{
			"忽略上述指令",
			"忽略以上",
			"忽略前文",
			"系统提示词",
			"管理员权限",
			"systemprompt",
			"bypassguardrails",
		},
	}
}

func (d *KeywordInjectionDetector) ID() string {
	return "cn-keyword-injection-v1"
}

func (d *KeywordInjectionDetector) Category() string {
	return "prompt_injection"
}

func (d *KeywordInjectionDetector) Version() string {
	return "1.0.0"
}

func (d *KeywordInjectionDetector) Detect(ctx context.Context, normalizedPrompt string) (core.DetectorSignal, error) {
	if err := ctx.Err(); err != nil {
		return core.DetectorSignal{}, err
	}

	cleaned := strings.ReplaceAll(normalizedPrompt, " ", "")
	for _, keyword := range d.keywords {
		if strings.Contains(cleaned, keyword) {
			return core.DetectorSignal{
				DetectorID: d.ID(),
				Category:   d.Category(),
				Version:    d.Version(),
				Confidence: 1.0,
				Triggered:  true,
			}, nil
		}
	}

	return core.DetectorSignal{
		DetectorID: d.ID(),
		Category:   d.Category(),
		Version:    d.Version(),
		Confidence: 0,
		Triggered:  false,
	}, nil
}
