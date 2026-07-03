package prompts

const DefaultPresetKey = "default"

type Input struct {
	PresetKey string
}

func normalizeInput(input Input) Input {
	if input.PresetKey != DefaultPresetKey {
		input.PresetKey = DefaultPresetKey
	}
	return input
}
