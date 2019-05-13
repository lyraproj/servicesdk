package wf

import "github.com/lyraproj/issue/issue"

const (
	ActionExecutionError     = `WF_ACTION_EXECUTION_ERROR`
	BadParameter             = `WF_BAD_PARAMETER`
	ConditionSyntaxError     = `WF_CONDITION_SYNTAX_ERROR`
	ConditionMissingRp       = `WF_CONDITION_MISSING_RP`
	ConditionInvalidName     = `WF_CONDITION_INVALID_NAME`
	ConditionUnexpectedEnd   = `WF_CONDITION_UNEXPECTED_END`
	ElementNotParameter      = `WF_ELEMENT_NOT_PARAMETER`
	FieldTypeMismatch        = `WF_FIELD_TYPE_MISMATCH`
	IllegalIterationStyle    = `WF_ILLEGAL_ITERATION_STYLE`
	IllegalOperation         = `WF_ILLEGAL_OPERATION`
	InvalidFunction          = `WF_INVALID_FUNCTION`
	InvalidTypeName          = `WF_INVALID_TYPE_NAME`
	IteratorNotOneStep       = `WF_ITERATOR_NOT_ONE_STEP`
	MissingRequiredField     = `WF_MISSING_REQUIRED_FIELD`
	MissingRequiredFunction  = `WF_MISSING_REQUIRED_FUNCTION`
	NoDefinition             = `WF_NO_DEFINITION`
	NoServerBuilderInContext = `WF_NO_SERVER_BUILDER_IN_CONTEXT`
	NotStep                  = `WF_NOT_STEP`
	NotStepDefinition        = `WF_NOT_STEP_DEFINITION`
	StepBuildError           = `WF_STEP_BUILD_ERROR`
	StepNoName               = `WF_STEP_NO_NAME`
	StateCreationError       = `WF_STATE_CREATION_ERROR`
)

func init() {
	issue.Hard(ActionExecutionError, `error while executing %{step}`)
	issue.Hard2(BadParameter, `%{step}: element %{name} is not a valid %{parameterType} parameter`, issue.HF{`step`: issue.Label})
	issue.Hard(ConditionSyntaxError, `syntax error in condition '%{text}' at position %{pos}`)
	issue.Hard(ConditionMissingRp, `expected right parenthesis in condition '%{text}' at position %{pos}`)
	issue.Hard(ConditionInvalidName, `invalid name '%{name}' in condition '%{text}' at position %{pos}`)
	issue.Hard(ConditionUnexpectedEnd, `unexpected end of condition '%{text}' at position %{pos}`)
	issue.Hard(ElementNotParameter, `expected field %{field} element to be a Parameter, got %{type}`)
	issue.Hard(FieldTypeMismatch, `expected field %{field} to be a %{expected}, got %{actual}`)
	issue.Hard(IllegalIterationStyle, `no such iteration style '%{style}'`)
	issue.Hard(IllegalOperation, `no such operation '%{operation}'`)
	issue.Hard(InvalidFunction, `invalid function '%{function}'. Expected one of 'create', 'read', 'update', or 'delete'`)
	issue.Hard(InvalidTypeName, `invalid type name '%{name}'. A type name must consist of one to many capitalized segments separated with '::'`)
	issue.Hard(IteratorNotOneStep, `an iterator must have exactly one step`)
	issue.Hard(MissingRequiredField, `missing required field '%{field}'`)
	issue.Hard(MissingRequiredFunction, `missing required '%{function}'`)
	issue.Hard(NoDefinition, `expected step to contain a definition block`)
	issue.Hard(NoServerBuilderInContext, `no ServerBuilder has been registered with the evaluation context`)
	issue.Hard2(NotStep, `block may only contain workflow steps. %{actual} is not supported here`,
		issue.HF{`actual`: issue.UcAnOrA})
	issue.Hard(NotStepDefinition, `a step definition must be a hash`)
	issue.Hard(StepBuildError, `error while building %{step}`)
	issue.Hard(StepNoName, `an step must have a name`)
	issue.Hard(StateCreationError, `error while creating %{step} state`)
}
