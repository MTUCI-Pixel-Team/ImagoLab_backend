package pictureGeneration

import (
	"fmt"
	"strings"

	pg "github.com/prorok210/WS_Client-for_runware.ai-"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s", e.Message)
}

type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

type RequestRules struct {
	taskTypes         []string
	outputTypes       []string
	outputFormats     []string
	minPromtLength    int
	maxPromtLength    int
	max_height        int
	min_height        int
	max_width         int
	min_width         int
	allowed_models    []string
	max_steps         int
	max_CFGScale      float64
	max_numberResults int
}

func DefaultValidationRules() *RequestRules {
	return &RequestRules{
		taskTypes:      []string{"imageInference"},
		outputTypes:    []string{"URL"},
		outputFormats:  []string{"JPG"},
		minPromtLength: 4,
		maxPromtLength: 2000,
		min_height:     512,
		max_height:     2048,
		min_width:      512,
		max_width:      2048,
		// allowed_models:    []string{"civitai:25694@143906"},
		max_steps:         100,
		max_CFGScale:      30.0,
		max_numberResults: 10,
	}
}

func ValidateRequest(request pg.ReqMessage) error {
	rules := DefaultValidationRules()
	var errors ValidationErrors

	if err := ValidateTaskType(request.TaskType, rules); err != nil {
		errors = append(errors, *err)
	}
	if err := ValidateOutputType(request.OutputType[0], rules); err != nil {
		errors = append(errors, *err)
	}
	if request.OutputFormat != "" {
		if err := ValidateOutputFormat(request.OutputFormat, rules); err != nil {
			errors = append(errors, *err)
		}
	}

	if len(request.PositivePrompt) < rules.minPromtLength || len(request.PositivePrompt) > rules.maxPromtLength {
		errors = append(errors, ValidationError{
			Field:   "positivePrompt",
			Message: "Invalid prompt length",
		})
	}

	if len(request.NegativePrompt) != 0 {
		if len(request.NegativePrompt) < rules.minPromtLength || len(request.NegativePrompt) > rules.maxPromtLength {
			errors = append(errors, ValidationError{
				Field:   "negativePrompt",
				Message: "Invalid prompt length",
			})
		}
	}

	if request.Height > rules.max_height || request.Height < rules.min_height || request.Height%64 != 0 {
		errors = append(errors, ValidationError{
			Field:   "height",
			Message: "Invalid height",
		})
	}

	if request.Width > rules.max_width || request.Width < rules.min_width || request.Width%64 != 0 {
		errors = append(errors, ValidationError{
			Field:   "width",
			Message: "Invalid width",
		})
	}

	if request.Steps > rules.max_steps || request.Steps < 0 {
		errors = append(errors, ValidationError{
			Field:   "steps",
			Message: "Invalid steps",
		})
	}

	if request.CFGScale > 0.0 {
		if request.CFGScale > rules.max_CFGScale {
			errors = append(errors, ValidationError{
				Field:   "CFGScale",
				Message: "Invalid CFGScale",
			})
		}
	}

	if request.NumberResults > rules.max_numberResults || request.NumberResults < 1 {
		errors = append(errors, ValidationError{
			Field:   "numberResults",
			Message: "Invalid number of results",
		})
	}

	if len(errors) > 0 {
		return errors
	}
	return nil

}

func ValidateTaskType(taskType string, rules *RequestRules) *ValidationError {
	for _, t := range rules.taskTypes {
		if t == taskType {
			return nil
		}
	}
	return &ValidationError{
		Field:   "taskType",
		Message: "Invalid task type",
	}
}

func ValidateOutputType(outputType string, rules *RequestRules) *ValidationError {
	for _, t := range rules.outputTypes {
		if t == outputType {
			return nil
		}
	}
	return &ValidationError{
		Field:   "outputType",
		Message: "Invalid output type",
	}
}

func ValidateOutputFormat(outputFormat string, rules *RequestRules) *ValidationError {
	for _, t := range rules.outputFormats {
		if t == outputFormat {
			return nil
		}
	}
	return &ValidationError{
		Field:   "outputFormat",
		Message: "Invalid output format",
	}
}

// func ValidateModel(model string, rules *RequestRules) *ValidationError {
// 	for _, t := range rules.allowed_models {
// 		if t == model {
// 			return nil
// 		}
// 	}
// 	return &ValidationError{
// 		Field:   "model",
// 		Message: "Invalid model",
// 	}
// }
