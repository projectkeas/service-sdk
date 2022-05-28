package opa

import (
	"context"
	"fmt"
	"strconv"

	"github.com/open-policy-agent/opa/rego"
)

const SERVICE_NAME string = "OPA"

type OPAService struct {
	policies map[string]rego.PreparedEvalQuery
}

func (opa *OPAService) AddOrUpdatePolicy(namespace string, name string, outputParameters map[string]interface{}, policy string) error {
	if opa.policies == nil {
		opa.policies = map[string]rego.PreparedEvalQuery{}
	}

	outputQuery := ""
	defaultValues := ""

	for key, value := range outputParameters {
		outputQuery += fmt.Sprintf("%s = data.%s.%s\n", key, namespace, key)

		str, isString := value.(string)

		if isString {
			defaultValues += fmt.Sprintf("default %s = \"%s\"\n", key, str)
		} else {
			b, isBool := value.(bool)
			if isBool {
				defaultValues += fmt.Sprintf("default %s = %s\n", key, strconv.FormatBool(b))
				continue
			}

			i, isInt := value.(int)
			if isInt {
				defaultValues += fmt.Sprintf("default %s = %d\n", key, i)
				continue
			}
		}
	}

	completePolicy := fmt.Sprintf("package %s\n\n%s\n%s", namespace, defaultValues, policy)

	//fmt.Println(completePolicy)

	query, err := rego.New(
		rego.Query(outputQuery),
		rego.Module(namespace+".rego", completePolicy),
	).PrepareForEval(context.TODO())

	if err != nil {
		return err
	}

	opa.policies[name] = query

	return nil
}

func (opa OPAService) EvaluatePolicy(policy string, input interface{}) (rego.ResultSet, error) {

	if opa.policies == nil {
		return nil, fmt.Errorf("policy not found: %s", policy)
	}

	executablePolicy, found := opa.policies[policy]

	if !found {
		return nil, fmt.Errorf("policy not found: %s", policy)
	}

	result, err := executablePolicy.Eval(context.TODO(), rego.EvalInput(input))

	if err != nil {
		return nil, err
	} else if len(result) == 0 {
		return nil, fmt.Errorf("no results returned from policy for the specified input. policy: %s", policy)
	}

	return result, nil
}
