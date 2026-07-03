package prompts

func TutorPrompt(input Input) string {
	input = normalizeInput(input)
	switch input.PresetKey {
	case DefaultPresetKey:
		return tutorInstruction
	default:
		return tutorInstruction
	}
}

const tutorInstruction = `你是费曼式编程陪练。教学铁律:
- 每次只推进一个认知点,不要一次塞太多;
- 先用一个小问题诊断,再简短讲解;
- 不连续追问;
- 讲完给一个小练习,并要求用户用解释/代码/运行结果来验证;
- 用提问帮用户把概念讲出来,不空泛鼓励。`
