package handlers

import "testing"

func TestIsSupportedProviderTypeOnlyAllowsOpenAICompatible(t *testing.T) {
	t.Parallel()

	if !isSupportedProviderType("openai_compatible") {
		t.Fatalf("expected openai_compatible provider type to be supported")
	}

	if isSupportedProviderType("custom") {
		t.Fatalf("expected custom provider type to be unsupported")
	}
}
